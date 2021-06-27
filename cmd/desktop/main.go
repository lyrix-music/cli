package main

import (
	"fmt"
	"github.com/srevinsaju/lyrix/lyrixd/cmd/desktop/daemon"
	"github.com/srevinsaju/lyrix/lyrixd/cmd/desktop/logging"
	"github.com/srevinsaju/lyrix/lyrixd/cmd/desktop/meta"
	"github.com/srevinsaju/lyrix/lyrixd/cmd/desktop/routes"
	"github.com/srevinsaju/lyrix/lyrixd/config"
	"github.com/webview/webview"
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
	}

	// launch the daemon to discover changes in the current listening song
	go daemon.Start()

	// create the web server and launch it as a goroutine
	app := routes.BuildServer(auth)
	newAddress, err := GetNewAddress()
	if err != nil {
		logger.Fatal(err)
	}
	go func() {
		app.Listen(newAddress)
	}()

	logger.Infof("Attempting to use '%s'", newAddress)
	// create the web application instance
	debug := true
	w := webview.New(debug)
	defer w.Destroy()
	w.SetTitle("Lyrixd")
	w.SetSize(600, 800, webview.HintNone)
	suffix := ""
	if auth == nil {
		suffix = "login"
	}
	w.Navigate(fmt.Sprintf("http://%s/%s", newAddress, suffix))
	w.Run()
}
