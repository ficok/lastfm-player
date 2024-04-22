package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2/theme"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

const LASTFM_API_KEY = "069b66ee4d6a7f5e860db3af52c19ab0"

// player controller - current playing track info
var playerCtrl *CtrlVolume

var seekStep = 5

func playTrack(track Track) {
	trackLocation := getTrackLocation(track.ID)
	f, err := os.Open(trackLocation)

	if err != nil {
		log.Fatal(err)
	}

	// decode MP3 track into a playable stream
	streamer, _, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	// fill player controller with track info
	playerCtrl.Title = track.Title
	playerCtrl.Artist = track.Artist
	playerCtrl.ID = track.ID

	artistNameText.Set(track.Artist)
	trackTitleText.Set(track.Title)
	setCoverArt()

	// clear any playing song from the speaker
	speaker.Clear()

	// get new track sample rate and update the speaker SR if needed
	// decodedStream, err := gomp3.NewDecoder(f)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if decodedStream.SampleRate() != int(sampleRate) {
	// 	playerCtrl.Streamer = beep.Resample(4, beep.SampleRate(decodedStream.SampleRate()), sampleRate, streamer)
	// } else {
	// 	playerCtrl.Streamer = streamer
	// }

	//TODO: cannot use Resample, it can't convert to StreamSeeker that we need for current position in track/total track length
	playerCtrl.Streamer = streamer

	speaker.Play(playerCtrl)
	setPauseStatus(false)
}

// main func to change pause status and update play/pause button
// should only be changed with this
func setPauseStatus(paused bool) {
	if playerCtrl.Streamer == nil {
		return
	}

	speaker.Lock()
	playerCtrl.Paused = paused
	if paused {
		playPauseBtn.SetIcon(theme.MediaPlayIcon())
	} else {
		playPauseBtn.SetIcon(theme.MediaPauseIcon())
	}
	speaker.Unlock()
	fmt.Println("INFO[Player]: paused:", playerCtrl.Paused)
}

// wrapper func for setPauseStatus to toggle status
func togglePlay() {
	if playlistIndex == -1 {
		nextTrack()
		return
	}

	setPauseStatus(!playerCtrl.Paused)
}

// main func to set global playlist index, should only be changed with this
func setPlaylistIndex(idx int) {
	playlistIndex = idx
	fmt.Println("INFO[Player]: playlistIndex:", playlistIndex)
}

func nextTrack() {
	if playlistIndex+1 == len(playlist) {
		return
	}
	playlistList.Select(playlistIndex + 1)
}

func previousTrack() {
	if playlistIndex-1 < 0 {
		return
	}

	playlistList.Select(playlistIndex - 1)
}

func seekFwd() {
	if playerCtrl.Streamer == nil {
		return
	}
	// lock the speaker. important because we're using the current position.
	// if unlocked, it changes
	speaker.Lock()
	// getting the streamer info
	currentPosition := playerCtrl.Streamer.Position()
	totalLen := playerCtrl.Streamer.Len()
	currentTimeInt := currentPosition / sampleRate.N(time.Second)
	totalTimeInt := totalLen / sampleRate.N(time.Second)

	// if seeking get's us past the end of streamer, skip it
	if currentTimeInt+seekStep >= totalTimeInt {
		// don't forget to unlock streamer, otherwise it stops and
		// we can't unlock it!
		speaker.Unlock()
		return
	}

	// set the streamer's new position and unlock it
	playerCtrl.Streamer.Seek(currentPosition + seekStep*sampleRate.N(time.Second))
	speaker.Unlock()
}

func seekBwd() {
	if playerCtrl.Streamer == nil {
		return
	}
	// lock the speaker. important because we're using the current position.
	// if unlocked, it changes
	speaker.Lock()
	// getting the streamer info
	currentPosition := playerCtrl.Streamer.Position()
	totalLen := playerCtrl.Streamer.Len()
	currentTimeInt := currentPosition / sampleRate.N(time.Second)
	totalTimeInt := totalLen / sampleRate.N(time.Second)

	// if seeking get's us before the start of streamer, skip it
	if currentTimeInt-seekStep >= totalTimeInt {
		// don't forget to unlock streamer, otherwise it stops and
		// we can't unlock it!
		speaker.Unlock()
		return
	}

	// set the streamer's new position and unlock it
	playerCtrl.Streamer.Seek(currentPosition - seekStep*sampleRate.N(time.Second))
	speaker.Unlock()
}

func playThread() {
	for {
		fmt.Println("INFO[playThread]: waiting for play request")
		// while waiting, timeThread can keep on working
		// 1. wait for a new play request
		id := <-playChannel
		// when a new song is requested, timeChannel is blocked until the song starts
		timeChannelStop <- true
		fmt.Println("INFO[playThread]: started working")
		// 2. stat the file
		trackLocation := getTrackLocation(playlist[id].ID)
		// 3. if available, play
		if _, statErr := os.Stat(trackLocation); statErr == nil {
			setPlaylistIndex(id)
			playTrack(playlist[playlistIndex])
			// start the time thread
			timeChannelGo <- true

			// 4. if unavailable, send to download queue and wait on semaphore
		} else {
			// send to front of the queue
			request := Pair{idx: id, priority: true}
			pushFront(request)
			// signal the downloadThread to proceed
			fmt.Println("INFO[playThread]: requesting download of track", playlist[id].ID)
			dldChannel <- true
			// wait for the track to become available
			fmt.Println("INFO[playThread]: waiting for downloadThread...")
			status := <-playChannel
			if status == 0 {
				log.Panic(errors.New("ERROR[playThread]: track unplayable"))
			}

			// 4. play track
			fmt.Println("INFO[playThread]: finally playing!")
			setPlaylistIndex(id)
			playTrack(playlist[playlistIndex])
			// start the time thread
			timeChannelGo <- true
		}
	}
}

func trackTime() {
	for {
		select {
		case <-timeChannelStop:
			<-timeChannelGo
		default:
		}

		// printing currently playing time info
		if playerCtrl.Streamer == nil {
			continue
		}

		currentTimeInt := playerCtrl.Streamer.Position() / sampleRate.N(time.Second)
		totalTimeInt := playerCtrl.Streamer.Len() / sampleRate.N(time.Second)

		currentTime := getTimeString(currentTimeInt)
		totalTime := getTimeString(totalTimeInt)

		trackTimeText.Set(fmt.Sprintf("%s/%s", currentTime, totalTime))

		// playing the next song
		if playerCtrl.Streamer.Position() == playerCtrl.Streamer.Len() {
			playlistList.Select(playlistIndex + 1)
		}
	}
}

func getTimeString(time int) string {
	minutes := fmt.Sprint(time / 60)
	seconds := fmt.Sprint(time % 60)

	if len(seconds) == 1 {
		seconds = "0" + seconds
	}

	return fmt.Sprint(minutes, ":", seconds)
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

func getTrackLocation(videoID string) string {
	return fmt.Sprint(tracksDir, videoID, ".mp3")
}
