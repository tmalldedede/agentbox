package session

import (
	"fmt"
	"sync"
	"time"
)

// Store 会话存储接口
type Store interface {
	Create(session *Session) error
	Get(id string) (*Session, error)
	List(filter *ListFilter) ([]*Session, error)
	Count(filter *ListFilter) (int, error)
	Update(session *Session) error
	Delete(id string) error

	// Execution 相关
	CreateExecution(exec *Execution) error
	GetExecution(id string) (*Execution, error)
	ListExecutions(sessionID string) ([]*Execution, error)
	UpdateExecution(exec *Execution) error
}

// MemoryStore 内存存储实现 (开发/测试用)
type MemoryStore struct {
	mu         sync.RWMutex
	sessions   map[string]*Session
	executions map[string]*Execution
}

// NewMemoryStore 创建内存存储
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions:   make(map[string]*Session),
		executions: make(map[string]*Execution),
	}
}

// Create 创建会话
func (s *MemoryStore) Create(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[session.ID]; exists {
		return fmt.Errorf("session already exists: %s", session.ID)
	}

	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	s.sessions[session.ID] = session
	return nil
}

// Get 获取会话
func (s *MemoryStore) Get(id string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	return session, nil
}

// List 列出会话
func (s *MemoryStore) List(filter *ListFilter) ([]*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Session, 0)
	for _, session := range s.sessions {
		// 应用过滤器
		if filter != nil {
			if filter.Agent != "" && session.Agent != filter.Agent {
				continue
			}
			if filter.Status != "" && session.Status != filter.Status {
				continue
			}
		}
		result = append(result, session)
	}

	// 应用分页
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		end := filter.Offset + filter.Limit
		if start >= len(result) {
			return []*Session{}, nil
		}
		if end > len(result) {
			end = len(result)
		}
		result = result[start:end]
	}

	return result, nil
}

// Count 统计会话数量
func (s *MemoryStore) Count(filter *ListFilter) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, session := range s.sessions {
		if filter != nil {
			if filter.Agent != "" && session.Agent != filter.Agent {
				continue
			}
			if filter.Status != "" && session.Status != filter.Status {
				continue
			}
		}
		count++
	}
	return count, nil
}

// Update 更新会话
func (s *MemoryStore) Update(session *Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[session.ID]; !exists {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	session.UpdatedAt = time.Now()
	s.sessions[session.ID] = session
	return nil
}

// Delete 删除会话
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	delete(s.sessions, id)
	return nil
}

// CreateExecution 创建执行记录
func (s *MemoryStore) CreateExecution(exec *Execution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.executions[exec.ID]; exists {
		return fmt.Errorf("execution already exists: %s", exec.ID)
	}

	exec.StartedAt = time.Now()
	s.executions[exec.ID] = exec
	return nil
}

// GetExecution 获取执行记录
func (s *MemoryStore) GetExecution(id string) (*Execution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exec, ok := s.executions[id]
	if !ok {
		return nil, fmt.Errorf("execution not found: %s", id)
	}
	return exec, nil
}

// ListExecutions 列出会话的执行记录
func (s *MemoryStore) ListExecutions(sessionID string) ([]*Execution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Execution, 0)
	for _, exec := range s.executions {
		if exec.SessionID == sessionID {
			result = append(result, exec)
		}
	}
	return result, nil
}

// UpdateExecution 更新执行记录
func (s *MemoryStore) UpdateExecution(exec *Execution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.executions[exec.ID]; !exists {
		return fmt.Errorf("execution not found: %s", exec.ID)
	}

	s.executions[exec.ID] = exec
	return nil
}
