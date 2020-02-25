package sorcer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	uri, _ := url.Parse("http://listen.samcloud.com/webapi/station/100903/history")
	query := uri.Query()
	query.Add("token", "030c8d06bdd9e82eae632eaff484df864c54f14c")
	query.Add("top", "5")
	query.Add("mediaTypeCodes", "MUS,COM,NWS,INT")
	query.Add("format", "json")
	query.Add("_", strconv.FormatInt(time.Now().Unix(), 10))
	uri.RawQuery = query.Encode()

	return uri.String()
}

// ParseTrackInfo parses the provided bytes into a TrackInfo
func (s Atmospheres) ParseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	recentSongs := []sorcerRadioSong{}
	err := json.Unmarshal(raw, &recentSongs)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal sorcer radio history: %w", err)
		return nil, err
	}

	info := &models.TrackInfo{}
	if len(recentSongs) > 0 {
		currentSong := recentSongs[0]
		info.Title = currentSong.Title
		info.Artist = currentSong.Artist
		info.Album = currentSong.Album

		durationRegexp := regexp.MustCompile(`^PT(?:(\d+)H)?(?:(\d+)M)?([0-9\.]+)S$`)
		matches := durationRegexp.FindAllStringSubmatch(currentSong.Duration, -1)

		hours, err := strconv.Atoi(matches[0][1])
		if err != nil {
			hours = 0
		}

		minutes, err := strconv.Atoi(matches[0][2])
		if err != nil {
			minutes = 0
		}

		seconds, err := strconv.ParseFloat(matches[0][3], 10)
		if err != nil {
			fmt.Println("failed to parse seconds", err)
			seconds = 0
		}

		info.Duration = float64(hours*60*60) + float64(minutes*60) + seconds

		unixStr := strings.Split(strings.Trim(currentSong.DatePlayed, "\\/Date()"), "+")[0]
		unixMillisecs, err := strconv.ParseInt(unixStr, 10, 64)
		if err != nil {
			err = fmt.Errorf("failed to parse Sorcer atmospheres started at: %w", err)
			return info, err
		}
		startedAt := time.Unix(unixMillisecs/1000, 0)
		info.StartedAt = startedAt

	}
	return info, nil

}

type sorcerRadioSong struct {
	Title       string `json:"Title"`
	Album       string `json:"Album"`
	Artist      string `json:"Artist"`
	Duration    string `json:"Duration"`
	DatePlayed  string `json:"DatePlayed"`
	MediaItemID string `json:"MediaItemId"`
}
