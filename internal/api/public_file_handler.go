package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/apperr"
	"github.com/tmalldedede/agentbox/internal/config"
)

// UploadedFile 已上传的文件信息（API 响应）
type UploadedFile struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Size       int64       `json:"size"`
	MimeType   string      `json:"mime_type"`
	TaskID     string      `json:"task_id,omitempty"`
	Purpose    FilePurpose `json:"purpose"`
	UploadedAt time.Time   `json:"uploaded_at"`
	ExpiresAt  time.Time   `json:"expires_at,omitempty"`
}

// PublicFileHandler 独立文件上传处理器
type PublicFileHandler struct {
	uploadDir      string
	retentionHours int
	maxFileSize    int64
	store          FileStore
	cleanupStop    chan struct{}
}

// NewPublicFileHandler 创建独立文件处理器
func NewPublicFileHandler(cfg config.FilesConfig, store FileStore) *PublicFileHandler {
	os.MkdirAll(cfg.UploadDir, 0755)
	return &PublicFileHandler{
		uploadDir:      cfg.UploadDir,
		retentionHours: cfg.RetentionHours,
		maxFileSize:    cfg.MaxFileSize,
		store:          store,
		cleanupStop:    make(chan struct{}),
	}
}

// StartCleanup 启动过期文件清理 goroutine
func (h *PublicFileHandler) StartCleanup(interval time.Duration) {
	if h.retentionHours <= 0 {
		return // 0 表示永不过期
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.cleanupExpiredFiles()
			case <-h.cleanupStop:
				return
			}
		}
	}()
}

// Stop 停止清理 goroutine
func (h *PublicFileHandler) Stop() {
	close(h.cleanupStop)
}

// cleanupExpiredFiles 删除过期且未被 task 引用的文件
func (h *PublicFileHandler) cleanupExpiredFiles() {
	expired, err := h.store.ListExpired(time.Now())
	if err != nil {
		log.Error("file cleanup: failed to list expired", "error", err)
		return
	}

	removed := 0
	for _, record := range expired {
		// 有关联 task 的文件不删除
		if record.TaskID != "" {
			continue
		}

		// 删除磁盘文件
		fileDir := filepath.Join(h.uploadDir, record.ID)
		os.RemoveAll(fileDir)

		// 更新数据库状态
		h.store.UpdateStatus(record.ID, FileStatusExpired)
		removed++
	}

	if removed > 0 {
		log.Info("file cleanup completed", "removed", removed, "retention_hours", h.retentionHours)
	}
}

// RegisterRoutes 注册路由
func (h *PublicFileHandler) RegisterRoutes(r *gin.RouterGroup) {
	files := r.Group("/files")
	{
		files.GET("", h.List)
		files.POST("", h.Upload)
		files.GET("/:id", h.Get)
		files.DELETE("/:id", h.Delete)
		files.GET("/:id/download", h.Download)
	}
}

// List 列出已上传的文件
// GET /api/v1/files
func (h *PublicFileHandler) List(c *gin.Context) {
	records, err := h.store.List(&FileListFilter{Status: FileStatusActive})
	if err != nil {
		HandleError(c, apperr.Wrap(err, "failed to list files"))
		return
	}

	files := make([]UploadedFile, 0, len(records))
	for _, r := range records {
		files = append(files, recordToUploadedFile(r))
	}

	Success(c, files)
}

