package files

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// DefaultMaxReadSize 默认最大读取大小 1MB
	DefaultMaxReadSize = 1024 * 1024
	// MaxUploadSize 最大上传大小 100MB
	MaxUploadSize = 100 * 1024 * 1024
)

// Manager 文件管理器
type Manager struct{}

// NewManager 创建文件管理器
func NewManager() *Manager {
	return &Manager{}
}

// List 列出目录内容
func (m *Manager) List(workspacePath, relativePath string, recursive bool) (*FileList, error) {
	// 安全检查：确保路径在工作空间内
	fullPath, err := m.safePath(workspacePath, relativePath)
	if err != nil {
		return nil, err
	}

	// 检查目录是否存在
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path not found: %s", relativePath)
		}
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", relativePath)
	}

	var files []FileInfo
	if recursive {
		files, err = m.listRecursive(workspacePath, fullPath)
	} else {
		files, err = m.listDir(workspacePath, fullPath)
	}
	if err != nil {
		return nil, err
	}

	return &FileList{
		Path:  relativePath,
		Files: files,
	}, nil
}

func (m *Manager) listDir(workspacePath, dirPath string) ([]FileInfo, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relPath, _ := filepath.Rel(workspacePath, filepath.Join(dirPath, entry.Name()))
		fileInfo := FileInfo{
			Name:       entry.Name(),
			Path:       "/" + relPath,
			ModifiedAt: info.ModTime(),
		}

		if entry.IsDir() {
			fileInfo.Type = FileTypeDirectory
			// 计算子项数量
			if subEntries, err := os.ReadDir(filepath.Join(dirPath, entry.Name())); err == nil {
				fileInfo.ChildrenCount = len(subEntries)
			}
		} else {
			fileInfo.Type = FileTypeFile
			fileInfo.Size = info.Size()
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

func (m *Manager) listRecursive(workspacePath, dirPath string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过错误的文件
		}

		// 跳过根目录本身
		if path == dirPath {
			return nil
		}

		relPath, _ := filepath.Rel(workspacePath, path)
		fileInfo := FileInfo{
			Name:       info.Name(),
			Path:       "/" + relPath,
			ModifiedAt: info.ModTime(),
		}

		if info.IsDir() {
			fileInfo.Type = FileTypeDirectory
		} else {
			fileInfo.Type = FileTypeFile
			fileInfo.Size = info.Size()
		}

		files = append(files, fileInfo)
		return nil
	})

	return files, err
}

// Read 读取文件内容
func (m *Manager) Read(workspacePath, relativePath string) (io.ReadCloser, os.FileInfo, error) {
	fullPath, err := m.safePath(workspacePath, relativePath)
	if err != nil {
		return nil, nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("file not found: %s", relativePath)
		}
		return nil, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return nil, nil, fmt.Errorf("path is a directory: %s", relativePath)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, info, nil
}

// ReadContent 读取文本文件内容
func (m *Manager) ReadContent(workspacePath, relativePath string, maxSize int64) (*FileContent, error) {
	if maxSize <= 0 {
		maxSize = DefaultMaxReadSize
	}

	reader, info, err := m.Read(workspacePath, relativePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// 判断是否需要截断
	truncated := info.Size() > maxSize
	readSize := info.Size()
	if truncated {
		readSize = maxSize
	}

	// 读取内容
	data := make([]byte, readSize)
	n, err := io.ReadFull(reader, data)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	data = data[:n]

	// 检测是否为文本文件
	encoding := "utf-8"
	content := string(data)
	if !isTextContent(data) {
		encoding = "base64"
		content = base64.StdEncoding.EncodeToString(data)
	}

	return &FileContent{
		Path:      relativePath,
		Content:   content,
		Size:      info.Size(),
		Encoding:  encoding,
		Truncated: truncated,
	}, nil
}

// Write 写入文件
func (m *Manager) Write(workspacePath, relativePath string, content io.Reader, size int64) (*FileUploadResult, error) {
	if size > MaxUploadSize {
		return nil, fmt.Errorf("file too large: %d bytes (max %d)", size, MaxUploadSize)
	}

	fullPath, err := m.safePath(workspacePath, relativePath)
	if err != nil {
		return nil, err
	}

	// 确保父目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建文件
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 写入内容
	written, err := io.Copy(file, content)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &FileUploadResult{
		Path:       relativePath,
		Size:       written,
		UploadedAt: time.Now(),
	}, nil
}

// Delete 删除文件或目录
func (m *Manager) Delete(workspacePath, relativePath string) error {
	fullPath, err := m.safePath(workspacePath, relativePath)
	if err != nil {
		return err
	}

	// 检查是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("path not found: %s", relativePath)
	}

	// 删除
	if err := os.RemoveAll(fullPath); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	return nil
}

// CreateDirectory 创建目录
func (m *Manager) CreateDirectory(workspacePath, relativePath string) (*FileInfo, error) {
	fullPath, err := m.safePath(workspacePath, relativePath)
	if err != nil {
		return nil, err
	}

	// 检查是否已存在
	if info, err := os.Stat(fullPath); err == nil {
		if info.IsDir() {
			return nil, fmt.Errorf("directory already exists: %s", relativePath)
		}
		return nil, fmt.Errorf("path exists but is not a directory: %s", relativePath)
	}

	// 创建目录
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	info, _ := os.Stat(fullPath)
	return &FileInfo{
		Name:       filepath.Base(relativePath),
		Path:       relativePath,
		Type:       FileTypeDirectory,
		ModifiedAt: info.ModTime(),
	}, nil
}

// safePath 安全路径检查，防止路径遍历攻击
func (m *Manager) safePath(workspacePath, relativePath string) (string, error) {
	// 清理路径
	cleanPath := filepath.Clean(relativePath)
	if cleanPath == "" || cleanPath == "." {
		cleanPath = "/"
	}

	// 移除开头的 /
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	// 构建完整路径
	fullPath := filepath.Join(workspacePath, cleanPath)

	// 确保路径在工作空间内
	absWorkspace, err := filepath.Abs(workspacePath)
	if err != nil {
		return "", fmt.Errorf("invalid workspace path: %w", err)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// 检查是否在工作空间内
	if !strings.HasPrefix(absPath, absWorkspace) {
		return "", fmt.Errorf("path traversal detected: %s", relativePath)
	}

	return absPath, nil
}

// isTextContent 检测是否为文本内容
func isTextContent(data []byte) bool {
	if len(data) == 0 {
		return true
	}

	// 检查是否包含空字节（二进制文件的特征）
	for _, b := range data {
		if b == 0 {
			return false
		}
	}

	// 简单的 UTF-8 有效性检查
	for i := 0; i < len(data); {
		if data[i] < 128 {
			// ASCII
			i++
		} else if data[i]&0xE0 == 0xC0 {
			// 2-byte UTF-8
			if i+1 >= len(data) || data[i+1]&0xC0 != 0x80 {
				return false
			}
			i += 2
		} else if data[i]&0xF0 == 0xE0 {
			// 3-byte UTF-8
			if i+2 >= len(data) || data[i+1]&0xC0 != 0x80 || data[i+2]&0xC0 != 0x80 {
				return false
			}
			i += 3
		} else if data[i]&0xF8 == 0xF0 {
			// 4-byte UTF-8
			if i+3 >= len(data) || data[i+1]&0xC0 != 0x80 || data[i+2]&0xC0 != 0x80 || data[i+3]&0xC0 != 0x80 {
				return false
			}
			i += 4
		} else {
			return false
		}
	}

	return true
}
