// Package feishu 飞书通道适配器
package feishu

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
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
	log = logger.Module("feishu")
}

// Config 飞书配置
type Config struct {
	Name              string `json:"name"`               // 配置名称
	AppID             string `json:"app_id"`
	AppSecret         string `json:"app_secret"`
	VerificationToken string `json:"verification_token"` // 用于验证事件来源
	EncryptKey        string `json:"encrypt_key"`        // 消息加密密钥（可选）
	BotName           string `json:"bot_name"`           // 机器人名称，用于检测 @
	DefaultAgentID    string `json:"default_agent_id"`   // 默认处理消息的 Agent
}

// Channel 飞书通道
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

// New 创建飞书通道
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
	return "feishu"
}

// Start 启动通道（飞书使用 webhook 推送，此方法仅初始化）
func (c *Channel) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	// 刷新 access token
	if err := c.refreshAccessToken(); err != nil {
		return fmt.Errorf("refresh access token: %w", err)
	}

	// 定时刷新 token
	go c.tokenRefreshLoop()

	log.Info("feishu channel started", "app_id", c.config.AppID)
	return nil
}

// Stop 停止通道
func (c *Channel) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	close(c.messages)
	log.Info("feishu channel stopped")
	return nil
}

// Send 发送消息
func (c *Channel) Send(ctx context.Context, req *channel.SendRequest) (*channel.SendResponse, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	// 构造消息体
	msgReq := map[string]interface{}{
		"receive_id": req.ChannelID,
		"msg_type":   "text",
		"content":    fmt.Sprintf(`{"text":"%s"}`, escapeJSON(req.Content)),
	}

	body, _ := json.Marshal(msgReq)

	// 发送请求
	httpReq, _ := http.NewRequestWithContext(ctx, "POST",
		"https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=chat_id",
		bytes.NewReader(body))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			MessageID string `json:"message_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("feishu api error: %d %s", result.Code, result.Msg)
	}

	return &channel.SendResponse{MessageID: result.Data.MessageID}, nil
}

// Messages 返回消息通道
func (c *Channel) Messages() <-chan *channel.Message {
	return c.messages
}

// GetConfig 获取配置（用于获取 default_agent_id 等）
func (c *Channel) GetConfig() *Config {
	return c.config
}

// HandleWebhook 处理飞书 webhook 回调（需要在 HTTP 路由中注册）
func (c *Channel) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// 尝试解密（如果配置了加密）
	var payload []byte
	if c.config.EncryptKey != "" {
		var encrypted struct {
			Encrypt string `json:"encrypt"`
		}
		if err := json.Unmarshal(body, &encrypted); err == nil && encrypted.Encrypt != "" {
			payload, err = c.decrypt(encrypted.Encrypt)
			if err != nil {
				log.Error("decrypt failed", "error", err)
				http.Error(w, "decrypt failed", http.StatusBadRequest)
				return
			}
		} else {
			payload = body
		}
	} else {
		payload = body
	}

	// 解析事件
	var event struct {
		Challenge string `json:"challenge"` // URL 验证
		Token     string `json:"token"`     // 验证 token
		Type      string `json:"type"`      // 事件类型
		Schema    string `json:"schema"`    // 2.0 schema
		Header    struct {
			EventType string `json:"event_type"`
			Token     string `json:"token"`
		} `json:"header"`
		Event json.RawMessage `json:"event"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Error("parse event failed", "error", err)
		http.Error(w, "parse failed", http.StatusBadRequest)
		return
	}

	// URL 验证挑战
	if event.Challenge != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"challenge": event.Challenge})
		return
	}

	// 验证 token
	token := event.Token
	if event.Header.Token != "" {
		token = event.Header.Token
	}
	if c.config.VerificationToken != "" && token != c.config.VerificationToken {
		log.Warn("invalid verification token")
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// 处理消息事件
	eventType := event.Type
	if event.Header.EventType != "" {
		eventType = event.Header.EventType
	}

	if eventType == "im.message.receive_v1" {
		c.handleMessage(event.Event)
	}

	w.WriteHeader(http.StatusOK)
}

