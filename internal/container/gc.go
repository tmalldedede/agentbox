package container

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/tmalldedede/agentbox/internal/logger"
)

// SessionLister 用于查询活跃会话的容器 ID，避免循环依赖
type SessionLister interface {
	ListContainerIDs(ctx context.Context) ([]string, error)
}

// GCConfig GC 配置
type GCConfig struct {
	Interval     time.Duration `json:"interval"`
	ContainerTTL time.Duration `json:"container_ttl"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// GCStats GC 运行统计
type GCStats struct {
	Running           bool      `json:"running"`
	LastRunAt         time.Time `json:"last_run_at"`
	NextRunAt         time.Time `json:"next_run_at"`
	ContainersRemoved int       `json:"containers_removed"`
	TotalRuns         int       `json:"total_runs"`
	Errors            []string  `json:"errors"`
	Config            GCConfig  `json:"config"`
}

// GarbageCollector 容器垃圾回收器
type GarbageCollector struct {
	containerMgr Manager
	sessionMgr   SessionLister
	config       GCConfig
	stopCh       chan struct{}
	stats        GCStats
	mu           sync.RWMutex
	logger       *slog.Logger
}

// NewGarbageCollector 创建垃圾回收器
func NewGarbageCollector(containerMgr Manager, sessionMgr SessionLister, config GCConfig) *GarbageCollector {
	return &GarbageCollector{
		containerMgr: containerMgr,
		sessionMgr:   sessionMgr,
		config:       config,
		stopCh:       make(chan struct{}),
		stats: GCStats{
			Errors: make([]string, 0),
			Config: config,
		},
		logger: logger.Module("gc"),
	}
}

// Start 启动 GC 后台 goroutine
func (gc *GarbageCollector) Start() {
	gc.mu.Lock()
	gc.stats.Running = true
	gc.stats.NextRunAt = time.Now().Add(gc.config.Interval)
	gc.mu.Unlock()

	gc.logger.Info("started",
		"interval", gc.config.Interval,
		"container_ttl", gc.config.ContainerTTL,
		"idle_timeout", gc.config.IdleTimeout,
	)

	// 启动时执行一次清理（Docker 不可用时跳过，不报错）
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := gc.RunOnce(ctx); err != nil {
			gc.logger.Warn("startup gc skipped", "error", err)
		}
	}()

	go gc.loop()
}

// Stop 停止 GC
func (gc *GarbageCollector) Stop() {
	gc.logger.Info("stopping")
	close(gc.stopCh)

	gc.mu.Lock()
	gc.stats.Running = false
	gc.mu.Unlock()
}

// Stats 获取 GC 统计信息
func (gc *GarbageCollector) Stats() GCStats {
	gc.mu.RLock()
	defer gc.mu.RUnlock()

	// 返回副本
	stats := gc.stats
	errors := make([]string, len(gc.stats.Errors))
	copy(errors, gc.stats.Errors)
	stats.Errors = errors
	return stats
}

// RunOnce 执行一次 GC 扫描
func (gc *GarbageCollector) RunOnce(ctx context.Context) error {
	gc.logger.Debug("gc scan started")

	// 获取所有 agentbox.managed 容器
	containers, err := gc.containerMgr.ListContainers(ctx)
	if err != nil {
		gc.recordError("list containers: " + err.Error())
		return err
	}

	// 获取所有活跃 Session 的容器 ID
	activeIDs, err := gc.sessionMgr.ListContainerIDs(ctx)
	if err != nil {
		gc.recordError("list session container IDs: " + err.Error())
		return err
	}

	activeSet := make(map[string]bool, len(activeIDs))
	for _, id := range activeIDs {
		activeSet[id] = true
	}

	now := time.Now()
	var removed int

	for _, ctr := range containers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		shouldRemove := false
		reason := ""

		// 1. 孤立容器：无对应 Session
		if !activeSet[ctr.ID] {
			shouldRemove = true
			reason = "orphaned (no session)"
		}

		// 2. 容器运行时间超过 TTL
		if !shouldRemove && ctr.Created > 0 {
			createdAt := time.Unix(ctr.Created, 0)
			if now.Sub(createdAt) > gc.config.ContainerTTL {
				shouldRemove = true
				reason = "exceeded TTL"
			}
		}

		// 3. 容器状态为 Exited 且超过 IdleTimeout
		if !shouldRemove && ctr.Status == StatusExited {
			// 使用容器创建时间 + TTL 作为近似退出时间
			// Docker API 中 Created 字段是容器创建时间
			createdAt := time.Unix(ctr.Created, 0)
			if now.Sub(createdAt) > gc.config.IdleTimeout {
				shouldRemove = true
				reason = "exited idle timeout"
			}
		}

		if shouldRemove {
			gc.logger.Info("removing container",
				"container_id", ctr.ID[:12],
				"reason", reason,
				"status", ctr.Status,
			)

			// 先停止再删除
			_ = gc.containerMgr.Stop(ctx, ctr.ID)
			if err := gc.containerMgr.Remove(ctx, ctr.ID); err != nil {
				gc.recordError("remove " + ctr.ID[:12] + ": " + err.Error())
				gc.logger.Error("failed to remove container", "container_id", ctr.ID[:12], "error", err)
			} else {
				removed++
			}
		}
	}

	// 更新统计
	gc.mu.Lock()
	gc.stats.LastRunAt = now
	gc.stats.NextRunAt = now.Add(gc.config.Interval)
	gc.stats.ContainersRemoved += removed
	gc.stats.TotalRuns++
	gc.mu.Unlock()

	gc.logger.Debug("gc scan completed",
		"scanned", len(containers),
		"removed", removed,
	)

	return nil
}

// GCCandidate 待清理容器候选
type GCCandidate struct {
	ContainerID string          `json:"container_id"`
	Name        string          `json:"name"`
	Image       string          `json:"image"`
	Status      ContainerStatus `json:"status"`
	CreatedAt   int64           `json:"created_at"`
	Reason      string          `json:"reason"`
}

// Preview 预览哪些容器会被 GC（不实际删除）
func (gc *GarbageCollector) Preview(ctx context.Context) ([]GCCandidate, error) {
	containers, err := gc.containerMgr.ListContainers(ctx)
	if err != nil {
		return nil, err
	}

	activeIDs, err := gc.sessionMgr.ListContainerIDs(ctx)
	if err != nil {
		return nil, err
	}

	activeSet := make(map[string]bool, len(activeIDs))
	for _, id := range activeIDs {
		activeSet[id] = true
	}

	now := time.Now()
	var candidates []GCCandidate

	for _, ctr := range containers {
		reason := ""

		if !activeSet[ctr.ID] {
			reason = "orphaned (no session)"
		}

		if reason == "" && ctr.Created > 0 {
			createdAt := time.Unix(ctr.Created, 0)
			if now.Sub(createdAt) > gc.config.ContainerTTL {
				reason = "exceeded TTL"
			}
		}

		if reason == "" && ctr.Status == StatusExited {
			createdAt := time.Unix(ctr.Created, 0)
			if now.Sub(createdAt) > gc.config.IdleTimeout {
				reason = "exited idle timeout"
			}
		}

		if reason != "" {
			candidates = append(candidates, GCCandidate{
				ContainerID: ctr.ID,
				Name:        ctr.Name,
				Image:       ctr.Image,
				Status:      ctr.Status,
				CreatedAt:   ctr.Created,
				Reason:      reason,
			})
		}
	}

	return candidates, nil
}

// UpdateConfig 热更新 GC 配置
func (gc *GarbageCollector) UpdateConfig(newConfig GCConfig) {
	gc.mu.Lock()
	gc.config = newConfig
	gc.stats.Config = newConfig
	gc.mu.Unlock()

	gc.logger.Info("config updated",
		"interval", newConfig.Interval,
		"container_ttl", newConfig.ContainerTTL,
		"idle_timeout", newConfig.IdleTimeout,
	)
}

func (gc *GarbageCollector) loop() {
	ticker := time.NewTicker(gc.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-gc.stopCh:
			gc.logger.Info("stopped")
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := gc.RunOnce(ctx); err != nil {
				gc.logger.Warn("gc scan skipped", "error", err)
			}
			cancel()
		}
	}
}

func (gc *GarbageCollector) recordError(msg string) {
	gc.mu.Lock()
	defer gc.mu.Unlock()

	gc.stats.Errors = append(gc.stats.Errors, msg)
	// 最多保留最近 20 条错误
	if len(gc.stats.Errors) > 20 {
		gc.stats.Errors = gc.stats.Errors[len(gc.stats.Errors)-20:]
	}
}
