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

var playlistPanelOn = true
var skipSliderVolumeUpdate = false

var mainApp fyne.App
var mainWindow, loginWindow fyne.Window
var artistNameText, trackTitleText, trackTimeText binding.String

var artistNameTextBox, trackTitleTextBox, trackTimeTextBox, blankTextBox *widget.Label
var coverArtImage *canvas.Image
var blankImage image.Image
var playlistList *widget.List
var volumeSlider *widget.Slider
var timeProgressBar *widget.Slider

var mainPanel *fyne.Container
var nowPlayingWindow *fyne.Container
var panelContents *fyne.Container
var mainToolbar, normalToolbar, extendedToolbar *widget.Toolbar
var mediaCtrlPnl, mediaCtrlPlayPnl, mediaCtrlPausePnl *widget.Toolbar

var volumeStatusIcon, lowVolumeIcon, medVolumeIcon, highVolumeIcon *widget.Icon
var volumeStatus *fyne.Container

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

	// volume slider
	volumeSlider = widget.NewSliderWithData(MIN_VOLUME, MAX_VOLUME, playerCtrl.VolumePercent)
	volumeSlider.Step = volumeStep
	volumeSlider.OnChanged = func(volume float64) {
		// if the volume was changed via shortcut or button, do not set volume
		// from here
		if !skipSliderVolumeUpdate {
			oldVolume, _ := playerCtrl.VolumePercent.Get()
			volumeChange := volume - oldVolume
			playerCtrl.setVolume(volumeChange)
		}
		// reset the value for future use
		skipSliderVolumeUpdate = false
	}

	// volume control
	lowVolumeIcon = widget.NewIcon(theme.VolumeMuteIcon())
	medVolumeIcon = widget.NewIcon(theme.VolumeDownIcon())
	highVolumeIcon = widget.NewIcon(theme.VolumeUpIcon())
	volumeStatusIcon = highVolumeIcon
	volumeStatus = container.NewHBox(
		container.NewGridWrap(fyne.NewSize(20, 20), volumeStatusIcon),
		container.NewGridWrap(fyne.NewSize(190, 20), volumeSlider),
		container.NewGridWrap(fyne.NewSize(20, 20), blankTextBox),
	)

	// MEDIA PANEL
	// media control buttons
	mediaCtrlPlayPnl = widget.NewToolbar(
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.MediaFastRewindIcon(), func() { seek(-seekStep) }),
		widget.NewToolbarAction(theme.MediaSkipPreviousIcon(), previousTrack),
		widget.NewToolbarAction(theme.MediaPlayIcon(), func() {
			togglePlay()
		}),
		widget.NewToolbarAction(theme.MediaSkipNextIcon(), nextTrack),
		widget.NewToolbarAction(theme.MediaFastForwardIcon(), func() { seek(seekStep) }),
		widget.NewToolbarSpacer(),
	)

	mediaCtrlPausePnl = widget.NewToolbar(
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.MediaFastRewindIcon(), func() { seek(-seekStep) }),
		widget.NewToolbarAction(theme.MediaSkipPreviousIcon(), previousTrack),
		widget.NewToolbarAction(theme.MediaPauseIcon(), func() {
			togglePlay()

		}),
		widget.NewToolbarAction(theme.MediaSkipNextIcon(), nextTrack),
		widget.NewToolbarAction(theme.MediaFastForwardIcon(), func() { seek(seekStep) }),
		widget.NewToolbarSpacer(),
	)

	mediaCtrlPnl = mediaCtrlPlayPnl

	// progress bar
	timeProgressBar = widget.NewSlider(0, 0)
	timeProgressBar.Step = 0.0
	timeProgressBar.Value = 0.0
	timeProgressBar.OnChanged = func(time float64) {
		fmt.Println("INFO[time progress bar]: disabled")
	}
	timeProgressBar.OnChangeEnded = func(time float64) {
		fmt.Println("INFO[time progress bar]: disabled")
	}

	// setting the media panel content
	// blank text boxes used to narrow space between components; looks nicer
	nowPlayingWindow = container.NewCenter(
		container.NewVBox(volumeStatus, blankTextBox, coverArtImage, artistNameTextBox, trackTitleTextBox,
			trackTimeTextBox, timeProgressBar, mediaCtrlPnl, blankTextBox),
	)

	// toolbar
	normalToolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			mainToolbar = extendedToolbar
			panelContents = container.NewBorder(mainToolbar, nil, nil, nil, mainPanel)
			mainWindow.SetContent(panelContents)
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.CancelIcon(), quit),
	)

	extendedToolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			mainToolbar = normalToolbar
			panelContents = container.NewBorder(mainToolbar, nil, nil, nil, mainPanel)
			mainWindow.SetContent(panelContents)
		}),
		widget.NewToolbarAction(theme.AccountIcon(), nil),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), refreshPlaylist),
		widget.NewToolbarAction(theme.ColorPaletteIcon(), togglePlaylistPanel),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.CancelIcon(), quit),
	)

	mainToolbar = normalToolbar

	// MAIN WINDOW
	// panel with main objects: playlist, playing window
	mainPanel = container.NewGridWithColumns(2, playlistList, nowPlayingWindow)

	// the general panel that will hold all the content
	panelContents = container.NewBorder(mainToolbar, nil, nil, nil, mainPanel)

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

