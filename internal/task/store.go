package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite" // SQLite 驱动（纯 Go，无需 CGO）
)

// Store 任务存储接口
type Store interface {
	// Create 创建任务
	Create(task *Task) error
	// Get 获取任务
	Get(id string) (*Task, error)
	// Update 更新任务
	Update(task *Task) error
	// Delete 删除任务
	Delete(id string) error
	// List 列出任务
	List(filter *ListFilter) ([]*Task, error)
	// Count 统计任务数量
	Count(filter *ListFilter) (int, error)
	// Close 关闭存储
	Close() error
}

// ListFilter 列表过滤器
type ListFilter struct {
	Status    []Status // 按状态过滤
	ProfileID string   // 按 Profile 过滤
	Limit     int      // 限制数量
	Offset    int      // 偏移量
	OrderBy   string   // 排序字段：created_at, started_at, completed_at
	OrderDesc bool     // 是否降序
}

// SQLiteStore SQLite 存储实现
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore 创建 SQLite 存储
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 启用 WAL 模式，提升并发性能
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return store, nil
}

// migrate 数据库迁移
func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		profile_id TEXT NOT NULL,
		profile_name TEXT,
		agent_type TEXT,
		prompt TEXT NOT NULL,
		input_json TEXT,
		output_config_json TEXT,
		webhook_url TEXT,
		timeout INTEGER DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'pending',
		session_id TEXT,
		error_message TEXT,
		result_json TEXT,
		metadata_json TEXT,
		created_at DATETIME NOT NULL,
		queued_at DATETIME,
		started_at DATETIME,
		completed_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_profile_id ON tasks(profile_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Create 创建任务
func (s *SQLiteStore) Create(task *Task) error {
	inputJSON, _ := json.Marshal(task.Input)
	outputConfigJSON, _ := json.Marshal(task.Output)
	resultJSON, _ := json.Marshal(task.Result)
	metadataJSON, _ := json.Marshal(task.Metadata)

	query := `
	INSERT INTO tasks (
		id, profile_id, profile_name, agent_type, prompt,
		input_json, output_config_json, webhook_url, timeout,
		status, session_id, error_message, result_json, metadata_json,
		created_at, queued_at, started_at, completed_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		task.ID, task.ProfileID, task.ProfileName, task.AgentType, task.Prompt,
		string(inputJSON), string(outputConfigJSON), task.WebhookURL, task.Timeout,
		string(task.Status), task.SessionID, task.ErrorMessage, string(resultJSON), string(metadataJSON),
		task.CreatedAt, nullTime(task.QueuedAt), nullTime(task.StartedAt), nullTime(task.CompletedAt),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrTaskExists
		}
		return err
	}
	return nil
}

// Get 获取任务
func (s *SQLiteStore) Get(id string) (*Task, error) {
	query := `
	SELECT id, profile_id, profile_name, agent_type, prompt,
		input_json, output_config_json, webhook_url, timeout,
		status, session_id, error_message, result_json, metadata_json,
		created_at, queued_at, started_at, completed_at
	FROM tasks WHERE id = ?
	`

	row := s.db.QueryRow(query, id)
	return s.scanTask(row)
}

// Update 更新任务
func (s *SQLiteStore) Update(task *Task) error {
	inputJSON, _ := json.Marshal(task.Input)
	outputConfigJSON, _ := json.Marshal(task.Output)
	resultJSON, _ := json.Marshal(task.Result)
	metadataJSON, _ := json.Marshal(task.Metadata)

	query := `
	UPDATE tasks SET
		profile_id = ?, profile_name = ?, agent_type = ?, prompt = ?,
		input_json = ?, output_config_json = ?, webhook_url = ?, timeout = ?,
		status = ?, session_id = ?, error_message = ?, result_json = ?, metadata_json = ?,
		queued_at = ?, started_at = ?, completed_at = ?
	WHERE id = ?
	`

	result, err := s.db.Exec(query,
		task.ProfileID, task.ProfileName, task.AgentType, task.Prompt,
		string(inputJSON), string(outputConfigJSON), task.WebhookURL, task.Timeout,
		string(task.Status), task.SessionID, task.ErrorMessage, string(resultJSON), string(metadataJSON),
		nullTime(task.QueuedAt), nullTime(task.StartedAt), nullTime(task.CompletedAt),
		task.ID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// Delete 删除任务
func (s *SQLiteStore) Delete(id string) error {
	result, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// List 列出任务
func (s *SQLiteStore) List(filter *ListFilter) ([]*Task, error) {
	query, args := s.buildListQuery(filter, false)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task, err := s.scanTaskRows(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// Count 统计任务数量
func (s *SQLiteStore) Count(filter *ListFilter) (int, error) {
	query, args := s.buildListQuery(filter, true)
	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	return count, err
}

// Close 关闭存储
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// buildListQuery 构建列表查询
func (s *SQLiteStore) buildListQuery(filter *ListFilter, countOnly bool) (string, []interface{}) {
	var query string
	var args []interface{}
	var conditions []string

	if countOnly {
		query = "SELECT COUNT(*) FROM tasks"
	} else {
		query = `
		SELECT id, profile_id, profile_name, agent_type, prompt,
			input_json, output_config_json, webhook_url, timeout,
			status, session_id, error_message, result_json, metadata_json,
			created_at, queued_at, started_at, completed_at
		FROM tasks
		`
	}

	if filter != nil {
		// 状态过滤
		if len(filter.Status) > 0 {
			placeholders := make([]string, len(filter.Status))
			for i, status := range filter.Status {
				placeholders[i] = "?"
				args = append(args, string(status))
			}
			conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
		}

		// Profile 过滤
		if filter.ProfileID != "" {
			conditions = append(conditions, "profile_id = ?")
			args = append(args, filter.ProfileID)
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if !countOnly {
		// 排序
		orderBy := "created_at"
		orderDir := "DESC"
		if filter != nil {
			if filter.OrderBy != "" {
				orderBy = filter.OrderBy
			}
			if !filter.OrderDesc {
				orderDir = "ASC"
			}
		}
		query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

		// 分页
		if filter != nil && filter.Limit > 0 {
			query += " LIMIT ?"
			args = append(args, filter.Limit)
			if filter.Offset > 0 {
				query += " OFFSET ?"
				args = append(args, filter.Offset)
			}
		}
	}

	return query, args
}

// scanTask 扫描单行结果
func (s *SQLiteStore) scanTask(row *sql.Row) (*Task, error) {
	var task Task
	var inputJSON, outputConfigJSON, resultJSON, metadataJSON sql.NullString
	var queuedAt, startedAt, completedAt sql.NullTime
	var status string

	err := row.Scan(
		&task.ID, &task.ProfileID, &task.ProfileName, &task.AgentType, &task.Prompt,
		&inputJSON, &outputConfigJSON, &task.WebhookURL, &task.Timeout,
		&status, &task.SessionID, &task.ErrorMessage, &resultJSON, &metadataJSON,
		&task.CreatedAt, &queuedAt, &startedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}

	task.Status = Status(status)

	// 解析 JSON 字段
	if inputJSON.Valid && inputJSON.String != "" && inputJSON.String != "null" {
		json.Unmarshal([]byte(inputJSON.String), &task.Input)
	}
	if outputConfigJSON.Valid && outputConfigJSON.String != "" && outputConfigJSON.String != "null" {
		json.Unmarshal([]byte(outputConfigJSON.String), &task.Output)
	}
	if resultJSON.Valid && resultJSON.String != "" && resultJSON.String != "null" {
		json.Unmarshal([]byte(resultJSON.String), &task.Result)
	}
	if metadataJSON.Valid && metadataJSON.String != "" && metadataJSON.String != "null" {
		json.Unmarshal([]byte(metadataJSON.String), &task.Metadata)
	}

	// 处理可空时间
	if queuedAt.Valid {
		task.QueuedAt = &queuedAt.Time
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return &task, nil
}

// scanTaskRows 扫描多行结果
func (s *SQLiteStore) scanTaskRows(rows *sql.Rows) (*Task, error) {
	var task Task
	var inputJSON, outputConfigJSON, resultJSON, metadataJSON sql.NullString
	var queuedAt, startedAt, completedAt sql.NullTime
	var status string

	err := rows.Scan(
		&task.ID, &task.ProfileID, &task.ProfileName, &task.AgentType, &task.Prompt,
		&inputJSON, &outputConfigJSON, &task.WebhookURL, &task.Timeout,
		&status, &task.SessionID, &task.ErrorMessage, &resultJSON, &metadataJSON,
		&task.CreatedAt, &queuedAt, &startedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}

	task.Status = Status(status)

	// 解析 JSON 字段
	if inputJSON.Valid && inputJSON.String != "" && inputJSON.String != "null" {
		json.Unmarshal([]byte(inputJSON.String), &task.Input)
	}
	if outputConfigJSON.Valid && outputConfigJSON.String != "" && outputConfigJSON.String != "null" {
		json.Unmarshal([]byte(outputConfigJSON.String), &task.Output)
	}
	if resultJSON.Valid && resultJSON.String != "" && resultJSON.String != "null" {
		json.Unmarshal([]byte(resultJSON.String), &task.Result)
	}
	if metadataJSON.Valid && metadataJSON.String != "" && metadataJSON.String != "null" {
		json.Unmarshal([]byte(metadataJSON.String), &task.Metadata)
	}

	// 处理可空时间
	if queuedAt.Valid {
		task.QueuedAt = &queuedAt.Time
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return &task, nil
}

// nullTime 将 *time.Time 转换为 sql.NullTime
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
