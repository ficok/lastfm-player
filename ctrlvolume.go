package main

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2/data/binding"
	"github.com/gopxl/beep"
)

type CtrlVolume struct {
	Title       string
	Artist      string
	ID          string
	Streamer    beep.StreamSeeker
	currentTime binding.Float
	Paused      bool
	Base        float64
	Volume      binding.Float
	Silent      bool
}

const volumeStep float64 = 0.1

// var baseVolume = 100

const MIN_VOLUME, MAX_VOLUME float64 = 0.0, 2.0

func (cv *CtrlVolume) Stream(samples [][2]float64) (n int, ok bool) {
	if cv.Streamer == nil {
		return 0, false
	}

	if cv.Paused {
		for i := range samples {
			samples[i] = [2]float64{}
		}
		return len(samples), true
	}

	n, ok = cv.Streamer.Stream(samples)
	gain := 0.0

	if !cv.Silent {
		volume, _ := cv.Volume.Get()
		gain = math.Pow(cv.Base, volume)
	}

	for i := range samples[:n] {
		samples[i][0] *= gain
		samples[i][1] *= gain
	}
	return n, ok
}

func (cv *CtrlVolume) Err() error {
	if cv.Streamer == nil {
		return nil
	}
	return cv.Streamer.Err()
}

func (cv *CtrlVolume) changeVolume(value float64) {
	fmt.Println("INFO[changeVolume]")
	oldVolume, _ := cv.Volume.Get()
	newVolume := roundFloat(oldVolume+value, 1)
	if newVolume < MIN_VOLUME || newVolume > MAX_VOLUME {
		return
	}

	fmt.Println("- old volume:", oldVolume, "\n- volume step:", value, "\n- new volume:", newVolume)
	cv.Volume.Set(newVolume)
	volume, _ := cv.Volume.Get()
	fmt.Println("- cv.Volume is", volume)
}

func (cv *CtrlVolume) raiseVolume() {
	/*
		because the value in the slider and the volume value of the playerCtrl are
		bound, if the value of volume was changed by the shortcut or the button,
		the widget will register it and will call the OnChangeEnded function.
		in other words, immediately after this function, OnChangeEnded (and
		therefore setVolume) is called, thus setting the volume twice.
		we don't want that, so we set this variable to true. OnChangeEnded is called,
		it sees that this value is true and does nothing.
	*/
	fmt.Println("INFO[raiseVolume]")

	oldVolume, _ := cv.Volume.Get()
	volume := roundFloat(oldVolume+volumeStep, 1)
	if volume > MAX_VOLUME {
		return
	} else {
		fmt.Println("- old volume:", oldVolume, "\n- volume step:", volumeStep, "\n- new volume:", volume)
		cv.Volume.Set(volume)
		newVolume, _ := cv.Volume.Get()
		fmt.Println("- cv.Volume is", newVolume)
	}
}

func (cv *CtrlVolume) lowerVolume() {
	/*
		because the value in the slider and the volume value of the playerCtrl are
		bound, if the value of volume was changed by the shortcut or the button,
		the widget will register it and will call the OnChangeEnded function.
		in other words, immediately after this function, OnChangeEnded (and
		therefore setVolume) is called, thus setting the volume twice.
		we don't want that, so we set this variable to true. OnChangeEnded is called,
		it sees that this value is true and does nothing.
	*/
	fmt.Println("INFO[lowerVolume]")

	oldVolume, _ := cv.Volume.Get()
	volume := roundFloat(oldVolume-volumeStep, 1)
	if volume < MIN_VOLUME {
		return
	} else {
		fmt.Println("- old volume:", oldVolume, "\n- volume step:", volumeStep, "\n- new volume:", volume)
		cv.Volume.Set(volume)
		newVolume, _ := cv.Volume.Get()
		fmt.Println("- cv.Volume is", newVolume)
	}
}

func (cv *CtrlVolume) setVolume(volume float64) {
	cv.Volume.Set(volume)
	newVolume, _ := cv.Volume.Get()
	fmt.Println("- cv.Volume is", newVolume)
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
