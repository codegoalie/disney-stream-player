# Disney Stream Player

Streams Disney music from known internet radio stations from the command line.
Close those browser tabs and easily skip to the next stream to hear the music
you want.

## Current Streams

- [Sorcer Radio Atmospheres](http://srsounds.com/popperSRloops.php)
- [DPark Radio background music](https://www.dparkradio.com/dparkradioplayerbm.html)
- [WDWNTunes](https://live365.com/station/WDWNTunes-a31769)

## Prerequisites

### Linux

```
sudo apt install libvlc-dev vlc libnotify-dev
```

## Installation

[Download the latest
release](https://github.com/codegoalie/disney-stream-player/releases) and add
to your path.

```
$ mv disney-stream-player ~/bin/
```

From source, `go run main.go` is sufficient to try things out. Or you can `go
build` and run the created executable. 

## Usage

No args are required. When the program is running, the current track info is
displayed.

```
$ disney-stream-player
Magic Kingdom Caribbean Plaza Area Loop pt1 - Magic Kingdom [Disney Parks] (1:03:51)
```

To change streams, use your media keys' "Next" button. The streams will cycle
through in the above order and start over once the last stream is "skipped".

## Contributing

The current status of this project is `just working`. Many band-aids and duct
tape were used. Music started playing and information I wanted to see was shown
then I paused. I'm very much open to accepting contributions. Check out the 
current issues and comment on any that you'd like to work on. Please, create a
new issue before working on something new; just to discuss before you sink time
into something that might have other implications.

Thanks in advance for your work and help!
