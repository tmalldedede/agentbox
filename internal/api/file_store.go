package api

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
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

// SQLiteFileStore SQLite 文件存储实现
type SQLiteFileStore struct {
	db *sql.DB
}

// NewSQLiteFileStore 创建 SQLite 文件存储
func NewSQLiteFileStore(dbPath string) (*SQLiteFileStore, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("failed to open file store db: %w", err)
	}

	store := &SQLiteFileStore{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate file store: %w", err)
	}

	return store, nil
}

func (s *SQLiteFileStore) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS files (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			size INTEGER NOT NULL DEFAULT 0,
			mime_type TEXT NOT NULL DEFAULT 'application/octet-stream',
			path TEXT NOT NULL,
			task_id TEXT NOT NULL DEFAULT '',
			purpose TEXT NOT NULL DEFAULT 'general',
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME
		);
		CREATE INDEX IF NOT EXISTS idx_files_task_id ON files(task_id);
		CREATE INDEX IF NOT EXISTS idx_files_status ON files(status);
		CREATE INDEX IF NOT EXISTS idx_files_expires_at ON files(expires_at);
	`)
	return err
}

func (s *SQLiteFileStore) Create(record *FileRecord) error {
	_, err := s.db.Exec(`
		INSERT INTO files (id, name, size, mime_type, path, task_id, purpose, status, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID, record.Name, record.Size, record.MimeType, record.Path,
		record.TaskID, record.Purpose, record.Status,
		record.CreatedAt, nullTimeValue(record.ExpiresAt),
	)
	return err
}

func (s *SQLiteFileStore) Get(id string) (*FileRecord, error) {
	row := s.db.QueryRow(`SELECT id, name, size, mime_type, path, task_id, purpose, status, created_at, expires_at FROM files WHERE id = ?`, id)
	return scanFileRecord(row)
}

func (s *SQLiteFileStore) List(filter *FileListFilter) ([]*FileRecord, error) {
	query := "SELECT id, name, size, mime_type, path, task_id, purpose, status, created_at, expires_at FROM files WHERE 1=1"
	args := []interface{}{}

	if filter != nil {
		if filter.TaskID != "" {
			query += " AND task_id = ?"
			args = append(args, filter.TaskID)
		}
		if filter.Purpose != "" {
			query += " AND purpose = ?"
			args = append(args, filter.Purpose)
		}
		if filter.Status != "" {
			query += " AND status = ?"
			args = append(args, filter.Status)
		}
	}

	query += " ORDER BY created_at DESC"

	if filter != nil && filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filter.Offset)
		}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*FileRecord
	for rows.Next() {
		r, err := scanFileRecordRow(rows)
		if err != nil {
			continue
		}
		records = append(records, r)
	}
	return records, nil
}

func (s *SQLiteFileStore) UpdateStatus(id string, status FileStatus) error {
	result, err := s.db.Exec("UPDATE files SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("file not found: %s", id)
	}
	return nil
}

func (s *SQLiteFileStore) BindTask(id string, taskID string, purpose FilePurpose) error {
	result, err := s.db.Exec("UPDATE files SET task_id = ?, purpose = ? WHERE id = ?", taskID, purpose, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("file not found: %s", id)
	}
	return nil
}

func (s *SQLiteFileStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM files WHERE id = ?", id)
	return err
}

func (s *SQLiteFileStore) ListExpired(before time.Time) ([]*FileRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, name, size, mime_type, path, task_id, purpose, status, created_at, expires_at
		FROM files
		WHERE status = 'active' AND expires_at IS NOT NULL AND expires_at < ?
		ORDER BY expires_at ASC`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*FileRecord
	for rows.Next() {
		r, err := scanFileRecordRow(rows)
		if err != nil {
			continue
		}
		records = append(records, r)
	}
	return records, nil
}

func (s *SQLiteFileStore) ListByTask(taskID string) ([]*FileRecord, error) {
	return s.List(&FileListFilter{TaskID: taskID, Status: FileStatusActive})
}

func (s *SQLiteFileStore) Close() error {
	return s.db.Close()
}

// helpers

func scanFileRecord(row *sql.Row) (*FileRecord, error) {
	r := &FileRecord{}
	var expiresAt sql.NullTime
	err := row.Scan(&r.ID, &r.Name, &r.Size, &r.MimeType, &r.Path,
		&r.TaskID, &r.Purpose, &r.Status, &r.CreatedAt, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("file not found")
		}
		return nil, err
	}
	if expiresAt.Valid {
		r.ExpiresAt = expiresAt.Time
	}
	return r, nil
}

func scanFileRecordRow(rows *sql.Rows) (*FileRecord, error) {
	r := &FileRecord{}
	var expiresAt sql.NullTime
	err := rows.Scan(&r.ID, &r.Name, &r.Size, &r.MimeType, &r.Path,
		&r.TaskID, &r.Purpose, &r.Status, &r.CreatedAt, &expiresAt)
	if err != nil {
		return nil, err
	}
	if expiresAt.Valid {
		r.ExpiresAt = expiresAt.Time
	}
	return r, nil
}

func nullTimeValue(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}
