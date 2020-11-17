package dpark

import (
	"github.com/codegoalie/stream-player/models"
)

const backgroundName = "DPark Radio Background"
const backgroundStreamURL = "https://str2b.openstream.co/578?aw_0_1st.collectionid=3127&aw_0_1st.publisherId=602"
const backgroundInfoURL = "https://c5.radioboss.fm/api/info/38"

// Background streams the background music channel from DPark Radio
type Background struct{}

// Name is the userpresentable name of the stream
func (b Background) Name() string {
	return backgroundName
}

// StreamURL provides the current URL to stream audio
func (b Background) StreamURL() string {
	return backgroundStreamURL
}

// InfoURL is the URL to fetch track data
func (b Background) InfoURL() string {
	return backgroundInfoURL
}

// ParseTrackInfo parses the provided bytes into a TrackInfo
func (b Background) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	return parseTrackInfo(raw)
}
