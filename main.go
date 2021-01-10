package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

const errGeneric = "Something went wrong!"
const errSaveFail = "Failed to save file!"

const isDebug = false
const dirVideo = "videos"
const dirImage = "images"

func main() {

	var url string

	log.Println("Instagram Image/Video Downloader")

	flag.StringVar(&url, "url", "", "Specify URL")

	flag.Parse()

	if url == "" {
		log.Println("Invalid URL")
		os.Exit(1)
	}

	requestURL := getRequestURL(url)
	makeRequest(requestURL)
}

func getRequestURL(url string) string {
	regExURL := regexp.MustCompile(`utm_source=ig_web_copy_link`)
	return regExURL.ReplaceAllString(url, "__a=1")
}

func makeRequest(requestURL string) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		if isDebug {
			log.Fatalln(err)
		}
		log.Println(errGeneric)
		os.Exit(1)
	}

	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		if isDebug {
			log.Fatalln(err)
		}
		log.Println(errGeneric)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var igResp Response
	e := json.Unmarshal(body, &igResp)
	if e != nil {
		if isDebug {
			log.Fatalln(e.Error())
		}
		log.Println(errGeneric)
	}

	media := igResp.Graphql.ShortcodeMedia

	if media.IsVideo {
		saveVideo(media.VideoURL, media.ID)
		return
	}

	imageURL := media.DisplayURL
	drLength := len(igResp.Graphql.DisplayResources)
	if drLength > 1 {
		imageURL = igResp.Graphql.DisplayResources[drLength-1].Src //Get HD Resolution
	}
	saveImage(imageURL, media.ID)
}

func saveVideo(videoURL, fileName string) {
	log.Println("Downloading Video...")

	createDirIfNotExist(dirVideo)

	fn := fileName + ".mp4"
	saveFile(videoURL, fn, dirVideo+"/")
}

func saveImage(imageURL, fileName string) {
	log.Println("Downloading Image...")

	createDirIfNotExist(dirImage)

	fn := fileName + ".jpg"
	saveFile(imageURL, fn, dirImage+"/")
}

func saveFile(fileURL string, fileName string, filePath string) {
	fileToDownload, err := http.Get(fileURL)
	if err != nil {
		if isDebug {
			log.Fatalln(err.Error())
		}
		log.Println(err.Error())
	}
	defer fileToDownload.Body.Close()

	file, err := os.Create(filePath + fileName)
	if err != nil {
		if isDebug {
			log.Fatalln(err.Error())
		}
		log.Println(errSaveFail)
	}
	defer file.Close()

	io.Copy(file, fileToDownload.Body)

	log.Printf("File %s is saved in %s", fileName, filePath)
}

func createDirIfNotExist(dirName string) {
	_, err := os.Stat(dirName)

	if os.IsNotExist(err) {
		dir := os.Mkdir(dirName, 0755)
		if dir != nil {
			log.Fatalln(err)
		}
	}
}

type Response struct {
	Graphql struct {
		DisplayResources []struct {
			ConfigHeight int64  `json:"config_height"`
			ConfigWidth  int64  `json:"config_width"`
			Src          string `json:"src"`
		} `json:"display_resources"`
		ShortcodeMedia struct {
			ID         string `json:"id"`
			DisplayURL string `json:"display_url"`
			IsVideo    bool   `json:"is_video"`
			VideoURL   string `json:"video_url"`
		} `json:"shortcode_media"`
	} `json:"graphql"`
}
