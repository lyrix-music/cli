package service

import (
	"bytes"
	"encoding/json"
	"github.com/lyrix-music/cli/types"
	"os"
	"os/exec"
)

type WindowsExporterSong struct {
	Title string `json:"title"`
	Artist string `json:"artist"`
}


func CheckForSongUpdatesWinRTExporter(ctx *Context, auth *types.UserInstance, cmd *exec.Cmd,  song *types.SongMeta) error {

	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logger.Warn("Couldn't detect any song, is any player running?", err)
		return err
	}

	wSong := &WindowsExporterSong{}
	out, err := cmd.Output()
	if err != nil {
		logger.Warn("No output received from win_exporter, skipping", err)
	}
	logger.Info(string(out))

	err = json.Unmarshal(out, wSong)
	if err != nil {
		logger.Warn("failed to parse output from win_exporter")
		panic(err)
	}


	svcSong := &ServiceSong{
		Artist:   wSong.Artist,
		Title:    wSong.Title,
		Status:   "Playing", // win_exporter doesn't return this yet
	}

	return checkForSongUpdates(ctx, auth, svcSong, song)

}