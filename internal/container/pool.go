// Package container provides container management functionality
package container

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"sort"
	"sync"
	"time"
)

// PooledContainer represents a container in the pool
type PooledContainer struct {
	ContainerID string            `json:"container_id"`
	ConfigHash  string            `json:"config_hash"`
	Image       string            `json:"image"`
	Labels      map[string]string `json:"labels"`
	CreatedAt   time.Time         `json:"created_at"`
	LastUsedAt  time.Time         `json:"last_used_at"`
	UseCount    int               `json:"use_count"`
}

// ContainerPool manages a pool of reusable containers
type ContainerPool struct {
	idle       map[string][]*PooledContainer // configHash -> containers
	mu         sync.RWMutex
	maxIdle    int           // Max idle containers per config
	maxTotal   int           // Max total idle containers
	idleTime   time.Duration // Time before container is removed from pool
	mgr        Manager       // Container manager for operations
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

// PoolConfig configuration for ContainerPool
type PoolConfig struct {
	MaxIdle    int           // Max idle containers per config (default: 3)
	MaxTotal   int           // Max total idle containers (default: 10)
	IdleTime   time.Duration // Idle time before removal (default: 10min)
	CleanupInterval time.Duration // Cleanup interval (default: 1min)
}

// DefaultPoolConfig returns default configuration
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxIdle:         3,
		MaxTotal:        10,
		IdleTime:        10 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
}

// NewContainerPool creates a new container pool
func NewContainerPool(mgr Manager, cfg *PoolConfig) *ContainerPool {
	if cfg == nil {
		cfg = DefaultPoolConfig()
	}
	if cfg.MaxIdle <= 0 {
		cfg.MaxIdle = 3
	}
	if cfg.MaxTotal <= 0 {
		cfg.MaxTotal = 10
	}
	if cfg.IdleTime <= 0 {
		cfg.IdleTime = 10 * time.Minute
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = 1 * time.Minute
	}

	pool := &ContainerPool{
		idle:     make(map[string][]*PooledContainer),
		maxIdle:  cfg.MaxIdle,
		maxTotal: cfg.MaxTotal,
		idleTime: cfg.IdleTime,
		mgr:      mgr,
		stopCh:   make(chan struct{}),
	}

	// Start cleanup goroutine
	pool.wg.Add(1)
	go pool.cleanupLoop(cfg.CleanupInterval)

	return pool
}

// Acquire gets a container from the pool matching the config hash
// Returns nil if no matching container is available
func (p *ContainerPool) Acquire(configHash string) *PooledContainer {
	p.mu.Lock()
	defer p.mu.Unlock()

	containers, ok := p.idle[configHash]
	if !ok || len(containers) == 0 {
		return nil
	}

	// Take the most recently used container (LIFO for better cache locality)
	container := containers[len(containers)-1]
	p.idle[configHash] = containers[:len(containers)-1]

	container.LastUsedAt = time.Now()
	container.UseCount++

	slog.Debug("container acquired from pool",
		"container_id", container.ContainerID[:12],
		"config_hash", configHash[:8],
		"use_count", container.UseCount,
	)

	return container
}

// Release returns a container to the pool
// Returns true if the container was added to the pool, false if it should be removed
func (p *ContainerPool) Release(container *PooledContainer) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if pool is at capacity for this config
	if len(p.idle[container.ConfigHash]) >= p.maxIdle {
		slog.Debug("pool at capacity for config, container not pooled",
			"container_id", container.ContainerID[:12],
			"config_hash", container.ConfigHash[:8],
		)
		return false
	}

	// Check total capacity
	total := 0
	for _, containers := range p.idle {
		total += len(containers)
	}
	if total >= p.maxTotal {
		slog.Debug("pool at total capacity, container not pooled",
			"container_id", container.ContainerID[:12],
			"total", total,
		)
		return false
	}

	container.LastUsedAt = time.Now()
	p.idle[container.ConfigHash] = append(p.idle[container.ConfigHash], container)

	slog.Debug("container released to pool",
		"container_id", container.ContainerID[:12],
		"config_hash", container.ConfigHash[:8],
		"pool_size", len(p.idle[container.ConfigHash]),
	)

	return true
}

// Remove removes a specific container from the pool (if present)
func (p *ContainerPool) Remove(containerID string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	for configHash, containers := range p.idle {
		for i, c := range containers {
			if c.ContainerID == containerID {
				p.idle[configHash] = append(containers[:i], containers[i+1:]...)
				return true
			}
		}
	}
	return false
}

// cleanupLoop periodically removes expired containers
func (p *ContainerPool) cleanupLoop(interval time.Duration) {
	defer p.wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.cleanup()
		}
	}
}

// cleanup removes expired containers from the pool
func (p *ContainerPool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	expired := []*PooledContainer{}

	for configHash, containers := range p.idle {
		active := []*PooledContainer{}
		for _, c := range containers {
			if now.Sub(c.LastUsedAt) > p.idleTime {
				expired = append(expired, c)
			} else {
				active = append(active, c)
			}
		}
		if len(active) > 0 {
			p.idle[configHash] = active
		} else {
			delete(p.idle, configHash)
		}
	}

	// Stop and remove expired containers
	if len(expired) > 0 {
		slog.Info("cleaning up expired pooled containers", "count", len(expired))
		go p.stopContainers(expired)
	}
}

