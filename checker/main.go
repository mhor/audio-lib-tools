package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	tag "github.com/dhowden/tag"
	color "github.com/fatih/color"
	cli "github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "audio-lib-checker"
	app.Version = "v0.0.4"
	app.Usage = "Check your audio library files tags errors"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "albums, a",
			Usage: "Check albums.",
		},
		cli.BoolFlag{
			Name:  "tracks, t",
			Usage: "Check tracks.",
		},
		cli.BoolFlag{
			Name:  "only-errors",
			Usage: "Show only errors.",
		},
		cli.IntFlag{
			Name:  "limit, l",
			Usage: "Limit number of errors.",
			Value: 0,
		},
	}

	app.Action = func(c *cli.Context) error {

		root := c.Args().Get(0)

		var checkAlbums = true
		var checkTracks = true

		if c.Bool("albums") == true || c.Bool("tracks") == true {
			checkAlbums = false
			checkTracks = false
		}

		if c.Bool("tracks") == true {
			checkTracks = true
		}

		if c.Bool("albums") == true {
			checkAlbums = true
		}

		if root == "" {
			color.Red("A root must be specified.")
			return nil
		}

		check(root, checkTracks, checkAlbums, c.Int("limit"), c.Bool("only-errors"))

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func check(root string, checkTracks bool, checkAlbums bool, limit int, onlyErrors bool) {
	var tracks []string
	var albums []string
	var errorCount = 0

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == true {
			if isAlbumDirectory(path) == true {
				albums = append(albums, path)
			}

			return nil
		}

		if isAudioFile(filepath.Ext(path)) == false {
			return nil
		}

		tracks = append(tracks, path)

		return nil
	})

	if checkAlbums == true {
		color.Green("\n// Check Albums //\n")

		var totalCheckedAlbums = 0
		var totalErroredAlbums = 0
		var totalWarningAlbums = 0
		for _, albumPath := range albums {
			errors, warnings, _ := checkAlbumRules(albumPath, onlyErrors)
			if len(errors) > 0 || len(warnings) > 0 {
				color.Cyan("Check directory %s", albumPath)
			}

			for _, reason := range errors {
				color.Red(reason)
			}

			for _, reason := range warnings {
				color.Yellow(reason)
			}

			errorCount += len(errors) + len(warnings)
			totalCheckedAlbums++
			totalErroredAlbums += len(errors)
			totalWarningAlbums += len(warnings)

			if limit > 0 && errorCount >= limit {
				red := color.New(color.FgRed)
				whiteBackground := red.Add(color.BgWhite)
				whiteBackground.Println("Error count limit reached")

				color.Green("\nTotal checked album: %d\n", totalCheckedAlbums)
				color.Red("Total errored album: %d\n", totalErroredAlbums)
				color.Yellow("Total warning album: %d\n", totalWarningAlbums)

				return
			}
		}

		color.Green("\nTotal checked album: %d\n", totalCheckedAlbums)
		color.Red("Total errored album: %d\n", totalErroredAlbums)
		color.Yellow("Total warning album: %d\n", totalWarningAlbums)
	}

	if checkTracks == true {
		color.Green("\n// Check Tracks //\n")

		var totalCheckedTracks = 0
		var totalErroredTracks = 0
		var totalWarningTracks = 0
		for _, trackPath := range tracks {

			errors, warnings, _ := checkTrackRules(trackPath, onlyErrors)
			if len(errors) > 0 || len(warnings) > 0 {
				color.Cyan("Check file %s", trackPath)
			}

			for _, reason := range errors {
				color.Red(reason)
			}

			for _, reason := range warnings {
				color.Yellow(reason)
			}

			totalCheckedTracks++
			errorCount += len(errors) + len(warnings)
			totalWarningTracks += len(warnings)
			totalErroredTracks += len(errors)
			if limit > 0 && errorCount >= limit {
				red := color.New(color.FgRed)
				whiteBackground := red.Add(color.BgWhite)
				whiteBackground.Print("Error count limit reached")

				break
			}
		}

		color.Green("\nTotal checked tracks: %d\n", totalCheckedTracks)
		color.Red("Total errored tracks: %d\n", totalErroredTracks)
		color.Yellow("Total warning tracks: %d\n", totalWarningTracks)
	}
}

func checkTrackRules(path string, onlyErrors bool) ([]string, []string, error) {
	var errors []string
	var warnings []string

	file, err := os.Open(path)
	defer file.Close()

	m, err := tag.ReadFrom(file)
	if err != nil {
		fmt.Printf("error reading file: %v\n", err)
		return nil, nil, nil
	}

	var errored bool
	var reason string

	errored, reason = missingTrackTagRule(path, m)
	if errored == true {
		errors = append(errors, reason)
	}

	errored, reason = missingAlbumTagRule(path, m)
	if errored == true {
		errors = append(errors, reason)
	}

	errored, reason = missingAlbumArtistTagRule(path, m)
	if errored == true && onlyErrors == false {
		warnings = append(errors, reason)
	}

	errored, reason = missingArtistTagRule(path, m)
	if errored == true {
		errors = append(errors, reason)
	}

	errored, reason = unknowTrackTagRule(path, m)
	if errored == true && onlyErrors == false {
		warnings = append(errors, reason)
	}

	errored, reason = unknowAlbumTagRule(path, m)
	if errored == true && onlyErrors == false {
		warnings = append(errors, reason)
	}

	errored, reason = unknowAlbumArtistTagRule(path, m)
	if errored == true && onlyErrors == false {
		warnings = append(errors, reason)
	}

	errored, reason = unknowArtistTagRule(path, m)
	if errored == true && onlyErrors == false {
		warnings = append(errors, reason)
	}

	return errors, warnings, nil
}

