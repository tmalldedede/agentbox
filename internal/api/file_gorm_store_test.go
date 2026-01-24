package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupFileTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	return db
}

func TestGormFileStore_Create(t *testing.T) {
	db := setupFileTestDB(t)
	store, err := NewGormFileStore(db)
	require.NoError(t, err)

	record := &FileRecord{
		ID:        "file-001",
		Name:      "test.txt",
		Size:      1024,
		MimeType:  "text/plain",
		Path:      "/tmp/files/test.txt",
		Purpose:   FilePurposeGeneral,
		Status:    FileStatusActive,
		CreatedAt: time.Now(),
	}

	err = store.Create(record)
	require.NoError(t, err)

	// Verify
	got, err := store.Get("file-001")
	require.NoError(t, err)
	assert.Equal(t, record.ID, got.ID)
	assert.Equal(t, record.Name, got.Name)
	assert.Equal(t, record.Size, got.Size)
	assert.Equal(t, record.MimeType, got.MimeType)
}

func TestGormFileStore_Get(t *testing.T) {
	db := setupFileTestDB(t)
	store, err := NewGormFileStore(db)
	require.NoError(t, err)

	// Get non-existent
	_, err = store.Get("non-existent")
	assert.Error(t, err)

	// Create and get
	record := &FileRecord{
		ID:        "file-get",
		Name:      "doc.pdf",
		Size:      2048,
		MimeType:  "application/pdf",
		Path:      "/tmp/files/doc.pdf",
		Purpose:   FilePurposeAttachment,
		Status:    FileStatusActive,
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.Create(record))

	got, err := store.Get("file-get")
	require.NoError(t, err)
	assert.Equal(t, "doc.pdf", got.Name)
	assert.Equal(t, FilePurposeAttachment, got.Purpose)
}

