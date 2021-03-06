package routes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/gofiber/fiber/v2"
	logger2 "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/django"
	"github.com/lyrix-music/cli/cmd/desktop/daemon"
	"github.com/lyrix-music/cli/cmd/desktop/logging"
	"github.com/lyrix-music/cli/mpris"
	"github.com/lyrix-music/cli/player"
	"github.com/lyrix-music/cli/service"
	"github.com/lyrix-music/cli/types"
	k "github.com/srevinsaju/korean-romanizer-go"
	sl "github.com/srevinsaju/swaglyrics-go"
	sltypes "github.com/srevinsaju/swaglyrics-go/types"
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

	app.Get("/similar", func(c *fiber.Ctx) error {
		// Render index
		localPlayer := daemon.GetPlayer()
		if localPlayer == nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		return c.Render("similar", fiber.Map{
			"queue_song_supported": localPlayer.SupportsQueueSong(),
		})
	})

	app.Get("/register", func(c *fiber.Ctx) error {
		// Render index
		return c.Render("register", fiber.Map{})
	})

	app.Get("/api/v1/user/logged-in", func(c *fiber.Ctx) error {
		loginStatus := daemon.GetAuth() != nil
		return c.JSON(map[string]bool{
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
		if err != nil {
			logger.Warn(err)
			return err
		}
		daemon.SetAuth(cfg)
		return c.SendStatus(fiber.StatusAccepted)
	})

	app.Post("/api/v1/prefs/scrobble/:enabled", func(c *fiber.Ctx) error {
		enabled := c.Params("enabled")
		daemon.SetScrobbleEnabled(enabled == "true")
		return c.SendStatus(fiber.StatusAccepted)
	})
	app.Post("/api/v1/prefs/romanize/:enabled", func(c *fiber.Ctx) error {
		enabled := c.Params("enabled")
		daemon.SetRomanizeEnabled(enabled == "true")
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

	app.Get("/api/v1/song/similar", func(c *fiber.Ctx) error {
		auth := daemon.GetAuth()
		if auth == nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		return c.JSON(player.GetSimilar(auth))
	})

	app.Get("/api/v1/player", func(c *fiber.Ctx) error {
		localPlayer := daemon.GetPlayer()
		if localPlayer == nil {
			logger.Warn("The received localPlayer is nil. No media players detected.")
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(localPlayer.GetIdentity())
	})

	app.Get("/api/v1/player/queue/similar", func(c *fiber.Ctx) error {
		// get the similar songs to the current playing song, and then
		// queue similar songs to local player
		auth := daemon.GetAuth()
		localPlayer := daemon.GetPlayer()
		if auth == nil || localPlayer == nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		similarSongs := player.GetSimilar(auth)
		service.QueueSimilarSongs(similarSongs, localPlayer)
		return c.SendStatus(fiber.StatusAccepted)
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

	app.Get("/api/v1beta/updates/lyrics", func(c *fiber.Ctx) error {
		s := daemon.GetSong()
		if s == nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		lyrics, err := sl.GetLyrics(sltypes.Song{Track: s.Track, Artist: s.Artist})
		if err != nil {
			logger.Warn(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if daemon.GetRomanizeEnabled() {
			var newLyrics []string
			lyricsInLines := strings.Split(lyrics, "\n")
			for i := range lyricsInLines {
				newLyrics = append(newLyrics, lyricsInLines[i])
				r := k.NewRomanizer(lyricsInLines[i])
				newLyrics = append(newLyrics, fmt.Sprintf("<span style='color: #999'>%s</span>", r.Romanize()))
			}
			lyrics = strings.Join(newLyrics, "\n")
		}
		l := &PlayerLyrics{
			Lyrics: lyrics,
			CJK:    false,
		}
		for _, v := range l.Lyrics {
			if k.IsHangul(v) {
				l.CJK = true
				break
			}
		}
		return c.JSON(l)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("home", fiber.Map{})
	})

	return app
}
