package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

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
	updateMediaCtrl()
}

// main func to change pause status and update play/pause button
// should only be changed with this
func setPauseStatus(paused bool) {
	if playerCtrl.Streamer == nil {
		return
	}

	speaker.Lock()
	playerCtrl.Paused = paused
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
	updateMediaCtrl()
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

func seek(change int) {
	if playerCtrl.Streamer == nil {
		return
	}

	speaker.Lock()

	currentSeconds := playerCtrl.currentTime
	totalSeconds := playerCtrl.totalTime

	newSeconds := int(currentSeconds) + change

	if newSeconds <= 0 {
		playerCtrl.Streamer.Seek(0 * sampleRate.N(time.Second))
	} else if newSeconds >= totalSeconds {
		playerCtrl.Streamer.Seek((totalSeconds - 1) * sampleRate.N(time.Second))
	} else {
		playerCtrl.Streamer.Seek(newSeconds * sampleRate.N(time.Second))
	}

	speaker.Unlock()
}

func playThread() {
	/*
		this part of the thread will try to download the first track of the playlist, in case
		it has not been downloaded yet
	*/
	// waiting for the signal that the playlist is ready (initGUI, after readPlaylist)
	fmt.Println("INFO[playThread]: waiting for playlist to become ready...")
	<-playChannel
	fmt.Println("INFO[playThread]: making the first track ready...")
	// trying to stat the song file
	trackLocation := getTrackLocation(playlist[0].ID)
	if _, statErr := os.Stat(trackLocation); statErr != nil {
		// if it doesn't exist, create a priority download request
		request := Pair{idx: 0, priority: true}
		// send the request
		pushFront(request)
		// signal to download thread to start downloading
		dldChannel <- true
		// wait for the download status from download thread
		status := <-playChannel
		if status == 0 {
			// couldn't download the song
			fmt.Println("WARNING[playThread]: couldn't download the first track in advance.")
		}
	}

	/*
		this part of the thread waits for a play request. it tries to play the requested track,
		but if it has not been downloaded, sends a download requst for it.
		then it waits for the request to be completed and plays the requestd track when it
		become available.
	*/
	for {
		fmt.Println("INFO[playThread]: waiting for play request")
		// while waiting, timeThread can keep on working
		// wait for a new play request
		id := <-playChannel
		// when a new song is requested, timeChannel is blocked until the song starts
		timeChannelStop <- true
		fmt.Println("INFO[playThread]: started working")
		// stat the file
		trackLocation := getTrackLocation(playlist[id].ID)
		// if available, play
		if _, statErr := os.Stat(trackLocation); statErr == nil {
			setPlaylistIndex(id)
			playTrack(playlist[playlistIndex])
			// tell the trackTime thread to continue working
			timeChannelGo <- true

			// if unavailable, send to download queue and wait while it downloads
		} else {
			// create a priority download request
			request := Pair{idx: id, priority: true}
			// send the request
			pushFront(request)
			// signal the downloadThread to start downloading
			fmt.Println("INFO[playThread]: requesting download of track", playlist[id].ID)
			dldChannel <- true
			// wait for the track to become available
			fmt.Println("INFO[playThread]: waiting for downloadThread...")
			status := <-playChannel
			if status == 0 {
				// panic if the song couldn't be downloaded
				log.Panic(errors.New("ERROR[playThread]: track unplayable"))
			}

			// play track
			fmt.Println("INFO[playThread]: finally playing!")
			setPlaylistIndex(id)
			playTrack(playlist[playlistIndex])
			// tell the time thread to continue playing
			timeChannelGo <- true
		}
	}
}

func trackTime() {
	/*
		playThread will send a stop signal when it is activated by a play request.
		when that happens, we will wait here for a continue signal.

		if the request has not been caught, that means we can safely continue working.
	*/
	for {
		select {
		case <-timeChannelStop:
			<-timeChannelGo
		default:
		}

		if playerCtrl.Streamer == nil {
			continue
		}

		// get the elapsed time in seconds from streamer and save it in playerCtrl.currentTime
		playerCtrl.currentTime = playerCtrl.Streamer.Position() / sampleRate.N(time.Second)
		/*
			set the progress bar value.
			we are not using a binding because this way, clicking and moving the slider is disabled
		*/
		timeProgressBar.Value = float64(playerCtrl.currentTime)
		// refresh to render the change
		timeProgressBar.Refresh()

		// update the time label
		trackTimeText.Set(fmt.Sprintf("%s/%s", getTimeString(playerCtrl.currentTime), getTimeString(playerCtrl.totalTime)))

		/*
			ensure we never skip a track
			without this, it sometimes happens that when the new track is started
			information about the current streamer position and length hasn't yet been
			updated. this causes the if to pass and plays the next song, thus skipping the
			real next song.

			this way the first time we try to play the next song, we send a skip signal
			to this thread. the next iteration will receive the signal and skip playing the next song.
		*/
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
