package profile

import "errors"

// Profile errors
var (
	ErrProfileNotFound        = errors.New("profile not found")
	ErrProfileAlreadyExists   = errors.New("profile already exists")
	ErrProfileIDRequired      = errors.New("profile ID is required")
	ErrProfileNameRequired    = errors.New("profile name is required")
	ErrProfileAdapterRequired = errors.New("profile adapter is required")
	ErrProfileInvalidAdapter  = errors.New("invalid adapter type")
	ErrProfileIsBuiltIn       = errors.New("cannot modify built-in profile")
	ErrProfileCircularExtends = errors.New("circular profile inheritance detected")
	ErrProfileParentNotFound  = errors.New("parent profile not found")
)
