package main

import (
	"flag"
	"os"
	"testing"
	"time"
)

// Helper to reset flags between tests
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
}

func TestParseAndValidateFlags_Defaults(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd"}

	cfg, requests, concurrency, outputJSON, err := parseAndValidateFlags()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cfg.URL != "http://localhost:8080" || requests != 100 || concurrency != 10 || cfg.ExpectedStatus != 200 || cfg.ExpectedBody != "" || cfg.Timeout != 5*time.Second || outputJSON != false {
		t.Errorf("Default flag values not parsed correctly: %+v, %d, %d, %v", cfg, requests, concurrency, outputJSON)
	}
}

func TestParseAndValidateFlags_CustomValues(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "-url=http://test", "-requests=42", "-concurrency=7", "-status=404", "-body=hello", "-timeout=2", "-json=true"}

	cfg, requests, concurrency, outputJSON, err := parseAndValidateFlags()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cfg.URL != "http://test" || requests != 42 || concurrency != 7 || cfg.ExpectedStatus != 404 || cfg.ExpectedBody != "hello" || cfg.Timeout != 2*time.Second || outputJSON != true {
		t.Errorf("Custom flag values not parsed correctly: %+v, %d, %d, %v", cfg, requests, concurrency, outputJSON)
	}
}

func TestParseAndValidateFlags_NegativeConcurrency(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "-concurrency=-1"}

	_, _, _, _, err := parseAndValidateFlags()
	if err == nil || err.Error() != "concurrency must be >= 1, got -1" {
		t.Errorf("Expected error for negative concurrency, got: %v", err)
	}
}

func TestParseAndValidateFlags_ZeroRequests(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "-requests=0"}

	_, _, _, _, err := parseAndValidateFlags()
	if err == nil || err.Error() != "requests must be >= 1, got 0" {
		t.Errorf("Expected error for zero requests, got: %v", err)
	}
}

func TestParseAndValidateFlags_TimeoutNegative(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "-timeout=-5"}

	_, _, _, _, err := parseAndValidateFlags()
	if err == nil || err.Error() != "timeout must be >= 1, got -5" {
		t.Errorf("Expected error for negative timeout, got: %v", err)
	}
}

func TestParseAndValidateFlags_ConfigConstruction(t *testing.T) {
	resetFlags()
	os.Args = []string{"cmd", "-url=http://test", "-status=201", "-body=abc", "-timeout=3"}

	cfg, _, _, _, err := parseAndValidateFlags()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if cfg.URL != "http://test" || cfg.ExpectedStatus != 201 || cfg.ExpectedBody != "abc" || cfg.Timeout != 3*time.Second {
		t.Errorf("Config not constructed correctly: %+v", cfg)
	}
}
