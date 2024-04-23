package main

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2/data/binding"
	"github.com/gopxl/beep"
)

type CtrlVolume struct {
	Title    string
	Artist   string
	ID       string
	Streamer beep.StreamSeeker
	Paused   bool
	Base     float64
	Volume   binding.Float
	Silent   bool
}

var volumeStep float64 = 0.2

// var baseVolume = 100

const MIN_VOLUME, MAX_VOLUME float64 = -2.0, 2.0

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

func (cv *CtrlVolume) raiseVolume() {
	dontFireVolumeChange = true
	fmt.Println("INFO[raiseVolume]")
	if oldVolume, _ := cv.Volume.Get(); oldVolume+volumeStep <= MAX_VOLUME {
		volume := oldVolume + volumeStep
		fmt.Println("- old volume:", oldVolume, "\n- volume step:", volumeStep, "\n- new volume:", volume)
		cv.Volume.Set(volume)
		newVolume, _ := cv.Volume.Get()
		fmt.Println("- cv.Volume is", newVolume)

	} else {
		volume := MAX_VOLUME
		fmt.Println("- old volume:", oldVolume, "\n- volume step:", volumeStep, "\n- new volume:", volume)
		cv.Volume.Set(volume)
		newVolume, _ := cv.Volume.Get()
		fmt.Println("- cv.Volume is", newVolume)
	}
}

func (cv *CtrlVolume) lowerVolume() {
	dontFireVolumeChange = true
	fmt.Println("INFO[lowerVolume]")
	if oldVolume, _ := cv.Volume.Get(); oldVolume-volumeStep >= MIN_VOLUME {
		volume := oldVolume - volumeStep
		fmt.Println("- old volume:", oldVolume, "\n- volume step:", volumeStep, "\n- new volume:", volume)
		cv.Volume.Set(volume)
		newVolume, _ := cv.Volume.Get()
		fmt.Println("- cv.Volume is", newVolume)
	} else {
		volume := MIN_VOLUME
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
