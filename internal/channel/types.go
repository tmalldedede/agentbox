// Package channel 提供多通道消息适配器
package channel

import (
	"context"
	"time"
)

// Message 统一消息格式
type Message struct {
	ID          string            // 消息 ID
	ChannelType string            // 通道类型: feishu, telegram, slack, etc.
	ChannelID   string            // 具体通道 ID (群组 ID 或用户 ID)
	SenderID    string            // 发送者 ID
	SenderName  string            // 发送者名称
	Content     string            // 消息内容（纯文本）
	ReplyTo     string            // 回复的消息 ID（可选）
	Mentions    []string          // @提到的用户
	Attachments []Attachment      // 附件列表
	Metadata    map[string]string // 原始元数据
	ReceivedAt  time.Time         // 接收时间
}

// Attachment 附件
type Attachment struct {
	ID       string // 附件 ID
	Name     string // 文件名
	Type     string // MIME 类型
	Size     int64  // 文件大小
	URL      string // 下载 URL
	LocalPath string // 本地缓存路径
}

// SendRequest 发送消息请求
type SendRequest struct {
	ChannelID   string       // 目标通道
	Content     string       // 消息内容
	ReplyTo     string       // 回复消息 ID（可选）
	Attachments []Attachment // 附件（可选）
}

// SendResponse 发送消息响应
type SendResponse struct {
	MessageID string // 发送成功后的消息 ID
}

// Channel 通道接口
type Channel interface {
	// Type 返回通道类型
	Type() string

	// Start 启动通道（开始接收消息）
	Start(ctx context.Context) error

	// Stop 停止通道
	Stop() error

	// Send 发送消息
	Send(ctx context.Context, req *SendRequest) (*SendResponse, error)

	// Messages 返回消息接收通道
	Messages() <-chan *Message
}

// MessageHandler 消息处理器
type MessageHandler func(ctx context.Context, msg *Message) error

// Config 通道配置
type Config struct {
	Type    string            `json:"type"`    // feishu, telegram, slack
	Enabled bool              `json:"enabled"` // 是否启用
	Options map[string]string `json:"options"` // 通道特定配置
}
