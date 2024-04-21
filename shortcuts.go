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
var seekFwdSC fyne.Shortcut
var seekBwdSC fyne.Shortcut
var nextSC fyne.Shortcut
var prevSC fyne.Shortcut

func initializeAndSetShortcuts() {
	mod = fyne.KeyModifierControl
	// initializing
	quitSC = &desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: mod}
	togglePlaySC = &desktop.CustomShortcut{KeyName: fyne.KeyP, Modifier: mod}
	seekFwdSC = &desktop.CustomShortcut{KeyName: fyne.KeyPeriod, Modifier: mod}
	seekBwdSC = &desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: mod}
	nextSC = &desktop.CustomShortcut{KeyName: fyne.KeyL, Modifier: mod}
	prevSC = &desktop.CustomShortcut{KeyName: fyne.KeyK, Modifier: mod}

	// setting
	mainWindow.Canvas().AddShortcut(quitSC, handleInput)
	mainWindow.Canvas().AddShortcut(togglePlaySC, handleInput)
	mainWindow.Canvas().AddShortcut(seekFwdSC, handleInput)
	mainWindow.Canvas().AddShortcut(seekBwdSC, handleInput)
	mainWindow.Canvas().AddShortcut(nextSC, handleInput)
	mainWindow.Canvas().AddShortcut(prevSC, handleInput)
}

// shortcut handler
func handleInput(sc fyne.Shortcut) {
	tokens := strings.Split(string(sc.ShortcutName()), "+")
	key := tokens[len(tokens)-1]

	// fmt.Println("DEBUG[handleInput]: shortcut:", key)

	switch key {
	case "Q":
		mainWindow.Close()
	case "P":
		togglePlay()
	case ".":
		seekForward()
	case ",":
		seekBackward()
	case "K":
		previousTrack()
	case "L":
		nextTrack()
	default:
		fmt.Println("INFO[handleInput]: not a shortcut!")
	}
}
