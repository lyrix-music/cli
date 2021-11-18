package service

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/godbus/dbus/v5"
	k "github.com/srevinsaju/korean-romanizer-go"
	dsClient "github.com/srevinsaju/rich-go/client"
	sl "github.com/srevinsaju/swaglyrics-go"
	sltypes "github.com/srevinsaju/swaglyrics-go/types"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/adrg/xdg"
	"github.com/fatih/color"
	"github.com/gen2brain/beeep"
	"github.com/lyrix-music/cli/config"
	"github.com/lyrix-music/cli/meta"
	"github.com/lyrix-music/cli/mpris"
	"github.com/lyrix-music/cli/player"
	"github.com/lyrix-music/cli/types"
	"github.com/urfave/cli/v2"
	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout)

type CliContext struct {
	ShowLyrics bool
}

type Context struct {
	LastFmEnabled      bool
	Predicted          bool
	Tui                bool
	Scrobble           bool
	DiscordIntegration bool
	AppIcon            string
	Romanize           bool
	Cli                *CliContext
}

type DaemonOptions struct {
}

func CheckForSongUpdatesDbus(ctx *Context, auth *types.UserInstance, pl *mpris.Player, song *types.SongMeta) error {

	metadata, ok := pl.GetMetadata()
	if !ok {
		return errors.New("player is no longer active")
	}

	// parse artist from mpris
	if metadata["xesam:artist"].Value() == nil || metadata["xesam:title"].Value() == nil {
		// wait for sometime
		return nil
	}
	artist := ""
	artists, ok := metadata["xesam:artist"].Value().([]string)
	if ok {
		artist = artists[0]
	} else {
		artist = metadata["xesam:artist"].Value().(string)
	}

	// get URL
	url, ok := metadata["xesam:url"].Value().(string)

	status := "Paused"
	if pl.GetPlaybackStatus() == "\"Playing\"" {
		status = "Playing"
	}

	title := metadata["xesam:title"].Value().(string)

	svcSong := &ServiceSong{
		Artist:   artist,
		Artists:  artists,
		Title:    title,
		Url:      url,
		Status:   status,
		Position: pl.GetPosition(),
	}

	return checkForSongUpdates(ctx, auth, svcSong, song)

}

