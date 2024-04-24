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

	timeProgressBar.Min = float64(0.0)
	playerCtrl.totalTime = playerCtrl.Streamer.Len() / sampleRate.N(time.Second)
	timeProgressBar.Max = float64(playerCtrl.totalTime)

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
	timeProgressBar.Value = float64(currentPosition + seekStep)
	timeProgressBar.Refresh()
	speaker.Unlock()
}

func seek(step int, origin uint) {
	if playerCtrl.Streamer == nil {
		return
	}

	skipTimeProgressBarUpdate <- true
	speaker.Lock()
	// currentPosition := playerCtrl.Streamer.Position()
	// currentSeconds := currentPosition / sampleRate.N(time.Second)
	currentSeconds, _ := playerCtrl.currentTime.Get()

	// totalLength := playerCtrl.Streamer.Len()
	// totalSeconds := totalLength / sampleRate.N(time.Second)
	totalSeconds := playerCtrl.totalTime

	newSeconds := int(currentSeconds) + step
	if newSeconds <= 0 || newSeconds >= totalSeconds {
		speaker.Unlock()
		return
	}

	playerCtrl.Streamer.Seek(newSeconds * sampleRate.N(time.Second))
	if origin != SLIDER {
		playerCtrl.currentTime.Set(float64(newSeconds))
	}
	speaker.Unlock()
	continueTrackingTime <- true
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
	// totalLen := playerCtrl.Streamer.Len()
	currentTimeInt := currentPosition / sampleRate.N(time.Second)
	// totalTimeInt := totalLen / sampleRate.N(time.Second)

	// if seeking get's us before the start of streamer, skip it
	if currentTimeInt-seekStep <= 0 {
		// don't forget to unlock streamer, otherwise it stops and
		// we can't unlock it!
		speaker.Unlock()
		return
	}

	// set the streamer's new position and unlock it
	playerCtrl.Streamer.Seek(currentPosition - seekStep*sampleRate.N(time.Second))
	timeProgressBar.Value = float64(currentPosition - seekStep)
	timeProgressBar.Refresh()
	speaker.Unlock()
}

func playThread() {
	// waiting for the playlist to become ready
	fmt.Println("INFO[playThread]: waiting for playlist to become ready...")
	<-playChannel
	fmt.Println("INFO[playThread]: making the first track ready...")
	trackLocation := getTrackLocation(playlist[0].ID)
	if _, statErr := os.Stat(trackLocation); statErr != nil {
		request := Pair{idx: 0, priority: true}
		pushFront(request)
		dldChannel <- true
		status := <-playChannel
		if status == 0 {
			fmt.Println("WARNING[playThread]: couldn't download the first track in advance.")
		}
	}

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
		if playerCtrl.Streamer == nil {
			continue
		}

		select {
		case <-timeChannelStop:
			<-timeChannelGo
		default:
		}

		select {
		case <-skipTimeProgressBarUpdate:
			<-continueTrackingTime
		default:
			/*
				this is still creating a problem.
				it sets the binding value, which triggers the OnChanged or OnChangeEnded
				function. is seeks to it's new position. this creates a short but obvious
				sound of seeking, similar to speeding up the play speed for a big factor.

				if we're using a binding, that means that whenever this function updates the value,
				both OnChanged and OnChangeEnded will be called and we must somehow skip this.

				not using a binding breaks seeking via slider.

				**TEMPORARY FIX**
				this thread is working and setting the current time too fast, either making synchronization
				between this thread and OnChanged function impossible, or making me look stupid.
				time.Sleep(time.Second) resolved the issue. now this thread ticks every 1 second, which is fine,
				because it only updates time, and it makes sense that it does that every 1 second.

				this enables using a simple if condition in the OnChangeEnded function, which skips sending a seek
				request, if the slider change came from this thread.
				seeking by clicking on the slider works fine, seeking by sliding the slider also mostly works fine:
				- if you drag and hold for too long, the position dot will escape the mouse
				- sometimes after using the slider to seek, the progress bar first updates back to where the trackTime
				  sets it, but immediately jumps back to the correct position
				- rarely just doesn't seek
				- seeking via buttons or shortcuts often breaks

				it depends on the timing of the seeking and trackTime; sometimes they clash.

				this is a bad, ugly fix, but it works for now.
			*/

			time.Sleep(time.Second)

			currentTimeInt := playerCtrl.Streamer.Position() / sampleRate.N(time.Second)
			totalTimeInt := playerCtrl.totalTime

			currentTime := getTimeString(currentTimeInt)
			totalTime := getTimeString(totalTimeInt)

			trackTimeText.Set(fmt.Sprintf("%s/%s", currentTime, totalTime))
			dontChange = true
			playerCtrl.currentTime.Set(float64(currentTimeInt))
		}

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
