package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"strings"
	"time"

	vlc "github.com/adrg/libvlc-go"
	"github.com/codegoalie/golibnotify"
	"github.com/codegoalie/stream-player/dpark"
	"github.com/codegoalie/stream-player/models"
	"github.com/codegoalie/stream-player/sorcer"
	"github.com/codegoalie/stream-player/utils"
	"github.com/codegoalie/stream-player/wdwnt"
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

type InfoFetcher interface {
	InfoURL() string
	ParseTrackInfo([]byte) (*models.TrackInfo, error)
}

// MediaSource is a streamable audio source which can fetch its own TrackInfo
type MediaSource interface {
	Name() string
	StreamURL() string
	InfoURL() string
	ParseTrackInfo([]byte) (*models.TrackInfo, error)
}

var medias = []MediaSource{
	sorcer.Atmospheres{},
	dpark.Background{},
	wdwnt.Tunes{},
}

func main() {
	quit := make(chan struct{})
	actions := make(chan mediaAction)
	mediaURLs := make(chan string)
	trackInfoFetchers := make(chan InfoFetcher)
	go listenForMediaKeys(actions)
	go playAudio(mediaURLs, quit)

	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	go pollForMetadataUpdates(writer, trackInfoFetchers, quit)

	currentMediaIndex := 0
	var currentMedia MediaSource

	for {
		currentMedia = medias[currentMediaIndex]
		mediaURLs <- currentMedia.StreamURL()
		fmt.Fprintf(writer, fmt.Sprintf("Loading %s...\n", currentMedia.Name()))
		trackInfoFetchers <- currentMedia
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

func pollForMetadataUpdates(writer io.Writer, trackInfoFetchers <-chan InfoFetcher, quit chan struct{}) {
	currentSong := &models.TrackInfo{}
	notifier := golibnotify.NewSimpleNotifier("Stream Player")
	defer notifier.Close()

	trackFetcher := <-trackInfoFetchers
	for {
		buf, err := utils.HTTPGet(trackFetcher.InfoURL())
		if err != nil {
			fmt.Fprintln(writer, "Error: "+err.Error())
			close(quit)
		}

		oldTitle := currentSong.Title
		currentSong, err = trackFetcher.ParseTrackInfo(buf.Bytes())
		if err != nil {
			fmt.Fprintln(writer, "Error: "+err.Error())
			close(quit)
		}

		duration := ""
		if currentSong.Duration > 0 {
			hours := math.Floor(currentSong.Duration / hourInSeconds)
			if hours > 0 {
				duration += fmt.Sprintf("%.f", hours) + ":"
			}
			minutes := math.Floor(math.Mod(currentSong.Duration, hourInSeconds) / 60)
			seconds := math.Mod(currentSong.Duration, 60)
			duration += fmt.Sprintf("%02.f:%02.f", minutes, seconds)
		}

		endsAt := currentSong.StartedAt.Add(time.Second * time.Duration(currentSong.Duration))
		left := time.Until(endsAt)

		msg := strings.Builder{}
		msg.WriteString(currentSong.Title)
		msg.WriteString(" - ")
		msg.WriteString(currentSong.Artist)
		if currentSong.Album != "" {
			msg.WriteString(" [")
			msg.WriteString(currentSong.Album)
			msg.WriteString("]")
		}

		if currentSong.Duration > 0 {
			msg.WriteString(" (")
			if left > 0 {
				msg.WriteString(fmt.Sprintf("%02.f:%02.f", math.Floor(left.Minutes()), math.Mod(left.Seconds(), 60)))
				msg.WriteString(" / ")
			}
			msg.WriteString(duration)
			msg.WriteString(")")
		}

		msg.WriteString("\n")

		fmt.Fprintf(writer, msg.String())

		if oldTitle != currentSong.Title {
			notifier.Update(
				currentSong.Title,
				currentSong.Artist,
				"",
			)
		}

		select {
		case trackFetcher = <-trackInfoFetchers:
		default:
			time.Sleep(time.Second * 5)
		}
	}
}
