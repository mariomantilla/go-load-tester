package config

import "time"

type RequestConfig struct {
	URL            string
	ExpectedStatus int
	ExpectedBody   string
	Timeout        time.Duration
	Concurrency    int
}
