package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var playlistPanelOn = true
var dontFireVolumeChange = false

var mainApp fyne.App
var mainWindow, loginWindow fyne.Window
var artistNameText, trackTitleText, trackTimeText binding.String

var artistNameTextBox, trackTitleTextBox, trackTimeTextBox, blankTextBox *widget.Label
var coverArtImage *canvas.Image
var blankImage image.Image
var playlistList *widget.List
var previousTrackBtn, playPauseBtn, nextTrackBtn *widget.Button
var seekFwdBtn, seekBwdBtn, lowerVolBtn, raiseVolBtn *widget.Button
var quitBtn, refreshBtn, logoutBtn, togglePlaylistPanelBtn *widget.Button
var volumeSlider *widget.Slider
var timeProgressBar *widget.Slider

var mainPanel *fyne.Container
var nowPlayingWindow *fyne.Container
var panelContents *fyne.Container
var settingsBtns *fyne.Container
var settingsPanel *fyne.Container

func initGUI() {
	// WINDOW INIT
	// initializing the main window
	mainWindow = mainApp.NewWindow("LastFM Player")
	mainWindow.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))

	// initializing the login window
	loginWindow = mainApp.NewWindow("LastFM Player - Login")
	loginWindow.Resize(fyne.NewSize(APP_WIDTH*0.5, APP_HEIGHT*0.2))
	// creating a new entry for username input
	usernameEntry := widget.NewEntry()
	// creating the login form
	loginForm := &widget.Form{
		Items: []*widget.FormItem{ // we can specify items in the constructor
			{Text: "Enter LastFM username", Widget: usernameEntry}},
		OnSubmit: func() {
			config.Username = usernameEntry.Text
			writeConfig()
			if _, err := os.Stat(playlistFile); err != nil {
				downloadPlaylist()
			}
			playlist = readPlaylist()
			playChannel <- 1
			loginWindow.Close()
			mainWindow.Show()
		},
	}
	// appending the login to the login window
	loginWindow.SetContent(loginForm)
	// ----------

	// PLAYLIST CREATION
	// creating the playlist widget
	playlistList = widget.NewList(playlistLen, playlistCreateItem, playlistUpdateItem)
	// setting the function that will be called when a playlist item is selected
	playlistList.OnSelected = playlistSelect
	// ----------

	// TRACK INFO AND ALBUM ART
	// creating track information
	artistNameText = binding.NewString()
	trackTitleText = binding.NewString()
	trackTimeText = binding.NewString()

	blankImageBytes, err := os.ReadFile(coversDir + "missing.png")
	if err != nil {
		log.Fatal(err)
	}
	blankImage, _, _ = image.Decode(bytes.NewReader(blankImageBytes))

	coverArtImage = canvas.NewImageFromImage(nil)
	coverArtImage.FillMode = canvas.ImageFillOriginal
	artistNameTextBox = widget.NewLabelWithData(artistNameText)
	artistNameTextBox.Alignment = fyne.TextAlignCenter
	trackTitleTextBox = widget.NewLabelWithData(trackTitleText)
	trackTitleTextBox.Alignment = fyne.TextAlignCenter
	trackTitleTextBox.TextStyle = fyne.TextStyle{Bold: true}
	trackTimeTextBox = widget.NewLabelWithData(trackTimeText)
	trackTimeTextBox.Alignment = fyne.TextAlignCenter
	blankTextBox = widget.NewLabel(" ")
	coverArtImage.Image = blankImage
	// ----------

	// SETTINGS PANEL
	// settings panel buttons
	quitBtn = widget.NewButtonWithIcon("", theme.CancelIcon(), quit)
	logoutBtn = widget.NewButtonWithIcon("", theme.LogoutIcon(), blank)
	refreshBtn = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), blank)
	raiseVolBtn = widget.NewButtonWithIcon("", theme.VolumeUpIcon(), func() { sendRequest(Request{VOL, 0, volumeStep, BTN}) })
	lowerVolBtn = widget.NewButtonWithIcon("", theme.VolumeDownIcon(), func() { sendRequest(Request{VOL, 0, -volumeStep, BTN}) })
	togglePlaylistPanelBtn = widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), togglePlaylistPanel)

	settingsBtns = container.NewGridWithColumns(8, quitBtn, logoutBtn, refreshBtn, togglePlaylistPanelBtn,
		blankTextBox, blankTextBox, blankTextBox, blankTextBox)

	// volume slider
	volumeSlider = widget.NewSliderWithData(MIN_VOLUME, MAX_VOLUME, playerCtrl.Volume)
	volumeSlider.Step = volumeStep
	volumeSlider.OnChangeEnded = func(volume float64) {
		oldVolume, _ := playerCtrl.Volume.Get()
		change := oldVolume - volume
		sendRequest(Request{VOL, 0, change, SLIDER})
	}

	settingsPanel = container.NewGridWithColumns(2, settingsBtns, volumeSlider)

	// MEDIA PANEL
	// media control buttons
	previousTrackBtn = widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), previousTrack)
	playPauseBtn = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), togglePlay)
	nextTrackBtn = widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), nextTrack)
	seekFwdBtn = widget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), func() { sendRequest(Request{SEEK, seekStep, 0, BTN}) })
	seekBwdBtn = widget.NewButtonWithIcon("", theme.MediaFastRewindIcon(), func() { sendRequest(Request{SEEK, -seekStep, 0, BTN}) })

	mediaCtrlPnl := container.NewGridWithColumns(5,
		seekBwdBtn,
		previousTrackBtn,
		playPauseBtn,
		nextTrackBtn,
		seekFwdBtn,
	)

	// progress bar
	timeProgressBar = widget.NewSlider(0, 0)
	timeProgressBar.SetValue(0.0)
	timeProgressBar.Step = 1.0
	timeProgressBar.OnChanged = func(position float64) {
		change := int(position) - (playerCtrl.Streamer.Position() / sampleRate.N(time.Second))
		sendRequest(Request{SEEK, change, 0, SLIDER})
	}

	// setting the media panel content
	// blank text boxes used to narrow space between components; looks nicer
	nowPlayingWindow = container.NewCenter(
		container.NewVBox(blankTextBox, coverArtImage, artistNameTextBox, trackTitleTextBox, trackTimeTextBox, timeProgressBar, mediaCtrlPnl, blankTextBox),
	)

	// MAIN WINDOW
	// panel with main objects: playlist, playing window
	mainPanel = container.NewGridWithColumns(2, playlistList, nowPlayingWindow)

	// the general panel that will hold all the content
	panelContents = container.NewBorder(settingsPanel, nil, nil, nil, mainPanel)

	// appending everything to the mainWindow
	mainWindow.SetContent(panelContents)
}

