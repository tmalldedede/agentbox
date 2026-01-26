// Package apperr provides application-level error handling
package apperr

import (
	"fmt"
	"net/http"
	"strings"
)

// FailoverReason indicates why a failover should occur
type FailoverReason string

const (
	// ReasonRateLimit - 429 Too Many Requests
	ReasonRateLimit FailoverReason = "rate_limit"

	// ReasonOverloaded - 503/529 Service Overloaded
	ReasonOverloaded FailoverReason = "overloaded"

	// ReasonTimeout - Request timeout
	ReasonTimeout FailoverReason = "timeout"

	// ReasonAuthFailed - 401/403 Authentication/Authorization failed
	ReasonAuthFailed FailoverReason = "auth_failed"

	// ReasonBadRequest - 400 Bad Request
	ReasonBadRequest FailoverReason = "bad_request"

	// ReasonContextWindow - Token limit exceeded
	ReasonContextWindow FailoverReason = "context_window"

	// ReasonNetworkError - Network/connection error
	ReasonNetworkError FailoverReason = "network_error"

	// ReasonServerError - 500/502/504 Server errors
	ReasonServerError FailoverReason = "server_error"

	// ReasonUnknown - Unknown error
	ReasonUnknown FailoverReason = "unknown"
)

// FailoverError represents an error that may trigger failover
type FailoverError struct {
	Reason    FailoverReason // Classification of the error
	Message   string         // Human-readable error message
	Err       error          // Original error
	Retryable bool           // Whether this error is retryable
	HTTPCode  int            // HTTP status code if applicable
}

// Error implements the error interface
func (e *FailoverError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Reason, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Reason, e.Message)
}

// Unwrap implements errors.Unwrap
func (e *FailoverError) Unwrap() error {
	return e.Err
}

// ShouldCooldown returns true if this error should trigger API key cooldown
func (e *FailoverError) ShouldCooldown() bool {
	switch e.Reason {
	case ReasonRateLimit, ReasonOverloaded, ReasonAuthFailed:
		return true
	default:
		return false
	}
}

// ShouldFailover returns true if this error should trigger failover to another key
func (e *FailoverError) ShouldFailover() bool {
	switch e.Reason {
	case ReasonRateLimit, ReasonOverloaded, ReasonAuthFailed, ReasonServerError:
		return true
	default:
		return false
	}
}

// ClassifyError classifies an error into a FailoverError
func ClassifyError(err error) *FailoverError {
	if err == nil {
		return nil
	}

	// Check if already a FailoverError
	if fe, ok := err.(*FailoverError); ok {
		return fe
	}

	// Check if it's an AppError
	if ae, ok := err.(*AppError); ok {
		return classifyFromAppError(ae)
	}

	// Classify from error message
	return classifyFromMessage(err)
}

// ClassifyHTTPStatus classifies an HTTP status code into a FailoverError
func ClassifyHTTPStatus(statusCode int, message string) *FailoverError {
	switch statusCode {
	case http.StatusTooManyRequests: // 429
		return &FailoverError{
			Reason:    ReasonRateLimit,
			Message:   message,
			Retryable: true,
			HTTPCode:  statusCode,
		}

	case http.StatusUnauthorized: // 401
		return &FailoverError{
			Reason:    ReasonAuthFailed,
			Message:   message,
			Retryable: false,
			HTTPCode:  statusCode,
		}

	case http.StatusForbidden: // 403
		return &FailoverError{
			Reason:    ReasonAuthFailed,
			Message:   message,
			Retryable: false,
			HTTPCode:  statusCode,
		}

	case http.StatusBadRequest: // 400
		// Check for context window errors
		if isContextWindowError(message) {
			return &FailoverError{
				Reason:    ReasonContextWindow,
				Message:   message,
				Retryable: false,
				HTTPCode:  statusCode,
			}
		}
		return &FailoverError{
			Reason:    ReasonBadRequest,
			Message:   message,
			Retryable: false,
			HTTPCode:  statusCode,
		}

	case http.StatusServiceUnavailable, 529: // 503, 529 (Overloaded)
		return &FailoverError{
			Reason:    ReasonOverloaded,
			Message:   message,
			Retryable: true,
			HTTPCode:  statusCode,
		}

	case http.StatusInternalServerError, // 500
		http.StatusBadGateway,    // 502
		http.StatusGatewayTimeout: // 504
		return &FailoverError{
			Reason:    ReasonServerError,
			Message:   message,
			Retryable: true,
			HTTPCode:  statusCode,
		}

	case http.StatusRequestTimeout: // 408
		return &FailoverError{
			Reason:    ReasonTimeout,
			Message:   message,
			Retryable: true,
			HTTPCode:  statusCode,
		}

	default:
		return &FailoverError{
			Reason:    ReasonUnknown,
			Message:   message,
			Retryable: false,
			HTTPCode:  statusCode,
		}
	}
}

