package sorcer

import "github.com/codegoalie/stream-player/models"

type Main struct{}

func (m Main) Name() string {
	return "Main Stream (Sorcer Radio)"
}

func (m Main) StreamURL() string {
	return "https://streaming.live365.com/a89268"
}

func (m Main) InfoURL() string {
	return "https://api.live365.com/station/a89268"
}

func (m Main) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	return parseLive365TrackInfo(raw)
}
