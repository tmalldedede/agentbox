package task

import (
	"time"
)

// TaskStats 任务统计
type TaskStats struct {
	Total       int            `json:"total"`
	ByStatus    map[Status]int `json:"by_status"`
	ByAgent     map[string]int `json:"by_agent"`
	AvgDuration float64        `json:"avg_duration_seconds"` // 已完成任务平均耗时
}

// Store 任务存储接口
type Store interface {
	// Create 创建任务
	Create(task *Task) error
	// Get 获取任务
	Get(id string) (*Task, error)
	// Update 更新任务
	Update(task *Task) error
	// Delete 删除任务
	Delete(id string) error
	// List 列出任务
	List(filter *ListFilter) ([]*Task, error)
	// Count 统计任务数量
	Count(filter *ListFilter) (int, error)
	// Stats 获取任务统计
	Stats() (*TaskStats, error)
	// Cleanup 清理旧任务
	Cleanup(before time.Time, statuses []Status) (int, error)
	// ClaimQueued 原子领取等待中的任务（用于多实例调度）
	ClaimQueued(limit int) ([]*Task, error)
	// Close 关闭存储
	Close() error
}

// ListFilter 列表过滤器
type ListFilter struct {
	Status    []Status // 按状态过滤
	AgentID   string   // 按 Agent 过滤
	Search    string   // 搜索 prompt 关键字
	Limit     int      // 限制数量
	Offset    int      // 偏移量
	OrderBy   string   // 排序字段：created_at, started_at, completed_at
	OrderDesc bool     // 是否降序
}
