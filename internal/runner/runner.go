package runner

import (
	"fmt"
	"loadtester/internal/client"
	"loadtester/internal/config"
	"loadtester/internal/stats"
	"sync"
	"time"
)

func RunLoadTest(config config.RequestConfig, numRequests int, concurrency int) stats.LoadTestStats {
	results := make(chan client.TestResult, numRequests)

	// Use a semaphore to limit concurrency
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	fmt.Printf("Starting load test: %d requests with %d concurrent workers\n",
		numRequests, concurrency)
	fmt.Printf("Target URL: %s\n", config.URL)
	fmt.Printf("Expected status: %d\n", config.ExpectedStatus)
	if config.ExpectedBody != "" {
		fmt.Printf("Expected body contains: %s\n", config.ExpectedBody)
	}
	fmt.Println("---")

	startTime := time.Now()

	progressChan := make(chan struct{}, numRequests)
	go func() {
		completed := 0
		for range progressChan {
			completed++
			if completed%10 == 0 || completed == numRequests {
				fmt.Printf("Progress: %d/%d requests completed\n", completed, numRequests)
			}
		}
	}()

	// Launch goroutines for concurrent requests
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}

			// Make request
			result := client.MakeRequest(config)
			results <- result
			progressChan <- struct{}{}
			// Release semaphore
			<-semaphore
		}()
	}

	// Close results channel when all requests complete
	go func() {
		wg.Wait()
		close(results)
		close(progressChan)
	}()

	return stats.CollectAndCalculateStats(results, startTime)
}
