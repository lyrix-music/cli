package daemon

import (
	"github.com/godbus/dbus/v5"
	"github.com/lyrix-music/cli/config"
	"github.com/lyrix-music/cli/mpris"
	"github.com/lyrix-music/cli/service"
	"github.com/lyrix-music/cli/types"
	"github.com/withmandala/go-log"
	"os"
	"path/filepath"
	"time"
)

var logger = log.New(os.Stdout)
var song *types.SongMeta
var auth *types.UserInstance
var player *mpris.Player
var dbusConn *dbus.Conn
var ctx *service.Context

// GetSong returns the last played song on the desktop
func GetSong() *types.SongMeta {
	return song
}

func SetScrobbleEnabled(enabled bool) {
	logger.Debug("Setting scrobble preference to", enabled)
	ctx.Scrobble = enabled
}

// SetAuth sets the authorization instance
func SetAuth(a *types.UserInstance) {
	auth = a
	config.Set("Username", a.Username)
	config.Set("AuthToken", a.Token)
	config.Set("Host", a.Host)
	config.Write()
}

// GetAuth returns the registerd authorization details
func GetAuth() *types.UserInstance {
	return auth
}

// GetPlayer returns the music player instance
func GetPlayer() *mpris.Player {
	return player
}

// SetPlayer sets the local player from a dbus identifier
func SetPlayer(dbusId string) string {
	logger.Debug("Changing player from %s to %s", GetPlayer(), dbusId)
	player = mpris.New(dbusConn, dbusId)
	return player.GetIdentity()
}

// Start loop the daemon process until all the parameters
// are successfull met
func Start() {
	var err error
	dbusConn, err = dbus.SessionBus()
	if err != nil {
		panic(err)
	}

	// try to get app icon
	appIcon := ""
	if os.Getenv("APPDIR") != "" {
		appIcon = filepath.Join(os.Getenv("APPDIR"), "lyrix-desktop.png")
	}

	ctx = &service.Context{
		LastFmEnabled: false,
		Predicted:     false,
		Tui:           false,
		Scrobble:      false,
		AppIcon:       appIcon,
	}
	song = &types.SongMeta{
		Playing:  false,
		Track:    "",
		Artist:   "",
		Source:   "",
		Url:      "",
		Scrobble: false,
	}

	for {
		daemon()
	}

}

func daemon() {
	var err error

	// start

	names, err := mpris.List(dbusConn)
	if err != nil {
		panic(err)
	}
	if len(names) == 0 {
		logger.Fatal("No media player found.")
	}

	mpDbusId := ""
	if len(names) == 1 {
		mpDbusId = names[0]
	}

	if mpDbusId == "" && player == nil {
		logger.Warn("Failed to detect media players")
		time.Sleep(10 * time.Second)
		return
	} else if player == nil {
		player = mpris.New(dbusConn, mpDbusId)
	}

	logger.Info("Media player identity:", player.GetIdentity())

	logger.Debug("Scrobbling enabled:", ctx.Scrobble)
	err = service.CheckForSongUpdates(ctx, auth, player, song)
	if err != nil {
		logger.Warn(err)
		return
	}
	time.Sleep(5 * time.Second)
}
