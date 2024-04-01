package main

import (
	"encoding/json"
	// "errors"
	"fmt"
	"log"
	"os"

	// "os/exec"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/gopxl/beep"

	// "github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	// gomp3 "github.com/hajimehoshi/go-mp3"
	"github.com/itchyny/gojq"
)

// struct for playlist tracks
type Track struct {
	Title  string
	Artist string
	ID     string
}

type Pair struct {
	idx      int
	priority bool
}

const APP_WIDTH = 500
const APP_HEIGHT = 500

// player controller - current playing track info
var playerCtrl *CtrlVolume

// download queue - holds tracks that need to be downloaded
var downloadQueue *DoubleList

var sampleRate = beep.SampleRate(48000)

var tracksDir string = "tracks"
var outputFilename string = fmt.Sprintf("%s/%%(id)s", tracksDir)

// global GUI elements
var playlistList *widget.List
var previousTrackBtn, playPauseBtn, nextTrackBtn *widget.Button

// main playlist slice
var playlist []Track

// current position in playlist
var playlistIndex int = -1

// play and download thread sync channels
var playChannel chan int
var dldChannel chan bool

// read playlist JSON and return a slice of tracks
// TODO: replace with call to get new playlist
func readPlaylist() []Track {
	playlistTracksJSON, err := os.ReadFile("mix.json")
	if err != nil {
		log.Fatal(err)
	}

	// unpack playlist JSON into map
	var playlistTracksMap map[string]any
	json.Unmarshal(playlistTracksJSON, &playlistTracksMap)

	// query to parse JSON into tracks
	playlistQuery, err := gojq.Parse(".playlist[]")
	if err != nil {
		log.Fatalln(err)
	}

	var playlistSlice []Track

	// run the query on JSON, iterate through tracks and add to slice
	iter := playlistQuery.Run(playlistTracksMap)
	for trackMap, ok := iter.Next(); ok; trackMap, ok = iter.Next() {
		trackTitleQuery, _ := gojq.Parse(".name")
		trackTitle, _ := trackTitleQuery.Run(trackMap).Next()

		trackArtistQuery, _ := gojq.Parse(".artists[0].name")
		trackArtist, _ := trackArtistQuery.Run(trackMap).Next()

		trackIDQuery, _ := gojq.Parse(".playlinks[0].id")
		trackID, _ := trackIDQuery.Run(trackMap).Next()

		newTrack := Track{trackTitle.(string), trackArtist.(string), trackID.(string)}
		playlistSlice = append(playlistSlice, newTrack)
	}

	return playlistSlice
}

// get track filename from videoID
func getTrackLocation(videoID string) string {
	return fmt.Sprint(tracksDir, "/", videoID, ".mp3")
}

func playlistSelect(idx int) {
	playChannel <- idx
}

func main() {
	// configure app and mainWindow
	app := app.New()
	mainWindow := app.NewWindow("LastFM Player")
	mainWindow.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))

	// init speaker
	err := speaker.Init(sampleRate, sampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}

	// init player controller
	playerCtrl = &CtrlVolume{
		Streamer: nil,
		Paused:   true,
		Silent:   false,
		Base:     2.0,
		Volume:   0.0,
	}

	// init download queue
	downloadQueue = &DoubleList{
		first: nil,
		last:  nil,
		empty: true,
		size:  0,
	}

	// get playlist data as a slice
	playlist = readPlaylist()

	// make list with playlist data as labels
	playlistList = widget.NewList(
		func() int {
			return len(playlist)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("track")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(fmt.Sprintf("%d. %s - %s", i+1, playlist[i].Artist, playlist[i].Title))
		})

	// initialize communication channels
	dldChannel = make(chan bool, 1)
	playChannel = make(chan int, 1)

	// start play and download threads
	go downloadThread()
	go playThread()

	// func to run when list object is selected
	playlistList.OnSelected = playlistSelect

	previousTrackBtn = widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), previousTrack)
	playPauseBtn = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), togglePlay)
	nextTrackBtn = widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), nextTrack)

	buttonPnl := container.NewGridWithColumns(3,
		previousTrackBtn,
		playPauseBtn,
		nextTrackBtn,
	)

	mainPnl := container.NewBorder(nil, buttonPnl, nil, nil, playlistList)

	mainWindow.SetContent(mainPnl)

	mainWindow.Show()
	app.Run()
}
