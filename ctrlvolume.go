package main

import (
	"fmt"
	"math"

	"github.com/gopxl/beep"
)

type CtrlVolume struct {
	Title    string
	Artist   string
	ID       string
	Streamer beep.StreamSeeker
	Paused   bool
	Base     float64
	Volume   float64
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
		gain = math.Pow(cv.Base, cv.Volume)
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
	if cv.Volume+volumeStep <= MAX_VOLUME {
		cv.Volume += volumeStep
		volumeSlider.Value = cv.Volume
		fmt.Println("INFO[ctrlVolume]:\n- cv.Volume is", cv.Volume, "\n- volumeSlider.Value is:", volumeSlider.Value)
		volumeSlider.Refresh()
	} else {
		cv.Volume = MAX_VOLUME
		volumeSlider.Value = cv.Volume
		fmt.Println("INFO[ctrlVolume]:\n- cv.Volume is", cv.Volume, "\n- volumeSlider.Value is:", volumeSlider.Value)
		volumeSlider.Refresh()
	}
}

func (cv *CtrlVolume) lowerVolume() {
	if cv.Volume-volumeStep >= MIN_VOLUME {
		cv.Volume -= volumeStep
		volumeSlider.Value = cv.Volume
		fmt.Println("INFO[ctrlVolume]:\n- cv.Volume is", cv.Volume, "\n- volumeSlider.Value is:", volumeSlider.Value)
		volumeSlider.Refresh()
	} else {
		cv.Volume = MIN_VOLUME
		volumeSlider.Value = cv.Volume
		fmt.Println("INFO[ctrlVolume]:\n- cv.Volume is", cv.Volume, "\n- volumeSlider.Value is:", volumeSlider.Value)
		volumeSlider.Refresh()
	}
}

func (cv *CtrlVolume) setVolume(value float64) {
	cv.Volume = value
	volumeSlider.SetValue(cv.Volume)
	fmt.Println("INFO[ctrlVolume]:\n- cv.Volume is", cv.Volume, "\n- volumeSlider.Value is:", volumeSlider.Value)
}