func checkForSongUpdates(ctx *Context, auth *types.UserInstance, m *ServiceSong, song *types.SongMeta) error {
	if m == nil {
		return errors.New("player is no longer active")
	}

	if m.Title == "" || (m.Artist == "" && len(m.Artists) == 0) {
		// wait for sometime
		return nil
	}

	artist := ""
	if len(m.Artists) >= 1 {
		artist = m.Artists[0]
	} else {
		artist = m.Artist
	}
	source := "local"
	url := m.Url
	if strings.HasPrefix(url, "https://music.youtube.com/") {
		// this is a song played on music.youtube.com
		artist = strings.Replace(artist, " - Topic", "", 1)
		source = "music.youtube.com"
	}
	title := m.Title
	playerIsPlaying := m.Status == "Playing"
	artist = strings.Replace(artist, " - Topic", "", -1)

	position := m.Position

	isRepeat := position < song.Position && song.Artist == artist && song.Track == title
	if playerIsPlaying && (song.Artist != artist || song.Track != title || !song.Playing || isRepeat) {
		color.Green("%s by %s", title, artist)
		color.HiBlack("%s\n", source)
		if isRepeat {
			color.HiBlack("on Repeat.")
		}

		song.Position = position
		song.IsRepeat = isRepeat
		if auth != nil {
			player.PlayingSongHandler(
				auth,
				&types.SongMeta{
					Track:    title,
					Artist:   artist,
					Source:   source,
					Url:      url,
					Scrobble: ctx.Scrobble,
					IsRepeat: isRepeat,
					Position: position},
			)
			go func() {
				if ctx.DiscordIntegration && meta.DiscordApplicationId != "" {
					appName := "Local Player"
					appId := "lyrix"
					if strings.HasPrefix(url, "https://music.youtube.com/") {
						appId = "youtube-music"
						appName = "Youtube Music"
					} else if strings.HasPrefix(url, "https://open.spotify.com/") {
						appId = "spotify"
						appName = "Spotify"
					}
					logger.Info(url)

					t := time.Now()
					info := title
					if isRepeat {
						info += " - on Repeat"
					}
					err := dsClient.SetActivity(dsClient.Activity{
						State:      artist,
						Details:    info,
						LargeImage: appId,
						LargeText:  appName,
						SmallImage: "lyrix",
						SmallText:  "Lyrix Music",
						Timestamps: &dsClient.Timestamps{
							Start: &t,
						},
					})
					if err != nil {
						logger.Debug("Failed to set discord ")
					}
				}
			}()
		}

		if song.Track != "" {
			// the previous saved song was not a blank song
			// implies that the user had not paused the song, but song changed naturally

			message := ""
			if ctx.Scrobble {
				message = "Now scrobbling"
			} else {
				message = "Now playing"
			}

			if isRepeat {
				message += " on repeat"
			}
			err := beeep.Notify(fmt.Sprintf("%s by %s", title, artist), message, ctx.AppIcon)
			if err != nil {
				logger.Warn(err)
			}
		}
		song.Artist = artist
		song.Track = title
		song.Playing = true

		if ctx.Cli != nil {
			if ctx.Cli.ShowLyrics {
				lyrics, err := sl.GetLyrics(sltypes.Song{Track: song.Track, Artist: song.Artist})
				if err != nil {
					logger.Warn(err)
					return nil
				}
				if ctx.Romanize {
					var newLyrics []string
					lyricsInLines := strings.Split(lyrics, "\n")
					for i := range lyricsInLines {
						fmt.Println(lyricsInLines[i])
						r := k.NewRomanizer(lyricsInLines[i])
						color.HiBlack(r.Romanize())
					}
					lyrics = strings.Join(newLyrics, "\n")
				} else {
					fmt.Println(lyrics)
				}
			}
		}

	} else if m.Status == "Paused" && song.Playing {
		fmt.Println("Playback is paused now")
		song.Playing = false
		if auth != nil {
			player.NotPlayingSongHandler(auth)
		}
		if ctx.DiscordIntegration && meta.DiscordApplicationId != "" {
			err := dsClient.SetActivity(dsClient.Activity{
				State:      "",
				Details:    "",
				LargeImage: "",
				LargeText:  "",
				SmallImage: "",
				SmallText:  "",
			})
			if err != nil {
				logger.Warn("There was an error while clearing discord activity")
			}

		}

	}

	return nil
}

func QueueSimilarSongs(similarSongs []types.SongMeta, pl *mpris.Player) {
	usr, _ := user.Current()
	dir := usr.HomeDir

	// we support elisa music player only at the moment
	if pl.GetIdentity() != "\"Elisa\"" {
		return
	}
	if len(similarSongs) == 0 {
		return
	}

	for i := range similarSongs {
		go func(j int) {
			if similarSongs[j].Track == "" || similarSongs[j].Artist == "" {
				return
			}
			searchString := fmt.Sprintf("%s %s", similarSongs[j].Track, similarSongs[j].Artist)
			searchCommand := exec.Command(
				"baloosearch",
				"-d", strings.Replace(xdg.UserDirs.Music, "~", dir, -1),
				searchString)
			out, err := searchCommand.CombinedOutput()
			if err != nil {
				logger.Warnf("Failed to execute '%s'", searchCommand.String())
				logger.Warn(err, fmt.Sprintf("%s", out))
				return
			}
			output := string(out[:])
			s := strings.Split(output, "\n")[0]
			if strings.HasPrefix(s, "Elapsed") {
				// baloosearch didnt give a valid output
				// just suggest this song to the user

				color.HiBlack("Suggestion:")
				color.Yellow("%s by %s", similarSongs[j].Track, similarSongs[j].Artist)
				fmt.Println("")

				return
			}
			// the baloosearch found an answer
			color.HiBlack("Queued:")
			color.Green("%s by %s", similarSongs[j].Track, similarSongs[j].Artist)
			fmt.Println("")
			err = exec.Command("elisa", s).Run()
			if err != nil {
				logger.Warn(err)
			}

		}(i)

	}
}

