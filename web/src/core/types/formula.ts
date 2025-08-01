/**
 * 公式系统相关类型定义
 */

import type { BaseValueObject } from '@univerjs/preset-sheets-core'

// 公式函数类型
export type FormulaFunction = (...args: any[]) => any

// 公式参数定义
export interface FormulaParameter {
  name: string
  detail: string
  example: string
  require: number // 1=必需, 0=可选
  repeat: number // 1=可重复, 0=不可重复
  type?: FormulaParameterType
}

// 参数类型
export enum FormulaParameterType {
  NUMBER = 'number',
  STRING = 'string',
  BOOLEAN = 'boolean',
  RANGE = 'range',
  ARRAY = 'array',
  ANY = 'any'
}

// 公式分类
export enum FormulaCategory {
  MATH = 'math',
  STATISTICAL = 'statistical',
  FINANCIAL = 'financial',
  LOGICAL = 'logical',
  TEXT = 'text',
  DATE = 'date',
  LOOKUP = 'lookup',
  CUSTOM = 'custom'
}

// 公式信息
export interface FormulaInfo {
  name: string
  category: FormulaCategory
  description: string
  abstract: string
  syntax: string
  parameters: FormulaParameter[]
  examples?: FormulaExample[]
  links?: FormulaLink[]
  version?: string
  author?: string
  deprecated?: boolean
}

// 公式示例
export interface FormulaExample {
  formula: string
  result: any
  description?: string
}

// 公式链接
export interface FormulaLink {
  title: string
  url: string
}

// 公式注册配置
export interface FormulaRegistration {
  name: string
  function: FormulaFunction
  info: FormulaInfo
  locales?: Record<string, any>
}

// 公式执行上下文
export interface FormulaContext {
  workbookId: string
  sheetId: string
  row: number
  col: number
  dependencies?: string[]
}

// 公式计算结果
export interface FormulaResult {
  value: any
  error?: FormulaError
  dependencies?: string[]
  calculationTime?: number
}

// 公式错误类型
export enum FormulaErrorType {
  SYNTAX_ERROR = '#SYNTAX!',
  NAME_ERROR = '#NAME?',
  VALUE_ERROR = '#VALUE!',
  NUM_ERROR = '#NUM!',
  DIV_ZERO_ERROR = '#DIV/0!',
  REF_ERROR = '#REF!',
  NULL_ERROR = '#NULL!',
  CIRCULAR_REF = '#CIRCULAR!'
}

// 公式错误
export interface FormulaError {
  type: FormulaErrorType
  message: string
  position?: number
  details?: any
}

// 公式状态
export interface FormulaState {
  registeredFunctions: Map<string, FormulaRegistration>
  calculationCache: Map<string, FormulaResult>
  dependencyGraph: Map<string, Set<string>>
  isCalculating: boolean
  calculationQueue: string[]
}

// 公式事件类型
export enum FormulaEventType {
  FUNCTION_REGISTERED = 'function_registered',
  FUNCTION_UNREGISTERED = 'function_unregistered',
  CALCULATION_START = 'calculation_start',
  CALCULATION_END = 'calculation_end',
  CALCULATION_ERROR = 'calculation_error',
  DEPENDENCY_CHANGED = 'dependency_changed'
}

// 公式事件
export interface FormulaEvent {
  type: FormulaEventType
  functionName?: string
  cellId?: string
  error?: FormulaError
  result?: FormulaResult
  timestamp: number
}

// 自定义公式基类接口
export interface ICustomFormula {
  name: string
  category: FormulaCategory
  calculate(...args: BaseValueObject[]): BaseValueObject
  getInfo(): FormulaInfo
}

// 公式插件接口
export interface IFormulaPlugin {
  name: string
  version: string
  functions: FormulaRegistration[]
  install(): Promise<void>
  uninstall(): Promise<void>
}