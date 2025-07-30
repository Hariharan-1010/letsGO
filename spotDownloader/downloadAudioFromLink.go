package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/kkdai/youtube/v2"
)

func sanitizeFileName(name string) string {
	illegal := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, ch := range illegal {
		name = strings.ReplaceAll(name, ch, "_")
	}
	return name
}

func DownloadAudio(videoURL string) {
	client := youtube.Client{}

	video, err := client.GetVideo(videoURL)
	if err != nil {
		log.Fatalf("error fetching video: %v", err)
	}

	formats := video.Formats.Type("audio")
	if len(formats) == 0 {
		log.Fatal("no audio-only formats available")
	}

	// Sort formats by bitrate (highest first)
	sort.Slice(formats, func(i, j int) bool {
		return formats[i].Bitrate > formats[j].Bitrate
	})

	var bestAudio *youtube.Format
	var stream io.ReadCloser

	// Try formats until one works
	for _, f := range formats {
		s, _, err := client.GetStreamContext(context.Background(), video, &f)
		if err == nil {
			bestAudio = &f
			stream = s
			fmt.Printf("✅ Selected format: Bitrate: %d | Mime: %s\n", f.Bitrate, f.MimeType)
			break
		} else {
			fmt.Printf("⚠️ Skipping format (Bitrate: %d) due to error: %v\n", f.Bitrate, err)
		}
	}

	if bestAudio == nil || stream == nil {
		log.Fatal("no working audio format found (all formats returned 403/errors)")
	}
	defer stream.Close()

	// Create sanitized file name
	fileName := sanitizeFileName(video.Title) + ".mp4"
	outFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("error creating file: %v", err)
	}
	defer outFile.Close()

	// Download
	_, err = io.Copy(outFile, stream)
	if err != nil {
		log.Fatalf("error writing file: %v", err)
	}

	fmt.Println("✅ Download complete:", fileName)
}
