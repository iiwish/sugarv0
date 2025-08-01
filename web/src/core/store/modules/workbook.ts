/**
 * 工作簿状态管理
 */

import { defineStore } from 'pinia'
import type { 
  WorkbookData, 
  WorkbookState, 
  CellPosition, 
  CellRange,
  WorkbookAction 
} from '../../types/workbook'

interface WorkbookStoreState {
  // 当前工作簿
  currentWorkbook: WorkbookData | null
  // 工作簿列表
  workbooks: WorkbookData[]
  // 工作簿状态
  workbookState: WorkbookState
  // 历史记录
  history: {
    undo: any[]
    redo: any[]
    maxSize: number
  }
  // 加载状态
  loading: boolean
  // 错误信息
  error: string | null
}

export const useWorkbookStore = defineStore('workbook', {
  state: (): WorkbookStoreState => ({
    currentWorkbook: null,
    workbooks: [],
    workbookState: {
      activeSheetId: null,
      selectedRange: null,
      editingCell: null,
      isLoading: false,
      isDirty: false
    },
    history: {
      undo: [],
      redo: [],
      maxSize: 50
    },
    loading: false,
    error: null
  }),

  getters: {
    /**
     * 获取当前活动的工作表
     */
    activeSheet: (state) => {
      if (!state.currentWorkbook || !state.workbookState.activeSheetId) {
        return null
      }
      return state.currentWorkbook.sheets[state.workbookState.activeSheetId]
    },

    /**
     * 获取工作表列表
     */
    sheetList: (state) => {
      if (!state.currentWorkbook) return []
      
      return state.currentWorkbook.sheetOrder.map(sheetId => ({
        id: sheetId,
        name: state.currentWorkbook!.sheets[sheetId]?.name || sheetId,
        hidden: state.currentWorkbook!.sheets[sheetId]?.hidden || false
      }))
    },

    /**
     * 检查是否有未保存的更改
     */
    hasUnsavedChanges: (state) => state.workbookState.isDirty,

    /**
     * 检查是否可以撤销
     */
    canUndo: (state) => state.history.undo.length > 0,

    /**
     * 检查是否可以重做
     */
    canRedo: (state) => state.history.redo.length > 0,

    /**
     * 获取选中的单元格数据
     */
    selectedCellData: (state) => {
      const { selectedRange } = state.workbookState
      const sheet = state.currentWorkbook?.sheets[state.workbookState.activeSheetId || '']
      
      if (!selectedRange || !sheet) return null
      
      const { startRow, startCol } = selectedRange
      return sheet.cellData[startRow]?.[startCol] || null
    }
  },

  actions: {
    /**
     * 设置当前工作簿
     */
    setCurrentWorkbook(workbook: WorkbookData) {
      this.currentWorkbook = workbook
      this.workbookState.activeSheetId = workbook.sheetOrder[0] || null
      this.workbookState.isDirty = false
      this.clearHistory()
    },

    /**
     * 添加工作簿到列表
     */
    addWorkbook(workbook: WorkbookData) {
      const existingIndex = this.workbooks.findIndex(wb => wb.id === workbook.id)
      if (existingIndex >= 0) {
        this.workbooks[existingIndex] = workbook
      } else {
        this.workbooks.push(workbook)
      }
    },

    /**
     * 从列表中移除工作簿
     */
    removeWorkbook(workbookId: string) {
      this.workbooks = this.workbooks.filter(wb => wb.id !== workbookId)
      if (this.currentWorkbook?.id === workbookId) {
        this.currentWorkbook = null
        this.workbookState.activeSheetId = null
      }
    },

    /**
     * 设置活动工作表
     */
    setActiveSheet(sheetId: string) {
      if (this.currentWorkbook?.sheets[sheetId]) {
        this.workbookState.activeSheetId = sheetId
        this.workbookState.selectedRange = null
        this.workbookState.editingCell = null
      }
    },

    /**
     * 设置选中范围
     */
    setSelectedRange(range: CellRange | null) {
      this.workbookState.selectedRange = range
      this.workbookState.editingCell = null
    },

    /**
     * 设置编辑单元格
     */
    setEditingCell(position: CellPosition | null) {
      this.workbookState.editingCell = position
    },

    /**
     * 更新单元格数据
     */
    updateCellData(row: number, col: number, data: any) {
      if (!this.currentWorkbook || !this.workbookState.activeSheetId) return

      const sheet = this.currentWorkbook.sheets[this.workbookState.activeSheetId]
      if (!sheet) return

      // 保存到历史记录
      this.saveToHistory({
        type: 'cellUpdate',
        sheetId: this.workbookState.activeSheetId,
        row,
        col,
        oldData: sheet.cellData[row]?.[col],
        newData: data
      })

      // 确保行和列存在
      if (!sheet.cellData[row]) {
        sheet.cellData[row] = {}
      }

      // 更新数据
      if (data === null || data === undefined) {
        delete sheet.cellData[row][col]
      } else {
        sheet.cellData[row][col] = data
      }

      this.workbookState.isDirty = true
    },

    /**
     * 批量更新单元格数据
     */
    batchUpdateCells(updates: Array<{ row: number; col: number; data: any }>) {
      if (!this.currentWorkbook || !this.workbookState.activeSheetId) return

      const sheet = this.currentWorkbook.sheets[this.workbookState.activeSheetId]
      if (!sheet) return

      // 保存到历史记录
      this.saveToHistory({
        type: 'batchUpdate',
        sheetId: this.workbookState.activeSheetId,
        updates: updates.map(update => ({
          ...update,
          oldData: sheet.cellData[update.row]?.[update.col]
        }))
      })

      // 批量更新
      updates.forEach(({ row, col, data }) => {
        if (!sheet.cellData[row]) {
          sheet.cellData[row] = {}
        }

        if (data === null || data === undefined) {
          delete sheet.cellData[row][col]
        } else {
          sheet.cellData[row][col] = data
        }
      })

      this.workbookState.isDirty = true
    },

    /**
     * 添加工作表
     */
    addSheet(sheetData: any) {
      if (!this.currentWorkbook) return

      this.saveToHistory({
        type: 'addSheet',
        sheetId: sheetData.id
      })

      this.currentWorkbook.sheets[sheetData.id] = sheetData
      this.currentWorkbook.sheetOrder.push(sheetData.id)
      this.workbookState.isDirty = true
    },

    /**
     * 删除工作表
     */
    removeSheet(sheetId: string) {
      if (!this.currentWorkbook || this.currentWorkbook.sheetOrder.length <= 1) return

      this.saveToHistory({
        type: 'removeSheet',
        sheetId,
        sheetData: this.currentWorkbook.sheets[sheetId]
      })

      delete this.currentWorkbook.sheets[sheetId]
      this.currentWorkbook.sheetOrder = this.currentWorkbook.sheetOrder.filter(id => id !== sheetId)

      // 如果删除的是当前活动工作表，切换到第一个工作表
      if (this.workbookState.activeSheetId === sheetId) {
        this.workbookState.activeSheetId = this.currentWorkbook.sheetOrder[0] || null
      }

      this.workbookState.isDirty = true
    },

    /**
     * 重命名工作表
     */
    renameSheet(sheetId: string, newName: string) {
      if (!this.currentWorkbook?.sheets[sheetId]) return

      this.saveToHistory({
        type: 'renameSheet',
        sheetId,
        oldName: this.currentWorkbook.sheets[sheetId].name,
        newName
      })

      this.currentWorkbook.sheets[sheetId].name = newName
      this.workbookState.isDirty = true
    },

    /**
     * 设置加载状态
     */
    setLoading(loading: boolean) {
      this.loading = loading
      this.workbookState.isLoading = loading
    },

    /**
     * 设置错误信息
     */
    setError(error: string | null) {
      this.error = error
    },

    /**
     * 标记为已保存
     */
    markAsSaved() {
      this.workbookState.isDirty = false
      this.workbookState.lastSaved = new Date().toISOString()
    },

    /**
     * 保存到历史记录
     */
    saveToHistory(action: any) {
      this.history.undo.push(action)
      this.history.redo = [] // 清空重做栈

      // 限制历史记录大小
      if (this.history.undo.length > this.history.maxSize) {
        this.history.undo.shift()
      }
    },

    /**
     * 撤销操作
     */
    undo() {
      const action = this.history.undo.pop()
      if (!action) return

      this.history.redo.push(action)
      this.executeUndoAction(action)
    },

    /**
     * 重做操作
     */
    redo() {
      const action = this.history.redo.pop()
      if (!action) return

      this.history.undo.push(action)
      this.executeRedoAction(action)
    },

    /**
     * 清空历史记录
     */
    clearHistory() {
      this.history.undo = []
      this.history.redo = []
    },

    /**
     * 执行撤销动作
     */
    executeUndoAction(action: any) {
      // 根据动作类型执行相应的撤销操作
      switch (action.type) {
        case 'cellUpdate':
          // 实现单元格更新的撤销
          break
        case 'batchUpdate':
          // 实现批量更新的撤销
          break
        // 其他动作类型...
      }
    },

    /**
     * 执行重做动作
     */
    executeRedoAction(action: any) {
      // 根据动作类型执行相应的重做操作
      switch (action.type) {
        case 'cellUpdate':
          // 实现单元格更新的重做
          break
        case 'batchUpdate':
          // 实现批量更新的重做
          break
        // 其他动作类型...
      }
    }
  }
})