// stopContainers stops and removes containers (called outside of lock)
func (p *ContainerPool) stopContainers(containers []*PooledContainer) {
	ctx := context.Background()
	for _, c := range containers {
		if err := p.mgr.Stop(ctx, c.ContainerID); err != nil {
			slog.Debug("failed to stop pooled container", "container_id", c.ContainerID[:12], "error", err)
		}
		if err := p.mgr.Remove(ctx, c.ContainerID); err != nil {
			slog.Debug("failed to remove pooled container", "container_id", c.ContainerID[:12], "error", err)
		}
	}
}

// Stop stops the pool and cleans up all containers
func (p *ContainerPool) Stop() {
	close(p.stopCh)
	p.wg.Wait()

	// Drain all containers
	p.mu.Lock()
	allContainers := []*PooledContainer{}
	for _, containers := range p.idle {
		allContainers = append(allContainers, containers...)
	}
	p.idle = make(map[string][]*PooledContainer)
	p.mu.Unlock()

	if len(allContainers) > 0 {
		slog.Info("stopping all pooled containers", "count", len(allContainers))
		p.stopContainers(allContainers)
	}
}

// Stats returns pool statistics
type PoolStats struct {
	TotalIdle      int            `json:"total_idle"`
	ConfigCounts   map[string]int `json:"config_counts"`
	OldestIdleTime time.Duration  `json:"oldest_idle_time"`
}

func (p *ContainerPool) Stats() *PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := &PoolStats{
		ConfigCounts: make(map[string]int),
	}

	now := time.Now()
	var oldestTime time.Duration

	for configHash, containers := range p.idle {
		stats.TotalIdle += len(containers)
		stats.ConfigCounts[configHash[:8]] = len(containers)

		for _, c := range containers {
			idle := now.Sub(c.LastUsedAt)
			if idle > oldestTime {
				oldestTime = idle
			}
		}
	}

	stats.OldestIdleTime = oldestTime
	return stats
}

// GetAll returns all pooled containers (for debugging)
func (p *ContainerPool) GetAll() []*PooledContainer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var all []*PooledContainer
	for _, containers := range p.idle {
		all = append(all, containers...)
	}
	return all
}

// ComputeConfigHash generates a hash from container configuration
// This is used to match containers with the same configuration
func ComputeConfigHash(cfg *CreateConfig) string {
	// Create a normalized config for hashing
	normalized := struct {
		Image       string
		Env         []string // Sorted env vars (keys only, values may differ)
		Mounts      []string // Sorted mount targets
		NetworkMode string
		Privileged  bool
		CPULimit    float64
		MemoryLimit int64
	}{
		Image:       cfg.Image,
		Env:         extractEnvKeys(cfg.Env),
		Mounts:      extractMountTargets(cfg.Mounts),
		NetworkMode: cfg.NetworkMode,
		Privileged:  cfg.Privileged,
		CPULimit:    cfg.Resources.CPULimit,
		MemoryLimit: cfg.Resources.MemoryLimit,
	}

	data, _ := json.Marshal(normalized)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// extractEnvKeys extracts and sorts environment variable keys
func extractEnvKeys(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for k := range env {
		// Exclude sensitive keys from hash
		if k == "API_KEY" || k == "OPENAI_API_KEY" || k == "ANTHROPIC_API_KEY" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// extractMountTargets extracts and sorts mount target paths
func extractMountTargets(mounts []Mount) []string {
	targets := make([]string, 0, len(mounts))
	for _, m := range mounts {
		targets = append(targets, m.Target)
	}
	sort.Strings(targets)
	return targets
}

// PrepareForReuse prepares a pooled container for reuse
// This includes cleaning up any state from previous use
func (p *ContainerPool) PrepareForReuse(ctx context.Context, container *PooledContainer, newEnv map[string]string) error {
	// Run cleanup to remove any suspended processes
	cfg := DefaultProcessCleanupConfig()
	_, err := CleanupSuspendedProcesses(ctx, p.mgr, container.ContainerID, cfg)
	if err != nil {
		slog.Warn("failed to cleanup suspended processes before reuse",
			"container_id", container.ContainerID[:12],
			"error", err,
		)
		// Don't fail, just log warning
	}

	// TODO: Optionally clear workspace or reset state

	return nil
}

// PoolableCheck checks if a container is suitable for pooling
func PoolableCheck(cfg *CreateConfig) bool {
	// Don't pool containers with specific labels
	if cfg.Labels != nil {
		if _, ok := cfg.Labels["agentbox.no_pool"]; ok {
			return false
		}
	}

	// Only pool containers with standard images
	// Custom or temporary images shouldn't be pooled
	if cfg.Image == "" {
		return false
	}

	return true
}
