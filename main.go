package main

import (
	"fmt"
	"github.com/srevinsaju/lyrix/lyrixd/auth"

	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/srevinsaju/lyrix/lyrixd/mpris"
	"github.com/srevinsaju/lyrix/lyrixd/player"
	"github.com/srevinsaju/lyrix/lyrixd/types"
	"github.com/urfave/cli/v2"
	"github.com/withmandala/go-log"
    "github.com/fatih/color"
)

var logger = log.New(os.Stdout)

func CheckForSongUpdates(auth types.UserInstance, pl *mpris.Player, song *types.SongMeta) error {
    metadata := pl.GetMetadata()
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

	if playerIsPlaying && (song.Artist != artist || song.Track != title || !song.Playing) {
        color.Green("%s by %s", title, artist)
        color.HiBlack("%s\n", source)
        player.PlayingSongHandler(auth, &types.SongMeta{Track: title, Artist: artist, Source: source, Url: url})
		song.Artist = artist
		song.Track = title
		song.Playing = true
	} else if pl.GetPlaybackStatus() == "\"Paused\"" && song.Playing {
		fmt.Println("Playback is paused now")
		song.Playing = false
		player.NotPlayingSongHandler(auth)
	}
    return nil
}

func cleanup(auth types.UserInstance) {
	logger.Info("Cleaning up. Sending clear events")
	player.NotPlayingSongHandler(auth)
	logger.Info("Done.")
}

func StartDaemon(c *cli.Context) error {
    for {
        err := DaemonLoop()
        if err != nil {
            time.Sleep(5)
        }
    }
    return nil
}

func DaemonLoop() error {
	auth, err := LoadConfig()
	if err != nil {
		logger.Fatal(err)
	}
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		cleanup(auth)
		os.Exit(1)
	}()

	// id of the music player from dbus
	mpDbusId := ""
	userProvidedDbusId := os.Getenv("LYRIX_MUSIC_PLAYER_DBUS_ID")
    logger.Info("Waiting for a music player...")
	for {
		names, err := mpris.List(conn)
		if err != nil {
			panic(err)
		}
		if len(names) == 0 {
			logger.Fatal("No media player found.")
		}

		for i := range names {
            logger.Infof("Scanning %s", names[i])
			userRunningElisaPlayer := (strings.HasSuffix(names[i], "elisa") && (userProvidedDbusId == "" || strings.HasSuffix(userProvidedDbusId, "elisa")))
			if names[i] == userProvidedDbusId || userRunningElisaPlayer {
				mpDbusId = names[i]
				break
			}
		}
		if mpDbusId != "" {
			logger.Infof("Detected running %s player", mpDbusId)
			break
		}
		time.Sleep(5 * time.Second)
	}

	pl := mpris.New(conn, mpDbusId)

	logger.Info("Media player identity:", pl.GetIdentity())

	song := &types.SongMeta{}
	for {
		CheckForSongUpdates(auth, pl, song)
		time.Sleep(5 * time.Second)
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:   "Lyrix Daemon",
		Usage:  "A daemon for lyrix music network",
		Action: StartDaemon,
		Commands: []*cli.Command{
			{
				Name: "register",
				Action: func(c *cli.Context) error {
					auth.Register()
					return nil
				},
			},
			{
				Name: "reset-config",
				Action: func(c *cli.Context) error {
					_, configPath := GetLocalConfigPath()
					logger.Info("Removing old configuration files...")
					return os.RemoveAll(configPath)

				},
			},
			{
				Name:   "start",
				Action: StartDaemon,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}

}
