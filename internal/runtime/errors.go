package runtime

import "errors"

var (
	ErrRuntimeNotFound      = errors.New("runtime not found")
	ErrRuntimeIDRequired    = errors.New("runtime ID is required")
	ErrRuntimeNameRequired  = errors.New("runtime name is required")
	ErrRuntimeImageRequired = errors.New("runtime image is required")
	ErrRuntimeIsBuiltIn     = errors.New("cannot modify built-in runtime")
)
