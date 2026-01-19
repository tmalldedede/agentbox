package api

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/files"
)

// FileHandler 文件管理处理器
type FileHandler struct {
	sessionMgr  SessionGetter
	fileManager *files.Manager
}

// SessionGetter 获取 Session 的接口
type SessionGetter interface {
	GetWorkspace(sessionID string) (string, error)
}

// NewFileHandler 创建文件处理器
func NewFileHandler(sessionMgr SessionGetter) *FileHandler {
	return &FileHandler{
		sessionMgr:  sessionMgr,
		fileManager: files.NewManager(),
	}
}

// ListFiles 列出文件
// GET /sessions/:id/files
func (h *FileHandler) ListFiles(c *gin.Context) {
	sessionID := c.Param("id")

	var req files.ListFilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	if req.Path == "" {
		req.Path = "/"
	}

	workspace, err := h.sessionMgr.GetWorkspace(sessionID)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	fileList, err := h.fileManager.List(workspace, req.Path, req.Recursive)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Success(c, fileList)
}

// DownloadFile 下载文件
// GET /sessions/:id/files/*path
func (h *FileHandler) DownloadFile(c *gin.Context) {
	sessionID := c.Param("id")
	filePath := c.Param("path")

	if filePath == "" || filePath == "/" {
		BadRequest(c, "file path is required")
		return
	}

	workspace, err := h.sessionMgr.GetWorkspace(sessionID)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	reader, info, err := h.fileManager.Read(workspace, filePath)
	if err != nil {
		if isNotFoundError(err) {
			NotFound(c, err.Error())
		} else {
			InternalError(c, err.Error())
		}
		return
	}
	defer reader.Close()

	// 设置响应头
	filename := filepath.Base(filePath)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size()))

	// 流式传输文件
	c.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, reader)
		return err == nil
	})
}

// ReadFileContent 读取文件内容 (文本)
// GET /sessions/:id/files/*path/content
func (h *FileHandler) ReadFileContent(c *gin.Context) {
	sessionID := c.Param("id")
	filePath := c.Param("path")

	if filePath == "" || filePath == "/" {
		BadRequest(c, "file path is required")
		return
	}

	// 移除末尾的 /content (因为路由是 /*path 会捕获整个路径)
	// 实际上路由会单独处理，这里不需要

	var req files.ReadContentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	workspace, err := h.sessionMgr.GetWorkspace(sessionID)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	content, err := h.fileManager.ReadContent(workspace, filePath, req.MaxSize)
	if err != nil {
		if isNotFoundError(err) {
			NotFound(c, err.Error())
		} else {
			InternalError(c, err.Error())
		}
		return
	}

	Success(c, content)
}

// UploadFile 上传文件
// POST /sessions/:id/files
func (h *FileHandler) UploadFile(c *gin.Context) {
	sessionID := c.Param("id")

	workspace, err := h.sessionMgr.GetWorkspace(sessionID)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	// 获取上传路径
	targetPath := c.Query("path")
	if targetPath == "" {
		targetPath = "/"
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	// 构建目标路径
	if targetPath == "/" || targetPath == "" {
		targetPath = "/" + header.Filename
	} else if filepath.Ext(targetPath) == "" {
		// 如果目标路径是目录，追加文件名
		targetPath = filepath.Join(targetPath, header.Filename)
	}

	// 写入文件
	result, err := h.fileManager.Write(workspace, targetPath, file, header.Size)
	if err != nil {
		InternalError(c, err.Error())
		return
	}

	Created(c, result)
}

// DeleteFile 删除文件
// DELETE /sessions/:id/files/*path
func (h *FileHandler) DeleteFile(c *gin.Context) {
	sessionID := c.Param("id")
	filePath := c.Param("path")

	if filePath == "" || filePath == "/" {
		BadRequest(c, "file path is required")
		return
	}

	workspace, err := h.sessionMgr.GetWorkspace(sessionID)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	if err := h.fileManager.Delete(workspace, filePath); err != nil {
		if isNotFoundError(err) {
			NotFound(c, err.Error())
		} else {
			InternalError(c, err.Error())
		}
		return
	}

	Success(c, gin.H{"deleted": filePath})
}

// CreateDirectory 创建目录
// POST /sessions/:id/directories
func (h *FileHandler) CreateDirectory(c *gin.Context) {
	sessionID := c.Param("id")

	var req files.CreateDirectoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	workspace, err := h.sessionMgr.GetWorkspace(sessionID)
	if err != nil {
		NotFound(c, err.Error())
		return
	}

	dirInfo, err := h.fileManager.CreateDirectory(workspace, req.Path)
	if err != nil {
		if isConflictError(err) {
			Error(c, http.StatusConflict, err.Error())
		} else {
			InternalError(c, err.Error())
		}
		return
	}

	Created(c, dirInfo)
}

// isNotFoundError 检查是否为 not found 错误
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "not found") || contains(errStr, "does not exist")
}

// isConflictError 检查是否为冲突错误
func isConflictError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "already exists")
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
