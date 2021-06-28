package routes

import (
	"encoding/json"
	"github.com/godbus/dbus/v5"
	"github.com/gofiber/fiber/v2"
	logger2 "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/django"
	"github.com/srevinsaju/lyrix/lyrixd/cmd/desktop/daemon"
	"github.com/srevinsaju/lyrix/lyrixd/cmd/desktop/logging"
	"github.com/srevinsaju/lyrix/lyrixd/mpris"
	"github.com/srevinsaju/lyrix/lyrixd/types"
	sl "github.com/srevinsaju/swaglyrics-go"
	sltypes "github.com/srevinsaju/swaglyrics-go/types"
	"os"
	"path/filepath"
)

var logger = logging.GetLogger()


func BuildServer(cfg *types.UserInstance) *fiber.App {
	templatesDir := "templates"
	staticDir := "static"
	if os.Getenv("APPDIR") != "" {
		appDir := os.Getenv("APPDIR")
		templatesDir = filepath.Join(appDir, templatesDir)
		staticDir = filepath.Join(appDir, staticDir)
	}
	engine := django.New(templatesDir, ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Use(logger2.New())

	app.Static("/s", staticDir)

	app.Get("/login", func(c *fiber.Ctx) error {
		// Render index
		return c.Render("login", fiber.Map{})
	})

	app.Get("/register", func(c *fiber.Ctx) error {
		// Render index
		return c.Render("register", fiber.Map{})
	})

	app.Get("/api/v1/user/logged-in", func(c *fiber.Ctx) error {
		loginStatus := daemon.GetAuth() != nil
		return c.JSON(map[string]bool {
			"logged_in": loginStatus,
		})
	})

	app.Get("/api/v1/config", func(c *fiber.Ctx) error {
		if cfg != nil {
			return c.JSON(cfg)
		}
		return c.SendStatus(fiber.StatusNotFound)
	})

	app.Post("/api/v1/config", func(c *fiber.Ctx) error {
		err := json.Unmarshal(c.Body(), cfg)
		daemon.SetAuth(cfg)
		if err != nil {
			return err
		}
		return c.SendStatus(fiber.StatusAccepted)
	})

	app.Post("/api/v1/prefs/scrobble/:enabled", func(c *fiber.Ctx) error {
		enabled := c.Params("enabled")
		daemon.SetScrobbleEnabled(enabled == "true")
		return c.SendStatus(fiber.StatusAccepted)
	})

	app.Post("/api/v1/player", func(c *fiber.Ctx) error {
		// set the current music player listener
		playerReq := &PlayerChangeRequest{}
		err := json.Unmarshal(c.Body(), playerReq)
		if err != nil {
			return err
		}
		if playerReq.DbusId == "" {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		logger.Debug("Received dbus id", playerReq.DbusId)
		playerName := daemon.SetPlayer(playerReq.DbusId)
		return c.SendString(playerName)
	})

	app.Get("/api/v1/player", func(c *fiber.Ctx) error {
		player := daemon.GetPlayer()
		if player == nil {
			logger.Warn("The received player is nil. No media players detected")
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(player.GetIdentity())
	})

	app.Get("/api/v1/updates/players", func(c *fiber.Ctx) error {
		conn, err := dbus.SessionBus()
		if err != nil {
			panic(err)
		}
		names, err := mpris.List(conn)
		if err != nil {
			logger.Warn("Failed to fetch the players over mpris", err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(names)
	})

	app.Get("/api/v1/updates/song", func(c *fiber.Ctx) error {
		if daemon.GetSong() == nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return c.JSON(*daemon.GetSong())
	})

	app.Get("/api/v1/updates/lyrics", func(c *fiber.Ctx) error {
		s := daemon.GetSong()
		if s == nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		lyrics, err := sl.GetLyrics(sltypes.Song{Track: s.Track, Artist: s.Artist})
		if err != nil {
			logger.Warn(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.SendString(lyrics)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("home", fiber.Map{})
	})

	return app
}
