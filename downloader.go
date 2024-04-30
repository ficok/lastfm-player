package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
)

// number of tracks to download ahead
var nrOfTracks int = 3

func downloadThread() {
	/*
		NOTE: download is only triggered when the playThread sends something to download.
		explicitly requested tracks have download priority, so we download them first.

		after download attempt, send either 0 as a fail status, or 1 as a success status. the
		play thread handles the rest of the playing.

		next, download in advance.
		this can be done inside this function without creating a new thread. however, if
		a new tracks is requested immediatetely after completing a previous request, we need to
		prioritize it. if we don't want a new thread, we need to check for input from playthread
		and break out of the loop so we can start over and download the new track. that would
		require to check the channel without blocking and if we do catch a new request, we need to resend
		it to the download thread. however, another request can be added and then lost this way.
		i think it's more simple and robust to delegate downloading in advance to a new thread.
	*/
	for {
		fmt.Println("INFO[downloadThread]: waiting for signal from playThread")
		// wait for the signal from playThread
		<-dldChannel
		fmt.Println("INFO[downloadThread]: beginning work")
		// read the request
		request, err := readFront()
		if err != nil {
			log.Panic(err)
		}

		fmt.Println("INFO[downloadThread]: the requested pair is:", request)
		// we will start handling this request, so we pop it
		if !downloadQueue.empty {
			popFront()
		}
		fmt.Println("INFO[downloadThread]: starting download")
		// attempt the download of the track
		downloadTrack(playlist[request.idx].ID, &err)
		if err != nil {
			// if the download failed, inform the playthread
			playChannel <- 0
			fmt.Println(err)
		}
		fmt.Println("INFO[downloadThread]: signaling to playThread to start playing")
		// tell playThread the song is ready to be played
		playChannel <- 1
		// download in advance: send the next 3 tracks to be downloaded
		sent := 0
		for id := request.idx + 1; id < request.idx+4 && id < len(playlist); id++ {
			trackLocation := getTrackLocation(playlist[id].ID)
			// skip the already downloaded
			if _, statErr := os.Stat(trackLocation); statErr == nil {
				break
			}

			// create a request
			inAdvance := Pair{idx: id, priority: false}
			// send it
			pushBack(inAdvance)
			sent++
		}
		fmt.Println("INFO[downloadThread]: sent", sent, "tracks to download queue.")
		// concurrently download in advance
		go downloadTracksAhead(nrOfTracks)
	}
}

func downloadTracksAhead(nr int) {
	/*
		try to download nr tracks from the back of the queue. those tracks were sent
		for an in advance download.
	*/
	for i := 0; i < nr && !isEmpty(); i++ {
		// certainly won't be empty, because if it is empty,
		// the loop condition is unsatisfied and we exit the loop
		pair, _ := readBack()
		/*
			if pair.priority == true, it was explicitly requested by the user and the track
			was added to the front of the queue. this means that there are less than nr tracks
			in the queue that should be downloaded in advance.

			we will skip the priority pair, because by design it will be handled in downloadThread
		*/
		if pair.priority {
			break
		}
		// pop the request
		popBack()
		fmt.Println("INFO[downloadAhead]: downloading", playlist[pair.idx].ID)
		var err error
		// attempt the download
		downloadTrack(playlist[pair.idx].ID, &err)
		if err != nil {
			// print the info that the track couldn't be downloaded
			fmt.Println("WARNING[download ahead]: could not download", playlist[pair.idx].ID)
		}
	}
	fmt.Println("INFO[downloadAhead]: done.")
}

func downloadTrack(videoID string, err *error) {
	// create the URL from the video ID that we have
	videoURL := fmt.Sprintf("https://youtube.com/watch?v=%s", videoID)
	fmt.Println("INFO[downloadTrack]: now downloading", videoURL)
	// execute the ytdl program that will download and convert the video
	*err = exec.Command("./ytdl", "--extract-audio", "--audio-format", "mp3", "--output", outputFilename, videoURL).Run()

	if *err != nil {
		errorString := fmt.Sprintf("ERROR[downloadTrack]: could not download track %s", videoID)
		*err = errors.New(errorString)
	}
}
