package stats

import (
	"encoding/json"
	"fmt"
	"loadtester/internal/errors"
	"sort"
	"strings"
)

func PrintDetailedStats(stats LoadTestStats) {
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
			errorType errors.ErrorType
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

func PrintJSONStats(stats LoadTestStats) {
	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		fmt.Printf("Error creating JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}
