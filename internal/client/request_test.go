package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"loadtester/internal/config"
)

func TestMakeRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	}))
	defer server.Close()

	cfg := config.RequestConfig{
		URL:            server.URL,
		Timeout:        2 * time.Second,
		ExpectedStatus: http.StatusOK,
		ExpectedBody:   "Hello, World!",
		Concurrency:    1,
	}

	result := MakeRequest(cfg)

	if !result.Success {
		t.Errorf("Expected success, got failure: %v", result.ErrorMessage)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, result.StatusCode)
	}
	if result.ResponseSize != int64(len("Hello, World!")) {
		t.Errorf("Expected response size %d, got %d", len("Hello, World!"), result.ResponseSize)
	}
}

func TestMakeRequest_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := config.RequestConfig{
		URL:            server.URL,
		Timeout:        1 * time.Second,
		ExpectedStatus: http.StatusOK,
		ExpectedBody:   "",
		Concurrency:    1,
	}

	result := MakeRequest(cfg)

	if result.Success {
		t.Errorf("Expected failure due to timeout, got success")
	}
	if result.ErrorType == "" {
		t.Errorf("Expected error type for timeout, got none")
	}
}

func TestMakeRequest_BadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := config.RequestConfig{
		URL:            server.URL,
		Timeout:        2 * time.Second,
		ExpectedStatus: http.StatusOK,
		ExpectedBody:   "",
		Concurrency:    1,
	}

	result := MakeRequest(cfg)

	if result.Success {
		t.Errorf("Expected failure due to bad status, got success")
	}
	if result.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, result.StatusCode)
	}
}

func TestMakeRequest_BadBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Unexpected body"))
	}))
	defer server.Close()

	cfg := config.RequestConfig{
		URL:            server.URL,
		Timeout:        2 * time.Second,
		ExpectedStatus: http.StatusOK,
		ExpectedBody:   "Expected body",
		Concurrency:    1,
	}

	result := MakeRequest(cfg)

	if result.Success {
		t.Errorf("Expected failure due to bad body, got success")
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, result.StatusCode)
	}
}
