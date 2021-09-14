package sorcer

import (
	"github.com/codegoalie/stream-player/models"
)

const atmospheresName = "Atmospheres (Sorcer Radio)"
const atmospheresStreamURL = "https://samcloud.spacial.com/api/listen?sid=130157&m=sc&rid=273285"

type Atmospheres struct{}

// Name is the user presentable name for the stream
func (s Atmospheres) Name() string {
	return atmospheresName
}

// StreamURL provides the current URL to stream audio
func (s Atmospheres) StreamURL() string {
	return atmospheresStreamURL
}

// InfoURL is the URL to fetch track data
func (s Atmospheres) InfoURL() string {
	return infoURL("130157", "acce5d6b010ebf1438bc1990f4cd357556aecf3b")
}

// ParseTrackInfo parses the provided bytes into a TrackInfo
func (s Atmospheres) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	return parseTrackInfo(raw)
}
