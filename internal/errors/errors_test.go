package errors

import (
	"context"
	"errors"
	"net"
	"net/url"
	"testing"
)

func TestCategorizeError_Timeout(t *testing.T) {
	err := context.DeadlineExceeded
	etype, msg := CategorizeError(err, 200, 200, "", "")
	if etype != ErrorTypeTimeout {
		t.Errorf("Expected Timeout, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_DNS(t *testing.T) {
	dnsErr := &net.DNSError{Err: "no such host"}
	etype, msg := CategorizeError(dnsErr, 200, 200, "", "")
	if etype != ErrorTypeDNS {
		t.Errorf("Expected DNS, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_Connection(t *testing.T) {
	opErr := &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	etype, msg := CategorizeError(opErr, 200, 200, "", "")
	if etype != ErrorTypeConnection {
		t.Errorf("Expected Connection, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_URL(t *testing.T) {
	urlErr := &url.Error{Op: "parse", URL: ":://bad-url", Err: errors.New("invalid URL")}
	etype, msg := CategorizeError(urlErr, 200, 200, "", "")
	if etype != ErrorTypeURL {
		t.Errorf("Expected URL, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_TLS(t *testing.T) {
	tlsErr := errors.New("tls: handshake failure")
	etype, msg := CategorizeError(tlsErr, 200, 200, "", "")
	if etype != ErrorTypeTLS {
		t.Errorf("Expected TLS, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_Network(t *testing.T) {
	netErr := errors.New("some network error")
	etype, msg := CategorizeError(netErr, 200, 200, "", "")
	if etype != ErrorTypeNetwork {
		t.Errorf("Expected Network, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_ServerError(t *testing.T) {
	etype, msg := CategorizeError(nil, 500, 200, "", "")
	if etype != ErrorTypeServerError {
		t.Errorf("Expected Server Error, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_ClientError(t *testing.T) {
	etype, msg := CategorizeError(nil, 404, 200, "", "")
	if etype != ErrorTypeClientError {
		t.Errorf("Expected Client Error, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_Redirect(t *testing.T) {
	etype, msg := CategorizeError(nil, 302, 200, "", "")
	if etype != ErrorTypeRedirect {
		t.Errorf("Expected Redirect, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_HTTPStatus(t *testing.T) {
	etype, msg := CategorizeError(nil, 201, 200, "", "")
	if etype != ErrorTypeHTTPStatus {
		t.Errorf("Expected HTTP Status, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_BodyValidation(t *testing.T) {
	etype, msg := CategorizeError(nil, 200, 200, "expected", "not present")
	if etype != ErrorTypeBodyValidation {
		t.Errorf("Expected Body Validation, got %v", etype)
	}
	if msg == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestCategorizeError_None(t *testing.T) {
	etype, msg := CategorizeError(nil, 200, 200, "", "")
	if etype != ErrorTypeNone {
		t.Errorf("Expected None, got %v", etype)
	}
	if msg != "" {
		t.Errorf("Expected empty error message, got %v", msg)
	}
}
