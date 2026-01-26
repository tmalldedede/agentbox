// Package container provides container management functionality
package container

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ProcessCleanupConfig configuration for process cleanup
type ProcessCleanupConfig struct {
	CmdPrefixes []string      // Command prefixes to match (e.g., "codex", "claude")
	MaxAge      time.Duration // Maximum age of process before cleanup (0 = no age limit)
	DryRun      bool          // If true, only report without killing
}

// DefaultProcessCleanupConfig returns default configuration
func DefaultProcessCleanupConfig() *ProcessCleanupConfig {
	return &ProcessCleanupConfig{
		CmdPrefixes: []string{"codex", "claude", "opencode"},
		MaxAge:      0,
		DryRun:      false,
	}
}

// SuspendedProcess represents a suspended (T-state) process
type SuspendedProcess struct {
	PID     int
	PPID    int
	Command string
	State   string
	VSZ     int64
	RSS     int64
}

// CleanupSuspendedProcesses cleans up suspended (T-state/stopped) CLI processes in a container
// Returns the number of processes killed and any error
func CleanupSuspendedProcesses(ctx context.Context, mgr Manager, containerID string, cfg *ProcessCleanupConfig) (int, error) {
	if cfg == nil {
		cfg = DefaultProcessCleanupConfig()
	}

	log := slog.Default().With("module", "container.cleanup", "container_id", containerID[:12])

	// Execute ps command to get process list
	// Using ps -axo to get specific fields: PID, PPID, STAT, VSZ, RSS, COMM
	psCmd := []string{"ps", "-axo", "pid,ppid,stat,vsz,rss,comm"}
	result, err := mgr.Exec(ctx, containerID, psCmd)
	if err != nil {
		return 0, fmt.Errorf("failed to execute ps: %w", err)
	}

	if result.ExitCode != 0 {
		return 0, fmt.Errorf("ps command failed: %s", result.Stderr)
	}

	// Parse ps output
	suspended := parsePsOutput(result.Stdout, cfg.CmdPrefixes)
	if len(suspended) == 0 {
		log.Debug("no suspended processes found")
		return 0, nil
	}

	log.Info("found suspended processes", "count", len(suspended))

	if cfg.DryRun {
		for _, p := range suspended {
			log.Info("would kill suspended process", "pid", p.PID, "cmd", p.Command, "state", p.State)
		}
		return 0, nil
	}

	// Kill suspended processes
	killed := 0
	for _, p := range suspended {
		killCmd := []string{"kill", "-9", strconv.Itoa(p.PID)}
		killResult, err := mgr.Exec(ctx, containerID, killCmd)
		if err != nil {
			log.Warn("failed to kill process", "pid", p.PID, "error", err)
			continue
		}
		if killResult.ExitCode != 0 {
			log.Warn("kill command failed", "pid", p.PID, "stderr", killResult.Stderr)
			continue
		}
		log.Info("killed suspended process", "pid", p.PID, "cmd", p.Command)
		killed++
	}

	return killed, nil
}

// parsePsOutput parses ps -axo output and returns suspended processes matching prefixes
func parsePsOutput(output string, cmdPrefixes []string) []*SuspendedProcess {
	var result []*SuspendedProcess

	lines := strings.Split(output, "\n")
	// Skip header line
	if len(lines) <= 1 {
		return result
	}

	// Compile prefix matchers
	var prefixPatterns []*regexp.Regexp
	for _, prefix := range cmdPrefixes {
		// Match command that starts with or contains the prefix
		pattern := regexp.MustCompile(`(?i)(^|/)` + regexp.QuoteMeta(prefix))
		prefixPatterns = append(prefixPatterns, pattern)
	}

	// Parse each line (skip header)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		proc := parseProcessLine(line)
		if proc == nil {
			continue
		}

		// Check if process is in T (stopped) state
		// STAT column starts with T for stopped processes
		if !strings.HasPrefix(proc.State, "T") {
			continue
		}

		// Check if command matches any prefix
		matchesPrefix := false
		for _, pattern := range prefixPatterns {
			if pattern.MatchString(proc.Command) {
				matchesPrefix = true
				break
			}
		}

		if matchesPrefix {
			result = append(result, proc)
		}
	}

	return result
}

// parseProcessLine parses a single ps output line
// Format: PID PPID STAT VSZ RSS COMM
func parseProcessLine(line string) *SuspendedProcess {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return nil
	}

	pid, err := strconv.Atoi(fields[0])
	if err != nil {
		return nil
	}

	ppid, err := strconv.Atoi(fields[1])
	if err != nil {
		ppid = 0
	}

	vsz, err := strconv.ParseInt(fields[3], 10, 64)
	if err != nil {
		vsz = 0
	}

	rss, err := strconv.ParseInt(fields[4], 10, 64)
	if err != nil {
		rss = 0
	}

	// Command is the rest of the fields joined
	command := strings.Join(fields[5:], " ")

	return &SuspendedProcess{
		PID:     pid,
		PPID:    ppid,
		State:   fields[2],
		VSZ:     vsz,
		RSS:     rss,
		Command: command,
	}
}

