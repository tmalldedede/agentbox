package history

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/logger"
)

var log *slog.Logger

func init() {
	log = logger.Module("history")
}

// Manager 执行历史管理器
type Manager struct {
	store Store
}

// NewManager 创建历史管理器
func NewManager(store Store) *Manager {
	if store == nil {
		store = NewMemoryStore()
	}
	return &Manager{store: store}
}

// Record 记录一条执行历史
func (m *Manager) Record(entry *Entry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.StartedAt.IsZero() {
		entry.StartedAt = time.Now()
	}
	if entry.Status == "" {
		entry.Status = StatusPending
	}

	log.Info("recording execution",
		"id", entry.ID,
		"source_type", entry.SourceType,
		"source_id", entry.SourceID,
	)

	return m.store.Create(entry)
}

// RecordSession 记录 Session 执行
func (m *Manager) RecordSession(execID, sessionID, sessionName, engine, prompt string) (*Entry, error) {
	entry := &Entry{
		ID:         execID,
		SourceType: SourceSession,
		SourceID:   sessionID,
		SourceName: sessionName,
		Engine:     engine,
		Prompt:     prompt,
		Status:     StatusRunning,
		StartedAt:  time.Now(),
	}

	if err := m.store.Create(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// RecordAgent 记录 Agent 执行
func (m *Manager) RecordAgent(execID, agentID, agentName, engine, prompt string) (*Entry, error) {
	entry := &Entry{
		ID:         execID,
		SourceType: SourceAgent,
		SourceID:   agentID,
		SourceName: agentName,
		Engine:     engine,
		Prompt:     prompt,
		Status:     StatusRunning,
		StartedAt:  time.Now(),
	}

	if err := m.store.Create(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

// Complete 标记执行完成
func (m *Manager) Complete(id string, output string, usage *UsageInfo) error {
	entry, err := m.store.Get(id)
	if err != nil {
		return err
	}

	entry.Status = StatusCompleted
	entry.Output = output
	entry.Usage = usage
	now := time.Now()
	entry.EndedAt = &now

	log.Info("execution completed",
		"id", id,
		"duration", entry.Duration().String(),
	)

	return m.store.Update(entry)
}

// Fail 标记执行失败
func (m *Manager) Fail(id string, errMsg string, exitCode int) error {
	entry, err := m.store.Get(id)
	if err != nil {
		return err
	}

	entry.Status = StatusFailed
	entry.Error = errMsg
	entry.ExitCode = exitCode
	now := time.Now()
	entry.EndedAt = &now

	log.Warn("execution failed",
		"id", id,
		"error", errMsg,
		"exit_code", exitCode,
	)

	return m.store.Update(entry)
}

// Get 获取执行记录
func (m *Manager) Get(id string) (*Entry, error) {
	return m.store.Get(id)
}

// List 列出执行记录
func (m *Manager) List(filter *ListFilter) ([]*Entry, error) {
	return m.store.List(filter)
}

// Count 统计执行记录数量
func (m *Manager) Count(filter *ListFilter) (int, error) {
	return m.store.Count(filter)
}

// Delete 删除执行记录
func (m *Manager) Delete(id string) error {
	return m.store.Delete(id)
}

// GetStats 获取统计信息
func (m *Manager) GetStats(filter *ListFilter) (*Stats, error) {
	return m.store.GetStats(filter)
}
