package main

import (
	"time"
)

// Config 配置
type Config struct {
	BaseURL  string        // http://localhost:18080
	APIKey   string        // ab_xxx 或 JWT
	AgentID  string        // phishing-analyzer
	Workers  int           // 并行数
	Timeout  time.Duration // 单任务超时
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	File      string        `json:"file"`       // 文件名
	TaskID    string        `json:"task_id"`    // AgentBox Task ID
	Status    string        `json:"status"`     // completed/failed
	RiskLevel string        `json:"risk_level"` // critical/high/medium/low/safe
	RiskScore int           `json:"risk_score"` // 0-100
	Threats   []string      `json:"threats"`    // 威胁类型列表
	IOCs      []IOC         `json:"iocs"`       // 提取的 IOC
	Summary   string        `json:"summary"`    // AI 分析摘要
	Duration  time.Duration `json:"duration"`   // 分析耗时
	Error     string        `json:"error"`      // 错误信息（如有）
}

// IOC 威胁指标
type IOC struct {
	Type  string `json:"type"`  // url/domain/ip/hash/email
	Value string `json:"value"` // IOC 值
	Risk  string `json:"risk"`  // malicious/suspicious/safe
}

// BatchReport 批量报告
type BatchReport struct {
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
	TotalFiles  int               `json:"total_files"`
	Completed   int               `json:"completed"`
	Failed      int               `json:"failed"`
	RiskSummary map[string]int    `json:"risk_summary"` // {critical: 2, high: 5, ...}
	Results     []AnalysisResult  `json:"results"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
	Uptime  string `json:"uptime,omitempty"`
}

// UploadedFile 上传文件响应
type UploadedFile struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mime_type"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// Task 任务
type Task struct {
	ID         string                 `json:"id"`
	AgentID    string                 `json:"agent_id"`
	Status     string                 `json:"status"`
	Prompt     string                 `json:"prompt,omitempty"`
	Result     map[string]interface{} `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	TurnCount  int                    `json:"turn_count"`
	CreatedAt  time.Time              `json:"created_at"`
	StartedAt  *time.Time             `json:"started_at,omitempty"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	AgentID     string   `json:"agent_id"`
	Prompt      string   `json:"prompt"`
	Attachments []string `json:"attachments,omitempty"`
	Timeout     int      `json:"timeout,omitempty"`
}

// APIResponse 通用 API 响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// SSEEvent SSE 事件
type SSEEvent struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}
