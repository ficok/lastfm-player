package main

import (

	// "errors"
	"fmt"
	"log"

	// "os/exec"
	"time"

	"fyne.io/fyne/v2/app"
	"github.com/gopxl/beep"

	// "github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	// gomp3 "github.com/hajimehoshi/go-mp3"
)

const APP_WIDTH = 800
const APP_HEIGHT = 500

// download queue - holds tracks that need to be downloaded
var downloadQueue *DoubleList

var sampleRate = beep.SampleRate(48000)

var tracksDir string = "tracks/"
var coversDir string = "covers/"
var outputFilename string = fmt.Sprintf("%s%%(id)s", tracksDir)
var playlistFile string = "mix.json"
var configFile string = "config.toml"

// play and download thread sync channels
var playChannel chan int
var dldChannel chan bool
var timeChannelGo chan bool
var timeChannelStop chan bool

func init() {
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

	// initialize communication channels
	dldChannel = make(chan bool, 1)
	playChannel = make(chan int, 1)
	timeChannelGo = make(chan bool, 1)
	timeChannelStop = make(chan bool, 1)

	// start play and download threads
	go downloadThread()
	go playThread()
	go trackTime()

	playlist = []Track{}
}

func main() {
	mainApp = app.New()

	initGUI()

	if !validateConfig() {
		loginWindow.Show()
	} else {
		playlist = readPlaylist()
		mainWindow.Show()
	}

	mainApp.Run()
}
