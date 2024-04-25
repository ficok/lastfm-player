package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

var mod fyne.KeyModifier

// shortcut declarations
var quitSC fyne.Shortcut
var togglePlaySC fyne.Shortcut
var seekFwdSC, seekBwdSC fyne.Shortcut
var nextSC, prevSC fyne.Shortcut
var raiseVolSC, lowerVolSC, muteSC fyne.Shortcut
var togglePlaylistPanelSC fyne.Shortcut

func initializeAndSetShortcuts() {
	mod = fyne.KeyModifierControl
	// initializing
	quitSC = &desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: mod}
	togglePlaySC = &desktop.CustomShortcut{KeyName: fyne.KeySpace, Modifier: mod}
	seekFwdSC = &desktop.CustomShortcut{KeyName: fyne.KeyPeriod, Modifier: mod}
	seekBwdSC = &desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: mod}
	nextSC = &desktop.CustomShortcut{KeyName: fyne.KeySemicolon, Modifier: mod}
	prevSC = &desktop.CustomShortcut{KeyName: fyne.KeyApostrophe, Modifier: mod}
	raiseVolSC = &desktop.CustomShortcut{KeyName: fyne.KeyEqual, Modifier: mod}
	lowerVolSC = &desktop.CustomShortcut{KeyName: fyne.KeyMinus, Modifier: mod}
	muteSC = &desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: mod}
	togglePlaylistPanelSC = &desktop.CustomShortcut{KeyName: fyne.KeyP, Modifier: mod}

	// setting
	mainWindow.Canvas().AddShortcut(quitSC, handleInput)
	mainWindow.Canvas().AddShortcut(togglePlaySC, handleInput)
	mainWindow.Canvas().AddShortcut(seekFwdSC, handleInput)
	mainWindow.Canvas().AddShortcut(seekBwdSC, handleInput)
	mainWindow.Canvas().AddShortcut(nextSC, handleInput)
	mainWindow.Canvas().AddShortcut(prevSC, handleInput)
	mainWindow.Canvas().AddShortcut(raiseVolSC, handleInput)
	mainWindow.Canvas().AddShortcut(lowerVolSC, handleInput)
	mainWindow.Canvas().AddShortcut(muteSC, handleInput)
	mainWindow.Canvas().AddShortcut(togglePlaylistPanelSC, handleInput)
}

// shortcut handler
func handleInput(sc fyne.Shortcut) {
	tokens := strings.Split(string(sc.ShortcutName()), "+")
	key := tokens[len(tokens)-1]
	// fmt.Println("DEBUG[handleInput]: shortcut:", key)

	switch key {
	case "Q":
		mainWindow.Close()
	case "Space":
		togglePlay()
	case ".":
		seek(seekStep)
	case ",":
		seek(-seekStep)
	case ";":
		previousTrack()
	case "'":
		nextTrack()
	case "-":
		skipSliderVolumeUpdate = true
		playerCtrl.setVolume(-volumeStep)
	case "=":
		skipSliderVolumeUpdate = true
		playerCtrl.setVolume(volumeStep)
	case "M":
		skipSliderVolumeUpdate = true
		playerCtrl.mute()
	case "P":
		togglePlaylistPanel()
	default:
		fmt.Println("INFO[handleInput]:", key, "is not a shortcut!")
	}
}
