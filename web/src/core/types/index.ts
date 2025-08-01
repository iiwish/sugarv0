/**
 * 核心类型定义
 */

// 导出所有类型
export * from './workbook'
export * from './formula'
export * from './plugin'
export * from './events'

// 基础类型
export interface BaseConfig {
  id: string
  name: string
  version: string
  description?: string
}

// 错误类型
export interface AppError {
  code: string
  message: string
  details?: any
}

// 生命周期状态
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

// 日志级别
export enum LogLevel {
  DEBUG = 'debug',
  INFO = 'info',
  WARN = 'warn',
  ERROR = 'error'
}