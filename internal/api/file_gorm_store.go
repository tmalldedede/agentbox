package api

import (
	"fmt"
	"time"

	"github.com/tmalldedede/agentbox/internal/database"
	"gorm.io/gorm"
)

// GormFileStore GORM 文件存储实现
type GormFileStore struct {
	db *gorm.DB
}

// NewGormFileStore 创建 GORM 文件存储
func NewGormFileStore(db *gorm.DB) (*GormFileStore, error) {
	if err := db.AutoMigrate(&database.FileModel{}); err != nil {
		return nil, fmt.Errorf("failed to migrate files table: %w", err)
	}
	return &GormFileStore{db: db}, nil
}

// Create 创建文件记录
func (s *GormFileStore) Create(record *FileRecord) error {
	model := fileRecordToModel(record)
	return s.db.Create(model).Error
}

// Get 获取文件记录
func (s *GormFileStore) Get(id string) (*FileRecord, error) {
	var model database.FileModel
	if err := s.db.First(&model, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found")
		}
		return nil, err
	}
	return modelToFileRecord(&model), nil
}

// List 列出文件记录
func (s *GormFileStore) List(filter *FileListFilter) ([]*FileRecord, error) {
	query := s.db.Model(&database.FileModel{})

	if filter != nil {
		if filter.TaskID != "" {
			query = query.Where("task_id = ?", filter.TaskID)
		}
		if filter.Purpose != "" {
			query = query.Where("purpose = ?", filter.Purpose)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
			if filter.Offset > 0 {
				query = query.Offset(filter.Offset)
			}
		}
	}

	query = query.Order("created_at DESC")

	var models []database.FileModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	records := make([]*FileRecord, len(models))
	for i, model := range models {
		records[i] = modelToFileRecord(&model)
	}
	return records, nil
}

// UpdateStatus 更新文件状态
func (s *GormFileStore) UpdateStatus(id string, status FileStatus) error {
	result := s.db.Model(&database.FileModel{}).Where("id = ?", id).Update("status", string(status))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("file not found: %s", id)
	}
	return nil
}

// BindTask 绑定文件到任务
func (s *GormFileStore) BindTask(id string, taskID string, purpose FilePurpose) error {
	result := s.db.Model(&database.FileModel{}).Where("id = ?", id).Updates(map[string]interface{}{
		"task_id": taskID,
		"purpose": string(purpose),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("file not found: %s", id)
	}
	return nil
}

// Delete 删除文件记录
func (s *GormFileStore) Delete(id string) error {
	return s.db.Delete(&database.FileModel{}, "id = ?", id).Error
}

// ListExpired 列出过期文件
func (s *GormFileStore) ListExpired(before time.Time) ([]*FileRecord, error) {
	var models []database.FileModel
	err := s.db.Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?", "active", before).
		Order("expires_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	records := make([]*FileRecord, len(models))
	for i, model := range models {
		records[i] = modelToFileRecord(&model)
	}
	return records, nil
}

// ListByTask 列出任务关联的文件
func (s *GormFileStore) ListByTask(taskID string) ([]*FileRecord, error) {
	return s.List(&FileListFilter{TaskID: taskID, Status: FileStatusActive})
}

// Close 关闭存储（GORM 由外部管理）
func (s *GormFileStore) Close() error {
	return nil
}

// fileRecordToModel 转换 FileRecord 到 FileModel
func fileRecordToModel(r *FileRecord) *database.FileModel {
	model := &database.FileModel{
		BaseModel: database.BaseModel{
			ID:        r.ID,
			CreatedAt: r.CreatedAt,
		},
		Name:     r.Name,
		Size:     r.Size,
		MimeType: r.MimeType,
		Path:     r.Path,
		TaskID:   r.TaskID,
		Purpose:  string(r.Purpose),
		Status:   string(r.Status),
	}
	if !r.ExpiresAt.IsZero() {
		model.ExpiresAt = &r.ExpiresAt
	}
	return model
}

// modelToFileRecord 转换 FileModel 到 FileRecord
func modelToFileRecord(m *database.FileModel) *FileRecord {
	r := &FileRecord{
		ID:        m.ID,
		Name:      m.Name,
		Size:      m.Size,
		MimeType:  m.MimeType,
		Path:      m.Path,
		TaskID:    m.TaskID,
		Purpose:   FilePurpose(m.Purpose),
		Status:    FileStatus(m.Status),
		CreatedAt: m.CreatedAt,
	}
	if m.ExpiresAt != nil {
		r.ExpiresAt = *m.ExpiresAt
	}
	return r
}
