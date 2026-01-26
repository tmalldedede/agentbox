// Package task provides provider fallback execution support.
package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tmalldedede/agentbox/internal/agent"
	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/engine"
	"github.com/tmalldedede/agentbox/internal/session"
)

// ProviderKeyManager defines the interface for provider key management (used by fallback)
type ProviderKeyManager interface {
	GetEnvVarsWithKey(providerID string) (map[string]string, error)
	MarkProfileFailed(providerID, profileID string, reason string)
	MarkProfileSuccess(providerID, profileID string)
	GetRotatorStats(providerID string) map[string]interface{}
	GetDecryptedKey(providerID string) (string, error)
}

// FallbackExecutor handles provider fallback logic for task execution
type FallbackExecutor struct {
	agentMgr    *agent.Manager
	providerMgr ProviderKeyManager
	sessionMgr  *session.Manager

	// Metrics (protected by mutex for concurrent access)
	mu              sync.Mutex
	totalAttempts   int
	successAttempts int
	fallbackCount   int
}

// NewFallbackExecutor creates a new fallback executor
func NewFallbackExecutor(agentMgr *agent.Manager, providerMgr ProviderKeyManager, sessionMgr *session.Manager) *FallbackExecutor {
	return &FallbackExecutor{
		agentMgr:    agentMgr,
		providerMgr: providerMgr,
		sessionMgr:  sessionMgr,
	}
}

// FallbackResult contains the result of a fallback execution
type FallbackResult struct {
	// The provider that succeeded
	ProviderID string
	// The session created
	Session *session.Session
	// Execution result
	ExecResponse *session.ExecResponse
	// Total attempts made
	Attempts int
	// Whether fallback was used
	UsedFallback bool
	// Errors from each provider
	ProviderErrors map[string]error
	// Thread ID for multi-turn conversations
	ThreadID string
}

// ExecuteWithFallback tries to execute with the primary provider, falling back to alternatives on failure
func (e *FallbackExecutor) ExecuteWithFallback(ctx context.Context, task *Task, ag *agent.Agent) (*FallbackResult, error) {
	result := &FallbackResult{
		ProviderErrors: make(map[string]error),
	}

	// Build provider list: primary + fallbacks
	providers := []string{ag.ProviderID}
	if ag.FallbackEnabled && len(ag.FallbackProviderIDs) > 0 {
		providers = append(providers, ag.FallbackProviderIDs...)
	}

	var lastErr error

	for i, providerID := range providers {
		result.Attempts++
		e.incrementAttempts()

		if i > 0 {
			result.UsedFallback = true
			e.incrementFallback()
			log.Info("falling back to alternative provider",
				"task_id", task.ID,
				"provider", providerID,
				"attempt", i+1,
				"total_providers", len(providers))
		}

		// Try to execute with this provider
		sess, execResp, err := e.tryExecuteWithProvider(ctx, task, ag, providerID)
		if err == nil {
			// Success!
			result.ProviderID = providerID
			result.Session = sess
			result.ExecResponse = execResp
			if execResp != nil {
				result.ThreadID = execResp.ThreadID
			}
			e.incrementSuccess()

			// Mark provider key as successful
			if e.providerMgr != nil {
				e.providerMgr.MarkProfileSuccess(providerID, "")
			}

			log.Info("execution succeeded",
				"task_id", task.ID,
				"provider", providerID,
				"used_fallback", result.UsedFallback)

			return result, nil
		}

		// Record the error
		result.ProviderErrors[providerID] = err
		lastErr = err

		// Classify the error to determine if we should try the next provider
		engineErr := e.classifyError(err, providerID, ag.Model, ag.Adapter)

		// Mark provider key as failed
		if e.providerMgr != nil {
			e.providerMgr.MarkProfileFailed(providerID, "", string(engineErr.GetReason()))
		}

		log.Warn("provider execution failed",
			"task_id", task.ID,
			"provider", providerID,
			"error", err,
			"reason", engineErr.GetReason(),
			"should_switch", engineErr.ShouldSwitchProvider())

		// Check if we should try the next provider
		if !engineErr.ShouldSwitchProvider() {
			// Non-recoverable error (e.g., bad request, context window exceeded)
			return nil, fmt.Errorf("execution failed (non-recoverable): %w", err)
		}

		// Wait a bit before trying the next provider
		if i < len(providers)-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
		}
	}

	// All providers failed
	return nil, &ProviderFallbackError{
		Attempts:       result.Attempts,
		ProviderErrors: result.ProviderErrors,
		LastError:      lastErr,
	}
}

