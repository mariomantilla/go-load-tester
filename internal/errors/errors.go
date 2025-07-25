package errors

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

type ErrorType string

const (
	ErrorTypeNone           ErrorType = ""
	ErrorTypeDNS            ErrorType = "DNS"
	ErrorTypeConnection     ErrorType = "Connection"
	ErrorTypeTimeout        ErrorType = "Timeout"
	ErrorTypeTLS            ErrorType = "TLS"
	ErrorTypeURL            ErrorType = "URL"
	ErrorTypeNetwork        ErrorType = "Network"
	ErrorTypeServerError    ErrorType = "Server Error"
	ErrorTypeClientError    ErrorType = "Client Error"
	ErrorTypeRedirect       ErrorType = "Redirect"
	ErrorTypeHTTPStatus     ErrorType = "HTTP Status"
	ErrorTypeBodyValidation ErrorType = "Body Validation"
)

func CategorizeError(err error, statusCode int, expectedStatus int, expectedBody, responseBody string) (ErrorType, string) {
	if err != nil {
		// Network-level errors
		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				return ErrorTypeTimeout, fmt.Sprintf("Request timeout: %v", err)
			}
		}

		// DNS resolution errors
		if dnsErr, ok := err.(*net.DNSError); ok {
			return ErrorTypeDNS, fmt.Sprintf("DNS resolution failed: %v", dnsErr)
		}

		// Connection errors
		if opErr, ok := err.(*net.OpError); ok {
			if opErr.Op == "dial" {
				return ErrorTypeConnection, fmt.Sprintf("Connection failed: %v", err)
			}
		}

		// URL parsing errors
		if _, ok := err.(*url.Error); ok {
			return ErrorTypeURL, fmt.Sprintf("Invalid URL: %v", err)
		}

		// TLS errors
		if strings.Contains(err.Error(), "tls") || strings.Contains(err.Error(), "certificate") {
			return ErrorTypeTLS, fmt.Sprintf("TLS/SSL error: %v", err)
		}

		// Context cancellation (timeouts we set)
		if err == context.DeadlineExceeded {
			return ErrorTypeTimeout, "Request deadline exceeded"
		}

		// Generic network error
		return ErrorTypeNetwork, fmt.Sprintf("Network error: %v", err)
	}

	// HTTP-level errors (got response but wrong status)
	if statusCode != expectedStatus {
		if statusCode >= 500 {
			return ErrorTypeServerError, fmt.Sprintf("Server error (HTTP %d)", statusCode)
		} else if statusCode >= 400 {
			return ErrorTypeClientError, fmt.Sprintf("Client error (HTTP %d)", statusCode)
		} else if statusCode >= 300 {
			return ErrorTypeRedirect, fmt.Sprintf("Unexpected redirect (HTTP %d)", statusCode)
		} else {
			return ErrorTypeHTTPStatus, fmt.Sprintf("Unexpected status code: %d (expected %d)", statusCode, expectedStatus)
		}
	}

	// Body validation errors
	if expectedBody != "" && !strings.Contains(responseBody, expectedBody) {
		return ErrorTypeBodyValidation, fmt.Sprintf("Response body doesn't contain expected text: '%s'", expectedBody)
	}

	return ErrorTypeNone, "" // No error
}
