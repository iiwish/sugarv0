import { ref } from 'vue'
import { useWorkbookStore } from '@/core/store/modules/workbook'
import type { IWorkbookData, Univer, Workbook } from '@univerjs/core'
import { LocaleType } from '@univerjs/core'
import type { usePluginManager } from './usePluginManager'
import { UniverCorePlugin } from '@/plugins/univer-core'

const exampleWorkbookData: IWorkbookData = {
  id: 'workbook-01',
  sheetOrder: ['sheet-001'],
  name: 'Sugar表格应用',
  appVersion: '0.1.0',
  locale: LocaleType.ZH_CN,
  styles: {},
  sheets: {
    'sheet-001': {
      id: 'sheet-001',
      name: 'Sheet1',
      cellData: {
        '0': {
          '0': { v: 'Sugar表格应用' },
          '1': { v: 'LMDI 自定义公式示例' },
        },
        '1': {
          '0': { v: '基础期总指标值:' },
          '1': { v: 100 },
        },
        '2': {
          '0': { v: '对比期总指标值:' },
          '1': { v: 120 },
        },
        '3': {
          '0': { v: '基础期因素值:' },
          '1': { v: 10 },
        },
        '4': {
          '0': { v: '对比期因素值:' },
          '1': { v: 12 },
        },
        '5': {
          '0': { v: 'LMDI计算结果:' },
          '1': {
            v: '=LMDI(B2,B3,B4,B5)',
            f: '=LMDI(B2,B3,B4,B5)',
          },
        },
      },
    },
  },
}

/**
 * 工作簿管理器组合式函数
 * @param pluginManager - 插件管理器实例
 */
export function useWorkbookManager(pluginManager: any) {
  const workbookStore = useWorkbookStore()
  const currentWorkbook = ref(null)

  /**
   * 加载初始工作簿
   */
  async function loadInitialWorkbook() {
    return new Promise((resolve, reject) => {
      const eventBus = pluginManager.getEventBus()
      if (!eventBus) {
        return reject(new Error('事件总线未初始化'))
      }

      // 检查Univer核心插件是否已经初始化
      const univerCorePlugin = pluginManager.getPlugin('univer-core')
      if (!univerCorePlugin) {
        return reject(new Error('Univer核心插件未找到'))
      }

      // 检查是否已经有工作簿存在
      const existingWorkbook = univerCorePlugin.getCurrentWorkbook()
      if (existingWorkbook) {
        console.log('工作簿已存在，直接使用现有工作簿')
        currentWorkbook.value = existingWorkbook
        workbookStore.setCurrentWorkbook(existingWorkbook.getSnapshot())
        resolve(existingWorkbook)
        return
      }

      // 如果Univer API已经存在，直接创建工作簿
      const univerAPI = univerCorePlugin.getUniverAPI()
      if (univerAPI) {
        createWorkbookDirectly(univerCorePlugin, resolve, reject)
        return
      }

      // 否则监听Univer初始化完成事件
      eventBus.once('univer:initialized', async () => {
        createWorkbookDirectly(univerCorePlugin, resolve, reject)
      })
    })
  }

  /**
   * 直接创建工作簿
   */
  async function createWorkbookDirectly(univerCorePlugin: any, resolve: Function, reject: Function) {
    try {
      const workbook = await univerCorePlugin.createWorkbook(exampleWorkbookData)
      currentWorkbook.value = workbook

      // 更新Pinia状态
      workbookStore.setCurrentWorkbook(workbook.getSnapshot())
      
      resolve(workbook)
    } catch (error) {
      console.error('创建工作簿失败', error)
      reject(error)
    }
  }

  /**
   * 卸载当前工作簿
   */
  async function unloadWorkbook() {
    // unloadWorkbook 主要由 useApp 在 onBeforeUnmount 时调用
    // 而插件的停用/卸载逻辑由 usePluginManager 统一处理
    // 此处仅需清理 Pinia 状态和本地引用
    currentWorkbook.value = null
    workbookStore.setCurrentWorkbook(null)
    console.log('工作簿已卸载')
  }

  return {
    loadInitialWorkbook,
    unloadWorkbook,
  }
}