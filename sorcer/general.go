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

const (
	stationURL     = "http://listen.samcloud.com/webapi/station/%s/history"
	resultCount    = "5"
	mediaTypeCodes = "MUS,COM,NWS,INT"
	format         = "json"
)

type sorcerRadioSong struct {
	Title       string `json:"Title"`
	Album       string `json:"Album"`
	Artist      string `json:"Artist"`
	Duration    string `json:"Duration"`
	DatePlayed  string `json:"DatePlayed"`
	MediaItemID string `json:"MediaItemId"`
}

func infoURL(stationID, token string) string {
	uri, _ := url.Parse(fmt.Sprintf(stationURL, stationID))
	query := uri.Query()
	query.Add("token", token)
	query.Add("top", resultCount)
	query.Add("mediaTypeCodes", mediaTypeCodes)
	query.Add("format", format)
	query.Add("_", strconv.FormatInt(time.Now().Unix(), 10))
	uri.RawQuery = query.Encode()

	return uri.String()
}

func parseTrackInfo(raw []byte) (*models.TrackInfo, error) {
	recentSongs := []sorcerRadioSong{}
	err := json.Unmarshal(raw, &recentSongs)
	if err != nil {
		err = fmt.Errorf(
			"failed to unmarshal sorcer radio history: %w (%s)",
			err,
			string(raw),
		)
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
			err = fmt.Errorf("failed to parse Sorcer started at: %w", err)
			return info, err
		}
		startedAt := time.Unix(unixMillisecs/1000, 0)
		info.StartedAt = startedAt

	}

	return info, nil
}
