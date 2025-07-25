# Go Load Tester

Go Load Tester is a flexible HTTP load testing tool written in Go. It helps you evaluate the performance and reliability of HTTP endpoints by simulating configurable traffic profiles, reporting detailed statistics, and providing error breakdowns.

## Features

- **Configurable Target**: Set the URL, expected status code, and required substring in the response body.
- **Timeouts**: Specify a timeout for each request.
- **Concurrent Requests**: Control total requests and maximum concurrent workers.
- **Progress Reporting**: See progress updates during the test.
- **Detailed Statistics**: After the test, view total, successful, and failed requests, success rate, average, min, max, median, p95, and p99 response times, requests/sec, and total data transferred.
- **Breakdowns**: Get HTTP status code and error type breakdowns.
- **Output Formats**: Print results in human-readable or JSON format.

## Usage

1. **Clone the repository:**
   ```sh
   git clone https://github.com/mariomantilla/go-load-tester.git
   cd go-load-tester
   ```

2. **Build the project:**
   ```sh
   go build -o loadtester ./cmd/loadtester
   ```

3. **Run the tester:**
   ```sh
   ./loadtester [flags]
   ```

## Command-Line Flags

- `-url` (string): Target URL to test (default: `http://localhost:8080`)
- `-requests` (int): Total number of requests to send (default: `100`)
- `-concurrency` (int): Number of concurrent workers (default: `10`)
- `-status` (int): Expected HTTP status code (default: `200`)
- `-body` (string): Substring that must be present in the response body (default: `""`)
- `-timeout` (int): Request timeout in seconds (default: `5`)
- `-json` (bool): Output results in JSON format (default: `false`)

### Example

```sh
./loadtester -url http://localhost:8080 -requests 200 -concurrency 20 -status 200 -body "OK" -timeout 3 -json
```

## Output


After running, the tool prints:

- Target URL, expected status code, and expected body substring
- Progress updates during execution
- Summary including:
  - Total Requests
  - Successful and Failed Requests
  - Success Rate
  - Test Duration and Requests/sec
  - Data Transferred (MB)
  - Average, Median, Min, Max, 95th, and 99th percentile response times
  - HTTP Status Code Breakdown
  - Error Type Breakdown

If `-json` is used, all statistics are printed in JSON format for easy parsing.

