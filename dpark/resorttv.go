package dpark

import (
	"github.com/codegoalie/stream-player/models"
)

const resortName = "Resort TV (DPark Radio)"
const resortStreamURL = "https://cheetah.streemlion.com/Channel4?1631622328219"
const resortInfoURL = "https://c7.radioboss.fm/w/nowplayinginfo?u=208&nl=1&_=1605627484420"

// Resort streams the resort TV music channel from DPark Radio
type Resort struct{}

// Name is the user presentable name of the stream
func (b Resort) Name() string {
	return resortName
}

// StreamURL provides the current URL to stream audio
func (b Resort) StreamURL() string {
	return resortStreamURL
}

// InfoURL is the URL to fetch track data
func (b Resort) InfoURL() string {
	return resortInfoURL
}

// ParseTrackInfo parses the provided bytes into a TrackInfo
func (b Resort) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	return parseTrackInfo(raw)
}
