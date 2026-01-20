package api

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tmalldedede/agentbox/internal/apperr"
)

const (
	// DefaultUploadDir 默认上传目录
	DefaultUploadDir = "/tmp/agentbox/uploads"
	// MaxUploadSize 最大上传大小 100MB
	MaxUploadSize = 100 * 1024 * 1024
)

// UploadedFile 已上传的文件信息
type UploadedFile struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	MimeType   string    `json:"mime_type"`
	UploadedAt time.Time `json:"uploaded_at"`
	ExpiresAt  time.Time `json:"expires_at,omitempty"`
}

// PublicFileHandler 独立文件上传处理器
type PublicFileHandler struct {
	uploadDir string
}

// NewPublicFileHandler 创建独立文件处理器
func NewPublicFileHandler() *PublicFileHandler {
	// 确保上传目录存在
	os.MkdirAll(DefaultUploadDir, 0755)
	return &PublicFileHandler{
		uploadDir: DefaultUploadDir,
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
	entries, err := os.ReadDir(h.uploadDir)
	if err != nil {
		if os.IsNotExist(err) {
			Success(c, []UploadedFile{})
			return
		}
		HandleError(c, apperr.Wrap(err, "failed to read upload directory"))
		return
	}

	files := make([]UploadedFile, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 每个目录是一个上传的文件
		fileID := entry.Name()
		uploadedFile, err := h.getFileInfo(fileID)
		if err != nil {
			continue
		}
		files = append(files, *uploadedFile)
	}

	// 按上传时间倒序
	sort.Slice(files, func(i, j int) bool {
		return files[i].UploadedAt.After(files[j].UploadedAt)
	})

	Success(c, files)
}

// Upload 上传文件
// POST /api/v1/files
func (h *PublicFileHandler) Upload(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	// 检查文件大小
	if header.Size > MaxUploadSize {
		BadRequest(c, fmt.Sprintf("file too large: %d bytes (max %d)", header.Size, MaxUploadSize))
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

	// 保存文件
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

	// 保存元数据
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	uploadedFile := &UploadedFile{
		ID:         fileID,
		Name:       header.Filename,
		Size:       written,
		MimeType:   mimeType,
		UploadedAt: time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour), // 默认 24 小时后过期
	}

	Created(c, uploadedFile)
}

// Get 获取文件信息
// GET /api/v1/files/:id
func (h *PublicFileHandler) Get(c *gin.Context) {
	fileID := c.Param("id")

	uploadedFile, err := h.getFileInfo(fileID)
	if err != nil {
		HandleError(c, apperr.NotFound("file"))
		return
	}

	Success(c, uploadedFile)
}

// Delete 删除文件
// DELETE /api/v1/files/:id
func (h *PublicFileHandler) Delete(c *gin.Context) {
	fileID := c.Param("id")

	fileDir := filepath.Join(h.uploadDir, fileID)
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		HandleError(c, apperr.NotFound("file"))
		return
	}

	if err := os.RemoveAll(fileDir); err != nil {
		HandleError(c, apperr.Wrap(err, "failed to delete file"))
		return
	}

	Success(c, gin.H{"deleted": fileID})
}

// Download 下载文件
// GET /api/v1/files/:id/download
func (h *PublicFileHandler) Download(c *gin.Context) {
	fileID := c.Param("id")

	fileDir := filepath.Join(h.uploadDir, fileID)
	if _, err := os.Stat(fileDir); os.IsNotExist(err) {
		HandleError(c, apperr.NotFound("file"))
		return
	}

	// 找到目录中的文件
	entries, err := os.ReadDir(fileDir)
	if err != nil || len(entries) == 0 {
		HandleError(c, apperr.NotFound("file"))
		return
	}

	// 获取第一个文件
	var fileName string
	for _, entry := range entries {
		if !entry.IsDir() {
			fileName = entry.Name()
			break
		}
	}

	if fileName == "" {
		HandleError(c, apperr.NotFound("file"))
		return
	}

	filePath := filepath.Join(fileDir, fileName)

	// 设置响应头
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.File(filePath)
}

// getFileInfo 获取文件信息
func (h *PublicFileHandler) getFileInfo(fileID string) (*UploadedFile, error) {
	// 安全检查：防止路径遍历
	if strings.Contains(fileID, "/") || strings.Contains(fileID, "\\") || strings.Contains(fileID, "..") {
		return nil, fmt.Errorf("invalid file ID")
	}

	fileDir := filepath.Join(h.uploadDir, fileID)
	dirInfo, err := os.Stat(fileDir)
	if err != nil {
		return nil, err
	}

	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("not a directory")
	}

	// 找到目录中的文件
	entries, err := os.ReadDir(fileDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		return &UploadedFile{
			ID:         fileID,
			Name:       entry.Name(),
			Size:       info.Size(),
			MimeType:   getMimeType(entry.Name()),
			UploadedAt: dirInfo.ModTime(),
			ExpiresAt:  dirInfo.ModTime().Add(24 * time.Hour),
		}, nil
	}

	return nil, fmt.Errorf("no file in directory")
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
