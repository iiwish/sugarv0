/**
 * HTTP 请求工具
 * 基于 fetch API 的简单封装
 */

import type { ApiResponse } from '@/types/api'
import { useUserStore } from '@/stores/user'

// 请求配置接口
export interface RequestConfig {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  headers?: Record<string, string>
  body?: any
  timeout?: number
  donNotShowLoading?: boolean
}

// 默认配置
const DEFAULT_CONFIG: RequestConfig = {
  method: 'GET',
  timeout: 30000,
  donNotShowLoading: false
}

// 基础 URL - 在实际部署时可以通过环境变量配置
const BASE_URL = '/api'

/**
 * 创建请求实例
 */
class RequestService {
  private baseURL: string

  constructor(baseURL: string = BASE_URL) {
    this.baseURL = baseURL
  }

  /**
   * 发送请求
   */
  async request<T = any>(url: string, config: RequestConfig = {}): Promise<ApiResponse<T>> {
    const finalConfig = { ...DEFAULT_CONFIG, ...config }
    const fullUrl = url.startsWith('http') ? url : `${this.baseURL}${url}`

    // 准备请求头
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...finalConfig.headers
    }

    // 添加认证头
    const userStore = useUserStore()
    if (userStore.token) {
      headers['x-token'] = userStore.token
      headers['x-user-id'] = userStore.userInfo?.id?.toString() || ''
    }

    // 准备请求体
    let body: string | undefined
    if (finalConfig.body && finalConfig.method !== 'GET') {
      body = typeof finalConfig.body === 'string' 
        ? finalConfig.body 
        : JSON.stringify(finalConfig.body)
    }

    try {
      // 创建 AbortController 用于超时控制
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), finalConfig.timeout)

      const response = await fetch(fullUrl, {
        method: finalConfig.method,
        headers,
        body,
        signal: controller.signal
      })

      clearTimeout(timeoutId)

      // 检查响应状态
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      // 解析响应
      const data = await response.json()

      // 处理业务错误
      if (typeof data.code !== 'undefined' && data.code !== 0) {
        throw new Error(data.msg || '请求失败')
      }

      return data
    } catch (error) {
      console.error('请求失败:', error)
      
      // 处理不同类型的错误
      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          throw new Error('请求超时')
        }
        throw error
      }
      
      throw new Error('网络请求失败')
    }
  }

  /**
   * GET 请求
   */
  async get<T = any>(url: string, params?: Record<string, any>, config?: Omit<RequestConfig, 'method' | 'body'>): Promise<ApiResponse<T>> {
    let finalUrl = url
    
    if (params) {
      const searchParams = new URLSearchParams()
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          searchParams.append(key, String(value))
        }
      })
      const queryString = searchParams.toString()
      if (queryString) {
        finalUrl += (url.includes('?') ? '&' : '?') + queryString
      }
    }

    return this.request<T>(finalUrl, { ...config, method: 'GET' })
  }

  /**
   * POST 请求
   */
  async post<T = any>(url: string, data?: any, config?: Omit<RequestConfig, 'method'>): Promise<ApiResponse<T>> {
    return this.request<T>(url, { ...config, method: 'POST', body: data })
  }

  /**
   * PUT 请求
   */
  async put<T = any>(url: string, data?: any, config?: Omit<RequestConfig, 'method'>): Promise<ApiResponse<T>> {
    return this.request<T>(url, { ...config, method: 'PUT', body: data })
  }

  /**
   * DELETE 请求
   */
  async delete<T = any>(url: string, params?: Record<string, any>, config?: Omit<RequestConfig, 'method' | 'body'>): Promise<ApiResponse<T>> {
    let finalUrl = url
    
    if (params) {
      const searchParams = new URLSearchParams()
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          searchParams.append(key, String(value))
        }
      })
      const queryString = searchParams.toString()
      if (queryString) {
        finalUrl += (url.includes('?') ? '&' : '?') + queryString
      }
    }

    return this.request<T>(finalUrl, { ...config, method: 'DELETE' })
  }
}

// 创建默认实例
const service = new RequestService()

export default service
export { RequestService }