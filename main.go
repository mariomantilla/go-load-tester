package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

type ErrorType string

const (
	ErrorTypeNone           ErrorType = ""
	ErrorTypeDNS            ErrorType = "DNS"
	ErrorTypeConnection     ErrorType = "Connection"
	ErrorTypeTimeout        ErrorType = "Timeout"
	ErrorTypeTLS            ErrorType = "TLS"
	ErrorTypeURL            ErrorType = "URL"
	ErrorTypeNetwork        ErrorType = "Network"
	ErrorTypeServerError    ErrorType = "Server Error"
	ErrorTypeClientError    ErrorType = "Client Error"
	ErrorTypeRedirect       ErrorType = "Redirect"
	ErrorTypeHTTPStatus     ErrorType = "HTTP Status"
	ErrorTypeBodyValidation ErrorType = "Body Validation"
)

type RequestConfig struct {
	URL            string
	ExpectedStatus int
	ExpectedBody   string
	Timeout        time.Duration
	Concurrency    int
}

type TestResult struct {
	Success      bool
	StatusCode   int
	ResponseTime time.Duration
	ErrorType    ErrorType
	ErrorMessage string
	ResponseSize int64
}

type LoadTestStats struct {
	TotalRequests  int
	SuccessfulReqs int
	FailedReqs     int
	SuccessRate    float64
	AverageTime    time.Duration
	MinTime        time.Duration
	MaxTime        time.Duration
	MedianTime     time.Duration
	P95Time        time.Duration
	P99Time        time.Duration

	// Error breakdown
	ErrorBreakdown  map[ErrorType]int
	StatusBreakdown map[int]int

	// Performance insights
	TotalDataTransfer int64
	RequestsPerSecond float64
	TestDuration      time.Duration

	// Response time distribution
	ResponseTimes []time.Duration
}

func categorizeError(err error, statusCode int, expectedStatus int, expectedBody, responseBody string) (ErrorType, string) {
	if err != nil {
		// Network-level errors
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				return ErrorTypeTimeout, fmt.Sprintf("Request timeout: %v", err)
			}
		}

		// DNS resolution errors
		if dnsErr, ok := err.(*net.DNSError); ok {
			return ErrorTypeDNS, fmt.Sprintf("DNS resolution failed: %v", dnsErr)
		}

		// Connection errors
		if opErr, ok := err.(*net.OpError); ok {
			if opErr.Op == "dial" {
				return ErrorTypeConnection, fmt.Sprintf("Connection failed: %v", err)
			}
		}

		// URL parsing errors
		if _, ok := err.(*url.Error); ok {
			return ErrorTypeURL, fmt.Sprintf("Invalid URL: %v", err)
		}

		// TLS errors
		if strings.Contains(err.Error(), "tls") || strings.Contains(err.Error(), "certificate") {
			return ErrorTypeTLS, fmt.Sprintf("TLS/SSL error: %v", err)
		}

		// Context cancellation (timeouts we set)
		if err == context.DeadlineExceeded {
			return ErrorTypeTimeout, "Request deadline exceeded"
		}

		// Generic network error
		return ErrorTypeNetwork, fmt.Sprintf("Network error: %v", err)
	}

	// HTTP-level errors (got response but wrong status)
	if statusCode != expectedStatus {
		if statusCode >= 500 {
			return ErrorTypeServerError, fmt.Sprintf("Server error (HTTP %d)", statusCode)
		} else if statusCode >= 400 {
			return ErrorTypeClientError, fmt.Sprintf("Client error (HTTP %d)", statusCode)
		} else if statusCode >= 300 {
			return ErrorTypeRedirect, fmt.Sprintf("Unexpected redirect (HTTP %d)", statusCode)
		} else {
			return ErrorTypeHTTPStatus, fmt.Sprintf("Unexpected status code: %d (expected %d)", statusCode, expectedStatus)
		}
	}

	// Body validation errors
	if expectedBody != "" && !strings.Contains(responseBody, expectedBody) {
		return ErrorTypeBodyValidation, fmt.Sprintf("Response body doesn't contain expected text: '%s'", expectedBody)
	}

	return ErrorTypeNone, "" // No error
}

