package scraper

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

type ScrapResult struct {
	URL     string
	Content string
	Error   error
}

func ParallelScraper(urls []string, maxConcurrent int, timeout time.Duration) []ScrapResult {
	// define result
	results := make([]ScrapResult, 1)
	// define channels
	resultsChan := make(chan ScrapResult, len(urls))
	// define semaphore
	semaphore := make(chan struct{}, maxConcurrent)
	// define wait group
	var wg sync.WaitGroup

	// start here
	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// Rate limiting
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Create request with context
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				resultsChan <- ScrapResult{URL: url, Error: err}
				return
			}

			// Perform request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				resultsChan <- ScrapResult{URL: url, Error: err}
				return
			}
			defer resp.Body.Close()

			// Read body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				resultsChan <- ScrapResult{URL: url, Error: err}
				return
			}

			resultsChan <- ScrapResult{URL: url, Content: string(body), Error: nil}
		}(url)
	}

	// Close result channel when all gorutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect result
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}
