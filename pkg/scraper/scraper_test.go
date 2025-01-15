package scraper

import (
	"testing"
	"time"
)

func TestParallelScraper(t *testing.T) {
	urls := []string{
		"https://www.google.com",
		"https://www.github.com",
		"https://www.stackoverflow.com",
		"https://www.reddit.com",
	}

	results := ParallelScraper(urls, 2, 10*time.Second)
	if len(results) == 0 {
		t.Error("Expected results, got empty slice")
	} else {
		for _, result := range results {
			if result.Error == nil {
				t.Log(result)
			} else {
				t.Errorf("Error scraping %s: %v", result.URL, result.Error)
			}
		}
	}
}
