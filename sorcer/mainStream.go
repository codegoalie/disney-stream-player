package sorcer

import "github.com/codegoalie/stream-player/models"

type Main struct{}

func (m Main) Name() string {
	return "Sorcer Radio - Main Stream"
}

func (m Main) StreamURL() string {
	return "https://usa15.fastcast4u.com/proxy/wayarena?mp=/1"
}

func (m Main) InfoURL() string {
	return infoURL("67046", "4e2d422c81d81eff066a193572925fa52962dd32")
}

func (m Main) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	return parseTrackInfo(raw)
}
