package sorcer

import (
	"github.com/codegoalie/stream-player/models"
)

const atmospheresName = "Sorcer Radio Atmospheres"
const atmospheresStreamURL = "https://samcloud.spacial.com/api/listen?sid=100903&m=sc&rid=177361"

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
	return infoURL("100903", "030c8d06bdd9e82eae632eaff484df864c54f14c")
}

// ParseTrackInfo parses the provided bytes into a TrackInfo
func (s Atmospheres) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	return parseTrackInfo(raw)
}
