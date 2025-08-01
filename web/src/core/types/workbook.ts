/**
 * 工作簿相关类型定义
 */

// 工作簿数据结构
export interface WorkbookData {
  id: string
  name: string
  appVersion: string
  sheetOrder: string[]
  sheets: Record<string, SheetData>
  locale?: string
  theme?: string
  createdAt?: string
  updatedAt?: string
}

// 工作表数据结构
export interface SheetData {
  id: string
  name: string
  cellData: Record<string, Record<string, CellData>>
  rowCount?: number
  columnCount?: number
  defaultRowHeight?: number
  defaultColumnWidth?: number
  hidden?: boolean
  tabColor?: string
}

// 单元格数据结构
export interface CellData {
  v?: any // 值
  f?: string // 公式
  s?: string // 样式ID
  t?: CellValueType // 类型
  p?: any // 富文本
}

// 单元格值类型
export enum CellValueType {
  STRING = 's',
  NUMBER = 'n',
  BOOLEAN = 'b',
  ERROR = 'e',
  FORMULA = 'f'
}

// 工作簿配置
export interface WorkbookConfig {
  locale?: string
  theme?: string
  readonly?: boolean
  collaborative?: boolean
  autoSave?: boolean
  autoSaveInterval?: number
}

// 工作簿状态
export interface WorkbookState {
  activeSheetId: string | null
  selectedRange: CellRange | null
  editingCell: CellPosition | null
  isLoading: boolean
  isDirty: boolean
  lastSaved?: string
}

// 单元格位置
export interface CellPosition {
  row: number
  col: number
  sheetId: string
}

// 单元格范围
export interface CellRange {
  startRow: number
  startCol: number
  endRow: number
  endCol: number
  sheetId: string
}

// 工作簿操作类型
export enum WorkbookAction {
  CREATE = 'create',
  OPEN = 'open',
  SAVE = 'save',
  CLOSE = 'close',
  RENAME = 'rename',
  DELETE = 'delete'
}

// 工作簿事件
export interface WorkbookEvent {
  type: WorkbookAction
  workbookId: string
  data?: any
  timestamp: number
}