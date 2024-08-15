# yt-captions

## Overview

`yt-captions` is a learning project designed to download captions from YouTube videos and playlists.

Output format is XML.

It uses a part of the official YouTube API to handle some data fetching, like retrieving playlist information, but the actual download of captions is done using an unofficial API.

To be able to download a playlist you need to have a gcloud ADC setup locally https://cloud.google.com/docs/authentication/application-default-credentials.
You'll need the following scopes:
- https://www.googleapis.com/auth/youtube.readonly
- https://www.googleapis.com/auth/youtube.force-ssl

## Features

- Download captions for individual YouTube videos.
- Download captions for entire YouTube playlists.
- Handle multiple languages for captions.

## Project Structure

- `main.go`: Entry point of the application.
- `captions.go`: Contains functions and types related to downloading or listing captions.
- `downloader.go`: Manages the downloading process for videos and playlists.
- `util.go`: Utility functions used across the project.
- `Makefile`: Contains build and run commands.
- `README.md`: Project documentation.
- `go.mod` and `go.sum`: Go module files for dependency management.

## Usage

After building the project, you can use the binary to download captions:

```
./yt-captions <VIDEO_URL> ... --lang <LANGUAGE_CODE>
```

```
./yt-captions <PLAYLIST_URL> ... --lang <LANGUAGE_CODE>
```


Here is an example:
```
./yt-captions https://www.youtube.com/watch?v=JJIqTEYDQkU&t --lang en
```
