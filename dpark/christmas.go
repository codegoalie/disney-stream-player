package dpark

import (
	"github.com/codegoalie/stream-player/models"
)

const christmasName = "DPark Radio Christmas"
const christmasStreamURL = "https://str2b.openstream.co/1246?aw_0_1st.collectionid=4287&stationId=4287&publisherId=1270&k=1605627294"
const christmasInfoURL = "https://c11.radioboss.fm/w/nowplayinginfo?u=39"

// Christmas streams the christmas music channel from DPark Radio
type Christmas struct{}

// Name is the userpresentable name of the stream
func (b Christmas) Name() string {
	return christmasName
}

// StreamURL provides the current URL to stream audio
func (b Christmas) StreamURL() string {
	return christmasStreamURL
}

// InfoURL is the URL to fetch track data
func (b Christmas) InfoURL() string {
	return christmasInfoURL
}

// ParseTrackInfo parses the provided bytes into a TrackInfo
func (b Christmas) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	return parseTrackInfo(raw)
}
