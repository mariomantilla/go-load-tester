package client

import (
	"context"
	"crypto/tls"
	"io"
	"loadtester/internal/config"
	"loadtester/internal/errors"
	"net"
	"net/http"
	"time"
)

type TestResult struct {
	Success      bool
	StatusCode   int
	ResponseTime time.Duration
	ErrorType    errors.ErrorType
	ErrorMessage string
	ResponseSize int64
}

func MakeRequest(config config.RequestConfig) TestResult {
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
			MaxIdleConns:          config.Concurrency, // Limit max idle connections
			MaxIdleConnsPerHost:   config.Concurrency,
			// Skip TLS verification for testing (optional)
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", config.URL, nil)
	if err != nil {
		responseTime := time.Since(start)
		errorType, errorMsg := errors.CategorizeError(err, 0, config.ExpectedStatus, config.ExpectedBody, "")
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
		errorType, errorMsg := errors.CategorizeError(err, 0, config.ExpectedStatus, config.ExpectedBody, "")
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
		errorType, errorMsg := errors.CategorizeError(err, 0, config.ExpectedStatus, config.ExpectedBody, "")
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
	errorType, errorMsg := errors.CategorizeError(nil, resp.StatusCode, config.ExpectedStatus, config.ExpectedBody, bodyStr)

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
