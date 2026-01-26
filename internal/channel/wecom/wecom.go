// Package wecom 企业微信通道适配器
package wecom

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tmalldedede/agentbox/internal/channel"
	"github.com/tmalldedede/agentbox/internal/logger"
)

var log *slog.Logger

func init() {
	log = logger.Module("wecom")
}

// Config 企业微信配置
type Config struct {
	Name           string `json:"name"`             // 配置名称
	CorpID         string `json:"corp_id"`          // 企业 ID
	AgentID        int    `json:"agent_id"`         // 应用 AgentID
	Secret         string `json:"secret"`           // 应用 Secret
	Token          string `json:"token"`            // 回调 Token
	EncodingAESKey string `json:"encoding_aes_key"` // 回调 EncodingAESKey
	DefaultAgentID string `json:"default_agent_id"` // 默认处理消息的 AgentBox Agent
}

// Channel 企业微信通道
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

// New 创建企业微信通道
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
	return "wecom"
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

	log.Info("wecom channel started", "corp_id", c.config.CorpID, "agent_id", c.config.AgentID)
	return nil
}

// Stop 停止通道
func (c *Channel) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	close(c.messages)
	log.Info("wecom channel stopped")
	return nil
}

// Send 发送消息
func (c *Channel) Send(ctx context.Context, req *channel.SendRequest) (*channel.SendResponse, error) {
	token, err := c.getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get access token: %w", err)
	}

	// 判断是发送给用户还是群聊
	var msgReq map[string]interface{}
	if strings.HasPrefix(req.ChannelID, "ww") {
		// 群聊 ID 通常以 ww 开头
		msgReq = map[string]interface{}{
			"chatid":  req.ChannelID,
			"msgtype": "text",
			"text": map[string]string{
				"content": req.Content,
			},
		}
	} else {
		// 发送给用户
		msgReq = map[string]interface{}{
			"touser":  req.ChannelID,
			"msgtype": "text",
			"agentid": c.config.AgentID,
			"text": map[string]string{
				"content": req.Content,
			},
		}
	}

	body, _ := json.Marshal(msgReq)

	// 选择 API
	apiURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)
	if strings.HasPrefix(req.ChannelID, "ww") {
		apiURL = fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/appchat/send?access_token=%s", token)
	}

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		MsgID   string `json:"msgid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wecom api error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return &channel.SendResponse{MessageID: result.MsgID}, nil
}

// Messages 返回消息通道
func (c *Channel) Messages() <-chan *channel.Message {
	return c.messages
}

// GetConfig 获取配置
func (c *Channel) GetConfig() *Config {
	return c.config
}

// HandleWebhook 处理企业微信 webhook 回调
func (c *Channel) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// URL 验证
	if r.Method == "GET" {
		c.handleVerify(w, r)
		return
	}

	// 消息处理
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}

	// 验证签名
	msgSignature := r.URL.Query().Get("msg_signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")

	if !c.verifySignature(msgSignature, timestamp, nonce, string(body)) {
		log.Warn("invalid signature")
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// 解析消息
	var xmlMsg struct {
		XMLName    xml.Name `xml:"xml"`
		ToUserName string   `xml:"ToUserName"`
		Encrypt    string   `xml:"Encrypt"`
	}
	if err := xml.Unmarshal(body, &xmlMsg); err != nil {
		log.Error("parse xml failed", "error", err)
		http.Error(w, "parse failed", http.StatusBadRequest)
		return
	}

	// 解密消息
	decrypted, err := c.decryptMessage(xmlMsg.Encrypt)
	if err != nil {
		log.Error("decrypt message failed", "error", err)
		http.Error(w, "decrypt failed", http.StatusBadRequest)
		return
	}

	// 解析解密后的消息
	c.handleMessage(decrypted)

	w.WriteHeader(http.StatusOK)
}

// handleVerify 处理 URL 验证
func (c *Channel) handleVerify(w http.ResponseWriter, r *http.Request) {
	msgSignature := r.URL.Query().Get("msg_signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	echostr := r.URL.Query().Get("echostr")

	if !c.verifySignature(msgSignature, timestamp, nonce, echostr) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	// 解密 echostr
	decrypted, err := c.decryptMessage(echostr)
	if err != nil {
		http.Error(w, "decrypt failed", http.StatusBadRequest)
		return
	}

	w.Write(decrypted)
}

// handleMessage 处理消息
func (c *Channel) handleMessage(data []byte) {
	var msgEvent struct {
		XMLName      xml.Name `xml:"xml"`
		MsgType      string   `xml:"MsgType"`
		MsgID        string   `xml:"MsgId"`
		Content      string   `xml:"Content"`
		FromUserName string   `xml:"FromUserName"`
		CreateTime   int64    `xml:"CreateTime"`
		AgentID      int      `xml:"AgentID"`
		// 群聊相关
		ChatID string `xml:"ChatId"`
	}

	if err := xml.Unmarshal(data, &msgEvent); err != nil {
		log.Error("parse message failed", "error", err)
		return
	}

	// 只处理文本消息
	if msgEvent.MsgType != "text" {
		return
	}

	content := strings.TrimSpace(msgEvent.Content)
	if content == "" {
		return
	}

	// 确定聊天 ID（群聊用 ChatID，私聊用 FromUserName）
	chatID := msgEvent.FromUserName
	isGroup := false
	if msgEvent.ChatID != "" {
		chatID = msgEvent.ChatID
		isGroup = true
	}

	msg := &channel.Message{
		ID:          msgEvent.MsgID,
		ChannelType: "wecom",
		ChannelID:   chatID,
		SenderID:    msgEvent.FromUserName,
		Content:     content,
		ReceivedAt:  time.Unix(msgEvent.CreateTime, 0),
		Metadata: map[string]string{
			"msg_type":  msgEvent.MsgType,
			"chat_type": map[bool]string{true: "group", false: "single"}[isGroup],
		},
	}

	select {
	case c.messages <- msg:
		log.Debug("message received", "id", msg.ID, "content", msg.Content)
	default:
		log.Warn("message channel full, dropping message", "id", msg.ID)
	}
}

// verifySignature 验证签名
func (c *Channel) verifySignature(msgSignature, timestamp, nonce, data string) bool {
	params := []string{c.config.Token, timestamp, nonce, data}
	sort.Strings(params)
	str := strings.Join(params, "")
	hash := sha1.Sum([]byte(str))
	signature := fmt.Sprintf("%x", hash)
	return signature == msgSignature
}

// decryptMessage 解密消息（简化实现，生产环境需要完整的 AES 解密）
func (c *Channel) decryptMessage(encrypted string) ([]byte, error) {
	// TODO: 实现完整的企业微信消息解密
	// 这里需要使用 EncodingAESKey 进行 AES 解密
	// 参考: https://developer.work.weixin.qq.com/document/path/90968
	return []byte(encrypted), nil
}

// refreshAccessToken 刷新访问令牌
func (c *Channel) refreshAccessToken() error {
	url := fmt.Sprintf(
		"https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		c.config.CorpID, c.config.Secret,
	)

	resp, err := c.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("get token failed: %d %s", result.ErrCode, result.ErrMsg)
	}

	c.tokenMu.Lock()
	c.accessToken = result.AccessToken
	c.tokenExpireAt = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)
	c.tokenMu.Unlock()

	log.Debug("access token refreshed", "expire_in", result.ExpiresIn)
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
