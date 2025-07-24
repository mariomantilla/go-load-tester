# Go Load Tester

This project is a simple HTTP request load tester written in Go. Its purpose is to help you test the performance and reliability of HTTP endpoints by simulating different traffic profiles. You can configure the total number of requests to send and the maximum number of concurrent requests (concurrency) to simulate real-world load scenarios.

## Features

- **Configurable Target**: Set the URL, expected status code, and expected response body.
- **Timeouts**: Specify a timeout for each request.
- **Result Reporting**: See if the request succeeded, the status code, error (if any), and response time.
- **(Planned)**: Support for total requests and max concurrency to simulate load.

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
   ./load-tester
   ```

   By default, it sends a single request to `http://example.com` and checks for a 200 OK status and the presence of "Example Domain" in the response body.

## Configuration

Edit the `main.go` file to change the request parameters:

```go
config := RequestConfig{
    URL:            "http://example.com",
    ExpectedStatus: http.StatusOK,
    ExpectedBody:   "Example Domain",
    Timeout:        5 * time.Second,
}
```

## Roadmap

- [ ] Add support for specifying total number of requests
- [ ] Add support for setting maximum concurrency
- [ ] Aggregate and report statistics (success rate, average response time, etc.)
- [ ] Command-line flags for configuration

## Requirements

- Go 1.18 or newer

## License

MIT License
