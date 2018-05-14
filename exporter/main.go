package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tag "github.com/dhowden/tag"
	color "github.com/fatih/color"
	cli "github.com/urfave/cli"
)

//Artist struct
type Artist struct {
	Name   string  `json:"name"`
	Albums []Album `json:"-"`
}

//Album struct
type Album struct {
	Tracks []Track `json:"tracks"`
	Name   string  `json:"name"`
	Year   int     `json:"year"`
	Artist Artist  `json:"artist"`
}

//Track struct
type Track struct {
	Track  int    `json:"track"`
	Disc   int    `json:"disc"`
	Title  string `json:"title"`
	Album  Album  `json:"-"`
	Artist Artist `json:"artist"`
	Path   string `json:"path"`
}

//TrackFlat struct
type TrackFlat struct {
	Track       int
	Disc        int
	Title       string
	Album       string
	Artist      string
	AlbumArtist string
	Year        int
	Path        string
}

func main() {
	app := cli.NewApp()
	app.Name = "audio-lib-exporter"
	app.Version = "v0.0.2"
	app.Usage = "Check your audio library files tags errors"

	app.Action = func(c *cli.Context) error {

		root := c.Args().Get(0)
		if root == "" {
			color.Red("A root must be specified.")
			return nil
		}

		exportFile := c.Args().Get(1)
		if exportFile == "" {
			color.Red("Export file must be specified.")
			return nil
		}

		tf := extract(root)
		albums := transform(tf)

		json, _ := json.Marshal(albums)

		file, err := os.Create(exportFile)
		if err != nil {
			log.Fatal("Cannot create file", err)
		}

		file.WriteString(string(json))

		defer file.Close()

		color.Green("Success: %d albums successfully exported", len(albums))

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func extract(root string) []TrackFlat {
	var tracks []string

	t := []TrackFlat{}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() == true {
			return nil
		}

		if isAudioFile(filepath.Ext(path)) == false {
			return nil
		}

		tracks = append(tracks, path)

		return nil
	})

	for _, trackPath := range tracks {
		file, err := os.Open(trackPath)
		defer file.Close()

		m, err := tag.ReadFrom(file)
		if err != nil {
			fmt.Printf("error reading file: %v\n", err)
		}

		track, _ := m.Track()
		disc, _ := m.Disc()
		trackAbsPath, _ := filepath.Abs(trackPath)
		oTrack := TrackFlat{
			Track:       track,
			Disc:        disc,
			Title:       m.Title(),
			Album:       m.Album(),
			Artist:      m.Artist(),
			AlbumArtist: m.AlbumArtist(),
			Year:        m.Year(),
			Path:        trackAbsPath,
		}

		t = append(t, oTrack)
	}

	return t
}

func transform(tf []TrackFlat) []Album {
	mAlbums := map[string]*Album{}
	mArtists := map[string]*Artist{}
	for _, trackFlat := range tf {

		var exists bool
		var track *Track
		var album *Album
		var artistAlbum *Artist
		var artist *Artist

		slugArtist := trackFlat.Artist
		_, exists = mArtists[slugArtist]
		if exists == true {
			artist, _ = mArtists[slugArtist]
		} else {
			artist = &Artist{
				Name: trackFlat.Artist,
			}

			mArtists[slugArtist] = artist
		}

		slugAlbumgArtist := trackFlat.AlbumArtist
		_, exists = mArtists[slugAlbumgArtist]
		if exists == true {
			artistAlbum, _ = mArtists[slugAlbumgArtist]
		} else {
			artistAlbum = &Artist{
				Name: trackFlat.AlbumArtist,
			}

			mArtists[slugAlbumgArtist] = artistAlbum
		}

		slugAlbum := trackFlat.Album
		_, exists = mAlbums[slugAlbum]
		if exists == true {
			album, _ = mAlbums[slugAlbum]
		} else {
			album = &Album{
				Name:   trackFlat.Album,
				Year:   trackFlat.Year,
				Artist: *artistAlbum,
			}

			mAlbums[slugAlbum] = album
		}

		track = &Track{
			Track:  trackFlat.Track,
			Disc:   trackFlat.Disc,
			Title:  trackFlat.Title,
			Album:  *album,
			Artist: *artist,
			Path:   trackFlat.Path,
		}

		album.Tracks = append(album.Tracks, *track)
		if albumExists(*album, artistAlbum.Albums) == false {
			artistAlbum.Albums = append(artistAlbum.Albums, *album)
		}
	}

	albums := []Album{}
	for _, a := range mAlbums {
		albums = append(albums, *a)
	}

	return albums
}

func albumExists(album Album, albums []Album) bool {
	for _, a := range albums {
		if a.Name == album.Name {
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
