package stats

import (
	"loadtester/internal/client"
	"loadtester/internal/errors"
	"sort"
	"time"
)

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
	ErrorBreakdown  map[errors.ErrorType]int
	StatusBreakdown map[int]int

	// Performance insights
	TotalDataTransfer int64
	RequestsPerSecond float64
	TestDuration      time.Duration

	// Response time distribution
	ResponseTimes []time.Duration
}

func CollectAndCalculateStats(results chan client.TestResult, testStart time.Time) LoadTestStats {
	stats := LoadTestStats{
		MinTime:         time.Hour,
		ErrorBreakdown:  make(map[errors.ErrorType]int),
		StatusBreakdown: make(map[int]int),
		ResponseTimes:   make([]time.Duration, 0),
		TestDuration:    0,
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

	stats.TestDuration = time.Since(testStart)

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
