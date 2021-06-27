package main

import (
	"github.com/srevinsaju/lyrix/lyrixd/auth"
	"github.com/srevinsaju/lyrix/lyrixd/config"
	"github.com/srevinsaju/lyrix/lyrixd/meta"
	"github.com/srevinsaju/lyrix/lyrixd/service"

	"github.com/urfave/cli/v2"
	"github.com/withmandala/go-log"
	"os"
)

var logger = log.New(os.Stdout)


func main() {
	app := &cli.App{
		Name:   "Lyrix Daemon",
		Usage:  "A daemon for lyrix music network",
		Action: service.StartDaemon,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "lastfm-predict",
				Usage: "Use Last.fm suggestions to dynamically modify playlists " +
					   "according to your current playing track. (only KDE Elisa)",
			},

			&cli.BoolFlag{
				Name: "lastfm-scrobble",
				Usage: "Send your current listening song to last fm to get customized tracks",
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
