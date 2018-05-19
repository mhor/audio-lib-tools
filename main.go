package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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

	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		{
			Name:  "check",
			Usage: "Check your audio library files tags errors",
			Flags: []cli.Flag{
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
			},
			Action: func(c *cli.Context) error {
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
			},
		},
		{
			Name:  "export",
			Usage: "Export tracks to a json file",
			Action: func(c *cli.Context) error {
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
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
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

func getTrackMetaData(path string) (tag.Metadata, error) {
	file, err := os.Open(path)
	defer file.Close()

	m, err := tag.ReadFrom(file)
	if err != nil {

		return nil, err
	}

	return m, nil
}
