package wdwnt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/codegoalie/stream-player/models"
)

const tunesName = "WDWNTunes"
const tunesStreamURL = "https://streaming.live365.com/a31769"
const tunesInfoURL = "https://api.live365.com/station/a31769"

type Tunes struct{}

// Name is the userpresentable name of the stream
func (t Tunes) Name() string {
	return tunesName
}

// StreamURL provides the current URL to stream audio
func (t Tunes) StreamURL() string {
	return tunesStreamURL
}

// InfoURL is the URL to fetch track data
func (t Tunes) InfoURL() string {
	return tunesInfoURL
}

// ParseTrackInfo parses the provided bytes into a TrackInfo
func (t Tunes) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	resp := &wdwnTunesResponse{}
	err := json.Unmarshal(raw, &resp)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal WDWNTunes info: %w", err)
		return nil, err
	}
	info := &models.TrackInfo{}

	startedAt, err := time.Parse("2006-01-02 15:04:05-07:00", resp.CurrentTrack.Start)
	if err != nil {
		err = fmt.Errorf("failed to parse WDWNTunes started at info: %w", err)
		startedAt = time.Time{}
	}

	info.Title = resp.CurrentTrack.Title
	info.Artist = resp.CurrentTrack.Artist
	info.Album = ""
	info.Duration = resp.CurrentTrack.Duration
	info.StartedAt = startedAt

	return info, nil
}

type wdwnTunesResponse struct {
	CurrentTrack struct {
		Title    string  `json:"title"`
		Artist   string  `json:"artist"`
		Start    string  `json:"start"`
		Duration float64 `json:"duration"`
	} `json:"current-track"`
}
