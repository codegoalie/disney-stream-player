package sorcer

import (
	"github.com/codegoalie/stream-player/models"
)

const streamName = "Spa Day (Sorcer Radio)"
const spaStreamURL = "https://streaming.live365.com/a88328"
const spaHistoryURL = "https://api.live365.com/station/a88328"

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
	return parseLive365TrackInfo(raw)
}
