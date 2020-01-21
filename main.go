package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/0xAX/notificator"
	vlc "github.com/adrg/libvlc-go"
	"github.com/godbus/dbus/v5"
	"github.com/gosuri/uilive"
)

type mediaAction int

const (
	playPauseMediaAction mediaAction = iota
	nextMediaAction
	previousMediaAction
)

const hourInSeconds = 60 * 60

type trackInfo struct {
	Title     string
	Album     string
	Artist    string
	Duration  float64
	StartedAt time.Time
}

type sorcerRadioSong struct {
	Title       string `json:"Title"`
	Album       string `json:"Album"`
	Artist      string `json:"Artist"`
	Duration    string `json:"Duration"`
	DatePlayed  string `json:"DatePlayed"`
	MediaItemID string `json:"MediaItemId"`
}

type dParkResponse struct {
	NowPlaying string `json:"nowplaying"`
}

type infoFetcher struct {
	infoURL       string
	unmarshalJSON func([]byte, *trackInfo) error
}

type media struct {
	name          string
	streamURL     string
	infoURL       func() string
	unmarshalJSON func([]byte, *trackInfo) error
}

var medias = []media{
	{
		name:      "Sorcer Radio Atmospheres",
		streamURL: "https://samcloud.spacial.com/api/listen?sid=100903&m=sc&rid=177361",
		infoURL: func() string {
			uri, _ := url.Parse("http://listen.samcloud.com/webapi/station/100903/history")
			query := uri.Query()
			query.Add("token", "030c8d06bdd9e82eae632eaff484df864c54f14c")
			query.Add("top", "5")
			query.Add("mediaTypeCodes", "MUS,COM,NWS,INT")
			query.Add("format", "json")
			query.Add("_", strconv.FormatInt(time.Now().Unix(), 10))
			uri.RawQuery = query.Encode()

			return uri.String()
		},
		unmarshalJSON: func(raw []byte, info *trackInfo) error {
			recentSongs := []sorcerRadioSong{}
			err := json.Unmarshal(raw, &recentSongs)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal sorcer radio history: %w", err)
				return err
			}

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

			}
			return nil
		},
	},
	{
		name:      "DPark Radio Background",
		streamURL: "https://str2b.openstream.co/578?aw_0_1st.collectionid=3127&aw_0_1st.publisherId=602",
		infoURL:   func() string { return "https://c5.radioboss.fm/api/info/38" },
		unmarshalJSON: func(raw []byte, info *trackInfo) error {
			resp := &dParkResponse{}
			err := json.Unmarshal(raw, &resp)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal DPark Radio info: %w", err)
				return err
			}

			splits := strings.Split(resp.NowPlaying, " - ")
			if len(splits) < 3 {
				info.Title = resp.NowPlaying
				return nil
			}

			info.Title = splits[1]
			info.Artist = splits[2]
			info.Album = splits[0]

			return nil
		},
	},
}

func main() {
	quit := make(chan struct{})
	actions := make(chan mediaAction)
	mediaURLs := make(chan string)
	trackInfoFetchers := make(chan infoFetcher)
	go listenForMediaKeys(actions)
	go playAudio(mediaURLs, quit)

	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	go pollForMetadataUpdates(writer, trackInfoFetchers, quit)

	currentMediaIndex := 0
	var currentMedia media

	for {
		currentMedia = medias[currentMediaIndex]
		mediaURLs <- currentMedia.streamURL
		fmt.Fprintf(writer, fmt.Sprintf("Loading %s...\n", currentMedia.name))
		trackInfoFetchers <- struct {
			infoURL       string
			unmarshalJSON func([]byte, *trackInfo) error
		}{
			infoURL:       currentMedia.infoURL(),
			unmarshalJSON: currentMedia.unmarshalJSON,
		}
		select {
		case action := <-actions:
			switch action {
			case nextMediaAction:
				currentMediaIndex = (currentMediaIndex + 1) % len(medias)
			}
		case <-quit:
			return
		}
	}
}

