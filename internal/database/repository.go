package database

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Common errors
var (
	ErrNotFound      = errors.New("record not found")
	ErrAlreadyExists = errors.New("record already exists")
)

// ProfileRepository handles profile database operations
type ProfileRepository struct {
	db *gorm.DB
}

// NewProfileRepository creates a new ProfileRepository
func NewProfileRepository() *ProfileRepository {
	return &ProfileRepository{db: DB}
}

// Create creates a new profile
func (r *ProfileRepository) Create(model *ProfileModel) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	return r.db.Create(model).Error
}

// Get retrieves a profile by ID
func (r *ProfileRepository) Get(id string) (*ProfileModel, error) {
	var model ProfileModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves all profiles with optional filters
func (r *ProfileRepository) List(filters map[string]interface{}) ([]ProfileModel, error) {
	var models []ProfileModel
	query := r.db.Model(&ProfileModel{})

	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	err := query.Order("created_at DESC").Find(&models).Error
	return models, err
}

// Update updates a profile
func (r *ProfileRepository) Update(model *ProfileModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete deletes a profile
func (r *ProfileRepository) Delete(id string) error {
	return r.db.Delete(&ProfileModel{}, "id = ?", id).Error
}

// MCPServerRepository handles MCP server database operations
type MCPServerRepository struct {
	db *gorm.DB
}

// NewMCPServerRepository creates a new MCPServerRepository
func NewMCPServerRepository() *MCPServerRepository {
	return &MCPServerRepository{db: DB}
}

// Create creates a new MCP server
func (r *MCPServerRepository) Create(model *MCPServerModel) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	return r.db.Create(model).Error
}

// Get retrieves an MCP server by ID
func (r *MCPServerRepository) Get(id string) (*MCPServerModel, error) {
	var model MCPServerModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// GetByName retrieves an MCP server by name
func (r *MCPServerRepository) GetByName(name string) (*MCPServerModel, error) {
	var model MCPServerModel
	err := r.db.Where("name = ?", name).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves all MCP servers
func (r *MCPServerRepository) List() ([]MCPServerModel, error) {
	var models []MCPServerModel
	err := r.db.Order("name ASC").Find(&models).Error
	return models, err
}

// Update updates an MCP server
func (r *MCPServerRepository) Update(model *MCPServerModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete deletes an MCP server
func (r *MCPServerRepository) Delete(id string) error {
	return r.db.Delete(&MCPServerModel{}, "id = ?", id).Error
}

// SkillRepository handles skill database operations
type SkillRepository struct {
	db *gorm.DB
}

// NewSkillRepository creates a new SkillRepository
func NewSkillRepository() *SkillRepository {
	return &SkillRepository{db: DB}
}

// Create creates a new skill
func (r *SkillRepository) Create(model *SkillModel) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	return r.db.Create(model).Error
}

// Get retrieves a skill by ID
func (r *SkillRepository) Get(id string) (*SkillModel, error) {
	var model SkillModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// GetBySlug retrieves a skill by slug
func (r *SkillRepository) GetBySlug(slug string) (*SkillModel, error) {
	var model SkillModel
	err := r.db.Where("slug = ?", slug).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves all skills
func (r *SkillRepository) List() ([]SkillModel, error) {
	var models []SkillModel
	err := r.db.Order("name ASC").Find(&models).Error
	return models, err
}

// Update updates a skill
func (r *SkillRepository) Update(model *SkillModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete deletes a skill
func (r *SkillRepository) Delete(id string) error {
	return r.db.Delete(&SkillModel{}, "id = ?", id).Error
}

// CredentialRepository handles credential database operations
type CredentialRepository struct {
	db *gorm.DB
}

// NewCredentialRepository creates a new CredentialRepository
func NewCredentialRepository() *CredentialRepository {
	return &CredentialRepository{db: DB}
}

// Create creates a new credential
func (r *CredentialRepository) Create(model *CredentialModel) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	return r.db.Create(model).Error
}

// Get retrieves a credential by ID
func (r *CredentialRepository) Get(id string) (*CredentialModel, error) {
	var model CredentialModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves all credentials
func (r *CredentialRepository) List() ([]CredentialModel, error) {
	var models []CredentialModel
	err := r.db.Order("name ASC").Find(&models).Error
	return models, err
}

// Update updates a credential
func (r *CredentialRepository) Update(model *CredentialModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete deletes a credential
func (r *CredentialRepository) Delete(id string) error {
	return r.db.Delete(&CredentialModel{}, "id = ?", id).Error
}

// SessionRepository handles session database operations
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{db: DB}
}

// Create creates a new session
func (r *SessionRepository) Create(model *SessionModel) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	return r.db.Create(model).Error
}

// Get retrieves a session by ID
func (r *SessionRepository) Get(id string) (*SessionModel, error) {
	var model SessionModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves sessions with optional status filter
func (r *SessionRepository) List(status string, limit int) ([]SessionModel, error) {
	var models []SessionModel
	query := r.db.Model(&SessionModel{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&models).Error
	return models, err
}

// Update updates a session
func (r *SessionRepository) Update(model *SessionModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete deletes a session
func (r *SessionRepository) Delete(id string) error {
	return r.db.Delete(&SessionModel{}, "id = ?", id).Error
}

// UpdateStatus updates session status
func (r *SessionRepository) UpdateStatus(id, status string) error {
	return r.db.Model(&SessionModel{}).Where("id = ?", id).Update("status", status).Error
}

// WebhookRepository handles webhook database operations
type WebhookRepository struct {
	db *gorm.DB
}

// NewWebhookRepository creates a new WebhookRepository
func NewWebhookRepository() *WebhookRepository {
	return &WebhookRepository{db: DB}
}

// Create creates a new webhook
func (r *WebhookRepository) Create(model *WebhookModel) error {
	if model.ID == "" {
		model.ID = uuid.New().String()
	}
	return r.db.Create(model).Error
}

// Get retrieves a webhook by ID
func (r *WebhookRepository) Get(id string) (*WebhookModel, error) {
	var model WebhookModel
	err := r.db.Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &model, err
}

// List retrieves all webhooks
func (r *WebhookRepository) List() ([]WebhookModel, error) {
	var models []WebhookModel
	err := r.db.Order("name ASC").Find(&models).Error
	return models, err
}

// ListEnabled retrieves all enabled webhooks
func (r *WebhookRepository) ListEnabled() ([]WebhookModel, error) {
	var models []WebhookModel
	err := r.db.Where("is_enabled = ?", true).Find(&models).Error
	return models, err
}

// Update updates a webhook
func (r *WebhookRepository) Update(model *WebhookModel) error {
	model.UpdatedAt = time.Now()
	return r.db.Save(model).Error
}

// Delete deletes a webhook
func (r *WebhookRepository) Delete(id string) error {
	return r.db.Delete(&WebhookModel{}, "id = ?", id).Error
}

// Helper functions for JSON serialization
func ToJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func FromJSON(s string, v interface{}) error {
	if s == "" {
		return nil
	}
	return json.Unmarshal([]byte(s), v)
}