func makeRequest(config RequestConfig) TestResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second, // Connection timeout
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConns:          min(config.Concurrency, 20), // Limit max idle connections
			MaxIdleConnsPerHost:   min(config.Concurrency, 20),
			// Skip TLS verification for testing (optional)
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", config.URL, nil)
	if err != nil {
		responseTime := time.Since(start)
		errorType, errorMsg := categorizeError(err, 0, config.ExpectedStatus, config.ExpectedBody, "")
		return TestResult{
			Success:      false,
			StatusCode:   0,
			ResponseTime: responseTime,
			ErrorType:    errorType,
			ErrorMessage: errorMsg,
		}
	}

	// Add User-Agent for identification
	req.Header.Set("User-Agent", "Go-Load-Tester/1.0")

	// Make the request
	resp, err := client.Do(req)
	responseTime := time.Since(start)

	if err != nil {
		errorType, errorMsg := categorizeError(err, 0, config.ExpectedStatus, config.ExpectedBody, "")
		return TestResult{
			Success:      false,
			StatusCode:   0,
			ResponseTime: responseTime,
			ErrorType:    errorType,
			ErrorMessage: errorMsg,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		errorType, errorMsg := categorizeError(err, 0, config.ExpectedStatus, config.ExpectedBody, "")
		return TestResult{
			Success:      false,
			StatusCode:   resp.StatusCode,
			ResponseTime: responseTime,
			ErrorType:    errorType,
			ErrorMessage: errorMsg,
			ResponseSize: int64(len(body)),
		}
	}

	bodyStr := string(body)
	errorType, errorMsg := categorizeError(nil, resp.StatusCode, config.ExpectedStatus, config.ExpectedBody, bodyStr)

	success := errorType == ""

	return TestResult{
		Success:      success,
		StatusCode:   resp.StatusCode,
		ResponseTime: responseTime,
		ErrorType:    errorType,
		ErrorMessage: errorMsg,
		ResponseSize: int64(len(body)),
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
	for range numRequests {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}

			// Make request
			result := makeRequest(config)
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

	return collectAndCalculateStats(results, startTime)
}

func collectAndCalculateStats(results chan TestResult, testStart time.Time) LoadTestStats {
	stats := LoadTestStats{
		MinTime:         time.Hour,
		ErrorBreakdown:  make(map[ErrorType]int),
		StatusBreakdown: make(map[int]int),
		ResponseTimes:   make([]time.Duration, 0),
		TestDuration:    time.Since(testStart),
	}
	var totalTime time.Duration

	for result := range results {
		stats.TotalRequests++
		stats.ResponseTimes = append(stats.ResponseTimes, result.ResponseTime)
		stats.TotalDataTransfer += result.ResponseSize

		if result.Success {
			stats.SuccessfulReqs++
		} else {
			stats.FailedReqs++
			// Track error types
			if result.ErrorType != "" {
				stats.ErrorBreakdown[result.ErrorType]++
			}
		}

		if result.StatusCode > 0 {
			stats.StatusBreakdown[result.StatusCode]++
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
		stats.RequestsPerSecond = float64(stats.TotalRequests) / stats.TestDuration.Seconds()

		// Calculate percentiles
		sort.Slice(stats.ResponseTimes, func(i, j int) bool {
			return stats.ResponseTimes[i] < stats.ResponseTimes[j]
		})

		if len(stats.ResponseTimes) > 0 {
			stats.MedianTime = percentile(stats.ResponseTimes, 50)
			stats.P95Time = percentile(stats.ResponseTimes, 95)
			stats.P99Time = percentile(stats.ResponseTimes, 99)
		}
	}

	return stats
}

func percentile(sortedTimes []time.Duration, p int) time.Duration {
	if len(sortedTimes) == 0 {
		return 0
	}
	index := (len(sortedTimes) * p / 100)
	if index >= len(sortedTimes) {
		index = len(sortedTimes) - 1
	}
	return sortedTimes[index]
}

func printDetailedStats(stats LoadTestStats) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("LOAD TEST RESULTS")
	fmt.Println(strings.Repeat("=", 60))

	// Summary
	fmt.Printf("Total Requests:     %d\n", stats.TotalRequests)
	fmt.Printf("Successful:         %d (%.2f%%)\n", stats.SuccessfulReqs, stats.SuccessRate)
	fmt.Printf("Failed:             %d (%.2f%%)\n", stats.FailedReqs, 100-stats.SuccessRate)
	fmt.Printf("Test Duration:      %v\n", stats.TestDuration)
	fmt.Printf("Requests/sec:       %.2f\n", stats.RequestsPerSecond)
	fmt.Printf("Data Transferred:   %.2f MB\n", float64(stats.TotalDataTransfer)/(1024*1024))

	// Response Time Statistics
	fmt.Println("\nResponse Time Statistics:")
	fmt.Printf("  Average:          %v\n", stats.AverageTime)
	fmt.Printf("  Median (50th):    %v\n", stats.MedianTime)
	fmt.Printf("  95th percentile:  %v\n", stats.P95Time)
	fmt.Printf("  99th percentile:  %v\n", stats.P99Time)
	fmt.Printf("  Min:              %v\n", stats.MinTime)
	fmt.Printf("  Max:              %v\n", stats.MaxTime)

	// Status Code Breakdown
	if len(stats.StatusBreakdown) > 0 {
		fmt.Println("\nHTTP Status Code Breakdown:")
		var statusCodes []int
		for code := range stats.StatusBreakdown {
			statusCodes = append(statusCodes, code)
		}
		sort.Ints(statusCodes)

		for _, code := range statusCodes {
			count := stats.StatusBreakdown[code]
			percentage := float64(count) / float64(stats.TotalRequests) * 100
			fmt.Printf("  %d: %d (%.2f%%)\n", code, count, percentage)
		}
	}

	// Error Breakdown
	if len(stats.ErrorBreakdown) > 0 {
		fmt.Println("\nError Type Breakdown:")
		type errorStat struct {
			errorType ErrorType
			count     int
		}

		var errorStats []errorStat
		for errorType, count := range stats.ErrorBreakdown {
			errorStats = append(errorStats, errorStat{errorType, count})
		}

		// Sort by count (descending)
		sort.Slice(errorStats, func(i, j int) bool {
			return errorStats[i].count > errorStats[j].count
		})

		for _, stat := range errorStats {
			percentage := float64(stat.count) / float64(stats.TotalRequests) * 100
			fmt.Printf("  %s: %d (%.2f%%)\n", stat.errorType, stat.count, percentage)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

func printJSONStats(stats LoadTestStats) {
	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		fmt.Printf("Error creating JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

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

	config := RequestConfig{
		URL:            *url,
		ExpectedStatus: *expectedCode,
		ExpectedBody:   *expectedBody,
		Timeout:        time.Duration(*timeout) * time.Second,
	}

	stats := runLoadTest(config, *requests, *concurrency)

	if *outputJSON {
		printJSONStats(stats)
	} else {
		printDetailedStats(stats)
	}

}
