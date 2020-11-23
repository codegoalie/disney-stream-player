package dpark

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codegoalie/stream-player/models"
)

type dParkResponse struct {
	NowPlaying string `json:"nowplaying"`
}

func parseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	resp := &dParkResponse{}
	err := json.Unmarshal(raw, &resp)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal DPark Radio info: %w (%s)", err, string(raw))
		return nil, err
	}

	info := &models.TrackInfo{}

	splits := strings.Split(resp.NowPlaying, " - ")
	if len(splits) < 3 {
		info.Title = resp.NowPlaying
		return info, nil
	}

	info.Title = splits[1]
	info.Artist = splits[2]
	info.Album = splits[0]
	info.Duration = 0
	info.StartedAt = time.Time{}

	return info, nil
}
