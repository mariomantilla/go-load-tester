package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
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

func main() {
	config := RequestConfig{
		URL:            "http://example.com",
		ExpectedStatus: http.StatusOK,
		ExpectedBody:   "Example Domain",
		Timeout:        5 * time.Second,
	}

	result := makeRequest(config)

	if result.Success {
		fmt.Printf("Request succeeded with status code %d in %v\n", result.StatusCode, result.ResponseTime)
	} else {
		fmt.Printf("Request failed with status code %d, error: %v, response time: %v\n", result.StatusCode, result.Error, result.ResponseTime)
	}
}
