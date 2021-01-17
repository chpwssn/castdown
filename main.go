package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
	var configFilePath string
	flag.StringVar(&configFilePath, "f", "./config.json", "Specify config file. Default is ./config.json")

	flag.Usage = func() {
		fmt.Printf("Usage of our Program: \n")
		fmt.Printf("./go-project -n username\n")
		// flag.PrintDefaults()  // prints default usage
	}
	flag.Parse()
	configFile, err := os.Open(configFilePath)
	if err != nil {
		fmt.Println("Error: no config.json supplied")
		fmt.Println(err)
		os.Exit(2)
	}
	defer configFile.Close()
	byteValue, _ := ioutil.ReadAll(configFile)
	var config Config
	json.Unmarshal([]byte(byteValue), &config)

	var wg sync.WaitGroup

	for _, podcast := range config.Podcasts {
		wg.Add(1)
		go processFeed(podcast.Name, podcast.URL, &config, &wg)
	}

	wg.Wait()
}

func processFeed(podcastName string, feedURL string, config *Config, wg *sync.WaitGroup) {
	defer wg.Done()

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return
	}
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
			var descriptionPath string
			var metaDataPath string
			switch enclosure.Type {
			case "audio/mpeg":
				path = filepath.Join(podcastDir, fmt.Sprintf("%s.mp3", safeTitle))
				descriptionPath = filepath.Join(podcastDir, fmt.Sprintf("%s_description.html", safeTitle))
				metaDataPath = filepath.Join(podcastDir, fmt.Sprintf("%s.json", safeTitle))
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
			WriteFile(descriptionPath, item.Description)
			_, err = os.Stat(metaDataPath)
			if err != nil {
				metaData, err := json.Marshal(item)
				if err == nil {
					WriteFile(metaDataPath, string(metaData))
					fmt.Println("Writing metadata")
					SetDates(metaDataPath, *item.PublishedParsed)
				} else {
					fmt.Println("Could not marshal meta data", err)
				}
			}
			if item.PublishedParsed != nil {
				fmt.Println("Setting time to", item.PublishedParsed)
				SetDates(path, *item.PublishedParsed)
				SetDates(descriptionPath, *item.PublishedParsed)
			} else {
				fmt.Println("Unable to parse publish date", item.Published)
			}
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

func WriteFile(filepath string, body string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.WriteString(out, body)
	if err != nil {
		return err
	}

	return nil
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
