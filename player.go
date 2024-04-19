package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"fyne.io/fyne/v2/theme"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/itchyny/gojq"
)

const LASTFM_API_KEY = "069b66ee4d6a7f5e860db3af52c19ab0"

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

func playThread() {
	for {
		fmt.Println("INFO[playThread]: waiting for play request")
		// 1. wait for a new play request
		id := <-playChannel
		fmt.Println("INFO[playThread]: started working")
		// 2. stat the file
		trackLocation := getTrackLocation(playlist[id].ID)
		// 3. if available, play
		if _, statErr := os.Stat(trackLocation); statErr == nil {
			setPlaylistIndex(id)
			playTrack(playlist[playlistIndex])

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
		}
	}
}

func trackTime() {
	for {
		if playerCtrl.Streamer == nil {
			continue
		}

		currentTimeInt := playerCtrl.Streamer.Position() / sampleRate.N(time.Second)
		totalTimeInt := playerCtrl.Streamer.Len() / sampleRate.N(time.Second)

		currentTime := getTimeString(currentTimeInt)
		totalTime := getTimeString(totalTimeInt)

		trackTimeText.Set(fmt.Sprintf("%s/%s", currentTime, totalTime))
	}
}

func setCoverArt() {
	if playerCtrl == nil {
		return
	}

	var coverArtBytes []byte

	coverArtPath := getCoverPath(playerCtrl.ID)
	if fileExists(coverArtPath) {
		coverArtBytes, err := os.ReadFile(coverArtPath)
		if err != nil {
			log.Fatal(err)
		}

		coverArtImage.Image, _, err = image.Decode(bytes.NewReader(coverArtBytes))
		if err != nil {
			log.Fatal(err)
		}
		coverArtImage.Refresh()

		return
	}

	coverArtBytes = downloadCoverArt(playerCtrl.Artist, playerCtrl.Title)
	if coverArtBytes == nil {
		coverArtImage.Image = blankImage
		coverArtImage.Refresh()
		return
	}

	// decode bytes into image and set it
	var format string
	var err error
	coverArtImage.Image, format, err = image.Decode(bytes.NewReader(coverArtBytes))
	if err != nil {
		log.Fatal(err)
	}
	coverArtImage.Refresh()

	// save the cover
	coverFile, err := os.Create(coverArtPath)
	if err != nil {
		log.Fatal(err)
	}

	if format == "jpeg" {
		err = jpeg.Encode(coverFile, coverArtImage.Image, nil)
	} else if format == "png" {
		err = png.Encode(coverFile, coverArtImage.Image)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func downloadCoverArt(artist, track string) []byte {
	// get track info, along with cover, from lastfm
	req, err := http.NewRequest("GET", "https://ws.audioscrobbler.com/2.0/?method=track.getinfo", nil)
	if err != nil {
		log.Fatal(err)
	}

	params := req.URL.Query()
	params.Add("api_key", LASTFM_API_KEY)
	params.Add("artist", artist)
	params.Add("track", track)
	params.Add("format", "json")
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	trackInfoJSON, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var trackInfoMap map[string]any
	json.Unmarshal(trackInfoJSON, &trackInfoMap)

	coverArtQuery, err := gojq.Parse(".track.album.image[-1].\"#text\"")
	if err != nil {
		log.Fatalln(err)
	}

	coverArtLink, ok := coverArtQuery.Run(trackInfoMap).Next()

	if coverArtLink == nil || !ok {
		return nil
	}

	req, err = http.NewRequest("GET", coverArtLink.(string), nil)
	if err != nil {
		log.Fatal(err)
	}

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	reader := bufio.NewReader(res.Body)
	coverArtBytes, _ := io.ReadAll(reader)

	return coverArtBytes

}

func getTimeString(time int) string {
	minutes := fmt.Sprint(time / 60)
	seconds := fmt.Sprint(time % 60)

	if len(seconds) == 1 {
		seconds = "0" + seconds
	}

	return fmt.Sprint(minutes, ":", seconds)
}

func getCoverPath(trackID string) string {
	return fmt.Sprint(coversDir, trackID, ".jpg")
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}
