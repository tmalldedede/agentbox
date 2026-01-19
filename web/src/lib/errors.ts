/**
 * 基础 API 错误类
 */
export class ApiError extends Error {
  constructor(
    public code: number,
    message: string,
    public details?: unknown
  ) {
    super(message)
    this.name = 'ApiError'
    Object.setPrototypeOf(this, ApiError.prototype)
  }
}

/**
 * 网络错误
 */
export class NetworkError extends Error {
  constructor(message = '网络连接失败，请检查您的网络设置') {
    super(message)
    this.name = 'NetworkError'
    Object.setPrototypeOf(this, NetworkError.prototype)
  }
}

/**
 * 认证错误
 */
export class AuthenticationError extends ApiError {
  constructor(message = '未授权，请重新登录') {
    super(401, message)
    this.name = 'AuthenticationError'
    Object.setPrototypeOf(this, AuthenticationError.prototype)
  }
}

/**
 * 权限错误
 */
export class ForbiddenError extends ApiError {
  constructor(message = '没有权限访问此资源') {
    super(403, message)
    this.name = 'ForbiddenError'
    Object.setPrototypeOf(this, ForbiddenError.prototype)
  }
}

/**
 * 资源不存在错误
 */
export class NotFoundError extends ApiError {
  constructor(message = '请求的资源不存在') {
    super(404, message)
    this.name = 'NotFoundError'
    Object.setPrototypeOf(this, NotFoundError.prototype)
  }
}

/**
 * 验证错误
 */
export class ValidationError extends ApiError {
  constructor(
    message = '数据验证失败',
    public fields?: Record<string, string[]>
  ) {
    super(422, message, fields)
    this.name = 'ValidationError'
    Object.setPrototypeOf(this, ValidationError.prototype)
  }
}

/**
 * 服务器错误
 */
export class ServerError extends ApiError {
  constructor(message = '服务器错误，请稍后重试') {
    super(500, message)
    this.name = 'ServerError'
    Object.setPrototypeOf(this, ServerError.prototype)
  }
}

/**
 * 超时错误
 */
export class TimeoutError extends Error {
  constructor(message = '请求超时，请稍后重试') {
    super(message)
    this.name = 'TimeoutError'
    Object.setPrototypeOf(this, TimeoutError.prototype)
  }
}

/**
 * 根据 HTTP 状态码创建相应的错误
 */
export function createHttpError(status: number, message: string, details?: unknown): ApiError {
  switch (status) {
    case 401:
      return new AuthenticationError(message)
    case 403:
      return new ForbiddenError(message)
    case 404:
      return new NotFoundError(message)
    case 422:
      return new ValidationError(message, details as Record<string, string[]>)
    case 500:
    case 502:
    case 503:
    case 504:
      return new ServerError(message)
    default:
      return new ApiError(status, message, details)
  }
}

/**
 * 错误处理工具函数
 */
export function getErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return error.message
  }
  if (error instanceof NetworkError) {
    return error.message
  }
  if (error instanceof Error) {
    return error.message
  }
  if (typeof error === 'string') {
    return error
  }
  return '发生了未知错误'
}

/**
 * 判断是否为可重试的错误
 */
export function isRetryableError(error: unknown): boolean {
  if (error instanceof NetworkError) {
    return true
  }
  if (error instanceof TimeoutError) {
    return true
  }
  if (error instanceof ServerError) {
    return true
  }
  return false
}
