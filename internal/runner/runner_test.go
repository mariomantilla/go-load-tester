package runner

import (
	"loadtester/internal/client"
	"loadtester/internal/config"
	"sync/atomic"
	"testing"
	"time"
)

// Mock MakeRequest to count concurrent calls
var concurrentCalls int32
var maxConcurrent int32

func mockMakeRequest(cfg config.RequestConfig) client.TestResult {
	atomic.AddInt32(&concurrentCalls, 1)
	curr := atomic.LoadInt32(&concurrentCalls)
	for {
		prev := atomic.LoadInt32(&maxConcurrent)
		if curr > prev {
			atomic.CompareAndSwapInt32(&maxConcurrent, prev, curr)
		} else {
			break
		}
	}
	time.Sleep(10 * time.Millisecond) // Simulate work
	atomic.AddInt32(&concurrentCalls, -1)
	return client.TestResult{Success: true, StatusCode: 200, ResponseTime: 10 * time.Millisecond}
}

func TestRunLoadTest_ConcurrencyLimit(t *testing.T) {
	atomic.StoreInt32(&concurrentCalls, 0)
	atomic.StoreInt32(&maxConcurrent, 0)

	cfg := config.RequestConfig{URL: "http://test", Timeout: 1 * time.Second, ExpectedStatus: 200, Concurrency: 5}
	numRequests := 20
	concurrency := 3

	stats := RunLoadTest(cfg, numRequests, concurrency, mockMakeRequest)

	if maxConcurrent > int32(concurrency) {
		t.Errorf("Concurrency limit exceeded: max %d, expected %d", maxConcurrent, concurrency)
	}
	if stats.TotalRequests != numRequests {
		t.Errorf("Expected %d requests, got %d", numRequests, stats.TotalRequests)
	}
}

func TestRunLoadTest_AllSuccess(t *testing.T) {
	cfg := config.RequestConfig{URL: "http://test", Timeout: 1 * time.Second, ExpectedStatus: 200, Concurrency: 2}
	numRequests := 10
	concurrency := 2

	stats := RunLoadTest(cfg, numRequests, concurrency, func(cfg config.RequestConfig) client.TestResult {
		return client.TestResult{Success: true, StatusCode: 200, ResponseTime: 5 * time.Millisecond}
	})

	if stats.SuccessfulReqs != numRequests {
		t.Errorf("Expected all requests successful, got %d", stats.SuccessfulReqs)
	}
	if stats.FailedReqs != 0 {
		t.Errorf("Expected 0 failed requests, got %d", stats.FailedReqs)
	}
}
