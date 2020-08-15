package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

// Config struct which describes the json file format
type Config struct {
	DataDir  string    `json:"dataDir"`
	Podcasts []Podcast `json:"podcasts"`
}

// Podcast struct
type Podcast struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func main() {
	configFile, err := os.Open("./config.json")
	if err != nil {
		fmt.Println("Error: no config.json supplied")
		fmt.Println(err)
		os.Exit(2)
	}
	defer configFile.Close()
	byteValue, _ := ioutil.ReadAll(configFile)
	var config Config
	json.Unmarshal([]byte(byteValue), &config)

	for _, podcast := range config.Podcasts {
		processFeed(podcast.Name, podcast.URL, &config)
	}
}

func processFeed(podcastName string, feedURL string, config *Config) {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(feedURL)
	fmt.Println(feed.Title)
	for _, element := range feed.Items {
		getItem(podcastName, element, config)
	}
}
func getItem(podcast string, item *gofeed.Item, config *Config) bool {
	savable := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	safeCast := savable.ReplaceAllString(podcast, "_")
	safeTitle := savable.ReplaceAllString(item.Title, "_")
	podcastDir := filepath.Join(config.DataDir, safeCast)
	fmt.Printf("Getting %s\n", savable.ReplaceAllString(item.Title, "_"))

	_, err := os.Stat(podcastDir)
	if err != nil {
		os.MkdirAll(podcastDir, 0755)
	}
	for _, enclosure := range item.Enclosures {
		if strings.Contains(enclosure.Type, "audio") {
			var path string
			switch enclosure.Type {
			case "audio/mpeg":
				path = filepath.Join(podcastDir, fmt.Sprintf("%s.mp3", safeTitle))
				break
			default:
				fmt.Println("I don't know what to do with type", enclosure.Type, enclosure.URL)
				continue
			}
			if path == "" {
				return false
			}
			fmt.Printf("Downloading %s ...", path)
			_, err := os.Stat(path)
			if err != nil {
				DownloadFile(path, enclosure.URL)
				fmt.Println("done.")
			} else {
				fmt.Println("exists.")
			}
			fmt.Println("Setting time to", item.PublishedParsed)
			SetDates(path, *item.PublishedParsed)
		}
	}
	return true
}

func md5Str(data string) string {
	md5 := md5.New()
	dataBytes := []byte(data)
	md5.Write(dataBytes)
	return fmt.Sprintf("%x", md5.Sum(nil))
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func SetDates(filepath string, date time.Time) error {
	os.Chtimes(filepath, date, date)
	return nil
}
