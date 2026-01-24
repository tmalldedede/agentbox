package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmalldedede/agentbox/internal/apperr"
)

// Response 通用响应
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Type    string      `json:"type,omitempty"` // 错误类型
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Created 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "created",
		Data:    data,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// HandleError 处理 AppError 或普通 error
// 这是推荐的错误处理方式
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	// 检查是否为 AppError
	if appErr, ok := err.(*apperr.AppError); ok {
		c.JSON(appErr.Code, Response{
			Code:    appErr.Code,
			Message: appErr.Message,
			Type:    string(appErr.Type),
		})
		return
	}

	// 普通错误，返回 500 — 打印日志方便调试
	log.Error("unhandled error (500)", "error", err.Error(), "path", c.Request.URL.Path, "method", c.Request.Method)
	c.JSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		Message: err.Error(),
		Type:    string(apperr.TypeInternal),
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Forbidden 403 错误
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound 404 错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalError 500 错误
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// Pagination 分页信息
type Pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// PaginatedResponse 分页响应
type PaginatedResponse struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// SuccessWithPagination 带分页的成功响应
func SuccessWithPagination(c *gin.Context, data interface{}, total, limit, offset int) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Code:    0,
		Message: "success",
		Data:    data,
		Pagination: &Pagination{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error string `json:"error"`
}
