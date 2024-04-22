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

var volumeStep = 0.2
var baseVolume = 100

const MIN_VOLUME, MAX_VOLUME = 0, 150

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
	fmt.Println("INFO[ctrlVolume]: volume is", cv.Volume)
	if cv.Volume+volumeStep <= 2.0 {
		cv.Volume += volumeStep
	}
}

func (cv *CtrlVolume) lowerVolume() {
	fmt.Println("INFO[ctrlVolume]: volume is", cv.Volume)
	if cv.Volume-volumeStep >= -2.9 {
		cv.Volume -= volumeStep
	}
}
