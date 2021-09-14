package sorcer

import (
	"github.com/codegoalie/stream-player/models"
)

const mochaName = "Mocha (Sorcer Radio)"
const mochaStreamURL = "https://samcloud.spacial.com/api/listen?sid=100903&m=sc&rid=177361"

type Mocha struct{}

// Name is the user presentable name for the stream
func (s Mocha) Name() string {
	return mochaName
}

// StreamURL provides the current URL to stream audio
func (s Mocha) StreamURL() string {
	return mochaStreamURL
}

// InfoURL is the URL to fetch track data
func (s Mocha) InfoURL() string {
	return infoURL("100903", "030c8d06bdd9e82eae632eaff484df864c54f14c")
}

// ParseTrackInfo parses the provided bytes into a TrackInfo
func (s Mocha) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	return parseTrackInfo(raw)
}
