package main

import (
	"flag"
	"fmt"
	"loadtester/internal/client"
	"loadtester/internal/config"
	"loadtester/internal/runner"
	"loadtester/internal/stats"
	"os"
	"time"
)

func parseAndValidateFlags() (config.RequestConfig, int, int, bool, error) {
	url := flag.String("url", "http://localhost:8080", "Target URL to test")
	requests := flag.Int("requests", 100, "Total number of requests")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent workers")
	expectedCode := flag.Int("status", 200, "Expected HTTP status code")
	expectedBody := flag.String("body", "", "Expected response body content")
	timeout := flag.Int("timeout", 5, "Request timeout in seconds")
	outputJSON := flag.Bool("json", false, "Output results in JSON format")

	flag.Parse()

	// Validation
	if *requests < 1 {
		return config.RequestConfig{}, 0, 0, false, fmt.Errorf("requests must be >= 1, got %d", *requests)
	}
	if *concurrency < 1 {
		return config.RequestConfig{}, 0, 0, false, fmt.Errorf("concurrency must be >= 1, got %d", *concurrency)
	}
	if *timeout < 1 {
		return config.RequestConfig{}, 0, 0, false, fmt.Errorf("timeout must be >= 1, got %d", *timeout)
	}

	cfg := config.RequestConfig{
		URL:            *url,
		ExpectedStatus: *expectedCode,
		ExpectedBody:   *expectedBody,
		Timeout:        time.Duration(*timeout) * time.Second,
	}
	return cfg, *requests, *concurrency, *outputJSON, nil
}

func main() {
	cfg, requests, concurrency, outputJSON, err := parseAndValidateFlags()
	if err != nil {
		// Print error and exit with non-zero code
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	results_stats := runner.RunLoadTest(cfg, requests, concurrency, client.MakeRequest)

	if outputJSON {
		stats.PrintJSONStats(results_stats)
	} else {
		stats.PrintDetailedStats(results_stats)
	}
}
