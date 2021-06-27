package main

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/godbus/dbus/v5"
	"github.com/srevinsaju/lyrix/lyrixd/mpris"
	"github.com/srevinsaju/lyrix/lyrixd/service"
	"github.com/srevinsaju/lyrix/lyrixd/types"
	"github.com/withmandala/go-log"
	"os"
	"time"
)

var logger = log.New(os.Stdout)
var song *types.SongMeta
var auth *types.UserInstance

func GetSong() *types.SongMeta {
	return song
}

func SetAuth(a *types.UserInstance) {
	auth = a
}

func GetAuth() *types.UserInstance {
	return auth
}
func DaemonWrapper() {
	for {
		if auth == nil {
			time.Sleep(5 * time.Second)
			continue
		}
		Daemon()
	}

}

func Daemon() {
	ctx := &service.Context{
		LastFmEnabled: false,
		Predicted:     false,
		Tui:           false,
		Scrobble:      false,
	}
	song = &types.SongMeta{
		Playing:  false,
		Track:    "",
		Artist:   "",
		Source:   "",
		Url:      "",
		Scrobble: false,
	}
	// start
	conn, err := dbus.SessionBus()
	names, err := mpris.List(conn)
	if err != nil {
		panic(err)
	}
	if len(names) == 0 {
		logger.Fatal("No media player found.")
	}

	mpDbusId := ""
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
		return
	}


	pl := mpris.New(conn, mpDbusId)

	logger.Info("Media player identity:", pl.GetIdentity())

	// end


	for {
		err := service.CheckForSongUpdates(ctx, auth, pl, song)
		if err != nil {
			logger.Warn(err)
			break
		}
		time.Sleep(5 * time.Second)
		logger.Info("song!!", song)

	}
}