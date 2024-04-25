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

/*
ISSUE:
seeking backward repeatedly when the new position would be less than 0
for some reasong sets the progressbar value to max briefly (streaming not
affected, just looks weird)
TODO:
remove old seeking functions, unify the interface

TRY: not using a current time binding, but refreshing the time progress bar position
with timeProgressBar.Value and timeProgressBar.Refresh() (this way, the progress bar
is unclickable)
*/
func seek(change int) {
	if playerCtrl.Streamer == nil {
		return
	}

	speaker.Lock()

	currentSeconds, _ := playerCtrl.currentTime.Get()
	totalSeconds := playerCtrl.totalTime

	newSeconds := int(currentSeconds) + change

	if newSeconds <= 0 || newSeconds >= totalSeconds {
		speaker.Unlock()
		return
	}

	playerCtrl.Streamer.Seek(newSeconds * sampleRate.N(time.Second))

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
		select {
		case <-timeChannelStop:
			<-timeChannelGo
		default:
		}

		// printing currently playing time info
		if playerCtrl.Streamer == nil {
			continue
		}

		// currentTimeInt := playerCtrl.Streamer.Position() / sampleRate.N(time.Second)
		// totalTimeInt := playerCtrl.Streamer.Len() / sampleRate.N(time.Second)

		// currentTime := getTimeString(currentTimeInt)
		// totalTime := getTimeString(totalTimeInt)

		currentTime := playerCtrl.Streamer.Position() / sampleRate.N(time.Second)
		playerCtrl.currentTime.Set(float64(currentTime))
		totalTime := playerCtrl.totalTime

		trackTimeText.Set(fmt.Sprintf("%s/%s", getTimeString(currentTime), getTimeString(totalTime)))

		// ensure we never skip a track
		select {
		case <-playingNextTrackChannel:
		// if something is received, that means that in the previous iteration, a new
		// track request was sent, so we should skip the check altogether
		default:
			// playing the next song
			if playerCtrl.Streamer.Position() == playerCtrl.Streamer.Len() {
				// inform the channel that we're already playing the next song
				playingNextTrackChannel <- true
				// change this to nextTrack? shouldn't break anything
				playlistList.Select(playlistIndex + 1)
			}
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
