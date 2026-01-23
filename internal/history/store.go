package history

import (
	"sort"
	"sync"
	"time"

	"github.com/tmalldedede/agentbox/internal/apperr"
)

// Store 历史记录存储接口
type Store interface {
	Create(entry *Entry) error
	Get(id string) (*Entry, error)
	List(filter *ListFilter) ([]*Entry, error)
	Count(filter *ListFilter) (int, error)
	Update(entry *Entry) error
	Delete(id string) error
	GetStats(filter *ListFilter) (*Stats, error)
}

// MemoryStore 内存存储实现
type MemoryStore struct {
	mu      sync.RWMutex
	entries map[string]*Entry
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		entries: make(map[string]*Entry),
	}
}

// Create 创建执行记录
func (s *MemoryStore) Create(entry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[entry.ID]; exists {
		return apperr.AlreadyExists("history entry")
	}

	if entry.StartedAt.IsZero() {
		entry.StartedAt = time.Now()
	}
	s.entries[entry.ID] = entry
	return nil
}

// Get 获取执行记录
func (s *MemoryStore) Get(id string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[id]
	if !ok {
		return nil, apperr.NotFound("history entry")
	}
	return entry, nil
}

// List 列出执行记录
func (s *MemoryStore) List(filter *ListFilter) ([]*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Entry, 0)
	for _, entry := range s.entries {
		if !matchFilter(entry, filter) {
			continue
		}
		result = append(result, entry)
	}

	// 按开始时间倒序排列 (最新的在前)
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartedAt.After(result[j].StartedAt)
	})

	// 应用分页
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		end := filter.Offset + filter.Limit
		if start >= len(result) {
			return []*Entry{}, nil
		}
		if end > len(result) {
			end = len(result)
		}
		result = result[start:end]
	}

	return result, nil
}

// Count 统计执行记录数量
func (s *MemoryStore) Count(filter *ListFilter) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, entry := range s.entries {
		if matchFilter(entry, filter) {
			count++
		}
	}
	return count, nil
}

// Update 更新执行记录
func (s *MemoryStore) Update(entry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[entry.ID]; !exists {
		return apperr.NotFound("history entry")
	}

	s.entries[entry.ID] = entry
	return nil
}

// Delete 删除执行记录
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[id]; !exists {
		return apperr.NotFound("history entry")
	}

	delete(s.entries, id)
	return nil
}

// GetStats 获取统计信息
func (s *MemoryStore) GetStats(filter *ListFilter) (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &Stats{
		BySource: make(map[string]int),
		ByEngine: make(map[string]int),
	}

	for _, entry := range s.entries {
		if !matchFilter(entry, filter) {
			continue
		}

		stats.TotalExecutions++

		switch entry.Status {
		case StatusCompleted:
			stats.CompletedCount++
		case StatusFailed:
			stats.FailedCount++
		}

		if entry.Usage != nil {
			stats.TotalInputTokens += entry.Usage.InputTokens
			stats.TotalOutputTokens += entry.Usage.OutputTokens
		}

		stats.BySource[string(entry.SourceType)]++
		if entry.Engine != "" {
			stats.ByEngine[entry.Engine]++
		}
	}

	return stats, nil
}

// matchFilter 检查是否匹配过滤条件
func matchFilter(entry *Entry, filter *ListFilter) bool {
	if filter == nil {
		return true
	}

	if filter.SourceType != "" && entry.SourceType != filter.SourceType {
		return false
	}
	if filter.SourceID != "" && entry.SourceID != filter.SourceID {
		return false
	}
	if filter.Engine != "" && entry.Engine != filter.Engine {
		return false
	}
	if filter.Status != "" && entry.Status != filter.Status {
		return false
	}

	return true
}
