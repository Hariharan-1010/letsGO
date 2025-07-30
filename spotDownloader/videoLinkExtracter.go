package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func ExtractVideoLink(trackURL string) (string, error) {
	fmt.Printf("Downloading track: %s\n", trackURL)

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	videoTitle := ""
	if err := chromedp.Run(ctx,
		chromedp.Navigate(trackURL),
		chromedp.Sleep(3*time.Second),
		chromedp.Evaluate(`document.title`, &videoTitle),
	); err != nil {
		return "", fmt.Errorf("failed to load Spotify track: %w", err)
	}

	fmt.Println("Track title:", videoTitle)

	cleanedTitle := strings.NewReplacer(
		" - ", " ", ":", " ", "/", "", "\\", "",
		"?", "", "*", "", "\"", "", "|", " ",
		"Spotify", "").Replace(videoTitle)

	query := strings.Join(strings.Fields(cleanedTitle), "%20")
	youtubeSearchURL := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", query)
	fmt.Println("YouTube search URL:", youtubeSearchURL)

	videoURL := ""
	if err := chromedp.Run(ctx,
		chromedp.Navigate(youtubeSearchURL),
		chromedp.Sleep(3*time.Second),
		// Click the first *actual video* (force watch page, not radio)
		chromedp.Click(`ytd-video-renderer a#thumbnail`, chromedp.NodeVisible),
		chromedp.Sleep(3*time.Second),
		// Extract canonical watch URL
		chromedp.Evaluate(`window.location.href`, &videoURL),
	); err != nil {
		return "", fmt.Errorf("failed to navigate to YouTube video: %w", err)
	}

	// Strip any playlist params (RDU, start_radio, etc.)
	parts := strings.Split(videoURL, "&")
	cleanVideoURL := parts[0]

	fmt.Println("Final clean Video URL:", cleanVideoURL)
	return cleanVideoURL, nil
}
