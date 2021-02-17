package sorcer

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/codegoalie/stream-player/models"
)

const streamName = "Spa Day"
const spaStreamURL = "https://cheetah.streemlion.com:1705/stream"
const spaHistoryURL = "https://cheetah.streemlion.com:1330/api/v2/history/?limit=1&offset=0&server=2"

type SpaDay struct{}

func (s SpaDay) Name() string {
	return streamName
}

func (s SpaDay) StreamURL() string {
	return spaStreamURL
}

func (s SpaDay) InfoURL() string {
	return spaHistoryURL
}

// type TrackInfo struct {
// 	Title     string
// 	Album     string
// 	Artist    string
// 	Duration  float64
// 	StartedAt time.Time
// }

func (s SpaDay) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	resp := streemlionHistory{}
	err := json.Unmarshal(raw, &resp)
	if err != nil {
		err = fmt.Errorf(
			"failed to unmarshal sorcer radio history: %w (%s)",
			err,
			string(raw),
		)
		return nil, err
	}

	if len(resp.Results) < 1 {
		return nil, errors.New("no track results found")
	}

	track := resp.Results[0]
	info := models.TrackInfo{
		Title:     track.Title,
		Artist:    track.Author,
		Duration:  track.Length / 1000,
		StartedAt: time.Unix(track.TS/1000, 0),
	}

	return &info, nil
}

type streemlionHistory struct {
	Count   int    `json:"count"`
	Next    string `json:"next"`
	Results []struct {
		ID            int     `json:"id"`
		ImgURL        string  `json:"img_url"`
		TS            int64   `json:"ts"`
		Metadata      string  `json:"metadata"`
		Author        string  `json:"author"`
		Title         string  `json:"title"`
		PlaylistTitle string  `json:"playlist_title"`
		DJName        string  `json:"dj_name"`
		Listeners     int     `json:"n_listeners"`
		Length        float64 `json:"length"`
		ImgMediumURL  string  `json:"img_medium_url"`
		ImgLargeURL   string  `json:"img_large_url"`
		AllMusicID    int     `json:"all_music_id"`
	} `json:"results"`
}
