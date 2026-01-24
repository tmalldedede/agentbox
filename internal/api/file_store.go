package api

import (
	"time"
)

// FilePurpose 文件用途
type FilePurpose string

const (
	FilePurposeAttachment FilePurpose = "attachment" // Task 输入附件
	FilePurposeOutput     FilePurpose = "output"     // Task 输出文件
	FilePurposeGeneral    FilePurpose = "general"    // 通用上传
)

// FileStatus 文件状态
type FileStatus string

const (
	FileStatusActive  FileStatus = "active"
	FileStatusExpired FileStatus = "expired"
	FileStatusDeleted FileStatus = "deleted"
)

// FileRecord 文件记录（数据库模型）
type FileRecord struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Size      int64       `json:"size"`
	MimeType  string      `json:"mime_type"`
	Path      string      `json:"path"`       // 文件在磁盘上的完整路径
	TaskID    string      `json:"task_id"`    // 关联的 Task ID（可为空）
	Purpose   FilePurpose `json:"purpose"`    // 文件用途
	Status    FileStatus  `json:"status"`     // 文件状态
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at,omitempty"`
}

// FileStore 文件存储接口
type FileStore interface {
	Create(record *FileRecord) error
	Get(id string) (*FileRecord, error)
	List(filter *FileListFilter) ([]*FileRecord, error)
	UpdateStatus(id string, status FileStatus) error
	BindTask(id string, taskID string, purpose FilePurpose) error
	Delete(id string) error
	ListExpired(before time.Time) ([]*FileRecord, error)
	ListByTask(taskID string) ([]*FileRecord, error)
	Close() error
}

// FileListFilter 文件列表过滤器
type FileListFilter struct {
	TaskID  string
	Purpose FilePurpose
	Status  FileStatus
	Limit   int
	Offset  int
}
