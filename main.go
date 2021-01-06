package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"strings"
	"time"

	vlc "github.com/adrg/libvlc-go/v3"
	"github.com/codegoalie/golibnotify"
	"github.com/codegoalie/stream-player/dpark"
	"github.com/codegoalie/stream-player/models"
	"github.com/codegoalie/stream-player/sorcer"
	"github.com/codegoalie/stream-player/utils"
	"github.com/codegoalie/stream-player/wdwnt"
	"github.com/godbus/dbus/v5"
	"github.com/gosuri/uilive"
)

var medias = []models.MediaSource{
	// sorcer.Seasons{},
	// dpark.Christmas{},
	sorcer.Atmospheres{},
	dpark.Background{},
	dpark.Resort{},
	wdwnt.Tunes{},
}

func main() {
	var currentMediaIndex int
	flag.IntVar(&currentMediaIndex, "s", 0, "index of stream to start on")
	flag.Parse()

	quit := make(chan struct{})
	actions := make(chan mediaAction)
	mediaURLs := make(chan string)

	go listenForMediaKeys(actions)
	go playAudio(mediaURLs, quit)

	writer := uilive.New()
	writer.Start()
	defer writer.Stop()
	trackInfoFetchers := make(chan models.InfoFetcher)

	go pollForMetadataUpdates(writer, trackInfoFetchers, quit)

	var currentMedia models.MediaSource

	currentMedia = medias[currentMediaIndex]
	mediaURLs <- currentMedia.StreamURL()
	fmt.Fprintf(writer, fmt.Sprintf("Loading %s...", currentMedia.Name()))
	writer.Flush()
	trackInfoFetchers <- currentMedia
	for {
		select {
		case action := <-actions:
			switch action {
			case nextMediaAction:
				currentMediaIndex = (currentMediaIndex + 1) % len(medias)
				currentMedia = medias[currentMediaIndex]
				mediaURLs <- currentMedia.StreamURL()
				fmt.Fprintf(writer, fmt.Sprintf("Loading %s...", currentMedia.Name()))
				writer.Flush()
				trackInfoFetchers <- currentMedia
			}
		case <-time.After(time.Second):
			trackInfoFetchers <- currentMedia
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
		log.Fatal("failed to init vlc", err)
	}
	defer vlc.Release()

	// Create a new player.
	player, err := vlc.NewPlayer()
	if err != nil {
		log.Fatal("failed to create new vlc player: ", err)
	}
	defer func() {
		player.Stop()
		player.Release()
	}()

	// Retrieve player event manager.
	manager, err := player.EventManager()
	if err != nil {
		log.Fatal("failed to get vlc player event manager", err)
	}

	// Register the media end reached event with the event manager.
	eventCallback := func(event vlc.Event, userData interface{}) {
		close(quit)
	}

	eventID, err := manager.Attach(vlc.MediaPlayerEndReached, eventCallback, nil)
	if err != nil {
		log.Fatal("failed to attach to media end reached event", err)
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
				log.Fatal("failed to load media from url", err)
			}

			// Start playing the media.
			err = player.Play()
			if err != nil {
				log.Fatal("failed to play media", err)
			}
		case <-quit:
			media.Release()
			break
		}
	}
}

func pollForMetadataUpdates(writer io.Writer, trackInfoFetchers <-chan models.InfoFetcher, quit chan struct{}) {
	currentSong := &models.TrackInfo{}
	notifier := golibnotify.NewSimpleNotifier("Stream Player")
	defer notifier.Close()

	var trackFetcher models.InfoFetcher
	var msg strings.Builder
	var oldTitle string
	lastFetchedAt := time.Time{}
	for {
		trackFetcher = <-trackInfoFetchers

		if lastFetchedAt.Before(time.Now().Add(-5 * time.Second)) {
			lastFetchedAt = time.Now()
			buf, err := utils.HTTPGet(trackFetcher.InfoURL())
			if err != nil {
				fmt.Fprintln(writer, "Error: "+err.Error())
				close(quit)
			}

			if len(buf.Bytes()) == 0 {
				fmt.Fprintf(writer, "Metadata fetch error")
				time.AfterFunc(time.Second, func() {
					fmt.Fprintf(writer, msg.String())
				})
				continue
			}

			oldTitle = currentSong.Title

			currentSong, err = trackFetcher.ParseTrackInfo(buf.Bytes())
			if err != nil {
				fmt.Fprintln(writer, "Error: "+err.Error())
				close(quit)
			}
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

		msg.WriteString(trackFetcher.Name() + "\n")
		msg.WriteString(currentSong.Title)

		if currentSong.Artist != "" {
			msg.WriteString(" - ")
			msg.WriteString(currentSong.Artist)
		}

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
		msg = strings.Builder{}

		if oldTitle != currentSong.Title {
			notifier.Update(
				currentSong.Title,
				currentSong.Artist,
				"",
			)
		}

	}
}

type mediaAction int

const (
	playPauseMediaAction mediaAction = iota
	nextMediaAction
	previousMediaAction
)

const hourInSeconds = 60 * 60