func playlistLen() int {
	return len(playlist)
}

func playlistCreateItem() fyne.CanvasObject {
	return widget.NewLabel("")
}

func playlistUpdateItem(idx int, item fyne.CanvasObject) {
	item.(*widget.Label).SetText(fmt.Sprintf("%d. %s - %s", idx+1, playlist[idx].Artist, playlist[idx].Title))
}

// placeholder function
func blank() {
}

func quit() {
	mainWindow.Close()
}

func togglePlaylistPanel() {
	if playlistPanelOn {
		// now it's false
		playlistPanelOn = !playlistPanelOn

		// draw gui w/o playlist panel
		mainPanel = container.NewGridWithColumns(1, nowPlayingWindow)
		settingsBtns = container.NewGridWithColumns(6, quitBtn, logoutBtn, refreshBtn, togglePlaylistPanelBtn,
			blankTextBox, blankTextBox)
		settingsPanel = container.NewGridWithColumns(3, settingsBtns, volumeSlider, blankTextBox)

		panelContents = container.NewBorder(settingsPanel, nil, nil, nil, mainPanel)
		mainWindow.SetContent(panelContents)
	} else {
		// now it's true
		playlistPanelOn = !playlistPanelOn
		// draw gui w/ playlist panel
		mainPanel = container.NewGridWithColumns(2, playlistList, nowPlayingWindow)
		settingsBtns = container.NewGridWithColumns(9, quitBtn, logoutBtn, refreshBtn, togglePlaylistPanelBtn,
			blankTextBox, blankTextBox, blankTextBox, blankTextBox, blankTextBox)
		settingsPanel = container.NewGridWithColumns(2, settingsBtns, volumeSlider)

		panelContents = container.NewBorder(settingsPanel, nil, nil, nil, mainPanel)
		mainWindow.SetContent(panelContents)
	}
}
