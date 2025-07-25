package main

import (
	"flag"
	"loadtester/internal/config"
	"loadtester/internal/runner"
	"loadtester/internal/stats"
	"time"
)

func main() {
	// Define command line flags
	var (
		url          = flag.String("url", "http://localhost:8080", "Target URL to test")
		requests     = flag.Int("requests", 100, "Total number of requests")
		concurrency  = flag.Int("concurrency", 10, "Number of concurrent workers")
		expectedCode = flag.Int("status", 200, "Expected HTTP status code")
		expectedBody = flag.String("body", "", "Expected response body content")
		timeout      = flag.Int("timeout", 5, "Request timeout in seconds")
		outputJSON   = flag.Bool("json", false, "Output results in JSON format")
	)

	flag.Parse()

	config := config.RequestConfig{
		URL:            *url,
		ExpectedStatus: *expectedCode,
		ExpectedBody:   *expectedBody,
		Timeout:        time.Duration(*timeout) * time.Second,
	}

	results_stats := runner.RunLoadTest(config, *requests, *concurrency)

	if *outputJSON {
		stats.PrintJSONStats(results_stats)
	} else {
		stats.PrintDetailedStats(results_stats)
	}

}