// CleanupZombieProcesses cleans up zombie (Z-state) processes in a container
func CleanupZombieProcesses(ctx context.Context, mgr Manager, containerID string) (int, error) {
	log := slog.Default().With("module", "container.cleanup", "container_id", containerID[:12])

	// Get zombie processes
	psCmd := []string{"ps", "-axo", "pid,ppid,stat,comm"}
	result, err := mgr.Exec(ctx, containerID, psCmd)
	if err != nil {
		return 0, fmt.Errorf("failed to execute ps: %w", err)
	}

	lines := strings.Split(result.Stdout, "\n")
	zombieCount := 0

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		// Z state indicates zombie
		if strings.HasPrefix(fields[2], "Z") {
			zombieCount++
			log.Info("found zombie process", "pid", fields[0], "ppid", fields[1], "comm", strings.Join(fields[3:], " "))
		}
	}

	if zombieCount > 0 {
		log.Warn("zombie processes detected (parent should reap)", "count", zombieCount)
	}

	return zombieCount, nil
}

// CleanupOrphanedProcesses finds processes whose parent is init (PID 1)
// These may indicate improperly cleaned up subprocesses
func CleanupOrphanedProcesses(ctx context.Context, mgr Manager, containerID string, cfg *ProcessCleanupConfig) (int, error) {
	if cfg == nil {
		cfg = DefaultProcessCleanupConfig()
	}

	log := slog.Default().With("module", "container.cleanup", "container_id", containerID[:12])

	psCmd := []string{"ps", "-axo", "pid,ppid,stat,comm"}
	result, err := mgr.Exec(ctx, containerID, psCmd)
	if err != nil {
		return 0, fmt.Errorf("failed to execute ps: %w", err)
	}

	// Compile prefix matchers
	var prefixPatterns []*regexp.Regexp
	for _, prefix := range cfg.CmdPrefixes {
		pattern := regexp.MustCompile(`(?i)(^|/)` + regexp.QuoteMeta(prefix))
		prefixPatterns = append(prefixPatterns, pattern)
	}

	lines := strings.Split(result.Stdout, "\n")
	orphanPids := []int{}

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		pid, _ := strconv.Atoi(fields[0])
		ppid, _ := strconv.Atoi(fields[1])
		command := strings.Join(fields[3:], " ")

		// Skip PID 1 itself
		if pid == 1 {
			continue
		}

		// Check if parent is init (PID 1) - indicates orphaned process
		if ppid == 1 {
			// Check if command matches any prefix
			for _, pattern := range prefixPatterns {
				if pattern.MatchString(command) {
					orphanPids = append(orphanPids, pid)
					log.Info("found orphaned process", "pid", pid, "comm", command)
					break
				}
			}
		}
	}

	if len(orphanPids) == 0 {
		return 0, nil
	}

	if cfg.DryRun {
		return 0, nil
	}

	// Kill orphaned processes
	killed := 0
	for _, pid := range orphanPids {
		killCmd := []string{"kill", "-9", strconv.Itoa(pid)}
		_, err := mgr.Exec(ctx, containerID, killCmd)
		if err == nil {
			killed++
		}
	}

	return killed, nil
}

// RunFullCleanup runs all cleanup routines
func RunFullCleanup(ctx context.Context, mgr Manager, containerID string, cfg *ProcessCleanupConfig) (*CleanupResult, error) {
	result := &CleanupResult{}

	// Cleanup suspended processes
	suspended, err := CleanupSuspendedProcesses(ctx, mgr, containerID, cfg)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("suspended cleanup failed: %v", err))
	}
	result.SuspendedKilled = suspended

	// Check for zombies (can't directly kill zombies, just report)
	zombies, err := CleanupZombieProcesses(ctx, mgr, containerID)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("zombie check failed: %v", err))
	}
	result.ZombiesFound = zombies

	// Cleanup orphaned processes
	orphans, err := CleanupOrphanedProcesses(ctx, mgr, containerID, cfg)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("orphan cleanup failed: %v", err))
	}
	result.OrphansKilled = orphans

	return result, nil
}

// CleanupResult contains results from cleanup operations
type CleanupResult struct {
	SuspendedKilled int
	ZombiesFound    int
	OrphansKilled   int
	Errors          []string
}

// Total returns total number of processes cleaned up
func (r *CleanupResult) Total() int {
	return r.SuspendedKilled + r.OrphansKilled
}

// HasErrors returns true if any errors occurred
func (r *CleanupResult) HasErrors() bool {
	return len(r.Errors) > 0
}
