package apperr

import (
	"errors"
	"net/http"
	"testing"
)

func TestClassifyHTTPStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected FailoverReason
	}{
		{"rate limit", http.StatusTooManyRequests, ReasonRateLimit},
		{"unauthorized", http.StatusUnauthorized, ReasonAuthFailed},
		{"forbidden", http.StatusForbidden, ReasonAuthFailed},
		{"service unavailable", http.StatusServiceUnavailable, ReasonOverloaded},
		{"internal error", http.StatusInternalServerError, ReasonServerError},
		{"bad gateway", http.StatusBadGateway, ReasonServerError},
		{"request timeout", http.StatusRequestTimeout, ReasonTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := ClassifyHTTPStatus(tt.status, "test message")
			if fe.Reason != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, fe.Reason)
			}
		})
	}
}

func TestClassifyError_FromMessage(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected FailoverReason
	}{
		{"rate limit", "rate limit exceeded", ReasonRateLimit},
		{"too many requests", "too many requests, try again", ReasonRateLimit},
		{"unauthorized", "unauthorized access", ReasonAuthFailed},
		{"invalid api key", "invalid api key provided", ReasonAuthFailed},
		{"timeout", "request timed out", ReasonTimeout},
		{"deadline exceeded", "context deadline exceeded", ReasonTimeout},
		{"connection refused", "connection refused", ReasonNetworkError},
		{"overloaded", "service overloaded", ReasonOverloaded},
		{"context window", "maximum context length exceeded", ReasonContextWindow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := ClassifyError(errors.New(tt.errMsg))
			if fe.Reason != tt.expected {
				t.Errorf("expected %s, got %s for message: %s", tt.expected, fe.Reason, tt.errMsg)
			}
		})
	}
}

func TestFailoverError_ShouldCooldown(t *testing.T) {
	tests := []struct {
		reason   FailoverReason
		expected bool
	}{
		{ReasonRateLimit, true},
		{ReasonOverloaded, true},
		{ReasonAuthFailed, true},
		{ReasonTimeout, false},
		{ReasonNetworkError, false},
		{ReasonUnknown, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.reason), func(t *testing.T) {
			fe := &FailoverError{Reason: tt.reason}
			if fe.ShouldCooldown() != tt.expected {
				t.Errorf("expected ShouldCooldown=%v for %s", tt.expected, tt.reason)
			}
		})
	}
}

func TestFailoverError_ShouldFailover(t *testing.T) {
	tests := []struct {
		reason   FailoverReason
		expected bool
	}{
		{ReasonRateLimit, true},
		{ReasonOverloaded, true},
		{ReasonAuthFailed, true},
		{ReasonServerError, true},
		{ReasonTimeout, false},
		{ReasonNetworkError, false},
		{ReasonBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.reason), func(t *testing.T) {
			fe := &FailoverError{Reason: tt.reason}
			if fe.ShouldFailover() != tt.expected {
				t.Errorf("expected ShouldFailover=%v for %s", tt.expected, tt.reason)
			}
		})
	}
}

func TestClassifyError_FromAppError(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected FailoverReason
	}{
		{"unauthorized", Unauthorized("test"), ReasonAuthFailed},
		{"forbidden", Forbidden("test"), ReasonAuthFailed},
		{"timeout", Timeout("test"), ReasonTimeout},
		{"unavailable", Unavailable("test"), ReasonOverloaded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fe := ClassifyError(tt.appErr)
			if fe.Reason != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, fe.Reason)
			}
		})
	}
}

func TestIsContextWindowError(t *testing.T) {
	tests := []struct {
		msg      string
		expected bool
	}{
		{"context_length_exceeded: too many tokens", true},
		{"maximum context length exceeded", true},
		{"prompt is too long", true},
		{"regular error message", false},
		{"token limit reached", true},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			result := isContextWindowError(tt.msg)
			if result != tt.expected {
				t.Errorf("expected %v for message: %s", tt.expected, tt.msg)
			}
		})
	}
}

func TestFailoverError_Error(t *testing.T) {
	fe := &FailoverError{
		Reason:  ReasonRateLimit,
		Message: "too many requests",
		Err:     errors.New("underlying error"),
	}

	errStr := fe.Error()
	if errStr == "" {
		t.Error("expected non-empty error string")
	}
}
