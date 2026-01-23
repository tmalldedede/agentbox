package task

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite" // SQLite 驱动（纯 Go，无需 CGO）
)

// TaskStats 任务统计
type TaskStats struct {
	Total     int            `json:"total"`
	ByStatus  map[Status]int `json:"by_status"`
	ByAgent   map[string]int `json:"by_agent"`
	AvgDuration float64      `json:"avg_duration_seconds"` // 已完成任务平均耗时
}

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
	// Stats 获取任务统计
	Stats() (*TaskStats, error)
	// Cleanup 清理旧任务
	Cleanup(before time.Time, statuses []Status) (int, error)
	// Close 关闭存储
	Close() error
}

// ListFilter 列表过滤器
type ListFilter struct {
	Status    []Status // 按状态过滤
	AgentID   string   // 按 Agent 过滤
	Search    string   // 搜索 prompt 关键字
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
	// 检查旧表 schema 是否兼容，不兼容则重建
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='attachments_json'`).Scan(&count)
	if err == nil && count == 0 {
		// 旧表 schema 不兼容（缺少 attachments_json 等新字段），重建
		s.db.Exec("DROP TABLE IF EXISTS tasks")
	}

	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		agent_id TEXT NOT NULL,
		agent_name TEXT,
		agent_type TEXT,
		prompt TEXT NOT NULL,
		attachments_json TEXT,
		output_files_json TEXT,
		turns_json TEXT,
		turn_count INTEGER DEFAULT 0,
		webhook_url TEXT,
		timeout INTEGER DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'pending',
		session_id TEXT,
		thread_id TEXT,
		error_message TEXT,
		result_json TEXT,
		metadata_json TEXT,
		created_at DATETIME NOT NULL,
		queued_at DATETIME,
		started_at DATETIME,
		completed_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_agent_id ON tasks(agent_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
	`

	_, err = s.db.Exec(schema)
	if err != nil {
		return err
	}

	// 增量迁移：添加 thread_id 列（已有表兼容）
	s.db.Exec("ALTER TABLE tasks ADD COLUMN thread_id TEXT")
	return nil
}

// Create 创建任务
func (s *SQLiteStore) Create(task *Task) error {
	attachmentsJSON, _ := json.Marshal(task.Attachments)
	outputFilesJSON, _ := json.Marshal(task.OutputFiles)
	turnsJSON, _ := json.Marshal(task.Turns)
	resultJSON, _ := json.Marshal(task.Result)
	metadataJSON, _ := json.Marshal(task.Metadata)

	query := `
	INSERT INTO tasks (
		id, agent_id, agent_name, agent_type, prompt,
		attachments_json, output_files_json, turns_json, turn_count,
		webhook_url, timeout,
		status, session_id, thread_id, error_message, result_json, metadata_json,
		created_at, queued_at, started_at, completed_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		task.ID, task.AgentID, task.AgentName, task.AgentType, task.Prompt,
		string(attachmentsJSON), string(outputFilesJSON), string(turnsJSON), task.TurnCount,
		task.WebhookURL, task.Timeout,
		string(task.Status), task.SessionID, task.ThreadID, task.ErrorMessage, string(resultJSON), string(metadataJSON),
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
	SELECT id, agent_id, agent_name, agent_type, prompt,
		attachments_json, output_files_json, turns_json, turn_count,
		webhook_url, timeout,
		status, session_id, thread_id, error_message, result_json, metadata_json,
		created_at, queued_at, started_at, completed_at
	FROM tasks WHERE id = ?
	`

	row := s.db.QueryRow(query, id)
	return s.scanTask(row)
}

