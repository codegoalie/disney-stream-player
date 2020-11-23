package models

// MediaSource is a streamable audio source which can fetch its own TrackInfo
type MediaSource interface {
	Name() string
	StreamURL() string
	InfoURL() string
	ParseTrackInfo([]byte) (*TrackInfo, error)
}
