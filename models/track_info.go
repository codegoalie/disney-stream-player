package models

import "time"

type TrackInfo struct {
	Title     string
	Album     string
	Artist    string
	Duration  float64
	StartedAt time.Time
}

type InfoFetcher interface {
	Name() string
	InfoURL() string
	ParseTrackInfo([]byte) (*TrackInfo, error)
}
