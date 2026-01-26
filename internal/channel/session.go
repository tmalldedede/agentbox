// Package channel 通道会话管理
package channel

import (
	"context"
	"sync"
	"time"
)

// ChannelSession 通道会话（跟踪多轮对话）
type ChannelSession struct {
	Key       string    // 会话键 (channelType:chatID:userID)
	TaskID    string    // 关联的 Task ID
	AgentID   string    // 使用的 Agent ID
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 最后活跃时间
}

// SessionManager 通道会话管理器
type SessionManager struct {
	sessions map[string]*ChannelSession // key -> session
	mu       sync.RWMutex
	timeout  time.Duration // 会话超时时间

	// 清理
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSessionManager 创建会话管理器
func NewSessionManager(timeout time.Duration) *SessionManager {
	if timeout == 0 {
		timeout = 5 * time.Minute // 默认 5 分钟超时
	}

	sm := &SessionManager{
		sessions: make(map[string]*ChannelSession),
		timeout:  timeout,
	}

	return sm
}

// Start 启动会话清理协程
func (sm *SessionManager) Start(ctx context.Context) {
	sm.ctx, sm.cancel = context.WithCancel(ctx)
	go sm.cleanupLoop()
}

// Stop 停止会话管理器
func (sm *SessionManager) Stop() {
	if sm.cancel != nil {
		sm.cancel()
	}
}

// GetSession 获取会话（如果存在且未过期）
func (sm *SessionManager) GetSession(key string) *ChannelSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[key]
	if !exists {
		return nil
	}

	// 检查是否过期
	if time.Since(session.UpdatedAt) > sm.timeout {
		return nil
	}

	return session
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(key, taskID, agentID string) *ChannelSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	session := &ChannelSession{
		Key:       key,
		TaskID:    taskID,
		AgentID:   agentID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	sm.sessions[key] = session
	log.Debug("channel session created", "key", key, "task_id", taskID)
	return session
}

// TouchSession 更新会话活跃时间
func (sm *SessionManager) TouchSession(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[key]; exists {
		session.UpdatedAt = time.Now()
	}
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(key string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.sessions, key)
	log.Debug("channel session deleted", "key", key)
}

// DeleteByTaskID 根据 TaskID 删除会话
func (sm *SessionManager) DeleteByTaskID(taskID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for key, session := range sm.sessions {
		if session.TaskID == taskID {
			delete(sm.sessions, key)
			log.Debug("channel session deleted by task_id", "key", key, "task_id", taskID)
			break
		}
	}
}

// cleanupLoop 定期清理过期会话
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.cleanup()
		}
	}
}

// cleanup 清理过期会话
func (sm *SessionManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for key, session := range sm.sessions {
		if now.Sub(session.UpdatedAt) > sm.timeout {
			delete(sm.sessions, key)
			log.Debug("channel session expired", "key", key, "task_id", session.TaskID)
		}
	}
}

// GetSessionKey 生成会话键
// 群聊：按用户隔离 (channelType:chatID:userID)
// 私聊：按聊天隔离 (channelType:chatID)
func GetSessionKey(channelType, chatID, userID string, isGroup bool) string {
	if isGroup {
		return channelType + ":" + chatID + ":" + userID
	}
	return channelType + ":" + chatID
}

// Stats 返回会话统计
func (sm *SessionManager) Stats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	active := 0
	for _, session := range sm.sessions {
		if time.Since(session.UpdatedAt) <= sm.timeout {
			active++
		}
	}

	return map[string]interface{}{
		"total":   len(sm.sessions),
		"active":  active,
		"timeout": sm.timeout.String(),
	}
}
