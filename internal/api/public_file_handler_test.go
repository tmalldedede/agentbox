package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupFileRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "agentbox-file-test-*")
	require.NoError(t, err)

	dbPath := filepath.Join(tmpDir, "test-files.db")
	store, err := NewSQLiteFileStore(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { store.Close() })

	handler := &PublicFileHandler{
		uploadDir:      tmpDir,
		retentionHours: 72,
		maxFileSize:    100 * 1024 * 1024,
		store:          store,
		cleanupStop:    make(chan struct{}),
	}

	router := gin.New()
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1)

	return router, tmpDir
}

func makeMultipartFile(t *testing.T, fieldName, fileName string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, fileName)
	require.NoError(t, err)

	_, err = part.Write(content)
	require.NoError(t, err)

	require.NoError(t, writer.Close())
	return body, writer.FormDataContentType()
}

// uploadFile is a helper that uploads and returns the parsed UploadedFile
func uploadFile(t *testing.T, router *gin.Engine, fileName string, content []byte) UploadedFile {
	t.Helper()
	body, ct := makeMultipartFile(t, "file", fileName, content)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)

	b, _ := json.Marshal(resp.Data)
	var f UploadedFile
	require.NoError(t, json.Unmarshal(b, &f))
	return f
}

func TestUploadFile(t *testing.T) {
	router, tmpDir := setupFileRouter(t)
	defer os.RemoveAll(tmpDir)

	t.Run("successful upload", func(t *testing.T) {
		content := []byte("Hello, World!")
		f := uploadFile(t, router, "test.txt", content)

		assert.NotEmpty(t, f.ID)
		assert.Equal(t, "test.txt", f.Name)
		assert.Equal(t, int64(len(content)), f.Size)
		assert.Equal(t, "text/plain", f.MimeType) // 基于扩展名检测
		assert.False(t, f.UploadedAt.IsZero())
		assert.False(t, f.ExpiresAt.IsZero())
	})

	t.Run("missing file field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/files", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("mime type consistency between upload and get", func(t *testing.T) {
		cases := []struct {
			name string
			mime string
		}{
			{"data.json", "application/json"},
			{"code.go", "text/x-go"},
			{"image.png", "image/png"},
			{"unknown.xyz", "application/octet-stream"},
		}
		for _, tc := range cases {
			f := uploadFile(t, router, tc.name, []byte("x"))
			// Upload 返回的 MIME 应该基于扩展名
			assert.Equal(t, tc.mime, f.MimeType, "Upload MIME for %s", tc.name)

			// GET 返回的 MIME 应该和 Upload 一致
			req := httptest.NewRequest(http.MethodGet, "/api/v1/files/"+f.ID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)

			var resp Response
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			b, _ := json.Marshal(resp.Data)
			var got UploadedFile
			require.NoError(t, json.Unmarshal(b, &got))
			assert.Equal(t, tc.mime, got.MimeType, "GET MIME for %s", tc.name)
			assert.Equal(t, f.MimeType, got.MimeType, "Upload vs GET MIME for %s", tc.name)
		}
	})
}

func TestListFiles(t *testing.T) {
	router, tmpDir := setupFileRouter(t)
	defer os.RemoveAll(tmpDir)

	t.Run("empty list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/files", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		files, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.Empty(t, files)
	})

	t.Run("returns uploaded files", func(t *testing.T) {
		uploadFile(t, router, "a.txt", []byte("aaa"))
		uploadFile(t, router, "b.txt", []byte("bbb"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/files", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		files, ok := resp.Data.([]interface{})
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(files), 2)
	})
}

func TestGetFile(t *testing.T) {
	router, tmpDir := setupFileRouter(t)
	defer os.RemoveAll(tmpDir)

	t.Run("get existing file", func(t *testing.T) {
		uploaded := uploadFile(t, router, "info.txt", []byte("info content"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/files/"+uploaded.ID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp Response
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		b, _ := json.Marshal(resp.Data)
		var f UploadedFile
		require.NoError(t, json.Unmarshal(b, &f))

		assert.Equal(t, uploaded.ID, f.ID)
		assert.Equal(t, "info.txt", f.Name)
		assert.Equal(t, int64(12), f.Size)
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/files/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDeleteFile(t *testing.T) {
	router, tmpDir := setupFileRouter(t)
	defer os.RemoveAll(tmpDir)

	t.Run("delete existing", func(t *testing.T) {
		uploaded := uploadFile(t, router, "bye.txt", []byte("goodbye"))

		// Delete
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/files/"+uploaded.ID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Verify gone
		req = httptest.NewRequest(http.MethodGet, "/api/v1/files/"+uploaded.ID, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete non-existent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/files/nope", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestDownloadFile(t *testing.T) {
	router, tmpDir := setupFileRouter(t)
	defer os.RemoveAll(tmpDir)

	t.Run("download existing", func(t *testing.T) {
		content := []byte("file bytes here")
		uploaded := uploadFile(t, router, "dl.txt", content)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/files/"+uploaded.ID+"/download", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Disposition"), "dl.txt")

		body, err := io.ReadAll(w.Body)
		require.NoError(t, err)
		assert.Equal(t, content, body)
	})

	t.Run("download non-existent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/files/nope/download", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestFileLifecycle(t *testing.T) {
	router, tmpDir := setupFileRouter(t)
	defer os.RemoveAll(tmpDir)

	content := []byte("lifecycle")

	// 1. Upload
	f := uploadFile(t, router, "life.txt", content)

	// 2. Get
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/"+f.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Download
	req = httptest.NewRequest(http.MethodGet, "/api/v1/files/"+f.ID+"/download", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, content, w.Body.Bytes())

	// 4. List includes it
	req = httptest.NewRequest(http.MethodGet, "/api/v1/files", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), f.ID)

	// 5. Delete
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/files/"+f.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 6. Confirm gone
	req = httptest.NewRequest(http.MethodGet, "/api/v1/files/"+f.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestConcurrentUploads(t *testing.T) {
	router, tmpDir := setupFileRouter(t)
	defer os.RemoveAll(tmpDir)

	const n = 10
	done := make(chan struct{}, n)
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			defer func() { done <- struct{}{} }()
			body, ct := makeMultipartFile(t, "file", fmt.Sprintf("c%d.txt", idx), []byte(fmt.Sprintf("data%d", idx)))
			req := httptest.NewRequest(http.MethodPost, "/api/v1/files", body)
			req.Header.Set("Content-Type", ct)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusCreated {
				errs <- fmt.Errorf("upload %d: status %d", idx, w.Code)
			}
		}(i)
	}

	for i := 0; i < n; i++ {
		<-done
	}
	close(errs)
	for err := range errs {
		t.Error(err)
	}

	// All should be listed
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	files, ok := resp.Data.([]interface{})
	require.True(t, ok)
	assert.Equal(t, n, len(files))
}

func TestFileOnDisk(t *testing.T) {
	router, tmpDir := setupFileRouter(t)
	defer os.RemoveAll(tmpDir)

	content := []byte("disk check")
	f := uploadFile(t, router, "ondisk.txt", content)

	// Verify on disk
	path := filepath.Join(tmpDir, f.ID, "ondisk.txt")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, content, data)
}

func TestMimeTypeDetection(t *testing.T) {
	cases := []struct {
		file string
		mime string
	}{
		{"a.txt", "text/plain"},
		{"a.json", "application/json"},
		{"a.pdf", "application/pdf"},
		{"a.zip", "application/zip"},
		{"a.png", "image/png"},
		{"a.jpg", "image/jpeg"},
		{"a.gif", "image/gif"},
		{"a.svg", "image/svg+xml"},
		{"a.go", "text/x-go"},
		{"a.py", "text/x-python"},
		{"a.rs", "text/x-rust"},
		{"a.ts", "text/typescript"},
		{"a.md", "text/markdown"},
		{"a.yaml", "text/yaml"},
		{"a.xyz", "application/octet-stream"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.mime, getMimeType(tc.file), "getMimeType(%q)", tc.file)
	}
}
