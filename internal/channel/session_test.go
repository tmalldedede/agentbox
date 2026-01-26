package channel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSessionManager_CreateAndGet(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)

	// 创建会话
	session := sm.CreateSession("feishu:chat123:user456", "task-abc", "agent-xyz")
	assert.NotNil(t, session)
	assert.Equal(t, "feishu:chat123:user456", session.Key)
	assert.Equal(t, "task-abc", session.TaskID)
	assert.Equal(t, "agent-xyz", session.AgentID)

	// 获取会话
	got := sm.GetSession("feishu:chat123:user456")
	assert.NotNil(t, got)
	assert.Equal(t, session.TaskID, got.TaskID)

	// 获取不存在的会话
	notFound := sm.GetSession("feishu:chat999:user999")
	assert.Nil(t, notFound)
}

func TestSessionManager_Touch(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)

	session := sm.CreateSession("test:key", "task-1", "agent-1")
	originalTime := session.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	sm.TouchSession("test:key")

	got := sm.GetSession("test:key")
	assert.True(t, got.UpdatedAt.After(originalTime))
}

func TestSessionManager_Delete(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)

	sm.CreateSession("test:key", "task-1", "agent-1")
	assert.NotNil(t, sm.GetSession("test:key"))

	sm.DeleteSession("test:key")
	assert.Nil(t, sm.GetSession("test:key"))
}

func TestSessionManager_DeleteByTaskID(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)

	sm.CreateSession("test:key1", "task-1", "agent-1")
	sm.CreateSession("test:key2", "task-2", "agent-1")

	sm.DeleteByTaskID("task-1")

	assert.Nil(t, sm.GetSession("test:key1"))
	assert.NotNil(t, sm.GetSession("test:key2"))
}

func TestSessionManager_Expiration(t *testing.T) {
	// 使用很短的超时时间
	sm := NewSessionManager(50 * time.Millisecond)

	sm.CreateSession("test:key", "task-1", "agent-1")
	assert.NotNil(t, sm.GetSession("test:key"))

	// 等待超时
	time.Sleep(100 * time.Millisecond)

	// 超时后应该返回 nil
	assert.Nil(t, sm.GetSession("test:key"))
}

func TestSessionManager_Cleanup(t *testing.T) {
	sm := NewSessionManager(50 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sm.Start(ctx)

	sm.CreateSession("test:key1", "task-1", "agent-1")
	sm.CreateSession("test:key2", "task-2", "agent-1")

	// 等待超时 + 清理周期
	time.Sleep(150 * time.Millisecond)

	// 手动触发清理
	sm.cleanup()

	stats := sm.Stats()
	assert.Equal(t, 0, stats["total"])
}

func TestGetSessionKey(t *testing.T) {
	// 群聊：包含用户 ID
	groupKey := GetSessionKey("feishu", "chat123", "user456", true)
	assert.Equal(t, "feishu:chat123:user456", groupKey)

	// 私聊：不包含用户 ID
	p2pKey := GetSessionKey("feishu", "chat123", "user456", false)
	assert.Equal(t, "feishu:chat123", p2pKey)
}

func TestSessionManager_Stats(t *testing.T) {
	sm := NewSessionManager(5 * time.Minute)

	sm.CreateSession("test:key1", "task-1", "agent-1")
	sm.CreateSession("test:key2", "task-2", "agent-1")

	stats := sm.Stats()
	assert.Equal(t, 2, stats["total"])
	assert.Equal(t, 2, stats["active"])
	assert.Equal(t, "5m0s", stats["timeout"])
}
