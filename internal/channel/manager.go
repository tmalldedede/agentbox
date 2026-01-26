// Package channel 提供多通道消息管理
package channel

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tmalldedede/agentbox/internal/logger"
)

var log *slog.Logger

func init() {
	log = logger.Module("channel")
}

// Manager 通道管理器
type Manager struct {
	channels map[string]Channel // type -> channel
	handlers []MessageHandler   // 消息处理器列表

	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewManager 创建通道管理器
func NewManager() *Manager {
	return &Manager{
		channels: make(map[string]Channel),
		handlers: make([]MessageHandler, 0),
	}
}

// Register 注册通道
func (m *Manager) Register(ch Channel) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.channels[ch.Type()]; exists {
		return fmt.Errorf("channel type %s already registered", ch.Type())
	}

	m.channels[ch.Type()] = ch
	log.Info("channel registered", "type", ch.Type())
	return nil
}

// AddHandler 添加消息处理器
func (m *Manager) AddHandler(handler MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers = append(m.handlers, handler)
}

// Start 启动所有通道
func (m *Manager) Start(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.channels {
		if err := ch.Start(m.ctx); err != nil {
			return fmt.Errorf("start channel %s: %w", ch.Type(), err)
		}

		// 启动消息处理协程
		m.wg.Add(1)
		go m.processMessages(ch)
	}

	log.Info("channel manager started", "channels", len(m.channels))
	return nil
}

// Stop 停止所有通道
func (m *Manager) Stop() error {
	if m.cancel != nil {
		m.cancel()
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.channels {
		if err := ch.Stop(); err != nil {
			log.Error("stop channel failed", "type", ch.Type(), "error", err)
		}
	}

	m.wg.Wait()
	log.Info("channel manager stopped")
	return nil
}

// GetChannel 获取通道
func (m *Manager) GetChannel(channelType string) (Channel, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ch, ok := m.channels[channelType]
	return ch, ok
}

// Send 发送消息到指定通道
func (m *Manager) Send(ctx context.Context, channelType string, req *SendRequest) (*SendResponse, error) {
	ch, ok := m.GetChannel(channelType)
	if !ok {
		return nil, fmt.Errorf("channel type %s not found", channelType)
	}
	return ch.Send(ctx, req)
}

// processMessages 处理通道消息
func (m *Manager) processMessages(ch Channel) {
	defer m.wg.Done()

	for {
		select {
		case <-m.ctx.Done():
			return
		case msg, ok := <-ch.Messages():
			if !ok {
				return
			}
			m.handleMessage(msg)
		}
	}
}

// handleMessage 处理单条消息
func (m *Manager) handleMessage(msg *Message) {
	m.mu.RLock()
	handlers := make([]MessageHandler, len(m.handlers))
	copy(handlers, m.handlers)
	m.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(m.ctx, msg); err != nil {
			log.Error("handler error", "channel", msg.ChannelType, "message_id", msg.ID, "error", err)
		}
	}
}

// ListChannels 列出所有通道
func (m *Manager) ListChannels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	types := make([]string, 0, len(m.channels))
	for t := range m.channels {
		types = append(types, t)
	}
	return types
}