// Update 更新任务
func (s *SQLiteStore) Update(task *Task) error {
	attachmentsJSON, _ := json.Marshal(task.Attachments)
	outputFilesJSON, _ := json.Marshal(task.OutputFiles)
	turnsJSON, _ := json.Marshal(task.Turns)
	resultJSON, _ := json.Marshal(task.Result)
	metadataJSON, _ := json.Marshal(task.Metadata)

	query := `
	UPDATE tasks SET
		agent_id = ?, agent_name = ?, agent_type = ?, prompt = ?,
		attachments_json = ?, output_files_json = ?, turns_json = ?, turn_count = ?,
		webhook_url = ?, timeout = ?,
		status = ?, session_id = ?, thread_id = ?, error_message = ?, result_json = ?, metadata_json = ?,
		queued_at = ?, started_at = ?, completed_at = ?
	WHERE id = ?
	`

	result, err := s.db.Exec(query,
		task.AgentID, task.AgentName, task.AgentType, task.Prompt,
		string(attachmentsJSON), string(outputFilesJSON), string(turnsJSON), task.TurnCount,
		task.WebhookURL, task.Timeout,
		string(task.Status), task.SessionID, task.ThreadID, task.ErrorMessage, string(resultJSON), string(metadataJSON),
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

// Stats 获取任务统计
func (s *SQLiteStore) Stats() (*TaskStats, error) {
	stats := &TaskStats{
		ByStatus: make(map[Status]int),
		ByAgent:  make(map[string]int),
	}

	// 按状态统计
	rows, err := s.db.Query("SELECT status, COUNT(*) FROM tasks GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.ByStatus[Status(status)] = count
		stats.Total += count
	}

	// 按 Agent 统计（Top 10）
	rows2, err := s.db.Query("SELECT COALESCE(agent_name, agent_id), COUNT(*) FROM tasks GROUP BY agent_id ORDER BY COUNT(*) DESC LIMIT 10")
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var agentName string
		var count int
		if err := rows2.Scan(&agentName, &count); err != nil {
			return nil, err
		}
		stats.ByAgent[agentName] = count
	}

	// 平均执行时长（已完成任务）
	var avgDuration sql.NullFloat64
	err = s.db.QueryRow(`
		SELECT AVG(JULIANDAY(completed_at) - JULIANDAY(started_at)) * 86400
		FROM tasks WHERE status = 'completed' AND started_at IS NOT NULL AND completed_at IS NOT NULL
	`).Scan(&avgDuration)
	if err == nil && avgDuration.Valid {
		stats.AvgDuration = avgDuration.Float64
	}

	return stats, nil
}

// Cleanup 清理旧任务
func (s *SQLiteStore) Cleanup(before time.Time, statuses []Status) (int, error) {
	var args []interface{}
	query := "DELETE FROM tasks WHERE created_at < ?"
	args = append(args, before)

	if len(statuses) > 0 {
		placeholders := make([]string, len(statuses))
		for i, st := range statuses {
			placeholders[i] = "?"
			args = append(args, string(st))
		}
		query += fmt.Sprintf(" AND status IN (%s)", strings.Join(placeholders, ","))
	}

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
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
		SELECT id, agent_id, agent_name, agent_type, prompt,
			attachments_json, output_files_json, turns_json, turn_count,
			webhook_url, timeout,
			status, session_id, thread_id, error_message, result_json, metadata_json,
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

		// Agent 过滤
		if filter.AgentID != "" {
			conditions = append(conditions, "agent_id = ?")
			args = append(args, filter.AgentID)
		}

		// 搜索 prompt 关键字
		if filter.Search != "" {
			conditions = append(conditions, "prompt LIKE ?")
			args = append(args, "%"+filter.Search+"%")
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
	var agentName sql.NullString
	var threadID sql.NullString
	var attachmentsJSON, outputFilesJSON, turnsJSON, resultJSON, metadataJSON sql.NullString
	var queuedAt, startedAt, completedAt sql.NullTime
	var status string

	err := row.Scan(
		&task.ID, &task.AgentID, &agentName, &task.AgentType, &task.Prompt,
		&attachmentsJSON, &outputFilesJSON, &turnsJSON, &task.TurnCount,
		&task.WebhookURL, &task.Timeout,
		&status, &task.SessionID, &threadID, &task.ErrorMessage, &resultJSON, &metadataJSON,
		&task.CreatedAt, &queuedAt, &startedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}

	if agentName.Valid {
		task.AgentName = agentName.String
	}
	if threadID.Valid {
		task.ThreadID = threadID.String
	}
	task.Status = Status(status)

	// 解析 JSON 字段
	if attachmentsJSON.Valid && attachmentsJSON.String != "" && attachmentsJSON.String != "null" {
		json.Unmarshal([]byte(attachmentsJSON.String), &task.Attachments)
	}
	if outputFilesJSON.Valid && outputFilesJSON.String != "" && outputFilesJSON.String != "null" {
		json.Unmarshal([]byte(outputFilesJSON.String), &task.OutputFiles)
	}
	if turnsJSON.Valid && turnsJSON.String != "" && turnsJSON.String != "null" {
		json.Unmarshal([]byte(turnsJSON.String), &task.Turns)
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
	var agentName sql.NullString
	var threadID sql.NullString
	var attachmentsJSON, outputFilesJSON, turnsJSON, resultJSON, metadataJSON sql.NullString
	var queuedAt, startedAt, completedAt sql.NullTime
	var status string

	err := rows.Scan(
		&task.ID, &task.AgentID, &agentName, &task.AgentType, &task.Prompt,
		&attachmentsJSON, &outputFilesJSON, &turnsJSON, &task.TurnCount,
		&task.WebhookURL, &task.Timeout,
		&status, &task.SessionID, &threadID, &task.ErrorMessage, &resultJSON, &metadataJSON,
		&task.CreatedAt, &queuedAt, &startedAt, &completedAt,
	)
	if err != nil {
		return nil, err
	}

	if agentName.Valid {
		task.AgentName = agentName.String
	}
	if threadID.Valid {
		task.ThreadID = threadID.String
	}
	task.Status = Status(status)

	// 解析 JSON 字段
	if attachmentsJSON.Valid && attachmentsJSON.String != "" && attachmentsJSON.String != "null" {
		json.Unmarshal([]byte(attachmentsJSON.String), &task.Attachments)
	}
	if outputFilesJSON.Valid && outputFilesJSON.String != "" && outputFilesJSON.String != "null" {
		json.Unmarshal([]byte(outputFilesJSON.String), &task.OutputFiles)
	}
	if turnsJSON.Valid && turnsJSON.String != "" && turnsJSON.String != "null" {
		json.Unmarshal([]byte(turnsJSON.String), &task.Turns)
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
