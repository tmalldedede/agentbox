// Package plugin 提供简单的插件扩展机制
package plugin

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/tmalldedede/agentbox/internal/logger"
)

var log *slog.Logger

func init() {
	log = logger.Module("plugin")
}

// ToolFactory 工具工厂函数
type ToolFactory func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// HookHandler 钩子处理器
type HookHandler func(ctx context.Context, event string, data interface{}) error

// HTTPHandler HTTP 路由处理器
type HTTPHandler func(w http.ResponseWriter, r *http.Request)

// Plugin 插件接口
type Plugin interface {
	// Name 插件名称
	Name() string

	// Version 插件版本
	Version() string

	// Init 初始化插件
	Init(api *API) error

	// Shutdown 关闭插件
	Shutdown() error
}

// API 插件 API（暴露给插件的接口）
type API struct {
	manager *Manager
	plugin  Plugin
}

// RegisterTool 注册工具
func (a *API) RegisterTool(name string, factory ToolFactory) {
	a.manager.registerTool(a.plugin.Name(), name, factory)
}

// RegisterHook 注册钩子
func (a *API) RegisterHook(event string, handler HookHandler) {
	a.manager.registerHook(a.plugin.Name(), event, handler)
}

// RegisterHTTPRoute 注册 HTTP 路由
func (a *API) RegisterHTTPRoute(method, path string, handler HTTPHandler) {
	a.manager.registerHTTPRoute(a.plugin.Name(), method, path, handler)
}

// Tool 已注册的工具
type Tool struct {
	Plugin  string
	Name    string
	Factory ToolFactory
}

// Hook 已注册的钩子
type Hook struct {
	Plugin  string
	Event   string
	Handler HookHandler
}

// HTTPRoute 已注册的 HTTP 路由
type HTTPRoute struct {
	Plugin  string
	Method  string
	Path    string
	Handler HTTPHandler
}

// Manager 插件管理器
type Manager struct {
	plugins map[string]Plugin
	tools   map[string]*Tool     // "plugin.tool" -> Tool
	hooks   map[string][]*Hook   // event -> []Hook
	routes  []*HTTPRoute

	mu sync.RWMutex
}

// NewManager 创建插件管理器
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		tools:   make(map[string]*Tool),
		hooks:   make(map[string][]*Hook),
		routes:  make([]*HTTPRoute, 0),
	}
}

// Register 注册插件
func (m *Manager) Register(plugin Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := plugin.Name()
	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	api := &API{
		manager: m,
		plugin:  plugin,
	}

	if err := plugin.Init(api); err != nil {
		return fmt.Errorf("init plugin %s: %w", name, err)
	}

	m.plugins[name] = plugin
	log.Info("plugin registered", "name", name, "version", plugin.Version())
	return nil
}

// Unregister 注销插件
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if err := plugin.Shutdown(); err != nil {
		log.Error("shutdown plugin failed", "name", name, "error", err)
	}

	// 清理注册的工具
	for key := range m.tools {
		if m.tools[key].Plugin == name {
			delete(m.tools, key)
		}
	}

	// 清理注册的钩子
	for event, hooks := range m.hooks {
		filtered := make([]*Hook, 0)
		for _, h := range hooks {
			if h.Plugin != name {
				filtered = append(filtered, h)
			}
		}
		m.hooks[event] = filtered
	}

	// 清理注册的路由
	filteredRoutes := make([]*HTTPRoute, 0)
	for _, r := range m.routes {
		if r.Plugin != name {
			filteredRoutes = append(filteredRoutes, r)
		}
	}
	m.routes = filteredRoutes

	delete(m.plugins, name)
	log.Info("plugin unregistered", "name", name)
	return nil
}

// CallTool 调用工具
func (m *Manager) CallTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	m.mu.RLock()
	tool, exists := m.tools[name]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool.Factory(ctx, params)
}

// EmitHook 触发钩子
func (m *Manager) EmitHook(ctx context.Context, event string, data interface{}) error {
	m.mu.RLock()
	hooks := m.hooks[event]
	m.mu.RUnlock()

	for _, h := range hooks {
		if err := h.Handler(ctx, event, data); err != nil {
			log.Error("hook error", "plugin", h.Plugin, "event", event, "error", err)
		}
	}

	return nil
}

// GetRoutes 获取所有 HTTP 路由
func (m *Manager) GetRoutes() []*HTTPRoute {
	m.mu.RLock()
	defer m.mu.RUnlock()

	routes := make([]*HTTPRoute, len(m.routes))
	copy(routes, m.routes)
	return routes
}

// ListPlugins 列出所有插件
func (m *Manager) ListPlugins() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	return names
}

// ListTools 列出所有工具
func (m *Manager) ListTools() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.tools))
	for name := range m.tools {
		names = append(names, name)
	}
	return names
}

// Shutdown 关闭所有插件
func (m *Manager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, plugin := range m.plugins {
		if err := plugin.Shutdown(); err != nil {
			log.Error("shutdown plugin failed", "name", name, "error", err)
		}
	}

	m.plugins = make(map[string]Plugin)
	m.tools = make(map[string]*Tool)
	m.hooks = make(map[string][]*Hook)
	m.routes = make([]*HTTPRoute, 0)

	log.Info("plugin manager shutdown")
	return nil
}

// registerTool 注册工具（内部方法）
func (m *Manager) registerTool(plugin, name string, factory ToolFactory) {
	fullName := plugin + "." + name
	m.tools[fullName] = &Tool{
		Plugin:  plugin,
		Name:    name,
		Factory: factory,
	}
	log.Debug("tool registered", "plugin", plugin, "name", name)
}

// registerHook 注册钩子（内部方法）
func (m *Manager) registerHook(plugin, event string, handler HookHandler) {
	m.hooks[event] = append(m.hooks[event], &Hook{
		Plugin:  plugin,
		Event:   event,
		Handler: handler,
	})
	log.Debug("hook registered", "plugin", plugin, "event", event)
}

// registerHTTPRoute 注册 HTTP 路由（内部方法）
func (m *Manager) registerHTTPRoute(plugin, method, path string, handler HTTPHandler) {
	m.routes = append(m.routes, &HTTPRoute{
		Plugin:  plugin,
		Method:  method,
		Path:    path,
		Handler: handler,
	})
	log.Debug("route registered", "plugin", plugin, "method", method, "path", path)
}
