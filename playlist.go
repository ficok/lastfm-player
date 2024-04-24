package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/itchyny/gojq"
)

type Track struct {
	Title  string
	Artist string
	ID     string
}

// main playlist slice
var playlist []Track

// current position in playlist
var playlistIndex int = -1

func downloadPlaylist() {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.last.fm/player/station/user/%s/mix", config.Username), nil)
	if err != nil {
		log.Fatal(err)
	}

	params := req.URL.Query()
	req.URL.RawQuery = params.Encode()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	mixJSON, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(playlistFile, mixJSON, 0744)
	if err != nil {
		log.Fatal(err)
	}

	playlistReady <- true
}

// read playlist JSON and return a slice of tracks
// TODO: replace with call to get new playlist
func readPlaylist() []Track {
	playlistTracksJSON, err := os.ReadFile(playlistFile)
	if err != nil {
		log.Fatal(err)
	}

	// unpack playlist JSON into map
	var playlistTracksMap map[string]any
	json.Unmarshal(playlistTracksJSON, &playlistTracksMap)

	// query to parse JSON into tracks
	playlistQuery, err := gojq.Parse(".playlist[]")
	if err != nil {
		log.Fatalln(err)
	}

	var playlistSlice []Track

	// run the query on JSON, iterate through tracks and add to slice
	iter := playlistQuery.Run(playlistTracksMap)
	for trackMap, ok := iter.Next(); ok; trackMap, ok = iter.Next() {
		trackTitleQuery, _ := gojq.Parse(".name")
		trackTitle, _ := trackTitleQuery.Run(trackMap).Next()

		trackArtistQuery, _ := gojq.Parse(".artists[0].name")
		trackArtist, _ := trackArtistQuery.Run(trackMap).Next()

		trackIDQuery, _ := gojq.Parse(".playlinks[0].id")
		trackID, _ := trackIDQuery.Run(trackMap).Next()

		newTrack := Track{Title: trackTitle.(string), Artist: trackArtist.(string), ID: trackID.(string)}
		playlistSlice = append(playlistSlice, newTrack)
	}

	return playlistSlice
}

func playlistSelect(idx int) {
	playChannel <- idx
}