func cleanup(auth *types.UserInstance) {
	logger.Info("Cleaning up. Sending clear events")
	if auth != nil {
		player.NotPlayingSongHandler(auth)
	}
	logger.Info("Done.")
}

func StartDaemon(c *cli.Context) error {
	// TODO move this piece to lyrixd spec
	ctx := &Context{
		LastFmEnabled:      c.Bool("lastfm-predict"),
		Scrobble:           c.Bool("lastfm-scrobble"),
		Romanize:           c.Bool("romanize"),
		Cli:                &CliContext{c.Bool("show-lyrics")},
		DiscordIntegration: c.Bool("discord"),
	}
	if ctx.DiscordIntegration && meta.DiscordApplicationId != "" {
		logger.Info("Enabling discord integration")
		err := dsClient.Login(meta.DiscordApplicationId)
		if err != nil {
			logger.Warn("There was an error enabling discord integration. Kindly report this as a bug:", err)
		}
	} else {
		logger.Info("Discord integration disabled", ctx.DiscordIntegration, meta.DiscordApplicationId != "")
	}

	auth, err := config.Load(meta.AppName)
	if err != nil {
		logger.Warn(err)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	if auth != nil {
		go func() {
			<-ch
			cleanup(auth)
			os.Exit(1)
		}()
	}
	if runtime.GOOS == "windows" {
		Run()
		exe, err := os.Executable()
		if err != nil {
			logger.Fatal("Couldn't resolve the base filepath of the exe", err)
		}
		exporterPath := filepath.Join(filepath.Dir(exe), fmt.Sprintf("lyrix-windows_exporter%s.exe", meta.BuildVersion))
		if _, err := os.Stat(exporterPath); os.IsNotExist(err) {
			logger.Fatalf("Couldn't find lyrix-win_exporter.exe at %s, %s", exporterPath, err)
		}
		song := &types.SongMeta{}

		for {
			err := CheckForSongUpdatesWinRTExporter(ctx, auth, exporterPath, song)
			if err != nil {
				logger.Warn(err)
				break
			}
			time.Sleep(5 * time.Second)
		}

	} else {
		conn, err := dbus.SessionBus()
		if err != nil {
			panic(err)
		}

		// id of the music player from dbus
		mpDbusId := ""
		logger.Info("Waiting for a music player...")
		for {
			names, err := mpris.List(conn)
			if err != nil {
				panic(err)
			}
			if len(names) == 0 {
				logger.Fatal("No media player found.")
			}

			mpDbusId = ""
			if len(names) == 1 {
				mpDbusId = names[0]
			} else {
				prompt := &survey.Select{
					Message: "Lyrix found multiple players. Select one:",
					Options: names,
				}
				survey.AskOne(prompt, &mpDbusId)
			}

			if mpDbusId == "" {
				logger.Warn("Failed to detect media players")
				break
			}

			pl := mpris.New(conn, mpDbusId)

			logger.Info("Media player identity:", pl.GetIdentity())

			song := &types.SongMeta{}
			for {
				err := CheckForSongUpdatesDbus(ctx, auth, pl, song)
				if err != nil {
					logger.Warn(err)
					break
				}
				go func() {
					// el
					// logger.Info("pl.GetIdentity() == \"Elisa\" && ctx.LastFmEnabled && !ctx.Predicted", pl.GetIdentity() == "\"Elisa\"" && ctx.LastFmEnabled && !ctx.Predicted)
					if ctx.LastFmEnabled && !ctx.Predicted && song.Playing {
						ctx.Predicted = true
						if auth != nil {
							return
						}
						similarSongs := player.GetSimilar(auth)
						QueueSimilarSongs(similarSongs, pl)
					}
				}()
				time.Sleep(5 * time.Second)

			}
		}
	}

	return nil
}
