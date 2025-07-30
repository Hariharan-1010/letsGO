package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func main() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	url := `https://open.spotify.com/album/0Rkv5iqjF2uenfL0OVB8hg` // replace with your playlist
	trackLinks := make(map[string]bool)

	// Start browser and navigate
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
	); err != nil {
		log.Fatal(err)
	}

	maxScrolls := 30
	idleScrolls := 5

	for i := 0; i < maxScrolls; i++ {
		fmt.Printf("🔁 Scroll %d...\n", i+1)

		// Extract track links in view
		var nodes []*cdp.Node
		if err := chromedp.Run(ctx,
			chromedp.Nodes(`div[aria-colindex="2"] a[href^="/track/"]`, &nodes, chromedp.ByQueryAll),
		); err != nil {
			log.Fatal(err)
		}

		newLinks := 0
		for _, node := range nodes {
			for i := 0; i < len(node.Attributes); i += 2 {
				if node.Attributes[i] == "href" {
					href := node.Attributes[i+1]
					if strings.HasPrefix(href, "/track/") {
						fullURL := "https://open.spotify.com" + href
						if !trackLinks[fullURL] {
							trackLinks[fullURL] = true
							newLinks++
						}
					}
				}
			}
		}

		fmt.Printf("➕ Found %d new links (Total: %d)\n", newLinks, len(trackLinks))

		// Stop if no new links
		if newLinks == 0 {
			idleScrolls--
			if idleScrolls <= 0 {
				fmt.Println("✅ No new links after multiple scrolls. Stopping.")
				break
			}
		}

		// Scroll the correct container
		if err := chromedp.Run(ctx,
			chromedp.Evaluate(`document.querySelectorAll('[data-overlayscrollbars-viewport]')[1].scrollBy(0,1000);`, nil),
			chromedp.Sleep(1500*time.Millisecond),
		); err != nil {
			log.Fatal(err)
		}
	}

	// Print results
	fmt.Println("🎵 Final list of track links:")
	for link := range trackLinks {
		fmt.Println(link)
	}

	for link := range trackLinks {
		videoURL, err := ExtractVideoLink(link)
		if err != nil {
			log.Printf("Error extracting video link for %s: %v", link, err)
			continue
		}
		fmt.Println("Extracted Video URL:", videoURL)
		DownloadAudio(videoURL)
	}
}
