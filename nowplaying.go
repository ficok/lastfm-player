package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/itchyny/gojq"
)

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

func getCoverPath(trackID string) string {
	return fmt.Sprint(coversDir, trackID, ".jpg")
}
