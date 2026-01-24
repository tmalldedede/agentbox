package settings

import "time"

// Settings 业务配置（可运行时修改，存储在数据库）
type Settings struct {
	Agent   AgentSettings   `json:"agent"`
	Task    TaskSettings    `json:"task"`
	Batch   BatchSettings   `json:"batch"`
	Storage StorageSettings `json:"storage"`
	Notify  NotifySettings  `json:"notify"`
}

// AgentSettings Agent 默认配置
type AgentSettings struct {
	DefaultProviderID string `json:"default_provider_id"` // 默认 Provider ID
	DefaultModel      string `json:"default_model"`       // 默认模型
	DefaultRuntimeID  string `json:"default_runtime_id"`  // 默认 Runtime ID
	DefaultTimeout    int    `json:"default_timeout"`     // 默认超时（秒）
	SystemPrompt      string `json:"system_prompt"`       // 默认 System Prompt
}

// TaskSettings Task 配置
type TaskSettings struct {
	DefaultIdleTimeout int `json:"default_idle_timeout"` // 默认 Idle Timeout（秒）
	DefaultPollInterval int `json:"default_poll_interval"` // 默认轮询间隔（毫秒）
	MaxTurns           int `json:"max_turns"`             // 最大对话轮次
	MaxAttachments     int `json:"max_attachments"`       // 最大附件数
	MaxAttachmentSize  int `json:"max_attachment_size"`   // 单个附件最大大小（MB）
}

// BatchSettings Batch 配置
type BatchSettings struct {
	DefaultWorkers    int `json:"default_workers"`     // 默认 Worker 数量
	MaxWorkers        int `json:"max_workers"`         // 最大 Worker 数量
	MaxConcurrentBatches int `json:"max_concurrent_batches"` // 最大并行 Batch 数
	MaxRetries        int `json:"max_retries"`         // 最大重试次数
	RetryDelay        int `json:"retry_delay"`         // 重试延迟（秒）
	DeadLetterEnabled bool `json:"dead_letter_enabled"` // 是否启用 Dead Letter Queue
}

// StorageSettings 存储配置
type StorageSettings struct {
	HistoryRetentionDays int  `json:"history_retention_days"` // 历史记录保留天数
	SessionRetentionDays int  `json:"session_retention_days"` // Session 保留天数
	AutoCleanup          bool `json:"auto_cleanup"`           // 是否自动清理
}

// NotifySettings 通知配置
type NotifySettings struct {
	WebhookURL       string `json:"webhook_url"`        // Webhook URL
	WebhookSecret    string `json:"webhook_secret"`     // Webhook 签名密钥
	NotifyOnComplete bool   `json:"notify_on_complete"` // 任务完成时通知
	NotifyOnFailed   bool   `json:"notify_on_failed"`   // 任务失败时通知
	NotifyOnBatchComplete bool `json:"notify_on_batch_complete"` // Batch 完成时通知
}

// SettingItem 单个配置项（数据库存储格式）
type SettingItem struct {
	Key       string    `json:"key" gorm:"primaryKey"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Default 返回默认业务配置
func Default() *Settings {
	return &Settings{
		Agent: AgentSettings{
			DefaultProviderID: "",
			DefaultModel:      "",
			DefaultRuntimeID:  "",
			DefaultTimeout:    3600, // 1 hour
			SystemPrompt:      "",
		},
		Task: TaskSettings{
			DefaultIdleTimeout:  30,    // 30 seconds
			DefaultPollInterval: 500,   // 500ms
			MaxTurns:            100,
			MaxAttachments:      10,
			MaxAttachmentSize:   100,   // 100MB
		},
		Batch: BatchSettings{
			DefaultWorkers:       5,
			MaxWorkers:           50,
			MaxConcurrentBatches: 10,
			MaxRetries:           3,
			RetryDelay:           5,
			DeadLetterEnabled:    true,
		},
		Storage: StorageSettings{
			HistoryRetentionDays: 30,
			SessionRetentionDays: 7,
			AutoCleanup:          true,
		},
		Notify: NotifySettings{
			WebhookURL:            "",
			WebhookSecret:         "",
			NotifyOnComplete:      false,
			NotifyOnFailed:        true,
			NotifyOnBatchComplete: true,
		},
	}
}