// tryExecuteWithProvider attempts to execute a task with a specific provider
func (e *FallbackExecutor) tryExecuteWithProvider(ctx context.Context, task *Task, ag *agent.Agent, providerID string) (*session.Session, *session.ExecResponse, error) {
	// Determine workspace
	workspace := ag.Workspace
	if workspace == "" {
		workspace = fmt.Sprintf("agent-%s-%s", ag.ID, task.ID)
	}

	// Create session with this provider
	// We need to inject the provider's env vars
	createReq := &session.CreateRequest{
		AgentID:   ag.ID,
		Workspace: workspace,
	}

	// Get provider env vars
	if e.providerMgr != nil {
		envVars, err := e.providerMgr.GetEnvVarsWithKey(providerID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get env vars for provider %s: %w", providerID, err)
		}
		createReq.Env = envVars
	}

	// Create session
	sess, err := e.sessionMgr.Create(ctx, createReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Start session
	if err := e.sessionMgr.Start(ctx, sess.ID); err != nil {
		// Cleanup on failure
		e.sessionMgr.Stop(ctx, sess.ID)
		return nil, nil, fmt.Errorf("failed to start session: %w", err)
	}

	// Execute the task
	timeout := task.Timeout
	if timeout == 0 {
		timeout = 1800 // Default 30 minutes
	}

	execResp, err := e.sessionMgr.Exec(ctx, sess.ID, &session.ExecRequest{
		Prompt:  task.Prompt,
		Timeout: timeout,
	})
	if err != nil {
		// Cleanup on failure
		e.sessionMgr.Stop(ctx, sess.ID)
		return nil, nil, err
	}

	return sess, execResp, nil
}

// classifyError converts an error to an EngineError for decision making
func (e *FallbackExecutor) classifyError(err error, providerID, model, adapter string) *engine.EngineError {
	// Check if it's already an EngineError
	var engineErr *engine.EngineError
	if errors.As(err, &engineErr) {
		return engineErr
	}

	// Create new EngineError
	return engine.NewEngineError(err, providerID, model, adapter)
}

// Metrics helpers (thread-safe)
func (e *FallbackExecutor) incrementAttempts() {
	e.mu.Lock()
	e.totalAttempts++
	e.mu.Unlock()
}

func (e *FallbackExecutor) incrementSuccess() {
	e.mu.Lock()
	e.successAttempts++
	e.mu.Unlock()
}

func (e *FallbackExecutor) incrementFallback() {
	e.mu.Lock()
	e.fallbackCount++
	e.mu.Unlock()
}

// GetStats returns executor statistics
func (e *FallbackExecutor) GetStats() map[string]interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()

	return map[string]interface{}{
		"total_attempts":   e.totalAttempts,
		"success_attempts": e.successAttempts,
		"fallback_count":   e.fallbackCount,
	}
}

// AllProvidersFailed is returned when all providers in the fallback chain fail
var ErrAllProvidersFailed = errors.New("all providers failed")

// ProviderFallbackError contains information about a fallback chain failure
type ProviderFallbackError struct {
	Attempts       int
	ProviderErrors map[string]error
	LastError      error
}

func (e *ProviderFallbackError) Error() string {
	return fmt.Sprintf("all %d providers failed: %v", e.Attempts, e.LastError)
}

func (e *ProviderFallbackError) Unwrap() error {
	return e.LastError
}

// CheckProviderHealth checks if a provider is healthy (not in cooldown)
func (e *FallbackExecutor) CheckProviderHealth(providerID string) bool {
	if e.providerMgr == nil {
		return true
	}

	// Check if provider has any active profile
	stats := e.providerMgr.GetRotatorStats(providerID)
	if stats == nil {
		// No rotator configured, check single key mode
		_, err := e.providerMgr.GetDecryptedKey(providerID)
		return err == nil
	}

	// Check if all profiles are in cooldown
	allInCooldown, ok := stats["all_in_cooldown"].(bool)
	if ok && allInCooldown {
		return false
	}

	return true
}

// GetHealthyProviders returns a list of healthy providers from a fallback chain
func (e *FallbackExecutor) GetHealthyProviders(ag *agent.Agent) []string {
	providers := []string{ag.ProviderID}
	if ag.FallbackEnabled && len(ag.FallbackProviderIDs) > 0 {
		providers = append(providers, ag.FallbackProviderIDs...)
	}

	healthy := make([]string, 0, len(providers))
	for _, p := range providers {
		if e.CheckProviderHealth(p) {
			healthy = append(healthy, p)
		}
	}

	return healthy
}

// ProviderHealthStatus represents the health status of providers in a fallback chain
type ProviderHealthStatus struct {
	PrimaryProvider string           `json:"primary_provider"`
	FallbackEnabled bool             `json:"fallback_enabled"`
	Providers       []ProviderStatus `json:"providers"`
	HealthyCount    int              `json:"healthy_count"`
	TotalCount      int              `json:"total_count"`
}

type ProviderStatus struct {
	ID        string `json:"id"`
	Healthy   bool   `json:"healthy"`
	IsPrimary bool   `json:"is_primary"`
	Reason    string `json:"reason,omitempty"`
}

// GetProviderHealthStatus returns detailed health status for an agent's provider chain
func (e *FallbackExecutor) GetProviderHealthStatus(ag *agent.Agent) *ProviderHealthStatus {
	status := &ProviderHealthStatus{
		PrimaryProvider: ag.ProviderID,
		FallbackEnabled: ag.FallbackEnabled,
	}

	providers := []string{ag.ProviderID}
	if ag.FallbackEnabled && len(ag.FallbackProviderIDs) > 0 {
		providers = append(providers, ag.FallbackProviderIDs...)
	}

	status.TotalCount = len(providers)

	for i, providerID := range providers {
		ps := ProviderStatus{
			ID:        providerID,
			IsPrimary: i == 0,
		}

		if e.CheckProviderHealth(providerID) {
			ps.Healthy = true
			status.HealthyCount++
		} else {
			ps.Healthy = false
			// Try to get reason
			if e.providerMgr != nil {
				stats := e.providerMgr.GetRotatorStats(providerID)
				if stats != nil {
					if nextAvail, ok := stats["next_available_in"].(string); ok {
						ps.Reason = fmt.Sprintf("in cooldown, available in %s", nextAvail)
					}
				}
			}
			if ps.Reason == "" {
				ps.Reason = "provider unavailable"
			}
		}

		status.Providers = append(status.Providers, ps)
	}

	return status
}

// ShouldUseFallback determines if fallback should be used based on error type
func ShouldUseFallback(err error) bool {
	fe := apperr.ClassifyError(err)
	if fe == nil {
		return false
	}
	return fe.ShouldFailover()
}
