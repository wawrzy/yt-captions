package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"sync"

	"google.golang.org/api/youtube/v3"
)

type Downloader interface {
	Download(wg *sync.WaitGroup, targetPath string, language string)
	DestinationPath(dir string) string
}

type Video struct {
	VideoID string
}

func (v *Video) Download(wg *sync.WaitGroup, targetPath string, language string) {
	defer wg.Done()

	err := downloadVideoCaptions(v.VideoID, targetPath, language)
	if err != nil {
		log.Printf("Unable to download captions for %s: %v\n", v.VideoID, err)
	}
}

func (v *Video) DestinationPath(dir string) string {
	name := fmt.Sprintf("captions_%s.xml", v.VideoID)
	return filepath.Join(dir, name)
}

type Playlist struct {
	PlaylistID string
}

func (p *Playlist) Download(wg *sync.WaitGroup, targetPath string, language string) {
	defer wg.Done()

	service, err := youtube.NewService(context.Background())
	if err != nil {
		log.Printf("Unable to list captions: %v\n", err)
		return
	}

	if err := createFolderIfNotExists(targetPath); err != nil {
		log.Printf("Unable to create folder: %v\n", err)
		return
	}

	wg.Add(1)
	downloadPlaylistCaptions(wg, service, p.PlaylistID, targetPath, "")
}

func (p *Playlist) DestinationPath(path string) string {
	return filepath.Join(path, "playlist_"+p.PlaylistID)
}

func getDownloaderFromArg(arg string) (Downloader, error) {
	// CHeck if the input is a valid YouTube URL
	u, _ := url.Parse(arg)

	if u != nil {
		// If it is, extract the video ID
		q := u.Query()

		playListID := q.Get("list")

		if playListID != "" {
			return &Playlist{PlaylistID: playListID}, nil
		}

		videoID := q.Get("v")

		if videoID != "" {
			return &Video{VideoID: videoID}, nil
		}

		return nil, fmt.Errorf("invalid YouTube URL: %v", arg)
	}

	return &Video{VideoID: arg}, nil
}
