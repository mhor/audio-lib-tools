package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tag "github.com/dhowden/tag"
	color "github.com/fatih/color"
)

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

	m, err := getTrackMetaData(path)
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

	errored, reason = suspiciousVariousArtistsAlbumArtistTagRule(path, m)
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

	errored, reason = multipleAlbumArtistsRule(path, dirTracks)
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

func multipleAlbumArtistsRule(path string, tracks []tag.Metadata) (bool, string) {
	var isFirst = true
	var firstAlbumArtist string
	for _, track := range tracks {
		if isFirst == true {
			firstAlbumArtist = track.AlbumArtist()
			isFirst = false
		}

		if firstAlbumArtist != track.AlbumArtist() {

			return true, fmt.Sprintf("Directory contains multiple album artists names (%s != %s)", firstAlbumArtist, track.AlbumArtist())
		}
	}

	return false, ""
}

func noSameTrackNumberRule(path string, tracks []tag.Metadata) (bool, string) {
	var tracksNumbers []string
	for _, track := range tracks {

		trackNumber, _ := track.Track()
		trackDisc, _ := track.Disc()
		trackDiscAndNumber := fmt.Sprintf("%d-%d", trackDisc, trackNumber)

		if containsString(tracksNumbers, trackDiscAndNumber) == false {
			tracksNumbers = append(tracksNumbers, trackDiscAndNumber)
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
		return true, fmt.Sprintf("Album artist name should be unknow (%s).", track.AlbumArtist())
	}

	return false, ""
}

func suspiciousVariousArtistsAlbumArtistTagRule(path string, track tag.Metadata) (bool, string) {
	albumArtistName := sanitizeString(track.AlbumArtist())
	if true == isVariousArtists(albumArtistName) && track.AlbumArtist() != "Various Artists" {
		return true, fmt.Sprintf("Album artist name should be Various Artists (%s).", track.AlbumArtist())
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
		return true, fmt.Sprintf("Track name should be untitled (%s).", track.Title())
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
		return true, fmt.Sprintf("Album name should be untitled (%s).", track.Album())
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

func isVariousArtists(s string) bool {
	variousArtistsNames := []string{
		"Various Artists", "Artistiaid Amrywiol", "Amrywiol", "Danske Kunstnere",
		"Div. kunstnere", "Diverse danske artister", "Diverse kunstnere", "Diverse",
		"Diverse Interpreten", "Diversen", "versch. Künstler", "Verschiedene",
		"Verschiedene Interpreten", "Διάφοροι Καλλιτέχνες", "(various)",
		"[Various Artists]", "Assorted Artists", "Assorted Christian Artists",
		"Hairspray (Karaoke) Various Artists", "Miscellaneous", "Mixed Artists",
		"More artists", "Multiple Artists", "Sampler", "Time Life Music: Various Artists",
		"V. A.", "V. Artist", "V.A.", "V/a", "VA", "Varied", "Various",
		"Various (not original artists)", "Various (original artists)",
		"Various Artist - GoGo Wonderful", "Various Big Bands",
		"Various Celtic Artists", "Various Composers", "Various DJ's", "Various Items",
		"Various Military Bands", "Various Speakers", "Various Talented Artists", "Compilaciones",
		"Varios", "Varios artistas", "Erinevad", "Erinevad esitajad", "Eri esittäjiä",
		"Artistes divers", "Artistes variés", "Bande originale", "Collectif", "Collégiale",
		"Comp.", "Compilation", "div.", "Divers", "Multi-artistes",
		"Multi-interprètes", "Variées", "Variés", "A.A.V.V.", "AA.VV.", "AAVV", "Artisti Vari",
		"ヴァリアス", "ヴァリアス・アーティスト", "オムニバス", "さまざまなアーティスト", "여러 아티스트",
		"Diverse Artister", "Diverse Artiesten", "Diverse componisten", "Iedereen",
		"Diverse Artistar", "Różni", "Różni artyści", "Różni wykonawcy", "Wszyscy artysci",
		"Coletânea", "Vários", "Vários artistas", "Vários intérpretes", "razlichnye ispolniteli",
		"[различные исполнители]", "Различные исполнители", "различных исполнителей",
		"Разные артисты", "Razlièni izvajalci", "Blandade artister", "Blandat", "รวมศิลปิน",
		"หลากหลายศิลปิน", "Çeşitli sanatçılar", "Rizni vykonavci", "Rizni vykonavtsi",
		"різних виконавців", "Різні виконавці", "Hợp ca", "Nhiều ca sĩ", "Nhiều nghệ sĩ",
		"Tốp ca", "Tốp ca nam", "群星", "合輯", "Various 80's", "Artis JK", "Artistas Varios",
		"Artisti uniti per l'Emilia", "Assorted Artisits", "Concatenation Records", "Die Brandstifter",
		"Div. artister", "DMT[REC]", "Dominicanos Varios", "La Historia de la Fania",
		"Intérpretes Diversos", "MDB", "Multi Interprètes", "Multi‐interprètes", "Noevir", "OST",
		"Stockfisch", "Tabu Recz", "Ｖ．Ａ.", "VA", "val", "VAR", "Vari", "Vari Artisti", "varias",
		"Varias Artistas", "Varies", "Variois", "Varios artists", "Varios Intérpretes",
		"Various Arists", "various aritsts", "Various Artist", "various artiste",
		"Various Artistes", "Various Artists [DIY]", "華納群星", "Various Artitsts",
		"Various Artsits", "Various Compiled By", "Various DHM Artists", "VariousArtist",
		"Variuos", "Variuos Artists", "Varius", "Varius Artist", "Varius Artists", "Varoius",
		"Varoius Artists", "Verious", "Vrious Artists", "VV AA", "ぱにぽにだっしゅ！", "世界小姐",
		"原声带", "多位艺术家", "多位艺术家", "多位藝術家", "影视原声", "歌手",
	}

	for _, n := range variousArtistsNames {
		if s == sanitizeString(n) {
			return true
		}
	}

	return false
}
