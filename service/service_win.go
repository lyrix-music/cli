package service

import (
	"bytes"
	"encoding/json"
	"github.com/lyrix-music/cli/types"
	"os"
        "strings"
	"os/exec"
)

type WindowsExporterSong struct {
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Source    string `json:"source"`
	IsPlaying bool   `json:"is_playing"`
}

func CheckForSongUpdatesWinRTExporter(ctx *Context, auth *types.UserInstance, exporterPath string, song *types.SongMeta) error {
	cmd := exec.Command(exporterPath)

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logger.Warn("Couldn't detect any song, is any player running?", err)
		return err
	}

	wSong := &WindowsExporterSong{}
	out := cmdOutput.Bytes()
	if err != nil {
		logger.Warn("No output received from win_exporter, skipping", err)
	}
	logger.Info(string(out))

	err = json.Unmarshal(out, wSong)
	if err != nil {
		logger.Warn("failed to parse output from win_exporter")
		panic(err)
	}
	status := "Paused"
	if wSong.IsPlaying {
		status = "Playing"
	}

	source := "local"
	if source == "Groove Music" {
		source = "groove-music"
	} else if source == "VLC" {
		source = "vlc"
	} else {
		source = strings.Replace(" ", "-", source, -1)
	}

	svcSong := &ServiceSong{
		Artist: wSong.Artist,
		Title:  wSong.Title,
		Source: source,

		Status: status,
	}

	return checkForSongUpdates(ctx, auth, svcSong, song)

}
