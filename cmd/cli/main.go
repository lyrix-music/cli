package main

import (
	"github.com/lyrix-music/cli/auth"
	"github.com/lyrix-music/cli/config"
	"github.com/lyrix-music/cli/meta"
	"github.com/lyrix-music/cli/service"

	"github.com/urfave/cli/v2"
	"github.com/withmandala/go-log"
	"os"
)

var logger = log.New(os.Stdout)

func main() {
	if os.Getenv("DEBUG") == "1" {
		logger = logger.WithDebug()
	}
	app := &cli.App{
		Name:   "Lyrix Daemon",
		Usage:  "A daemon for lyrix music network",
		Action: service.StartDaemon,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "romanize",
				Usage: "Romanize support CJK characters",
			},
			&cli.BoolFlag{
				Name:  "show-lyrics",
				Usage: "Show lyrics on the terminal.",
			},
			&cli.BoolFlag{
				Name: "lastfm-predict",
				Usage: "Use Last.fm suggestions to dynamically modify playlists " +
					"according to your current playing track. (only KDE Elisa)",
			},
			&cli.BoolFlag{
				Name:  "lastfm-scrobble",
				Usage: "Send your current listening song to last fm to get customized tracks",
			},
			&cli.BoolFlag{
				Name: "discord",
			},
		},
		Commands: []*cli.Command{
			{
				Name: "register",
				Action: func(c *cli.Context) error {
					auth.Register()
					return nil
				},
			},
			{
				Name: "login",
				Action: func(c *cli.Context) error {
					auth.Login()
					return nil
				},
			},
			{
				Name: "reset-config",
				Action: func(c *cli.Context) error {
					_, configPath := config.GetPath(meta.AppName)
					logger.Info("Removing old configuration files...")
					return os.RemoveAll(configPath)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logger.Fatal(err)
	}

}
