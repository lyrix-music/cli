package service

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strings"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/adrg/xdg"
	"github.com/fatih/color"
	"github.com/gen2brain/beeep"
	"github.com/godbus/dbus/v5"
	"github.com/lyrix-music/cli/config"
	"github.com/lyrix-music/cli/meta"
	"github.com/lyrix-music/cli/mpris"
	"github.com/lyrix-music/cli/player"
	"github.com/lyrix-music/cli/types"
	"github.com/urfave/cli/v2"
	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout)

type Context struct {
	LastFmEnabled bool
	Predicted     bool
	Tui           bool
	Scrobble      bool
	AppIcon       string
}
type DaemonOptions struct {
}

func CheckForSongUpdates(ctx *Context, auth *types.UserInstance, pl *mpris.Player, song *types.SongMeta) error {

	metadata, ok := pl.GetMetadata()
	if !ok {
		return errors.New("player is no longer active")
	}
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

	url, ok := metadata["xesam:url"].Value().(string)
	source := "local"

	if ok && strings.HasPrefix(url, "https://music.youtube.com/") {
		// this is a song played on music.youtube.com
		artist = strings.Replace(artist, " - Topic", "", 1)
		source = "music.youtube.com"
	}
	title := metadata["xesam:title"].Value().(string)
	playerIsPlaying := pl.GetPlaybackStatus() == "\"Playing\""
	artist = strings.Replace(artist, " - Topic", "", -1)

	position := pl.GetPosition()

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

	} else if pl.GetPlaybackStatus() == "\"Paused\"" && song.Playing {
		fmt.Println("Playback is paused now")
		song.Playing = false
		if auth != nil {
			player.NotPlayingSongHandler(auth)
		}

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
		LastFmEnabled: c.Bool("lastfm-predict"),
		Scrobble:      c.Bool("lastfm-scrobble"),
	}

	auth, err := config.Load(meta.AppName)
	if err != nil {
		logger.Warn(err)
	}
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
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
			err := CheckForSongUpdates(ctx, auth, pl, song)
			if err != nil {
				logger.Warn(err)
				break
			}
			time.Sleep(5 * time.Second)

		}
	}

	return nil
}
