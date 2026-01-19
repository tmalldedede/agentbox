import {
  ApiError,
  NetworkError,
  TimeoutError,
  createHttpError,
} from './errors'

interface RequestConfig extends RequestInit {
  timeout?: number
  baseURL?: string
}

/**
 * 创建一个带超时的 fetch 请求
 */
async function fetchWithTimeout(
  url: string,
  options: RequestConfig = {}
): Promise<Response> {
  const { timeout = 30000, ...fetchOptions } = options

  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), timeout)

  try {
    const response = await fetch(url, {
      ...fetchOptions,
      signal: controller.signal,
    })
    clearTimeout(timeoutId)
    return response
  } catch (error) {
    clearTimeout(timeoutId)

    if (error instanceof Error) {
      if (error.name === 'AbortError') {
        throw new TimeoutError()
      }
      if (error.message.includes('fetch')) {
        throw new NetworkError()
      }
    }
    throw error
  }
}

/**
 * API 响应接口
 */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

/**
 * 增强的请求函数
 */
export async function request<T>(
  url: string,
  options: RequestConfig = {}
): Promise<T> {
  const { baseURL = '/api/v1', ...requestOptions } = options

  // 构建完整 URL
  const fullUrl = url.startsWith('http') ? url : `${baseURL}${url}`

  // 设置默认 headers
  const headers = {
    'Content-Type': 'application/json',
    ...requestOptions.headers,
  }

  try {
    // 发送请求
    const response = await fetchWithTimeout(fullUrl, {
      ...requestOptions,
      headers,
    })

    // 处理 HTTP 错误状态码
    if (!response.ok) {
      let errorMessage = `请求失败: ${response.status} ${response.statusText}`
      let errorDetails: unknown

      try {
        const errorData: ApiResponse = await response.json()
        errorMessage = errorData.message || errorMessage
        errorDetails = errorData.data
      } catch {
        // 如果无法解析 JSON，使用默认错误消息
      }

      throw createHttpError(response.status, errorMessage, errorDetails)
    }

    // 解析响应
    const data: ApiResponse<T> = await response.json()

    // 检查业务错误码
    if (data.code !== 0) {
      throw new ApiError(data.code, data.message || '请求失败', data.data)
    }

    return data.data as T
  } catch (error) {
    // 如果已经是自定义错误，直接抛出
    if (
      error instanceof ApiError ||
      error instanceof NetworkError ||
      error instanceof TimeoutError
    ) {
      throw error
    }

    // 其他未知错误
    if (error instanceof Error) {
      throw new ApiError(500, error.message)
    }

    throw new ApiError(500, '发生了未知错误')
  }
}

/**
 * GET 请求
 */
export function get<T>(url: string, options?: RequestConfig): Promise<T> {
  return request<T>(url, { ...options, method: 'GET' })
}

/**
 * POST 请求
 */
export function post<T>(
  url: string,
  data?: unknown,
  options?: RequestConfig
): Promise<T> {
  return request<T>(url, {
    ...options,
    method: 'POST',
    body: data ? JSON.stringify(data) : undefined,
  })
}

/**
 * PUT 请求
 */
export function put<T>(
  url: string,
  data?: unknown,
  options?: RequestConfig
): Promise<T> {
  return request<T>(url, {
    ...options,
    method: 'PUT',
    body: data ? JSON.stringify(data) : undefined,
  })
}

/**
 * DELETE 请求
 */
export function del<T>(url: string, options?: RequestConfig): Promise<T> {
  return request<T>(url, { ...options, method: 'DELETE' })
}

/**
 * PATCH 请求
 */
export function patch<T>(
  url: string,
  data?: unknown,
  options?: RequestConfig
): Promise<T> {
  return request<T>(url, {
    ...options,
    method: 'PATCH',
    body: data ? JSON.stringify(data) : undefined,
  })
}