// Upload 上传文件
// POST /api/v1/files
func (h *PublicFileHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	// 检查文件大小
	if h.maxFileSize > 0 && header.Size > h.maxFileSize {
		BadRequest(c, fmt.Sprintf("file too large: %d bytes (max %d)", header.Size, h.maxFileSize))
		return
	}

	// 生成唯一 ID
	fileID := uuid.New().String()

	// 创建文件目录
	fileDir := filepath.Join(h.uploadDir, fileID)
	if err := os.MkdirAll(fileDir, 0755); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to create upload directory"))
		return
	}

	// 保存文件到磁盘
	filePath := filepath.Join(fileDir, header.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		os.RemoveAll(fileDir)
		HandleError(c, apperr.Wrap(err, "failed to create file"))
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		os.RemoveAll(fileDir)
		HandleError(c, apperr.Wrap(err, "failed to write file"))
		return
	}

	// 检测 MIME 类型
	mimeType := getMimeType(header.Filename)
	if mimeType == "application/octet-stream" {
		if ct := header.Header.Get("Content-Type"); ct != "" && ct != "application/octet-stream" {
			mimeType = ct
		}
	}

	now := time.Now()
	var expiresAt time.Time
	if h.retentionHours > 0 {
		expiresAt = now.Add(time.Duration(h.retentionHours) * time.Hour)
	}

	// 写入数据库
	record := &FileRecord{
		ID:        fileID,
		Name:      header.Filename,
		Size:      written,
		MimeType:  mimeType,
		Path:      filePath,
		Purpose:   FilePurposeGeneral,
		Status:    FileStatusActive,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	if err := h.store.Create(record); err != nil {
		os.RemoveAll(fileDir)
		HandleError(c, apperr.Wrap(err, "failed to save file record"))
		return
	}

	Created(c, recordToUploadedFile(record))
}

// Get 获取文件信息
// GET /api/v1/files/:id
func (h *PublicFileHandler) Get(c *gin.Context) {
	fileID := c.Param("id")

	record, err := h.store.Get(fileID)
	if err != nil {
		HandleError(c, apperr.NotFound("file"))
		return
	}

	if record.Status != FileStatusActive {
		HandleError(c, apperr.NotFound("file"))
		return
	}

	Success(c, recordToUploadedFile(record))
}

// Delete 删除文件
// DELETE /api/v1/files/:id
func (h *PublicFileHandler) Delete(c *gin.Context) {
	fileID := c.Param("id")

	record, err := h.store.Get(fileID)
	if err != nil {
		HandleError(c, apperr.NotFound("file"))
		return
	}

	// 删除磁盘文件
	fileDir := filepath.Join(h.uploadDir, fileID)
	os.RemoveAll(fileDir)

	// 从数据库删除
	if err := h.store.Delete(record.ID); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to delete file record"))
		return
	}

	Success(c, gin.H{"deleted": fileID})
}

// Download 下载文件
// GET /api/v1/files/:id/download
func (h *PublicFileHandler) Download(c *gin.Context) {
	fileID := c.Param("id")

	record, err := h.store.Get(fileID)
	if err != nil || record.Status != FileStatusActive {
		HandleError(c, apperr.NotFound("file"))
		return
	}

	// 验证磁盘文件存在
	if _, err := os.Stat(record.Path); os.IsNotExist(err) {
		HandleError(c, apperr.NotFound("file"))
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", record.Name))
	c.File(record.Path)
}

// GetUploadDir 返回上传目录（供 Task Manager 使用）
func (h *PublicFileHandler) GetUploadDir() string {
	return h.uploadDir
}

// BindFileToTask 将文件关联到 Task
func (h *PublicFileHandler) BindFileToTask(fileID, taskID string, purpose FilePurpose) error {
	return h.store.BindTask(fileID, taskID, purpose)
}

// ListByTask 列出 Task 关联的文件
func (h *PublicFileHandler) ListByTask(taskID string) ([]UploadedFile, error) {
	records, err := h.store.ListByTask(taskID)
	if err != nil {
		return nil, err
	}
	files := make([]UploadedFile, 0, len(records))
	for _, r := range records {
		files = append(files, recordToUploadedFile(r))
	}
	return files, nil
}

// recordToUploadedFile 将数据库记录转为 API 响应
func recordToUploadedFile(r *FileRecord) UploadedFile {
	return UploadedFile{
		ID:         r.ID,
		Name:       r.Name,
		Size:       r.Size,
		MimeType:   r.MimeType,
		TaskID:     r.TaskID,
		Purpose:    r.Purpose,
		UploadedAt: r.CreatedAt,
		ExpiresAt:  r.ExpiresAt,
	}
}

// getMimeType 根据文件扩展名获取 MIME 类型
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".gz":   "application/gzip",
		".tar":  "application/x-tar",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".go":   "text/x-go",
		".py":   "text/x-python",
		".rs":   "text/x-rust",
		".ts":   "text/typescript",
		".tsx":  "text/typescript-jsx",
		".md":   "text/markdown",
		".yaml": "text/yaml",
		".yml":  "text/yaml",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}
