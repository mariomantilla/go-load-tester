# Go Load Tester

This project is a simple HTTP request load tester written in Go. Its purpose is to help you test the performance and reliability of HTTP endpoints by simulating different traffic profiles. You can configure the total number of requests to send and the maximum number of concurrent requests (concurrency) to simulate real-world load scenarios.

## Features

- **Configurable Target**: Set the URL, expected status code, and expected response body (string that must contain to be considered valid).
- **Timeouts**: Specify a timeout for each request.
- **Concurrent Requests**: Set the total number of requests and the maximum number of concurrent workers.
- **Statistics**: After the test, see total requests, successful and failed requests, success rate, average, minimum, and maximum response times.

## Usage

1. **Clone the repository:**
   ```sh
   git clone https://github.com/yourusername/go-load-tester.git
   cd go-load-tester
   ```

2. **Build the project:**
   ```sh
   go build -o load-tester
   ```

3. **Run the tester:**
   ```sh
   ./load-tester [flags]
   ```

## Command-Line Flags

- `-url` (string): Target URL to test (default: `http://localhost:8080`)
- `-requests` (int): Total number of requests to send (default: `100`)
- `-concurrency` (int): Number of concurrent workers (default: `10`)
- `-status` (int): Expected HTTP status code (default: `200`)
- `-body` (string): Expected response body content (default: `""`)
- `-timeout` (int): Request timeout in seconds (default: `5`)

### Example

```sh
./load-tester -url http://localhost:8080 -requests 200 -concurrency 20 -status 200 -body "OK" -timeout 3
```

## Output

After running, the tool prints:

- The URL and expectations being tested
- Progress and completion time
- A summary including:
  - Total Requests
  - Successful and Failed Requests
  - Success Rate
  - Average, Min, and Max Response Times

