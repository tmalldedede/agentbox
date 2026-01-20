package profile

import "github.com/tmalldedede/agentbox/internal/apperr"

// Profile errors - 使用 apperr 提供正确的 HTTP 状态码
var (
	ErrProfileNotFound        = apperr.NotFound("profile")
	ErrProfileAlreadyExists   = apperr.AlreadyExists("profile")
	ErrProfileIDRequired      = apperr.Validation("profile ID is required")
	ErrProfileNameRequired    = apperr.Validation("profile name is required")
	ErrProfileAdapterRequired = apperr.Validation("profile adapter is required")
	ErrProfileInvalidAdapter  = apperr.Validation("invalid adapter type")
	ErrProfileIsBuiltIn       = apperr.BadRequest("cannot modify built-in profile")
	ErrProfileCircularExtends = apperr.BadRequest("circular profile inheritance detected")
	ErrProfileParentNotFound  = apperr.NotFound("parent profile")
)