func checkAlbumRules(path string, onlyErrors bool) ([]string, []string, error) {
	var errors []string
	var warnings []string
	var dirTracks []tag.Metadata

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() == true {
			return nil
		}

		if isAudioFile(filepath.Ext(path)) == false {
			return nil
		}

		file, err := os.Open(path)
		defer file.Close()

		m, err := tag.ReadFrom(file)
		if err != nil {
			fmt.Printf("error reading file: %v\n", err)
			return nil
		}

		dirTracks = append(dirTracks, m)

		return nil
	})

	if err != nil {
		panic(err)
	}

	var errored bool
	var reason string

	errored, reason = multipleAlbumNameRule(path, dirTracks)
	if errored == true {
		errors = append(errors, reason)
	}

	errored, reason = noSameTrackNumberRule(path, dirTracks)
	if errored == true {
		errors = append(errors, reason)
	}

	return errors, warnings, nil
}

func multipleAlbumNameRule(path string, tracks []tag.Metadata) (bool, string) {
	var isFirst = true
	var firstAlbumName string
	for _, track := range tracks {
		if isFirst == true {
			firstAlbumName = track.Album()
			isFirst = false
		}

		if firstAlbumName != track.Album() {

			return true, fmt.Sprintf("Directory contains multiple album names (%s != %s)", firstAlbumName, track.Album())
		}
	}

	return false, ""
}

func noSameTrackNumberRule(path string, tracks []tag.Metadata) (bool, string) {
	var tracksNumbers []int
	for _, track := range tracks {

		trackNumber, _ := track.Track()
		if containsInt(tracksNumbers, trackNumber) == false {
			tracksNumbers = append(tracksNumbers, trackNumber)
		} else {
			return true, "Directory contains same track number"
		}
	}

	return false, ""
}

func missingArtistTagRule(path string, track tag.Metadata) (bool, string) {
	artistName := sanitizeString(track.Artist())
	if "" == artistName {
		return true, "Artist name is empty."
	}

	return false, ""
}

func unknowArtistTagRule(path string, track tag.Metadata) (bool, string) {
	artistName := sanitizeString(track.Artist())
	if true == isUnknow(artistName) {
		return true, fmt.Sprintf("Artist name should be unknow (%s).", track.Artist())
	}

	return false, ""
}

func missingAlbumArtistTagRule(path string, track tag.Metadata) (bool, string) {
	albumArtistName := sanitizeString(track.AlbumArtist())
	if "" == albumArtistName {
		return true, "Album artist name is empty."
	}
	return false, ""
}

func unknowAlbumArtistTagRule(path string, track tag.Metadata) (bool, string) {
	albumArtistName := sanitizeString(track.AlbumArtist())
	if true == isUnknow(albumArtistName) {
		return true, fmt.Sprintf("Album artist name should be unknow(%s).", track.AlbumArtist())
	}

	return false, ""
}

func missingTrackTagRule(path string, track tag.Metadata) (bool, string) {
	trackName := sanitizeString(track.Title())
	if "" == trackName {
		return true, "Track name is empty."
	}

	return false, ""
}

func unknowTrackTagRule(path string, track tag.Metadata) (bool, string) {
	trackName := sanitizeString(track.Title())
	if true == isUnknow(trackName) {
		return true, fmt.Sprintf("Track name should be unknow (%s).", track.Title())
	}

	return false, ""
}

func missingAlbumTagRule(path string, track tag.Metadata) (bool, string) {
	albumName := sanitizeString(track.Album())
	if "" == albumName {
		return true, "Album name is empty."
	}

	return false, ""
}

func unknowAlbumTagRule(path string, track tag.Metadata) (bool, string) {
	albumName := sanitizeString(track.Album())
	if true == isUnknow(albumName) {
		return true, fmt.Sprintf("Album name should be unknow (%s).", track.Album())
	}

	return false, ""
}

func isUnknow(s string) bool {
	s = sanitizeString(s)
	if true == strings.Contains(s, "unknow") {
		return true

	}

	if true == strings.Contains(s, "untitled") {
		return true
	}

	if true == strings.Contains(s, "track") {
		return true
	}

	return false
}

func isAlbumDirectory(path string) bool {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	}

	for _, file := range files {
		if file.IsDir() == false && isAudioFile(filepath.Ext(file.Name())) == true {
			return true
		}
	}

	return false
}

func isAudioFile(extension string) bool {
	switch extension {
	case
		".aac",
		".mp4",
		".m4a",
		".ogg",
		".oga",
		".wma",
		".wav",
		".mp3",
		".aif",
		".flac":
		return true
	}
	return false
}

func sanitizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func containsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
