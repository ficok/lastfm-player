# last.fm Player
[README na srpskom](README-sr.md). <br>
A university project for the Programming paradigms course at the Faculty of Mathematics, University of Belgrade.<br>
A GUI music player written in Go that uses the [last.fm](https://www.last.fm/) API to download a user's `mix.json` file, parses it and downloads and plays the tracks. <br> **A work in progress**.

## Shortcuts
- modifier key: **control**
- **q** quit
- **space** toggle play/pause
- **;** previous song
- **'** next song
- **-** lower volume
- **=** raise volume
- **m** mute
- **,** seek back 5 seconds
- **.** seek forward 5 seconds
- **p** show/hide playlist

## To-do
In relative order of implementation:
- [x] download thread
- [x] player thread 
- [x] polish dld and player threads 
- [x] polish double list 
- [x] download in advance
- [x] download and parse the `mix.json` for a specific user
- [x] play next track automatically after the current one has ended
- [x] seeking
- [ ] fix streamer struct to support both Len/Position and Resampling
- [x] currently playing song info: track title, artist name, elapsed/duration, album art
- [x] keyboard media control
- [x] move list to left and have song info on the right, buttons on the bottom
- [x] add volume control
- [x] add volume display (slider)
- [x] settings panel
- [x] upon login/start, immediately download the first track
- [ ] show download progress/downloaded indicator
- [x] change playlist panel size/just track info screen
- [ ] refresh playlist (and delete old tracks`)
- [ ] download new mix in advance, append to the current
<br>

*Maybe*

- [ ] keyboard control overhaul (use without mod key)
- [ ] scrobble to LastFM
