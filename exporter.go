package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
)

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

		m, err := getTrackMetaData(trackPath)
		if err != nil {
			fmt.Printf("error reading file: %v\n", err)

			continue
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

func copyAlbumCover(album Album, dir string) string {
	for _, track := range album.Tracks {
		m, err := getTrackMetaData(track.Path)
		if err != nil {
			fmt.Printf("error reading file: %v\n", err)

			continue
		}

		if nil == m.Picture() {
			continue
		}

		fileName := dir + "/" + uuid.Must(uuid.NewV4()).String() + "." + m.Picture().Ext
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal("Cannot create file", err)
		}

		file.Write(m.Picture().Data)

		absFilePath, _ := filepath.Abs(fileName)
		return absFilePath
	}

	return ""
}

func transform(tf []TrackFlat, exportAlbumCover bool, exportAlbumCoverDir string) []Album {
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

	if exportAlbumCover {
		os.MkdirAll(exportAlbumCoverDir, os.ModePerm)
	}

	albums := []Album{}
	for _, a := range mAlbums {

		if exportAlbumCover {
			a.CoverPath = copyAlbumCover(*a, exportAlbumCoverDir)
		}

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