func quit() {
	mainWindow.Close()
}

func togglePlaylistPanel() {
	if playlistPanelOn {
		// now it's false
		playlistPanelOn = !playlistPanelOn

		// draw gui w/o playlist panel
		mainPanel = container.NewGridWithColumns(1, nowPlayingWindow)

		panelContents = container.NewBorder(mainToolbar, nil, nil, nil, mainPanel)
		mainWindow.SetContent(panelContents)
	} else {
		// now it's true
		playlistPanelOn = !playlistPanelOn
		// draw gui w/ playlist panel
		mainPanel = container.NewGridWithColumns(2, playlistList, nowPlayingWindow)

		panelContents = container.NewBorder(mainToolbar, nil, nil, nil, mainPanel)
		mainWindow.SetContent(panelContents)
	}
}

func updateMediaCtrl() {
	if !playerCtrl.Paused {
		mediaCtrlPnl = mediaCtrlPausePnl
	} else {
		mediaCtrlPnl = mediaCtrlPlayPnl
	}

	nowPlayingWindow = container.NewCenter(
		container.NewVBox(volumeStatus, blankTextBox, coverArtImage, artistNameTextBox, trackTitleTextBox,
			trackTimeTextBox, timeProgressBar, mediaCtrlPnl, blankTextBox),
	)

	if playlistPanelOn {
		mainPanel = container.NewGridWithColumns(2, playlistList, nowPlayingWindow)
	} else {
		mainPanel = container.NewGridWithColumns(1, nowPlayingWindow)
	}

	panelContents = container.NewBorder(mainToolbar, nil, nil, nil, mainPanel)
	mainWindow.SetContent(panelContents)
}

func updateVolumeIcon() {
	currentVolume, _ := playerCtrl.VolumePercent.Get()
	if currentVolume < 33.0 {
		volumeStatusIcon = lowVolumeIcon
	} else if currentVolume >= 33.0 && currentVolume < 66.0 {
		volumeStatusIcon = medVolumeIcon
	} else {
		volumeStatusIcon = highVolumeIcon
	}

	volumeStatus = container.NewHBox(
		container.NewGridWrap(fyne.NewSize(20, 20), volumeStatusIcon),
		container.NewGridWrap(fyne.NewSize(190, 20), volumeSlider),
		container.NewGridWrap(fyne.NewSize(20, 20), blankTextBox),
	)

	nowPlayingWindow = container.NewCenter(
		container.NewVBox(volumeStatus, blankTextBox, coverArtImage, artistNameTextBox, trackTitleTextBox,
			trackTimeTextBox, timeProgressBar, mediaCtrlPnl, blankTextBox),
	)

	if playlistPanelOn {
		mainPanel = container.NewGridWithColumns(2, playlistList, nowPlayingWindow)
	} else {
		mainPanel = container.NewGridWithColumns(1, nowPlayingWindow)
	}

	panelContents = container.NewBorder(mainToolbar, nil, nil, nil, mainPanel)

	mainWindow.SetContent(panelContents)
}
