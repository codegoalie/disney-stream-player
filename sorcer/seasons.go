package sorcer

import (
	"github.com/codegoalie/stream-player/models"
)

const seasonsName = "Seasons (Sorcer Radio )"
const seasonsStreamURL = "http://19293.live.streamtheworld.com/SP_R2809833"

type Seasons struct{}

// Name is the user presentable name for the stream
func (s Seasons) Name() string {
	return seasonsName
}

// StreamURL provides the current URL to stream audio
func (s Seasons) StreamURL() string {
	return seasonsStreamURL
}

// InfoURL is the URL to fetch track data
func (s Seasons) InfoURL() string {
	return infoURL("104853", "254aec990e7d964645bc5fb68c58d45448f7719d")
}

// ParseTrackInfo parses the provided bytes into a TrackInfo
func (s Seasons) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	return parseTrackInfo(raw)
}
