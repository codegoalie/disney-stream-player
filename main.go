package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/0xAX/notificator"
	vlc "github.com/adrg/libvlc-go"
	"github.com/gosuri/uilive"
)

type song struct {
	Title       string `json:"Title"`
	Album       string `json:"Album"`
	Artist      string `json:"Artist"`
	Duration    string `json:"Duration"`
	DatePlayed  string `json:"DatePlayed"`
	MediaItemId string `json:"MediaItemId"`
}

func main() {
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

	// Add a media file from path or from URL.
	// Set player media from path:
	// media, err := player.LoadMediaFromPath("localpath/test.mp4")
	// Set player media from URL:
	media, err := player.LoadMediaFromURL("https://samcloud.spacial.com/api/listen?sid=100903&m=sc&rid=177361")
	if err != nil {
		log.Fatal(err)
	}
	defer media.Release()

	// Start playing the media.
	err = player.Play()
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve player event manager.
	manager, err := player.EventManager()
	if err != nil {
		log.Fatal(err)
	}

	// Register the media end reached event with the event manager.
	quit := make(chan struct{})
	eventCallback := func(event vlc.Event, userData interface{}) {
		close(quit)
	}

	eventID, err := manager.Attach(vlc.MediaPlayerEndReached, eventCallback, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer manager.Detach(eventID)

	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	go fetchTrackInfo(writer, quit)

	<-quit
}

func fetchTrackInfo(writer io.Writer, quit chan struct{}) {
	var currentSong song
	var notify *notificator.Notificator
	notify = notificator.New(notificator.Options{
		DefaultIcon: "icon/micke.png",
		AppName:     "Stream Player",
	})

	durationRegexp := regexp.MustCompile(`^PT(?:(\d+)H)?(?:(\d+)M)?([0-9\.]+)S$`)

	for {
		uri, _ := url.Parse("http://listen.samcloud.com/webapi/station/100903/history")
		query := uri.Query()
		query.Add("token", "030c8d06bdd9e82eae632eaff484df864c54f14c")
		query.Add("top", "5")
		query.Add("mediaTypeCodes", "MUS,COM,NWS,INT")
		query.Add("format", "json")
		query.Add("_", strconv.FormatInt(time.Now().Unix(), 10))
		uri.RawQuery = query.Encode()

		resp, err := defaultHTTPClient.Get(uri.String())
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

		recentSongs := []song{}
		err = json.Unmarshal(buf.Bytes(), &recentSongs)
		if err != nil {
			fmt.Fprintln(writer, "Error: "+err.Error())
			close(quit)
		}

		if len(recentSongs) > 0 {
			newSong := recentSongs[0].Title != currentSong.Title
			currentSong = recentSongs[0]

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

			duration := ""
			if hours > 0 {
				duration += fmt.Sprintf("%02d", hours) + ":"
			}
			duration += fmt.Sprintf("%02d:%02.f", minutes, seconds)

			msg := fmt.Sprintf(
				"%s - %s [%s] (%s)\n",
				currentSong.Title,
				currentSong.Artist,
				currentSong.Album,
				duration,
			)
			if newSong {
				fmt.Fprintf(writer, msg)
				notify.Push(
					currentSong.Title,
					currentSong.Artist,
					"https://prosamcloudmedia.blob.core.windows.net/67851-public/"+currentSong.MediaItemId+"_144x144.jpg",
					notificator.UR_NORMAL,
				)
			}
		}

		time.Sleep(time.Second * 5)
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
