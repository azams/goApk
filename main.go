package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"os"
	"time"
	"log"

	"github.com/cavaliercoder/grab"
	tm "github.com/buger/goterm"
)

func main() {
	var packageName string
	if len(os.Args) == 2 {
		packageName = os.Args[1]
	} else {
		help()
	}
	urlchrome, err := requestHTML("https://chrome.androidcontents.com/", "POST", packageName)
	if err != nil {
		fmt.Println(err.Error())
	}
	regex, err := regexp.Compile(`<a href="(.*?)">(\r\n.+)?<img src="(.*?)" height="20" width="20" style="width: 20px;">(.*?)<span class="dersize">(.*?)<\/span><\/a><\/li>`)
	if err != nil {
		fmt.Println(err.Error())
	}
	results := regex.FindAllStringSubmatch(urlchrome, -1)
	for _, download := range results {
		filename := strings.TrimSpace(download[4])
		linkdownload := strings.TrimSpace(download[1])
		fileSize := strings.TrimSpace(download[5])
		startDownload(linkdownload, filename, fileSize, packageName)
	}
}

func help() {
	fmt.Println("Usage: "+os.Args[0]+" [packagename]")
	os.Exit(1)
}

func requestHTML(baseUrl string, method string, packageName string) (result string, err error) {
	var client = &http.Client{}

	var param = url.Values{}
	param.Set("google_id", packageName)
	param.Set("x", "downloader")
	param.Set("tbi", "0")
	param.Set("av_u", "0")
	param.Set("device_id", "")
	param.Set("model", "")
	param.Set("hl", "en")
	param.Set("de_av", "")
	param.Set("android_ver", "0")
	var payload = bytes.NewBufferString(param.Encode())

	request, err := http.NewRequest(method, baseUrl, payload)
	if err != nil {
		return "", err
	}
	request.Header.Add("authority", "chrome.androidcontents.com")
	request.Header.Add("method", "POST")
	request.Header.Add("path", "/")
	request.Header.Add("scheme", "https")
	request.Header.Add("User-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.106 Safari/537.36")
	request.Header.Add("content-type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	resp, err := ioutil.ReadAll(response.Body)
	return string(resp), nil
}

func startDownload(link string, renameFile string, fileSize string, packageName string) {
	tm.Clear()
	client := grab.NewClient()
	req, _ := grab.NewRequest(".", link)

	fmt.Printf("Downloading %v...\n", renameFile)
	resp := client.Do(req)

	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			tm.MoveCursor(1,1)
			tm.Println("Package name: "+packageName)
			tm.Println("Name        : " + renameFile)
			tm.Println("File Size   : " + fileSize)
			tm.Println("Downloading "+fmt.Sprintf("%d", resp.BytesComplete())+ "/"+fmt.Sprintf("%d", resp.Size())+" ("+fmt.Sprintf("%f",100*resp.Progress())+"%)")
			tm.Flush()
		case <-resp.Done:
			break Loop
		}
	}

	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}
	err := os.Rename("./"+resp.Filename, "./"+renameFile)
    if err != nil {
        log.Fatal(err)
    }
	fmt.Printf("Download saved to ./%v \n", renameFile)
}