func listenForMediaKeys(actions chan<- mediaAction) {
	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	obj := conn.Object("org.gnome.SettingsDaemon.MediaKeys", "/org/gnome/SettingsDaemon/MediaKeys")
	call := obj.Call("org.gnome.SettingsDaemon.MediaKeys.GrabMediaPlayerKeys", 0, "dbus-test", uint32(0))
	if call.Err != nil {
		panic(call.Err)
	}

	if err = conn.AddMatchSignal(
		dbus.WithMatchObjectPath("/org/gnome/SettingsDaemon/MediaKeys/MediaPlayerKeyPressed"),
		dbus.WithMatchInterface("org.gnome.SettingsDaemon.MediaKeys"),
		dbus.WithMatchSender("org.gnome.SettingsDaemon.MediaKeys"),
	); err != nil {
		panic(err)
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for v := range c {
		if len(v.Body) < 2 {
			continue
		}

		if msg, ok := v.Body[1].(string); ok && msg == "Next" {
			actions <- nextMediaAction
		}
	}
}

func playAudio(nextMediaURL <-chan string, quit chan struct{}) {
	// Initialize libvlc. Additional command line arguments can be passed in
	// to libvlc by specifying them in the Init function.
	if err := vlc.Init("--no-video", "--quiet"); err != nil {
		log.Fatal(err)
	}
	defer vlc.Release()

	// Create a new player.
	player, err := vlc.NewPlayer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		player.Stop()
		player.Release()
	}()

	// Retrieve player event manager.
	manager, err := player.EventManager()
	if err != nil {
		log.Fatal(err)
	}

	// Register the media end reached event with the event manager.
	eventCallback := func(event vlc.Event, userData interface{}) {
		close(quit)
	}

	eventID, err := manager.Attach(vlc.MediaPlayerEndReached, eventCallback, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer manager.Detach(eventID)

	var media *vlc.Media
	for {

		select {
		case currentMediaURL := <-nextMediaURL:
			if media != nil {
				media.Release()
			}
			media, err = player.LoadMediaFromURL(currentMediaURL)
			if err != nil {
				log.Fatal(err)
			}

			// Start playing the media.
			err = player.Play()
			if err != nil {
				log.Fatal(err)
			}
		case <-quit:
			media.Release()
			break
		}
	}
}

func pollForMetadataUpdates(writer io.Writer, trackInfoFetchers <-chan infoFetcher, quit chan struct{}) {
	var currentSong trackInfo
	var notify *notificator.Notificator
	notify = notificator.New(notificator.Options{
		DefaultIcon: "icon/micke.png",
		AppName:     "Stream Player",
	})

	trackFetcher := <-trackInfoFetchers
	for {
		resp, err := defaultHTTPClient.Get(trackFetcher.infoURL)
		if err != nil {
			fmt.Fprintln(writer, "Error: "+err.Error())
			close(quit)
		}

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			fmt.Fprintln(writer, "Error: "+err.Error())
			close(quit)
		}

		oldTitle := currentSong.Title
		err = trackFetcher.unmarshalJSON(buf.Bytes(), &currentSong)
		if err != nil {
			fmt.Fprintln(writer, "Error: "+err.Error())
			close(quit)
		}

		if oldTitle != currentSong.Title {
			duration := ""
			if currentSong.Duration > 0 {
				duration += " ("
				hours := math.Floor(currentSong.Duration / hourInSeconds)
				if hours > 0 {
					duration += fmt.Sprintf("%.f", hours) + ":"
				}
				minutes := math.Floor(math.Mod(currentSong.Duration, hourInSeconds) / 60)
				seconds := math.Mod(currentSong.Duration, 60)
				duration += fmt.Sprintf("%02.f:%02.f", minutes, seconds)
				duration += ")"
			}

			msg := fmt.Sprintf(
				"%s - %s [%s]%s\n",
				currentSong.Title,
				currentSong.Artist,
				currentSong.Album,
				duration,
			)

			fmt.Fprintf(writer, msg)
			notify.Push(
				currentSong.Title,
				currentSong.Artist,
				"",
				notificator.UR_NORMAL,
			)
		}

		select {
		case trackFetcher = <-trackInfoFetchers:
		default:
			time.Sleep(time.Second * 5)
		}
	}
}

var defaultHTTPClient = &http.Client{
	Timeout: time.Second * 30,
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Second * 10,
		}).Dial,
		TLSHandshakeTimeout: time.Second * 10,
	},
}