// handleMessage 处理消息事件
func (c *Channel) handleMessage(eventData json.RawMessage) {
	var msgEvent struct {
		Sender struct {
			SenderID struct {
				UserID  string `json:"user_id"`
				OpenID  string `json:"open_id"`
				UnionID string `json:"union_id"`
			} `json:"sender_id"`
			SenderType string `json:"sender_type"`
		} `json:"sender"`
		Message struct {
			MessageID   string `json:"message_id"`
			RootID      string `json:"root_id"`
			ParentID    string `json:"parent_id"`
			ChatID      string `json:"chat_id"`
			ChatType    string `json:"chat_type"` // p2p 或 group
			MessageType string `json:"message_type"`
			Content     string `json:"content"`
			Mentions    []struct {
				Key  string `json:"key"`
				ID   struct {
					UserID  string `json:"user_id"`
					OpenID  string `json:"open_id"`
					UnionID string `json:"union_id"`
				} `json:"id"`
				Name string `json:"name"`
			} `json:"mentions"`
		} `json:"message"`
	}

	if err := json.Unmarshal(eventData, &msgEvent); err != nil {
		log.Error("parse message event failed", "error", err)
		return
	}

	// 解析消息内容
	var content struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(msgEvent.Message.Content), &content); err != nil {
		log.Warn("parse message content failed", "error", err, "content", msgEvent.Message.Content)
		content.Text = msgEvent.Message.Content
	}

	// 检查是否 @ 了机器人（群组消息）
	isMentioned := false
	mentions := make([]string, 0)
	for _, m := range msgEvent.Message.Mentions {
		mentions = append(mentions, m.Name)
		if strings.Contains(m.Name, c.config.BotName) {
			isMentioned = true
		}
	}

	// 群组消息需要 @机器人 才处理
	if msgEvent.Message.ChatType == "group" && !isMentioned {
		return
	}

	// 去掉 @机器人 的部分
	text := content.Text
	for _, m := range msgEvent.Message.Mentions {
		text = strings.ReplaceAll(text, "@"+m.Name, "")
	}
	text = strings.TrimSpace(text)

	if text == "" {
		return
	}

	msg := &channel.Message{
		ID:          msgEvent.Message.MessageID,
		ChannelType: "feishu",
		ChannelID:   msgEvent.Message.ChatID,
		SenderID:    msgEvent.Sender.SenderID.OpenID,
		Content:     text,
		ReplyTo:     msgEvent.Message.ParentID,
		Mentions:    mentions,
		ReceivedAt:  time.Now(),
		Metadata: map[string]string{
			"chat_type":    msgEvent.Message.ChatType,
			"message_type": msgEvent.Message.MessageType,
		},
	}

	select {
	case c.messages <- msg:
		log.Debug("message received", "id", msg.ID, "content", msg.Content)
	default:
		log.Warn("message channel full, dropping message", "id", msg.ID)
	}
}

// refreshAccessToken 刷新访问令牌
func (c *Channel) refreshAccessToken() error {
	reqBody, _ := json.Marshal(map[string]string{
		"app_id":     c.config.AppID,
		"app_secret": c.config.AppSecret,
	})

	resp, err := c.client.Post(
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal",
		"application/json",
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
		Expire            int    `json:"expire"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if result.Code != 0 {
		return fmt.Errorf("get token failed: %d %s", result.Code, result.Msg)
	}

	c.tokenMu.Lock()
	c.accessToken = result.TenantAccessToken
	c.tokenExpireAt = time.Now().Add(time.Duration(result.Expire-60) * time.Second) // 提前 60 秒刷新
	c.tokenMu.Unlock()

	log.Debug("access token refreshed", "expire_in", result.Expire)
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

// decrypt AES 解密（飞书加密格式）
func (c *Channel) decrypt(encrypted string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}

	// 飞书使用 SHA256(key) 作为 AES 密钥
	keyHash := sha256.Sum256([]byte(c.config.EncryptKey))
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCBCDecrypter(block, iv)
	stream.CryptBlocks(ciphertext, ciphertext)

	// 去除 PKCS7 padding
	padding := int(ciphertext[len(ciphertext)-1])
	if padding > aes.BlockSize || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	return ciphertext[:len(ciphertext)-padding], nil
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
