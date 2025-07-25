package stats

import (
	"loadtester/internal/client"
	"loadtester/internal/errors"
	"testing"
	"time"
)

func makeResult(success bool, status int, respTime time.Duration, errType errors.ErrorType, respSize int64) client.TestResult {
	return client.TestResult{
		Success:      success,
		StatusCode:   status,
		ResponseTime: respTime,
		ErrorType:    errType,
		ResponseSize: respSize,
	}
}

func TestCollectAndCalculateStats_BasicStats(t *testing.T) {
	results := make(chan client.TestResult, 5)
	start := time.Now().Add(-2 * time.Second)

	results <- makeResult(true, 200, 100*time.Millisecond, errors.ErrorTypeNone, 100)
	results <- makeResult(true, 200, 200*time.Millisecond, errors.ErrorTypeNone, 200)
	results <- makeResult(false, 500, 300*time.Millisecond, errors.ErrorTypeServerError, 300)
	results <- makeResult(false, 404, 400*time.Millisecond, errors.ErrorTypeClientError, 400)
	results <- makeResult(true, 200, 500*time.Millisecond, errors.ErrorTypeNone, 500)
	close(results)

	stats := CollectAndCalculateStats(results, start)

	if stats.TotalRequests != 5 {
		t.Errorf("Expected 5 requests, got %d", stats.TotalRequests)
	}
	if stats.SuccessfulReqs != 3 {
		t.Errorf("Expected 3 successful requests, got %d", stats.SuccessfulReqs)
	}
	if stats.FailedReqs != 2 {
		t.Errorf("Expected 2 failed requests, got %d", stats.FailedReqs)
	}
	if stats.SuccessRate != 60 {
		t.Errorf("Expected success rate 60, got %f", stats.SuccessRate)
	}
	if stats.MinTime != 100*time.Millisecond {
		t.Errorf("Expected min time 100ms, got %v", stats.MinTime)
	}
	if stats.MaxTime != 500*time.Millisecond {
		t.Errorf("Expected max time 500ms, got %v", stats.MaxTime)
	}
	if stats.AverageTime != 300*time.Millisecond {
		t.Errorf("Expected average time 300ms, got %v", stats.AverageTime)
	}
	if stats.TotalDataTransfer != 1500 {
		t.Errorf("Expected total data transfer 1500, got %d", stats.TotalDataTransfer)
	}
	if stats.StatusBreakdown[200] != 3 || stats.StatusBreakdown[500] != 1 || stats.StatusBreakdown[404] != 1 {
		t.Errorf("Status breakdown incorrect: %+v", stats.StatusBreakdown)
	}
	if stats.ErrorBreakdown[errors.ErrorTypeServerError] != 1 || stats.ErrorBreakdown[errors.ErrorTypeClientError] != 1 {
		t.Errorf("Error breakdown incorrect: %+v", stats.ErrorBreakdown)
	}
}

func TestCollectAndCalculateStats_Percentiles(t *testing.T) {
	results := make(chan client.TestResult, 5)
	start := time.Now().Add(-1 * time.Second)

	results <- makeResult(true, 200, 100*time.Millisecond, errors.ErrorTypeNone, 100)
	results <- makeResult(true, 200, 200*time.Millisecond, errors.ErrorTypeNone, 200)
	results <- makeResult(true, 200, 300*time.Millisecond, errors.ErrorTypeNone, 300)
	results <- makeResult(true, 200, 400*time.Millisecond, errors.ErrorTypeNone, 400)
	results <- makeResult(true, 200, 500*time.Millisecond, errors.ErrorTypeNone, 500)
	close(results)

	stats := CollectAndCalculateStats(results, start)

	if stats.MedianTime != 300*time.Millisecond {
		t.Errorf("Expected median time 300ms, got %v", stats.MedianTime)
	}
	if stats.P95Time != 500*time.Millisecond {
		t.Errorf("Expected P95 time 500ms, got %v", stats.P95Time)
	}
	if stats.P99Time != 500*time.Millisecond {
		t.Errorf("Expected P99 time 500ms, got %v", stats.P99Time)
	}
}

func TestPercentile_EmptySlice(t *testing.T) {
	if percentile([]time.Duration{}, 50) != 0 {
		t.Error("Expected percentile of empty slice to be 0")
	}
}

func TestPercentile_Bounds(t *testing.T) {
	times := []time.Duration{10, 20, 30, 40, 50}
	if percentile(times, 100) != 50 {
		t.Errorf("Expected 100th percentile to be 50, got %v", percentile(times, 100))
	}
}
