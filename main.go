package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RequestConfig struct {
	URL            string
	ExpectedStatus int
	ExpectedBody   string
	Timeout        time.Duration
}

type TestResult struct {
	Success      bool
	StatusCode   int
	ResponseTime time.Duration
	Error        error
}

type LoadTestStats struct {
	TotalRequests  int
	SuccessfulReqs int
	FailedReqs     int
	SuccessRate    float64
	AverageTime    time.Duration
	MinTime        time.Duration
	MaxTime        time.Duration
}

func makeRequest(config RequestConfig) TestResult {
	startTime := time.Now()
	client := &http.Client{
		Timeout: config.Timeout,
	}

	resp, err := client.Get(config.URL)
	responseTime := time.Since(startTime)

	if err != nil {
		return TestResult{
			Success:      false,
			StatusCode:   0,
			ResponseTime: responseTime,
			Error:        err,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TestResult{
			Success:      false,
			StatusCode:   resp.StatusCode,
			ResponseTime: responseTime,
			Error:        err,
		}
	}

	success := resp.StatusCode == config.ExpectedStatus && strings.Contains(string(body), config.ExpectedBody)

	return TestResult{
		Success:      success,
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
		Error:        nil,
	}
}

// runLoadTest performs concurrent HTTP requests
func runLoadTest(config RequestConfig, numRequests int, concurrency int) LoadTestStats {
	results := make(chan TestResult, numRequests)

	// Use a semaphore to limit concurrency
	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	fmt.Printf("Starting load test: %d requests with %d concurrent workers\n",
		numRequests, concurrency)

	startTime := time.Now()

	// Launch goroutines
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}

			// Make request
			result := makeRequest(config)
			results <- result

			// Release semaphore
			<-semaphore
		}()
	}

	// Close results channel when all requests complete
	go func() {
		wg.Wait()
		close(results)
	}()

	stats := collectAndCalculateStats(results)

	totalTime := time.Since(startTime)
	fmt.Printf("Load test completed in %v\n", totalTime)

	return stats
}

func collectAndCalculateStats(results chan TestResult) LoadTestStats {
	stats := LoadTestStats{MinTime: time.Hour}
	var totalTime time.Duration

	for result := range results {
		stats.TotalRequests++
		if result.Success {
			stats.SuccessfulReqs++
		} else {
			stats.FailedReqs++
		}

		totalTime += result.ResponseTime
		if result.ResponseTime < stats.MinTime {
			stats.MinTime = result.ResponseTime
		}
		if result.ResponseTime > stats.MaxTime {
			stats.MaxTime = result.ResponseTime
		}
	}

	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulReqs) / float64(stats.TotalRequests) * 100
		stats.AverageTime = totalTime / time.Duration(stats.TotalRequests)
	}

	return stats
}

func printStats(stats LoadTestStats) {
	fmt.Println("\n=== LOAD TEST RESULTS ===")
	fmt.Printf("Total Requests:    %d\n", stats.TotalRequests)
	fmt.Printf("Successful:        %d\n", stats.SuccessfulReqs)
	fmt.Printf("Failed:            %d\n", stats.FailedReqs)
	fmt.Printf("Success Rate:      %.2f%%\n", stats.SuccessRate)
	fmt.Printf("Average Time:      %v\n", stats.AverageTime)
	fmt.Printf("Min Time:          %v\n", stats.MinTime)
	fmt.Printf("Max Time:          %v\n", stats.MaxTime)
	fmt.Println("========================")
}

func main() {
	config := RequestConfig{
		URL:            "http://example.com",
		ExpectedStatus: http.StatusOK,
		ExpectedBody:   "Example Domain",
		Timeout:        5 * time.Second,
	}

	// Run load test with 100 requests and 10 concurrent workers
	stats := runLoadTest(config, 1000, 100)
	printStats(stats)
}
