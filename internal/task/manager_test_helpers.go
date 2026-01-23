package task

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tmalldedede/agentbox/internal/session"
)

// SetUploadDir 设置基于目录扫描的文件路径解析器（仅供测试使用）
func (m *Manager) SetUploadDir(dir string) {
	m.filePathResolver = func(fileID string) (string, string, error) {
		fileDir := filepath.Join(dir, fileID)
		entries, err := os.ReadDir(fileDir)
		if err != nil {
			return "", "", fmt.Errorf("file %s not found: %w", fileID, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				return filepath.Join(fileDir, entry.Name()), entry.Name(), nil
			}
		}
		return "", "", fmt.Errorf("no file in directory for %s", fileID)
	}
}

// MountAttachmentsForTest 公开 mountAttachments 供外部测试调用
func (m *Manager) MountAttachmentsForTest(workspace string, fileIDs []string) {
	sess := &session.Session{
		Workspace: workspace,
	}
	m.mountAttachments(context.Background(), sess, fileIDs)
}

// UpdateTaskForTest 公开 store.Update 供外部测试调用
func (m *Manager) UpdateTaskForTest(task *Task) error {
	return m.store.Update(task)
}

// BroadcastEventForTest 公开 broadcastEvent 供外部测试调用
func (m *Manager) BroadcastEventForTest(taskID string, event *TaskEvent) {
	m.broadcastEvent(taskID, event)
}
