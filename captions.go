package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/youtube/v3"
)

const baseURL = "https://www.youtube.com/watch?v="

type ytInitialPlayerResponse struct {
	Captions struct {
		PlayerCaptionsTracklistRenderer struct {
			CaptionTracks []struct {
				BaseUrl      string `json:"baseUrl"`
				LanguageCode string `json:"languageCode"`
			} `json:"captionTracks"`
		} `json:"playerCaptionsTracklistRenderer"`
	} `json:"captions"`
}

type Caption struct {
	BaseUrl      string
	LanguageCode string
}

func (c *Caption) Download(targetPath string) error {
	resp, err := http.Get(c.BaseUrl)
	if err != nil {
		return fmt.Errorf("unable to download caption: %w", err)
	}

	defer resp.Body.Close()

	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("unable to create file: %w", err)
	}

	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("unable to write file: %w", err)
	}

	return nil
}

func listVideoCaptions(videoID string) ([]Caption, error) {
	resp, err := http.Get(baseURL + videoID)
	if err != nil {
		return nil, fmt.Errorf("unable to download video page: %w", err)
	}

	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	pageContent := string(content)

	// Find ytInitialPlayerResponse variable
	pageContentSplited := strings.Split(pageContent, "ytInitialPlayerResponse = ")
	if len(pageContentSplited) < 2 {
		return nil, fmt.Errorf("unable to find ytInitialPlayerResponse variable")
	}

	// Find the end of the variable
	pageContentSplited = strings.Split(pageContentSplited[1], ";</script>")
	if len(pageContentSplited) < 2 {
		return nil, fmt.Errorf("unable to find the end of the ytInitialPlayerResponse variable")
	}

	ytInitialPlayerResponse := ytInitialPlayerResponse{}
	err = json.Unmarshal([]byte(pageContentSplited[0]), &ytInitialPlayerResponse)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal ytInitialPlayerResponse: %w", err)
	}

	captions := make([]Caption, 0, len(ytInitialPlayerResponse.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks))
	for _, caption := range ytInitialPlayerResponse.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks {
		captions = append(captions, Caption{
			BaseUrl:      caption.BaseUrl,
			LanguageCode: caption.LanguageCode,
		})
	}

	return captions, nil
}

func downloadVideoCaptions(videoID string, targetPath string, language string) error {
	captions, err := listVideoCaptions(videoID)
	if err != nil {
		return fmt.Errorf("unable to list captions: %w", err)
	}

	for _, caption := range captions {
		if caption.LanguageCode == language {
			return caption.Download(targetPath)
		}
	}

	return fmt.Errorf("unable to find caption with language: %v", language)
}

func downloadPlaylistCaptions(wg *sync.WaitGroup, service *youtube.Service, playlistID string, targetPath string, nextPageToken string) error {
	defer wg.Done()

	part := []string{"snippet"}
	playlistItemsList := service.PlaylistItems.List(part).PlaylistId(playlistID)

	if nextPageToken != "" {
		playlistItemsList = playlistItemsList.PageToken(nextPageToken)
	}

	items, err := playlistItemsList.Do()
	if err != nil {
		return fmt.Errorf("unable to list playlist items: %w", err)
	}

	for _, item := range items.Items {
		wg.Add(1)

		go func(item *youtube.PlaylistItem) {
			defer wg.Done()

			// Sleep for 1-5 seconds to avoid rate limiting
			time.Sleep(time.Duration(1+time.Now().UnixNano()%5) * time.Second)

			name := fmt.Sprintf("captions_%s_%d.xml", item.Snippet.ResourceId.VideoId, item.Snippet.Position)
			targetVideoPath := filepath.Join(targetPath, name)

			// Check if the caption file already exists
			if _, err := os.Stat(targetVideoPath); err == nil {
				log.Printf("Caption file already exists: %v\n", targetVideoPath)
				return
			}

			log.Printf("Downloading captions for video: %v postion: %d\n", item.Snippet.ResourceId.VideoId, item.Snippet.Position)

			err := downloadVideoCaptions(item.Snippet.ResourceId.VideoId, targetVideoPath, "en")
			if err != nil {
				log.Printf("Unable to download caption for video: %v position: %d title: '%s' error: %v\n", item.Snippet.ResourceId.VideoId, item.Snippet.Position, item.Snippet.Title, err)
			}
		}(item)
	}

	if items.NextPageToken != "" {
		wg.Add(1)
		return downloadPlaylistCaptions(wg, service, playlistID, targetPath, items.NextPageToken)
	}

	return nil
}
