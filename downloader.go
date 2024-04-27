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
		// 1. wait for the signal from playThread
		<-dldChannel
		fmt.Println("INFO[downloadThread]: beginning work")
		// 2. read the index, pop it, download it
		request, err := readFront()
		if err != nil {
			log.Panic(err)
		}

		fmt.Println("INFO[downloadThread]: the requested pair is:", request)

		if !downloadQueue.empty {
			popFront()
		}
		fmt.Println("INFO[downloadThread]: starting download")
		downloadTrack(playlist[request.idx].ID, &err)
		if err != nil {
			playChannel <- 0
			fmt.Println(err)
		}
		// 3. signal playThread to play
		fmt.Println("INFO[downloadThread]: signaling to playThread to start playing")
		playChannel <- 1
		// 4. send next 3 tracks to the back of the queue for in advance download
		sent := 0
		for id := request.idx + 1; id < request.idx+4 && id < len(playlist); id++ {
			trackLocation := getTrackLocation(playlist[id].ID)
			// skip the already downloaded
			if _, statErr := os.Stat(trackLocation); statErr == nil {
				break
			}

			inAdvance := Pair{idx: id, priority: false}
			pushBack(inAdvance)
			sent++
		}
		fmt.Println("INFO[downloadThread]: sent", sent, "tracks to download queue.")
		// 5. concurrently download in advance
		go downloadTracksAhead(nrOfTracks)
	}
}

func downloadTracksAhead(nr int) {
	for i := 0; i < nr && !isEmpty(); i++ {
		// certainly won't be empty, because if it is empty,
		// the loop condition is unsatisfied and we exit the loop
		pair, _ := readBack()
		// if pair.priority == true, it was explicitly requested by the user. this is
		// handled in downloadQueue, so we exit because it was added to the front,
		// while here we handle tracks added from the back and therefore none such are left.
		if pair.priority {
			break
		}

		popBack()
		fmt.Println("INFO[downloadAhead]: downloading", playlist[pair.idx].ID)
		var err error
		downloadTrack(playlist[pair.idx].ID, &err)
		if err != nil {
			fmt.Println("WARNING[download ahead]: could not download", playlist[pair.idx].ID)
		}
	}
	fmt.Println("INFO[downloadAhead]: done.")
}

func downloadTrack(videoID string, err *error) {
	videoURL := fmt.Sprintf("https://youtube.com/watch?v=%s", videoID)
	fmt.Println("INFO[downloadTrack]: now downloading", videoURL)
	*err = exec.Command("./ytdl", "--extract-audio", "--audio-format", "mp3", "--output", outputFilename, videoURL).Run()

	if *err != nil {
		errorString := fmt.Sprintf("ERROR[downloadTrack]: could not download track %s", videoID)
		*err = errors.New(errorString)
	}
}
