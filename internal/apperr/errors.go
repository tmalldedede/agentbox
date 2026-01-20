package apperr

import (
	"fmt"
	"net/http"
)

// ErrorType 错误类型
type ErrorType string

const (
	// 通用错误类型
	TypeNotFound      ErrorType = "NOT_FOUND"
	TypeAlreadyExists ErrorType = "ALREADY_EXISTS"
	TypeValidation    ErrorType = "VALIDATION_ERROR"
	TypeUnauthorized  ErrorType = "UNAUTHORIZED"
	TypeForbidden     ErrorType = "FORBIDDEN"
	TypeInternal      ErrorType = "INTERNAL_ERROR"
	TypeBadRequest    ErrorType = "BAD_REQUEST"
	TypeConflict      ErrorType = "CONFLICT"
	TypeTimeout       ErrorType = "TIMEOUT"
	TypeUnavailable   ErrorType = "SERVICE_UNAVAILABLE"
)

// AppError 应用错误
type AppError struct {
	Code    int       // HTTP 状态码
	Type    ErrorType // 错误类型
	Message string    // 用户可见的错误消息
	Detail  string    // 详细错误信息（调试用）
	Err     error     // 原始错误
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	}
	return e.Message
}

// Unwrap 实现 errors.Unwrap 接口
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is 实现 errors.Is 接口
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Type == t.Type
	}
	return false
}

// WithDetail 添加详细信息
func (e *AppError) WithDetail(detail string) *AppError {
	e.Detail = detail
	return e
}

// WithError 添加原始错误
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	if e.Detail == "" && err != nil {
		e.Detail = err.Error()
	}
	return e
}

// ============== 构造函数 ==============

// NotFound 资源未找到错误
func NotFound(resource string) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Type:    TypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NotFoundf 资源未找到错误（带格式化）
func NotFoundf(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Type:    TypeNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

// AlreadyExists 资源已存在错误
func AlreadyExists(resource string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Type:    TypeAlreadyExists,
		Message: fmt.Sprintf("%s already exists", resource),
	}
}

// Validation 验证错误
func Validation(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Type:    TypeValidation,
		Message: message,
	}
}

// Validationf 验证错误（带格式化）
func Validationf(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Type:    TypeValidation,
		Message: fmt.Sprintf(format, args...),
	}
}

// BadRequest 错误请求
func BadRequest(message string) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Type:    TypeBadRequest,
		Message: message,
	}
}

// BadRequestf 错误请求（带格式化）
func BadRequestf(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    http.StatusBadRequest,
		Type:    TypeBadRequest,
		Message: fmt.Sprintf(format, args...),
	}
}

// Unauthorized 未授权错误
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "unauthorized"
	}
	return &AppError{
		Code:    http.StatusUnauthorized,
		Type:    TypeUnauthorized,
		Message: message,
	}
}

// Forbidden 禁止访问错误
func Forbidden(message string) *AppError {
	if message == "" {
		message = "forbidden"
	}
	return &AppError{
		Code:    http.StatusForbidden,
		Type:    TypeForbidden,
		Message: message,
	}
}

// Conflict 冲突错误
func Conflict(message string) *AppError {
	return &AppError{
		Code:    http.StatusConflict,
		Type:    TypeConflict,
		Message: message,
	}
}

// Internal 内部错误
func Internal(message string) *AppError {
	if message == "" {
		message = "internal server error"
	}
	return &AppError{
		Code:    http.StatusInternalServerError,
		Type:    TypeInternal,
		Message: message,
	}
}

// Internalf 内部错误（带格式化）
func Internalf(format string, args ...interface{}) *AppError {
	return &AppError{
		Code:    http.StatusInternalServerError,
		Type:    TypeInternal,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap 包装现有错误
func Wrap(err error, message string) *AppError {
	if err == nil {
		return nil
	}
	// 如果已经是 AppError，保留原有信息
	if appErr, ok := err.(*AppError); ok {
		return &AppError{
			Code:    appErr.Code,
			Type:    appErr.Type,
			Message: message,
			Detail:  appErr.Message,
			Err:     appErr.Err,
		}
	}
	return &AppError{
		Code:    http.StatusInternalServerError,
		Type:    TypeInternal,
		Message: message,
		Detail:  err.Error(),
		Err:     err,
	}
}

// Wrapf 包装现有错误（带格式化）
func Wrapf(err error, format string, args ...interface{}) *AppError {
	return Wrap(err, fmt.Sprintf(format, args...))
}

// Timeout 超时错误
func Timeout(message string) *AppError {
	if message == "" {
		message = "request timeout"
	}
	return &AppError{
		Code:    http.StatusRequestTimeout,
		Type:    TypeTimeout,
		Message: message,
	}
}

// Unavailable 服务不可用错误
func Unavailable(message string) *AppError {
	if message == "" {
		message = "service unavailable"
	}
	return &AppError{
		Code:    http.StatusServiceUnavailable,
		Type:    TypeUnavailable,
		Message: message,
	}
}

// ============== 辅助函数 ==============

// IsNotFound 判断是否为未找到错误
func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == TypeNotFound
	}
	return false
}

// IsAlreadyExists 判断是否为已存在错误
func IsAlreadyExists(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == TypeAlreadyExists
	}
	return false
}

// IsValidation 判断是否为验证错误
func IsValidation(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == TypeValidation
	}
	return false
}

// GetHTTPCode 获取错误的 HTTP 状态码
func GetHTTPCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return http.StatusInternalServerError
}

// GetType 获取错误类型
func GetType(err error) ErrorType {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type
	}
	return TypeInternal
}
