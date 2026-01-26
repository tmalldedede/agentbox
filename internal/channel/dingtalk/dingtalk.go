// Package dingtalk 钉钉通道适配器
package dingtalk

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tmalldedede/agentbox/internal/channel"
	"github.com/tmalldedede/agentbox/internal/logger"
)

var log *slog.Logger

func init() {
	log = logger.Module("dingtalk")
}

// Config 钉钉配置
type Config struct {
	Name           string `json:"name"`             // 配置名称
	AppKey         string `json:"app_key"`          // 应用 AppKey
	AppSecret      string `json:"app_secret"`       // 应用 AppSecret
	AgentID        int64  `json:"agent_id"`         // 应用 AgentID
	RobotCode      string `json:"robot_code"`       // 机器人 Code
	DefaultAgentID string `json:"default_agent_id"` // 默认处理消息的 AgentBox Agent
}

// Channel 钉钉通道
type Channel struct {
	config   *Config
	messages chan *channel.Message

	// 访问令牌
	accessToken   string
	tokenExpireAt time.Time
	tokenMu       sync.RWMutex

	// HTTP 客户端
	client *http.Client

	// 运行状态
	ctx    context.Context
	cancel context.CancelFunc
}

// New 创建钉钉通道
func New(cfg *Config) *Channel {
	return &Channel{
		config:   cfg,
		messages: make(chan *channel.Message, 100),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Type 返回通道类型
func (c *Channel) Type() string {
	return "dingtalk"
}

// Start 启动通道
func (c *Channel) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	// 刷新 access token
	if err := c.refreshAccessToken(); err != nil {
		return fmt.Errorf("refresh access token: %w", err)
	}

	// 定时刷新 token
	go c.tokenRefreshLoop()

	log.Info("dingtalk channel started", "app_key", c.config.AppKey, "agent_id", c.config.AgentID)
	return nil
}

// Stop 停止通道
func (c *Channel) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	close(c.messages)
	log.Info("dingtalk channel stopped")
	return nil
}

// Send 发送消息
func (c *Channel) Send(ctx context.Context, req *channel.SendRequest) (*channel.SendResponse, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	// 判断是群聊还是单聊
	var apiURL string
	var msgReq map[string]interface{}

	if strings.HasPrefix(req.ChannelID, "cid") {
		// 群聊消息
		apiURL = "https://api.dingtalk.com/v1.0/robot/groupMessages/send"
		msgReq = map[string]interface{}{
			"robotCode":              c.config.RobotCode,
			"openConversationId":     req.ChannelID,
			"msgKey":                 "sampleText",
			"msgParam":               fmt.Sprintf(`{"content":"%s"}`, escapeJSON(req.Content)),
		}
	} else {
		// 单聊消息
		apiURL = "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
		msgReq = map[string]interface{}{
			"robotCode": c.config.RobotCode,
			"userIds":   []string{req.ChannelID},
			"msgKey":    "sampleText",
			"msgParam":  fmt.Sprintf(`{"content":"%s"}`, escapeJSON(req.Content)),
		}
	}

	body, _ := json.Marshal(msgReq)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-acs-dingtalk-access-token", token)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ProcessQueryKey string `json:"processQueryKey"`
		// 错误响应
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if result.Code != "" {
		return nil, fmt.Errorf("dingtalk api error: %s %s", result.Code, result.Message)
	}

	return &channel.SendResponse{MessageID: result.ProcessQueryKey}, nil
}

// Messages 返回消息通道
func (c *Channel) Messages() <-chan *channel.Message {
	return c.messages
}

// GetConfig 获取配置
func (c *Channel) GetConfig() *Config {
	return c.config
}

