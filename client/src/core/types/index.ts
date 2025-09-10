/**
 * 核心类型定义
 */

// 生命周期状态枚举
export enum LifecycleState {
  CREATED = 'created',
  INITIALIZING = 'initializing',
  INITIALIZED = 'initialized',
  STARTING = 'starting',
  STARTED = 'started',
  STOPPING = 'stopping',
  STOPPED = 'stopped',
  DESTROYED = 'destroyed'
}

// 基础接口
export interface BaseEntity {
  id: string
  createdAt?: string
  updatedAt?: string
}

// 错误类型
export interface AppError extends Error {
  code?: string
  details?: any
}

// 配置接口
export interface AppConfig {
  name: string
  version: string
  environment: 'development' | 'production'
  debug: boolean
  features: {
    darkMode: boolean
    notifications: boolean
    analytics: boolean
    autoSave: boolean
  }
  api: {
    baseURL: string
    timeout: number
    retryAttempts: number
  }
  ui: {
    theme: 'light' | 'dark'
    language: string
    pageSize: number
    animationDuration: number
  }
}