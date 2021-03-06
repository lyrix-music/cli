package main

import (
	"fmt"
	"github.com/lyrix-music/cli/cmd/desktop/daemon"
	"github.com/lyrix-music/cli/cmd/desktop/logging"
	"github.com/lyrix-music/cli/cmd/desktop/meta"
	"github.com/lyrix-music/cli/cmd/desktop/routes"
	"github.com/lyrix-music/cli/config"
	"github.com/lyrix-music/cli/types"
	"github.com/webview/webview"
	"os"
	"path/filepath"
)

var logger = logging.GetLogger()

func main() {
	auth, err := config.Load(meta.AppName)
	if err != nil {
		logger.Fatal(err)
		return
	}
	if auth != nil {
		daemon.SetAuth(auth)
	} else {
		auth = &types.UserInstance{}
	}

	// launch the daemon to discover changes in the current listening song
	go daemon.Start()

	// create the web server and launch it as a goroutine
	app := routes.BuildServer(auth)
	newAddress, err := GetNewAddress()
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("Attempting to use '%s'", newAddress)

	if os.Getenv("LYRIX_SERVER_ONLY") == "1" {
		app.Listen(newAddress)
		return
	}

	go func() {
		app.Listen(newAddress)
	}()

	// create the web application instance
	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Lyrixd")
	w.SetSize(600, 800, webview.HintNone)
	err = w.Bind("open", open)
	if err != nil {
		logger.Fatal(err)
	}
	err = w.Bind("iswebview", func() bool { return true })
	if err != nil {
		logger.Fatal(err)
	}

	suffix := ""
	if auth == nil {
		suffix = "login"
	}
	w.Navigate(fmt.Sprintf("http://%s/%s", newAddress, suffix))
	windowIcon := "AppDir/lyrix-desktop.png"
	if os.Getenv("APPDIR") != "" {
		appDir := os.Getenv("APPDIR")
		windowIcon = filepath.Join(appDir, "lyrix-desktop.png")
	}
	err = setWindowIcon(w, windowIcon)
	if err != nil {
		logger.Warn("Error while setting window icon", err)
	}
	w.Run()

}