// classifyFromAppError classifies from an AppError
func classifyFromAppError(ae *AppError) *FailoverError {
	switch ae.Type {
	case TypeUnauthorized:
		return &FailoverError{
			Reason:    ReasonAuthFailed,
			Message:   ae.Message,
			Err:       ae,
			Retryable: false,
			HTTPCode:  ae.Code,
		}
	case TypeForbidden:
		return &FailoverError{
			Reason:    ReasonAuthFailed,
			Message:   ae.Message,
			Err:       ae,
			Retryable: false,
			HTTPCode:  ae.Code,
		}
	case TypeTimeout:
		return &FailoverError{
			Reason:    ReasonTimeout,
			Message:   ae.Message,
			Err:       ae,
			Retryable: true,
			HTTPCode:  ae.Code,
		}
	case TypeUnavailable:
		return &FailoverError{
			Reason:    ReasonOverloaded,
			Message:   ae.Message,
			Err:       ae,
			Retryable: true,
			HTTPCode:  ae.Code,
		}
	default:
		return &FailoverError{
			Reason:    ReasonUnknown,
			Message:   ae.Message,
			Err:       ae,
			Retryable: false,
			HTTPCode:  ae.Code,
		}
	}
}

// classifyFromMessage classifies based on error message patterns
func classifyFromMessage(err error) *FailoverError {
	msg := strings.ToLower(err.Error())

	// Rate limit patterns
	rateLimitPatterns := []string{
		"rate limit",
		"too many requests",
		"quota exceeded",
		"throttl",
		"limit exceeded",
	}
	for _, p := range rateLimitPatterns {
		if strings.Contains(msg, p) {
			return &FailoverError{
				Reason:    ReasonRateLimit,
				Message:   err.Error(),
				Err:       err,
				Retryable: true,
			}
		}
	}

	// Auth failed patterns
	authPatterns := []string{
		"unauthorized",
		"authentication",
		"invalid api key",
		"api key invalid",
		"invalid_api_key",
		"forbidden",
		"access denied",
	}
	for _, p := range authPatterns {
		if strings.Contains(msg, p) {
			return &FailoverError{
				Reason:    ReasonAuthFailed,
				Message:   err.Error(),
				Err:       err,
				Retryable: false,
			}
		}
	}

	// Context window patterns
	if isContextWindowError(msg) {
		return &FailoverError{
			Reason:    ReasonContextWindow,
			Message:   err.Error(),
			Err:       err,
			Retryable: false,
		}
	}

	// Timeout patterns
	timeoutPatterns := []string{
		"timeout",
		"timed out",
		"deadline exceeded",
	}
	for _, p := range timeoutPatterns {
		if strings.Contains(msg, p) {
			return &FailoverError{
				Reason:    ReasonTimeout,
				Message:   err.Error(),
				Err:       err,
				Retryable: true,
			}
		}
	}

	// Network error patterns
	networkPatterns := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"network unreachable",
		"eof",
		"broken pipe",
	}
	for _, p := range networkPatterns {
		if strings.Contains(msg, p) {
			return &FailoverError{
				Reason:    ReasonNetworkError,
				Message:   err.Error(),
				Err:       err,
				Retryable: true,
			}
		}
	}

	// Overloaded patterns
	overloadPatterns := []string{
		"overloaded",
		"service unavailable",
		"capacity",
		"try again later",
	}
	for _, p := range overloadPatterns {
		if strings.Contains(msg, p) {
			return &FailoverError{
				Reason:    ReasonOverloaded,
				Message:   err.Error(),
				Err:       err,
				Retryable: true,
			}
		}
	}

	// Default to unknown
	return &FailoverError{
		Reason:    ReasonUnknown,
		Message:   err.Error(),
		Err:       err,
		Retryable: false,
	}
}

// isContextWindowError checks if the message indicates a context window error
func isContextWindowError(msg string) bool {
	patterns := []string{
		"context_length_exceeded",
		"context window",
		"maximum context length",
		"token limit",
		"max tokens",
		"input too long",
		"prompt is too long",
		"maximum number of tokens",
	}
	lowerMsg := strings.ToLower(msg)
	for _, p := range patterns {
		if strings.Contains(lowerMsg, p) {
			return true
		}
	}
	return false
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(msg string) *FailoverError {
	return &FailoverError{
		Reason:    ReasonRateLimit,
		Message:   msg,
		Retryable: true,
		HTTPCode:  http.StatusTooManyRequests,
	}
}

// NewAuthFailedError creates an auth failed error
func NewAuthFailedError(msg string) *FailoverError {
	return &FailoverError{
		Reason:    ReasonAuthFailed,
		Message:   msg,
		Retryable: false,
		HTTPCode:  http.StatusUnauthorized,
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(msg string) *FailoverError {
	return &FailoverError{
		Reason:    ReasonTimeout,
		Message:   msg,
		Retryable: true,
		HTTPCode:  http.StatusRequestTimeout,
	}
}

// NewContextWindowError creates a context window error
func NewContextWindowError(msg string) *FailoverError {
	return &FailoverError{
		Reason:    ReasonContextWindow,
		Message:   msg,
		Retryable: false,
		HTTPCode:  http.StatusBadRequest,
	}
}
