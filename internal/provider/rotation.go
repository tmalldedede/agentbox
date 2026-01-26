// Package provider defines API Provider configurations for AgentBox.
package provider

import (
	"sync"
	"time"

	"github.com/tmalldedede/agentbox/internal/apperr"
)

// AuthProfile represents a single API key credential with rotation support
type AuthProfile struct {
	ID            string    `json:"id"`             // Unique profile ID
	ProviderID    string    `json:"provider_id"`    // Associated provider
	EncryptedKey  string    `json:"encrypted_key"`  // Encrypted API key
	KeyMasked     string    `json:"key_masked"`     // Masked display (e.g., "sk-...xxx")
	Priority      int       `json:"priority"`       // Lower = higher priority (0 is highest)
	IsEnabled     bool      `json:"is_enabled"`     // Whether this profile is active
	CooldownUntil time.Time `json:"cooldown_until"` // Cooldown expiry time
	FailCount     int       `json:"fail_count"`     // Consecutive failure count
	SuccessCount  int       `json:"success_count"`  // Total success count
	LastUsedAt    time.Time `json:"last_used_at"`   // Last usage timestamp
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ProfileRotator manages API key rotation for a provider
type ProfileRotator struct {
	profiles         []*AuthProfile
	mu               sync.RWMutex
	defaultCooldown  time.Duration // Default cooldown duration after failure
	maxFailCount     int           // Max failures before extended cooldown
	extendedCooldown time.Duration // Extended cooldown after max failures
}

// RotatorConfig configuration for ProfileRotator
type RotatorConfig struct {
	DefaultCooldown  time.Duration // Default: 60s
	MaxFailCount     int           // Default: 3
	ExtendedCooldown time.Duration // Default: 5min
}

// DefaultRotatorConfig returns default configuration
func DefaultRotatorConfig() *RotatorConfig {
	return &RotatorConfig{
		DefaultCooldown:  60 * time.Second,
		MaxFailCount:     3,
		ExtendedCooldown: 5 * time.Minute,
	}
}

// NewProfileRotator creates a new profile rotator
func NewProfileRotator(profiles []*AuthProfile, cfg *RotatorConfig) *ProfileRotator {
	if cfg == nil {
		cfg = DefaultRotatorConfig()
	}
	return &ProfileRotator{
		profiles:         profiles,
		defaultCooldown:  cfg.DefaultCooldown,
		maxFailCount:     cfg.MaxFailCount,
		extendedCooldown: cfg.ExtendedCooldown,
	}
}

// GetActiveProfile returns the best available profile (not in cooldown)
// Returns nil if no profiles are available
func (r *ProfileRotator) GetActiveProfile() *AuthProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	var best *AuthProfile

	for _, p := range r.profiles {
		if !p.IsEnabled {
			continue
		}
		// Skip profiles in cooldown
		if now.Before(p.CooldownUntil) {
			continue
		}
		// Select by priority (lower = better), then by fail count (lower = better)
		if best == nil {
			best = p
		} else if p.Priority < best.Priority {
			best = p
		} else if p.Priority == best.Priority && p.FailCount < best.FailCount {
			best = p
		}
	}

	return best
}

// MarkFailed records a failure for a profile and applies cooldown
func (r *ProfileRotator) MarkFailed(profileID string, reason apperr.FailoverReason) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.profiles {
		if p.ID == profileID {
			p.FailCount++
			p.UpdatedAt = time.Now()

			// Determine cooldown duration
			cooldown := r.defaultCooldown
			if p.FailCount >= r.maxFailCount {
				cooldown = r.extendedCooldown
			}

			// Rate limit errors get longer cooldown
			if reason == apperr.ReasonRateLimit {
				cooldown = cooldown * 2
			}

			p.CooldownUntil = time.Now().Add(cooldown)
			return
		}
	}
}

// MarkSuccess records a successful use of a profile
func (r *ProfileRotator) MarkSuccess(profileID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.profiles {
		if p.ID == profileID {
			p.FailCount = 0 // Reset fail count on success
			p.SuccessCount++
			p.LastUsedAt = time.Now()
			p.UpdatedAt = time.Now()
			p.CooldownUntil = time.Time{} // Clear cooldown
			return
		}
	}
}

// ResetCooldown clears cooldown for a profile
func (r *ProfileRotator) ResetCooldown(profileID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.profiles {
		if p.ID == profileID {
			p.CooldownUntil = time.Time{}
			p.FailCount = 0
			p.UpdatedAt = time.Now()
			return
		}
	}
}

// GetAllProfiles returns a copy of all profiles
func (r *ProfileRotator) GetAllProfiles() []*AuthProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*AuthProfile, len(r.profiles))
	for i, p := range r.profiles {
		cp := *p
		result[i] = &cp
	}
	return result
}

// AddProfile adds a new profile
func (r *ProfileRotator) AddProfile(p *AuthProfile) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	r.profiles = append(r.profiles, p)
}

// RemoveProfile removes a profile by ID
func (r *ProfileRotator) RemoveProfile(profileID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.profiles {
		if p.ID == profileID {
			r.profiles = append(r.profiles[:i], r.profiles[i+1:]...)
			return true
		}
	}
	return false
}

// SetEnabled enables or disables a profile
func (r *ProfileRotator) SetEnabled(profileID string, enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.profiles {
		if p.ID == profileID {
			p.IsEnabled = enabled
			p.UpdatedAt = time.Now()
			return
		}
	}
}

// ActiveCount returns the number of available (not in cooldown) profiles
func (r *ProfileRotator) ActiveCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	count := 0
	for _, p := range r.profiles {
		if p.IsEnabled && now.After(p.CooldownUntil) {
			count++
		}
	}
	return count
}

// AllInCooldown returns true if all profiles are in cooldown
func (r *ProfileRotator) AllInCooldown() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.profiles) == 0 {
		return false
	}

	now := time.Now()
	for _, p := range r.profiles {
		if p.IsEnabled && now.After(p.CooldownUntil) {
			return false
		}
	}
	return true
}

// NextAvailableIn returns duration until next profile becomes available
// Returns 0 if a profile is currently available
func (r *ProfileRotator) NextAvailableIn() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	var minWait time.Duration = -1

	for _, p := range r.profiles {
		if !p.IsEnabled {
			continue
		}
		if now.After(p.CooldownUntil) {
			return 0 // Already available
		}
		wait := p.CooldownUntil.Sub(now)
		if minWait < 0 || wait < minWait {
			minWait = wait
		}
	}

	if minWait < 0 {
		return 0
	}
	return minWait
}
