package main

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var mainApp fyne.App
var mainWindow, loginWindow fyne.Window
var artistNameText, trackTitleText, trackTimeText binding.String

// trackTimeTextBox depracated in favor of progress bar
var artistNameTextBox, trackTitleTextBox, trackTimeTextBox, blankTextBox *widget.Label
var coverArtImage *canvas.Image
var blankImage image.Image
var playlistList *widget.List
var previousTrackBtn, playPauseBtn, nextTrackBtn *widget.Button
var seekFwdBtn, seekBwdBtn, lowerVolBtn, raiseVolBtn *widget.Button
var quitBtn, refreshBtn, logoutBtn *widget.Button
var volumeSlider *widget.Slider
var timeProgressBar *widget.ProgressBar

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
	// depracated in favor of progress bar
	// trackTimeTextBox = widget.NewLabelWithData(trackTimeText)
	// trackTimeTextBox.Alignment = fyne.TextAlignCenter
	blankTextBox = widget.NewLabel(" ")
	coverArtImage.Image = blankImage
	// ----------

	// SETTINGS PANEL
	// settings panel buttons
	quitBtn = widget.NewButtonWithIcon("", theme.CancelIcon(), quit)
	logoutBtn = widget.NewButtonWithIcon("", theme.LogoutIcon(), blank)
	refreshBtn = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), blank)
	raiseVolBtn = widget.NewButtonWithIcon("", theme.VolumeUpIcon(), playerCtrl.raiseVolume)
	lowerVolBtn = widget.NewButtonWithIcon("", theme.VolumeDownIcon(), playerCtrl.lowerVolume)

	settingsBtns := container.NewGridWithColumns(3, quitBtn, logoutBtn, refreshBtn)

	// volume slider
	volumeSlider = widget.NewSlider(MIN_VOLUME, MAX_VOLUME)
	volumeSlider.Step = volumeStep
	volumeSlider.SetValue(playerCtrl.Volume)
	volumeSlider.OnChanged = func(value float64) {
		playerCtrl.setVolume(value)
	}

	settingsPanel := container.NewGridWithColumns(2, settingsBtns, volumeSlider)

	// MEDIA PANEL
	// media control buttons
	previousTrackBtn = widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), previousTrack)
	playPauseBtn = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), togglePlay)
	nextTrackBtn = widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), nextTrack)
	seekFwdBtn = widget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), seekFwd)
	seekBwdBtn = widget.NewButtonWithIcon("", theme.MediaFastRewindIcon(), seekBwd)

	mediaCtrlPnl := container.NewGridWithColumns(5,
		seekBwdBtn,
		previousTrackBtn,
		playPauseBtn,
		nextTrackBtn,
		seekFwdBtn,
	)

	// progress bar
	timeProgressBar = widget.NewProgressBar()
	// initial value is 0
	timeProgressBar.Value = 0.0
	// initialising the progress bar text to nothing
	// check trackTime for formatting
	timeProgressBar.TextFormatter = func() string {
		return ""
	}

	// setting the media panel content
	// blank text boxes used to narrow space between components; looks nicer
	nowPlayingWindow := container.NewCenter(
		container.NewVBox(blankTextBox, coverArtImage, artistNameTextBox, trackTitleTextBox, timeProgressBar, mediaCtrlPnl, blankTextBox),
	)

	// MAIN WINDOW
	// panel with main objects: playlist, playing window
	mainPanel := container.NewGridWithColumns(2, playlistList, nowPlayingWindow)

	// the general panel that will hold all the content
	panelContents := container.NewBorder(settingsPanel, nil, nil, nil, mainPanel)

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
