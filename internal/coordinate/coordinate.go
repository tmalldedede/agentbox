// Package coordinate 提供跨会话协调功能
package coordinate

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/tmalldedede/agentbox/internal/logger"
	"github.com/tmalldedede/agentbox/internal/task"
)

var log *slog.Logger

func init() {
	log = logger.Module("coordinate")
}

// SessionInfo 会话信息（对外暴露的简化视图）
type SessionInfo struct {
	TaskID    string    `json:"task_id"`
	AgentID   string    `json:"agent_id"`
	AgentName string    `json:"agent_name"`
	Status    string    `json:"status"`
	Prompt    string    `json:"prompt"`      // 首轮 prompt（截断）
	TurnCount int       `json:"turn_count"`
	StartedAt time.Time `json:"started_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MessageRecord 消息记录
type MessageRecord struct {
	Role      string    `json:"role"` // user / assistant
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	TargetTaskID string `json:"target_task_id" binding:"required"`
	Message      string `json:"message" binding:"required"`
}

// Manager 协调管理器
type Manager struct {
	taskMgr *task.Manager
}

// NewManager 创建协调管理器
func NewManager(taskMgr *task.Manager) *Manager {
	return &Manager{
		taskMgr: taskMgr,
	}
}

// ListSessions 列出所有活跃会话
func (m *Manager) ListSessions(ctx context.Context) ([]*SessionInfo, error) {
	// 获取所有运行中和排队中的任务
	tasks, err := m.taskMgr.ListTasks(&task.ListFilter{
		Status: []task.Status{
			task.StatusQueued,
			task.StatusRunning,
		},
		Limit: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	sessions := make([]*SessionInfo, 0, len(tasks))
	for _, t := range tasks {
		prompt := t.Prompt
		if len(prompt) > 100 {
			prompt = prompt[:100] + "..."
		}

		// 使用最新的时间戳
		updatedAt := t.CreatedAt
		if t.StartedAt != nil {
			updatedAt = *t.StartedAt
		}
		if t.CompletedAt != nil {
			updatedAt = *t.CompletedAt
		}

		sessions = append(sessions, &SessionInfo{
			TaskID:    t.ID,
			AgentID:   t.AgentID,
			AgentName: t.AgentName,
			Status:    string(t.Status),
			Prompt:    prompt,
			TurnCount: t.TurnCount,
			StartedAt: t.CreatedAt,
			UpdatedAt: updatedAt,
		})
	}

	return sessions, nil
}

// GetSessionHistory 获取会话历史
func (m *Manager) GetSessionHistory(ctx context.Context, taskID string) ([]*MessageRecord, error) {
	t, err := m.taskMgr.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	records := make([]*MessageRecord, 0)

	// 添加初始 prompt
	records = append(records, &MessageRecord{
		Role:      "user",
		Content:   t.Prompt,
		Timestamp: t.CreatedAt,
	})

	// 如果首轮有结果
	if t.Result != nil && t.Result.Text != "" {
		ts := t.CreatedAt
		if t.CompletedAt != nil {
			ts = *t.CompletedAt
		}
		records = append(records, &MessageRecord{
			Role:      "assistant",
			Content:   t.Result.Text,
			Timestamp: ts,
		})
	}

	// 添加多轮对话
	for _, turn := range t.Turns {
		// 用户输入
		records = append(records, &MessageRecord{
			Role:      "user",
			Content:   turn.Prompt,
			Timestamp: turn.CreatedAt,
		})

		// Agent 输出
		if turn.Result != nil && turn.Result.Text != "" {
			records = append(records, &MessageRecord{
				Role:      "assistant",
				Content:   turn.Result.Text,
				Timestamp: turn.CreatedAt, // Turn 没有完成时间，用创建时间
			})
		}
	}

	return records, nil
}

// SendMessage 向另一个会话发送消息（追加一轮对话）
func (m *Manager) SendMessage(ctx context.Context, req *SendMessageRequest) error {
	log.Info("sending cross-session message",
		"target", req.TargetTaskID,
		"message_len", len(req.Message),
	)

	// 使用 appendTurn 机制
	_, err := m.taskMgr.CreateTask(&task.CreateTaskRequest{
		TaskID: req.TargetTaskID,
		Prompt: req.Message,
	})
	if err != nil {
		return fmt.Errorf("append turn: %w", err)
	}

	return nil
}

// GetSession 获取单个会话信息
func (m *Manager) GetSession(ctx context.Context, taskID string) (*SessionInfo, error) {
	t, err := m.taskMgr.GetTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	prompt := t.Prompt
	if len(prompt) > 100 {
		prompt = prompt[:100] + "..."
	}

	// 使用最新的时间戳
	updatedAt := t.CreatedAt
	if t.StartedAt != nil {
		updatedAt = *t.StartedAt
	}
	if t.CompletedAt != nil {
		updatedAt = *t.CompletedAt
	}

	return &SessionInfo{
		TaskID:    t.ID,
		AgentID:   t.AgentID,
		AgentName: t.AgentName,
		Status:    string(t.Status),
		Prompt:    prompt,
		TurnCount: t.TurnCount,
		StartedAt: t.CreatedAt,
		UpdatedAt: updatedAt,
	}, nil
}
