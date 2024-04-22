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
var artistNameTextBox, trackTitleTextBox, trackTimeTextBox *widget.Label
var coverArtImage *canvas.Image
var blankImage image.Image
var playlistList *widget.List
var previousTrackBtn, playPauseBtn, nextTrackBtn *widget.Button
var seekFwdBtn, seekBwdBtn, lowerVolBtn, raiseVolBtn *widget.Button

func initGUI() {
	mainWindow = mainApp.NewWindow("LastFM Player")
	mainWindow.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))

	loginWindow = mainApp.NewWindow("LastFM Player - Login")
	loginWindow.Resize(fyne.NewSize(APP_WIDTH*0.5, APP_HEIGHT*0.2))

	usernameEntry := widget.NewEntry()

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
			loginWindow.Close()
			mainWindow.Show()
		},
	}

	loginWindow.SetContent(loginForm)

	playlistList = widget.NewList(playlistLen, playlistCreateItem, playlistUpdateItem)

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
	coverArtImage.Image = blankImage

	playlistList.OnSelected = playlistSelect

	// setting shortcuts
	initializeAndSetShortcuts()

	// media control buttons
	previousTrackBtn = widget.NewButtonWithIcon("", theme.MediaSkipPreviousIcon(), previousTrack)
	playPauseBtn = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), togglePlay)
	nextTrackBtn = widget.NewButtonWithIcon("", theme.MediaSkipNextIcon(), nextTrack)
	seekFwdBtn = widget.NewButtonWithIcon("", theme.MediaFastForwardIcon(), seekFwd)
	seekBwdBtn = widget.NewButtonWithIcon("", theme.MediaFastRewindIcon(), seekBwd)

	mediaBtnPnl := container.NewGridWithColumns(5,
		seekBwdBtn,
		previousTrackBtn,
		playPauseBtn,
		nextTrackBtn,
		seekFwdBtn,
	)

	// settings panel buttons
	quitBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), blank)
	logoutBtn := widget.NewButtonWithIcon("", theme.LogoutIcon(), blank)
	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), blank)
	raiseVolBtn = widget.NewButtonWithIcon("", theme.VolumeUpIcon(), playerCtrl.raiseVolume)
	lowerVolBtn = widget.NewButtonWithIcon("", theme.VolumeDownIcon(), playerCtrl.lowerVolume)

	// top panel, with settings buttons
	settingsPanel := container.NewGridWithColumns(5, quitBtn, logoutBtn, refreshBtn, lowerVolBtn, raiseVolBtn)

	// volume buttons

	// media panel
	nowPlayingWindow := container.NewCenter(
		container.NewVBox(coverArtImage, artistNameTextBox, trackTitleTextBox, trackTimeTextBox, mediaBtnPnl),
	)
	// panel with main objects: playlist, playing window
	mainPanel := container.NewGridWithColumns(2, playlistList, nowPlayingWindow)

	// the general panel that will hold all the content
	panelContents := container.NewBorder(settingsPanel, nil, nil, nil, mainPanel)

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