// HandleWebhook 处理钉钉 webhook 回调
func (c *Channel) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// 验证签名
	timestamp := r.Header.Get("timestamp")
	sign := r.Header.Get("sign")
	if !c.verifySignature(timestamp, sign) {
		log.Warn("invalid signature")
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// 解析消息
	var event struct {
		// 消息类型
		MsgType string `json:"msgtype"`
		// 会话类型
		ConversationType string `json:"conversationType"` // 1: 单聊, 2: 群聊
		// 消息内容
		Text struct {
			Content string `json:"content"`
		} `json:"text"`
		// 消息 ID
		MsgID string `json:"msgId"`
		// 发送者
		SenderID       string `json:"senderStaffId"`
		SenderNick     string `json:"senderNick"`
		SenderCorpID   string `json:"senderCorpId"`
		// 会话 ID
		ConversationID string `json:"conversationId"`
		// 创建时间
		CreateAt int64 `json:"createAt"`
		// @机器人信息
		AtUsers []struct {
			DingtalkID string `json:"dingtalkId"`
		} `json:"atUsers"`
		IsAdmin bool `json:"isAdmin"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		log.Error("parse event failed", "error", err)
		http.Error(w, "parse failed", http.StatusBadRequest)
		return
	}

	// 只处理文本消息
	if event.MsgType != "text" {
		w.WriteHeader(http.StatusOK)
		return
	}

	content := strings.TrimSpace(event.Text.Content)
	if content == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 确定聊天类型
	isGroup := event.ConversationType == "2"
	chatID := event.ConversationID
	if !isGroup {
		chatID = event.SenderID
	}

	msg := &channel.Message{
		ID:          event.MsgID,
		ChannelType: "dingtalk",
		ChannelID:   chatID,
		SenderID:    event.SenderID,
		SenderName:  event.SenderNick,
		Content:     content,
		ReceivedAt:  time.UnixMilli(event.CreateAt),
		Metadata: map[string]string{
			"msg_type":        event.MsgType,
			"chat_type":       map[bool]string{true: "group", false: "single"}[isGroup],
			"conversation_id": event.ConversationID,
		},
	}

	select {
	case c.messages <- msg:
		log.Debug("message received", "id", msg.ID, "content", msg.Content)
	default:
		log.Warn("message channel full, dropping message", "id", msg.ID)
	}

	w.WriteHeader(http.StatusOK)
}

// verifySignature 验证签名
func (c *Channel) verifySignature(timestamp, sign string) bool {
	if timestamp == "" || sign == "" {
		return false
	}

	// 计算签名
	stringToSign := timestamp + "\n" + c.config.AppSecret
	h := hmac.New(sha256.New, []byte(c.config.AppSecret))
	h.Write([]byte(stringToSign))
	expectedSign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return sign == expectedSign
}

// refreshAccessToken 刷新访问令牌
func (c *Channel) refreshAccessToken() error {
	reqBody, _ := json.Marshal(map[string]string{
		"appKey":    c.config.AppKey,
		"appSecret": c.config.AppSecret,
	})

	resp, err := c.client.Post(
		"https://api.dingtalk.com/v1.0/oauth2/accessToken",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"accessToken"`
		ExpireIn    int    `json:"expireIn"`
		// 错误响应
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if result.Code != "" {
		return fmt.Errorf("get token failed: %s %s", result.Code, result.Message)
	}

	c.tokenMu.Lock()
	c.accessToken = result.AccessToken
	c.tokenExpireAt = time.Now().Add(time.Duration(result.ExpireIn-60) * time.Second)
	c.tokenMu.Unlock()

	log.Debug("access token refreshed", "expire_in", result.ExpireIn)
	return nil
}

// getAccessToken 获取访问令牌
func (c *Channel) getAccessToken() (string, error) {
	c.tokenMu.RLock()
	if time.Now().Before(c.tokenExpireAt) {
		token := c.accessToken
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()

	if err := c.refreshAccessToken(); err != nil {
		return "", err
	}

	c.tokenMu.RLock()
	token := c.accessToken
	c.tokenMu.RUnlock()
	return token, nil
}

// tokenRefreshLoop 定时刷新 token
func (c *Channel) tokenRefreshLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			if err := c.refreshAccessToken(); err != nil {
				log.Error("refresh token failed", "error", err)
			}
		}
	}
}

// escapeJSON 转义 JSON 字符串
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}
