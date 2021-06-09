package player

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/srevinsaju/lyrix/lyrixd/types"
	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout)
var client = &http.Client{}

func NotPlayingSongHandler(auth types.UserInstance) {
	song := &types.SongMeta{Track: "", Artist: ""}
	PlayingSongHandler(auth, song)
}

func PlayingSongHandler(auth types.UserInstance, song *types.SongMeta) {
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
