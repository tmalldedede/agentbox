package engine

import (
	"fmt"
	"sync"
)

// Registry Agent 注册表
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]Adapter),
	}
}

// Register 注册 Agent 适配器
func (r *Registry) Register(adapter Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.Name()] = adapter
}

// Get 获取 Agent 适配器
func (r *Registry) Get(name string) (Adapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, ok := r.adapters[name]
	if !ok {
		return nil, fmt.Errorf("agent not found: %s", name)
	}
	return adapter, nil
}

// List 列出所有已注册的 Agent
func (r *Registry) List() []Info {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Info, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		result = append(result, GetInfo(adapter))
	}
	return result
}

// Names 返回所有已注册的 Agent 名称
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	return names
}

// 全局默认注册表
var defaultRegistry = NewRegistry()

// DefaultRegistry 获取默认注册表
func DefaultRegistry() *Registry {
	return defaultRegistry
}

// Register 注册到默认注册表
func Register(adapter Adapter) {
	defaultRegistry.Register(adapter)
}

// Get 从默认注册表获取
func Get(name string) (Adapter, error) {
	return defaultRegistry.Get(name)
}

// List 从默认注册表列出
func List() []Info {
	return defaultRegistry.List()
}