func TestGormFileStore_List(t *testing.T) {
	db := setupFileTestDB(t)
	store, err := NewGormFileStore(db)
	require.NoError(t, err)

	// Create multiple files
	files := []struct {
		id      string
		taskID  string
		purpose FilePurpose
		status  FileStatus
	}{
		{"f1", "task-1", FilePurposeAttachment, FileStatusActive},
		{"f2", "task-1", FilePurposeOutput, FileStatusActive},
		{"f3", "task-2", FilePurposeAttachment, FileStatusActive},
		{"f4", "", FilePurposeGeneral, FileStatusExpired},
	}

	for _, f := range files {
		record := &FileRecord{
			ID:        f.id,
			Name:      f.id + ".txt",
			Size:      100,
			MimeType:  "text/plain",
			Path:      "/tmp/" + f.id,
			TaskID:    f.taskID,
			Purpose:   f.purpose,
			Status:    f.status,
			CreatedAt: time.Now(),
		}
		require.NoError(t, store.Create(record))
	}

	// List all
	result, err := store.List(nil)
	require.NoError(t, err)
	assert.Len(t, result, 4)

	// Filter by TaskID
	result, err = store.List(&FileListFilter{TaskID: "task-1"})
	require.NoError(t, err)
	assert.Len(t, result, 2)

	// Filter by Purpose
	result, err = store.List(&FileListFilter{Purpose: FilePurposeAttachment})
	require.NoError(t, err)
	assert.Len(t, result, 2)

	// Filter by Status
	result, err = store.List(&FileListFilter{Status: FileStatusExpired})
	require.NoError(t, err)
	assert.Len(t, result, 1)

	// With limit
	result, err = store.List(&FileListFilter{Limit: 2})
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestGormFileStore_UpdateStatus(t *testing.T) {
	db := setupFileTestDB(t)
	store, err := NewGormFileStore(db)
	require.NoError(t, err)

	record := &FileRecord{
		ID:        "file-status",
		Name:      "test.txt",
		Size:      100,
		MimeType:  "text/plain",
		Path:      "/tmp/test.txt",
		Purpose:   FilePurposeGeneral,
		Status:    FileStatusActive,
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.Create(record))

	// Update status
	err = store.UpdateStatus("file-status", FileStatusExpired)
	require.NoError(t, err)

	// Verify
	got, err := store.Get("file-status")
	require.NoError(t, err)
	assert.Equal(t, FileStatusExpired, got.Status)

	// Update non-existent
	err = store.UpdateStatus("non-existent", FileStatusDeleted)
	assert.Error(t, err)
}

func TestGormFileStore_BindTask(t *testing.T) {
	db := setupFileTestDB(t)
	store, err := NewGormFileStore(db)
	require.NoError(t, err)

	record := &FileRecord{
		ID:        "file-bind",
		Name:      "attachment.txt",
		Size:      100,
		MimeType:  "text/plain",
		Path:      "/tmp/attachment.txt",
		Purpose:   FilePurposeGeneral,
		Status:    FileStatusActive,
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.Create(record))

	// Bind to task
	err = store.BindTask("file-bind", "task-123", FilePurposeAttachment)
	require.NoError(t, err)

	// Verify
	got, err := store.Get("file-bind")
	require.NoError(t, err)
	assert.Equal(t, "task-123", got.TaskID)
	assert.Equal(t, FilePurposeAttachment, got.Purpose)

	// Bind non-existent
	err = store.BindTask("non-existent", "task-123", FilePurposeAttachment)
	assert.Error(t, err)
}

func TestGormFileStore_Delete(t *testing.T) {
	db := setupFileTestDB(t)
	store, err := NewGormFileStore(db)
	require.NoError(t, err)

	record := &FileRecord{
		ID:        "file-delete",
		Name:      "to-delete.txt",
		Size:      100,
		MimeType:  "text/plain",
		Path:      "/tmp/to-delete.txt",
		Purpose:   FilePurposeGeneral,
		Status:    FileStatusActive,
		CreatedAt: time.Now(),
	}
	require.NoError(t, store.Create(record))

	// Delete
	err = store.Delete("file-delete")
	require.NoError(t, err)

	// Verify deleted
	_, err = store.Get("file-delete")
	assert.Error(t, err)
}

func TestGormFileStore_ListExpired(t *testing.T) {
	db := setupFileTestDB(t)
	store, err := NewGormFileStore(db)
	require.NoError(t, err)

	now := time.Now()
	pastExpiry := now.Add(-1 * time.Hour)
	futureExpiry := now.Add(1 * time.Hour)

	files := []struct {
		id        string
		expiresAt time.Time
		status    FileStatus
	}{
		{"expired-1", pastExpiry, FileStatusActive},
		{"expired-2", pastExpiry, FileStatusActive},
		{"not-expired", futureExpiry, FileStatusActive},
		{"already-deleted", pastExpiry, FileStatusDeleted}, // should not be listed
	}

	for _, f := range files {
		record := &FileRecord{
			ID:        f.id,
			Name:      f.id + ".txt",
			Size:      100,
			MimeType:  "text/plain",
			Path:      "/tmp/" + f.id,
			Purpose:   FilePurposeGeneral,
			Status:    f.status,
			CreatedAt: now,
			ExpiresAt: f.expiresAt,
		}
		require.NoError(t, store.Create(record))
	}

	// List expired before now
	expired, err := store.ListExpired(now)
	require.NoError(t, err)
	assert.Len(t, expired, 2) // expired-1 and expired-2

	// Verify order (by expires_at ASC)
	for _, f := range expired {
		assert.True(t, f.ExpiresAt.Before(now))
		assert.Equal(t, FileStatusActive, f.Status)
	}
}

func TestGormFileStore_ListByTask(t *testing.T) {
	db := setupFileTestDB(t)
	store, err := NewGormFileStore(db)
	require.NoError(t, err)

	// Create files for different tasks
	files := []struct {
		id     string
		taskID string
		status FileStatus
	}{
		{"tf1", "task-abc", FileStatusActive},
		{"tf2", "task-abc", FileStatusActive},
		{"tf3", "task-abc", FileStatusDeleted}, // should not be listed
		{"tf4", "task-xyz", FileStatusActive},
	}

	for _, f := range files {
		record := &FileRecord{
			ID:        f.id,
			Name:      f.id + ".txt",
			Size:      100,
			MimeType:  "text/plain",
			Path:      "/tmp/" + f.id,
			TaskID:    f.taskID,
			Purpose:   FilePurposeAttachment,
			Status:    f.status,
			CreatedAt: time.Now(),
		}
		require.NoError(t, store.Create(record))
	}

	// List by task (only active)
	result, err := store.ListByTask("task-abc")
	require.NoError(t, err)
	assert.Len(t, result, 2)

	result, err = store.ListByTask("task-xyz")
	require.NoError(t, err)
	assert.Len(t, result, 1)

	result, err = store.ListByTask("non-existent")
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestGormFileStore_Close(t *testing.T) {
	db := setupFileTestDB(t)
	store, err := NewGormFileStore(db)
	require.NoError(t, err)

	err = store.Close()
	assert.NoError(t, err)
}
