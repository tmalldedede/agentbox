// Package engine provides error types for AI engine adapters.
package engine

import (
	"fmt"

	"github.com/tmalldedede/agentbox/internal/apperr"
)

// EngineError represents an error from an AI engine execution.
// It wraps FailoverError with additional context about the provider and model.
type EngineError struct {
	*apperr.FailoverError

	// Provider and model context
	ProviderID string `json:"provider_id,omitempty"`
	Model      string `json:"model,omitempty"`
	Adapter    string `json:"adapter,omitempty"`

	// Execution context
	SessionID   string `json:"session_id,omitempty"`
	ExecutionID string `json:"execution_id,omitempty"`
}

// Error implements the error interface
func (e *EngineError) Error() string {
	base := ""
	if e.FailoverError != nil {
		base = e.FailoverError.Error()
	}

	if e.ProviderID != "" {
		return fmt.Sprintf("[%s] %s", e.ProviderID, base)
	}
	return base
}

// Unwrap implements errors.Unwrap
func (e *EngineError) Unwrap() error {
	if e.FailoverError != nil {
		return e.FailoverError
	}
	return nil
}

// NewEngineError creates an EngineError from an error with context
func NewEngineError(err error, providerID, model, adapter string) *EngineError {
	fe := apperr.ClassifyError(err)
	return &EngineError{
		FailoverError: fe,
		ProviderID:    providerID,
		Model:         model,
		Adapter:       adapter,
	}
}

// NewEngineErrorFromHTTP creates an EngineError from HTTP status code
func NewEngineErrorFromHTTP(statusCode int, message, providerID, model, adapter string) *EngineError {
	fe := apperr.ClassifyHTTPStatus(statusCode, message)
	return &EngineError{
		FailoverError: fe,
		ProviderID:    providerID,
		Model:         model,
		Adapter:       adapter,
	}
}

// WithSession adds session context to the error
func (e *EngineError) WithSession(sessionID, executionID string) *EngineError {
	e.SessionID = sessionID
	e.ExecutionID = executionID
	return e
}

// ShouldSwitchProvider returns true if this error should trigger a switch to fallback provider
func (e *EngineError) ShouldSwitchProvider() bool {
	if e.FailoverError == nil {
		return false
	}

	// Switch provider on these conditions:
	// 1. Auth failed (key invalid) - need different key/provider
	// 2. Rate limit - provider is throttling
	// 3. Overloaded - provider capacity issue
	switch e.FailoverError.Reason {
	case apperr.ReasonAuthFailed,
		apperr.ReasonRateLimit,
		apperr.ReasonOverloaded:
		return true
	default:
		return false
	}
}

// ShouldRetry returns true if this error is retryable with the same provider
func (e *EngineError) ShouldRetry() bool {
	if e.FailoverError == nil {
		return false
	}
	return e.FailoverError.Retryable
}

// ShouldCooldownProvider returns true if the provider should enter cooldown
func (e *EngineError) ShouldCooldownProvider() bool {
	if e.FailoverError == nil {
		return false
	}
	return e.FailoverError.ShouldCooldown()
}

// IsAuthError returns true if this is an authentication/authorization error
func (e *EngineError) IsAuthError() bool {
	if e.FailoverError == nil {
		return false
	}
	return e.FailoverError.Reason == apperr.ReasonAuthFailed
}

// IsRateLimitError returns true if this is a rate limit error
func (e *EngineError) IsRateLimitError() bool {
	if e.FailoverError == nil {
		return false
	}
	return e.FailoverError.Reason == apperr.ReasonRateLimit
}

// IsContextWindowError returns true if this is a context window exceeded error
func (e *EngineError) IsContextWindowError() bool {
	if e.FailoverError == nil {
		return false
	}
	return e.FailoverError.Reason == apperr.ReasonContextWindow
}

// IsTimeoutError returns true if this is a timeout error
func (e *EngineError) IsTimeoutError() bool {
	if e.FailoverError == nil {
		return false
	}
	return e.FailoverError.Reason == apperr.ReasonTimeout
}

// GetReason returns the failover reason
func (e *EngineError) GetReason() apperr.FailoverReason {
	if e.FailoverError == nil {
		return apperr.ReasonUnknown
	}
	return e.FailoverError.Reason
}

// GetHTTPCode returns the HTTP status code if available
func (e *EngineError) GetHTTPCode() int {
	if e.FailoverError == nil {
		return 0
	}
	return e.FailoverError.HTTPCode
}
