package main

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2/data/binding"
	"github.com/gopxl/beep"
)

type CtrlVolume struct {
	Title         string
	Artist        string
	ID            string
	Streamer      beep.StreamSeeker
	Paused        bool
	Base          float64
	Volume        float64
	VolumePercent binding.Float
	Silent        bool
}

const volumeStep float64 = 1.0
const baseVolume float64 = 100.0

// var baseVolume = 100

const MIN_VOLUME, MAX_VOLUME float64 = 0.0, 150.0

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
		volume := cv.Volume
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

func (cv *CtrlVolume) setVolume(volumeChange float64) {
	fmt.Println("INFO[setVolume]")
	oldVolume, _ := cv.VolumePercent.Get()

	newVolume := oldVolume + volumeChange
	if newVolume < MIN_VOLUME || newVolume > MAX_VOLUME {
		return
	}

	cv.VolumePercent.Set(newVolume)
	playerCtrl.Volume = (newVolume - baseVolume) / 10
	fmt.Println("- cv.Volume is", newVolume)
}
