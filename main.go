package main

import (
	"log"
	"sync"

	"github.com/alecthomas/kong"
)

// Campaign 1: PL1tiwbzkOjQz7D0l_eLJGAISVtcL7oRu_
// Campaign 2: PLFY6DNpbV0R4QegED7mK4Yg6mptOC9PyT
// Campaign 3: PL1tiwbzkOjQydg3QOkBLG9OYqWJ0dwlxF

const baseDirDownload = "captions"

type DownloadCmd struct {
	VideoUrls []string `arg:"" required:"" name:"video-url-or-id" help:"Download captions for a YouTube video."`

	Lang string
}

func (d *DownloadCmd) Run() error {
	err := createFolderIfNotExists(baseDirDownload)
	if err != nil {
		log.Fatalf("Unable to create folder: %v\n", err)
	}

	language := d.Lang
	if language == "" {
		log.Println("No language specified, defaulting to English (en)")
		language = "en"
	}

	var wg sync.WaitGroup

	for _, videoUrl := range d.VideoUrls {
		downloader, err := getDownloaderFromArg(videoUrl)
		if err != nil {
			log.Printf("Unable to download captions for %s: %v\n", videoUrl, err)
			continue
		}

		wg.Add(1)
		go downloader.Download(&wg, downloader.DestinationPath(baseDirDownload), language)
	}

	wg.Wait()

	return nil
}

var CLI struct {
	Download DownloadCmd `cmd:"" help:"Download captions for a YouTube video."`
}

func main() {
	ctx := kong.Parse(&CLI)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
