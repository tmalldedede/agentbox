package files

import (
	"time"
)

// FileType 文件类型
type FileType string

const (
	FileTypeFile      FileType = "file"
	FileTypeDirectory FileType = "directory"
)

// FileInfo 文件信息
type FileInfo struct {
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	Type          FileType  `json:"type"`
	Size          int64     `json:"size,omitempty"`
	ModifiedAt    time.Time `json:"modified_at,omitempty"`
	ChildrenCount int       `json:"children_count,omitempty"` // 仅目录
}

// FileList 文件列表响应
type FileList struct {
	Path  string     `json:"path"`
	Files []FileInfo `json:"files"`
}

// FileContent 文件内容响应
type FileContent struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Size     int64  `json:"size"`
	Encoding string `json:"encoding"` // utf-8, base64
	Truncated bool  `json:"truncated"`
}

// FileUploadResult 文件上传结果
type FileUploadResult struct {
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// ListFilesRequest 列出文件请求
type ListFilesRequest struct {
	Path      string `form:"path"`      // 目录路径，默认 "/"
	Recursive bool   `form:"recursive"` // 是否递归
}

// ReadContentRequest 读取内容请求
type ReadContentRequest struct {
	MaxSize int64 `form:"max_size"` // 最大读取字节数，默认 1MB
}

// CreateDirectoryRequest 创建目录请求
type CreateDirectoryRequest struct {
	Path string `form:"path" binding:"required"` // 目录路径
}
