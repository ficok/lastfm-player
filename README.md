# last.fm Player
[README na srpskom](README-sr.md). <br>
A university project for the Programming paradigms course at the Faculty of Mathematics, University of Belgrade.<br>
A GUI music player written in Go that uses the [last.fm](https://www.last.fm/) API to download a user's `mix.json` file, parses it and downloads and plays the tracks. <br> **A work in progress**.

## To-do
In relative order of implementation:
- [x] download thread
- [x] player thread 
- [x] polish dld and player threads 
- [x] polish double list 
- [x] download in advance
- [ ] download and parse the `mix.json` for a specific user
- [ ] play next track automatically after the current one has ended
- [ ] fix streamer struct to support both Len/Position and Resampling
- [x] currently playing song info: track title, artist name, elapsed/duration, album art
- [ ] move list to left and have song info on the right, buttons on the bottom
- [ ] add volume control/display
- [ ] show download progress/downloaded indicator
- [ ] change playlist panel size/just track info screen
- [ ] refresh playlist
- [ ] download new mix in advance, append to the current
- [ ] keyboard media control
- [ ] visualizer
- [ ] delete tracks automatically
- [ ] play opus/webm file directly (if possible)
- [ ] scrobble to LastFM
