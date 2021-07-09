package player

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/lyrix-music/cli/types"
	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout)
var client = &http.Client{}

func NotPlayingSongHandler(auth *types.UserInstance) {
	song := &types.SongMeta{Track: "", Artist: "", Position: 0}
	PlayingSongHandler(auth, song)
}

func PlayingSongHandler(auth *types.UserInstance, song *types.SongMeta) {
	// do not do anything for a user who havent auth'd yet.
	if auth == nil {
		return
	}
	jsonStr, err := json.Marshal(song)
	if err != nil {
		logger.Warn(err)
		return
	}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/user/player/local/current_song", auth.Host),
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		logger.Warn(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.Token))

	resp, err := client.Do(req)

	if err != nil {
		logger.Warn(err)
	}
	if resp == nil {
		logger.Warn("Failed to complete request")
		return
	}
	defer resp.Body.Close()
}

func GetSimilar(auth *types.UserInstance) []types.SongMeta {
	// do not do anything for a user who havent auth'd yet.

	var songs []types.SongMeta

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/user/player/local/current_song/similar", auth.Host),
		nil,
	)
	if err != nil {
		logger.Warn(err)
		return songs
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.Token))

	resp, err := client.Do(req)

	if err != nil {
		logger.Warn(err)
		return songs
	}
	if resp == nil {
		logger.Warn("Failed to complete request")
		return songs
	}
	defer resp.Body.Close()

	similarSongsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warn(err)
		return songs
	}
	err = json.Unmarshal(similarSongsBytes, &songs)
	if err != nil {
		logger.Warn(err)
		return songs
	}
	return songs

}
