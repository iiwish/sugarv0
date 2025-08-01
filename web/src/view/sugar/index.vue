<template>
  <div class="sugar-app">
    <!-- 主要内容区域 -->
    <div class="main-content">
      <!-- Univer表格容器 -->
      <div id="univer-sheet-container" class="univer-container"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import { usePluginManager } from '@/composables/usePluginManager'
import { useWorkbookStore } from '@/core/store/modules/workbook'
import type { UniverContainerReadyEvent } from '@/core/types/events'
import { functionLmdiZhCN } from '@/plugins/custom-formulas/formulas/financial'

// 使用组合式函数
const pluginManager = usePluginManager()
const workbookStore = useWorkbookStore()

// 响应式数据
const isInitialized = ref(false)

// 示例工作簿数据
const exampleWorkbookData = {
  "id": "workbook-01",
  "sheetOrder": ["sheet-001"],
  "name": "Sugar表格应用",
  "appVersion": "0.1.0",
  "sheets": {
    "sheet-001": {
      "id": "sheet-001",
      "name": "Sheet1",
      "cellData": {
        "0": {
          "0": { "v": "Sugar表格应用" },
          "1": { "v": "LMDI 自定义公式示例" }
        },
        "1": {
          "0": { "v": "基础期总指标值:" },
          "1": { "v": 100 }
        },
        "2": {
          "0": { "v": "对比期总指标值:" },
          "1": { "v": 120 }
        },
        "3": {
          "0": { "v": "基础期因素值:" },
          "1": { "v": 10 }
        },
        "4": {
          "0": { "v": "对比期因素值:" },
          "1": { "v": 12 }
        },
        "5": {
          "0": { "v": "LMDI计算结果:" },
          "1": { 
            "v": "=LMDI(B2,B3,B4,B5)",
            "f": "=LMDI(B2,B3,B4,B5)"
          }
        }
      }
    }
  }
}

/**
 * 初始化应用
 */
async function initializeApp(): Promise<void> {
  try {
    // 初始化插件管理器
    await pluginManager.initialize()
    
    // 注册核心插件
    await pluginManager.registerCorePlugins()
    
    // 触发容器准备事件，让Univer核心插件初始化
    const containerReadyEvent: UniverContainerReadyEvent = {
      type: 'univer:container-ready',
      payload: {
        containerId: 'univer-sheet-container',
        locales: {
          zhCN: functionLmdiZhCN
        }
      },
      timestamp: Date.now()
    }
    
    const eventBus = pluginManager.getEventBus()
    if (!eventBus) {
      throw new Error('事件总线未初始化')
    }
    
    eventBus.emit('univer:container-ready', containerReadyEvent)
    
    // 监听Univer初始化完成事件
    eventBus.on('univer:initialized', handleUniverInitialized)
    
    // 监听工作簿创建事件
    eventBus.on('workbook:created', handleWorkbookCreated)
    
    isInitialized.value = true
    console.log('Sugar应用初始化完成')
  } catch (error) {
    console.error('Sugar应用初始化失败:', error)
  }
}

/**
 * 处理Univer初始化完成
 */
async function handleUniverInitialized(event: any): Promise<void> {
  try {
    // 获取Univer核心插件
    const univerCorePlugin = pluginManager.getPlugin('univer-core')
    if (!univerCorePlugin) {
      throw new Error('Univer核心插件未找到')
    }
    
    // 创建工作簿
    const workbook = await (univerCorePlugin as any).createWorkbook(exampleWorkbookData)
    
    // 更新状态管理
    workbookStore.setCurrentWorkbook({
      id: exampleWorkbookData.id,
      name: exampleWorkbookData.name,
      appVersion: exampleWorkbookData.appVersion,
      sheetOrder: exampleWorkbookData.sheetOrder,
      sheets: exampleWorkbookData.sheets
    })
    
    console.log('工作簿创建成功')
  } catch (error) {
    console.error('创建工作簿失败:', error)
  }
}

/**
 * 处理工作簿创建事件
 */
function handleWorkbookCreated(event: any): void {
  console.log('工作簿创建事件:', event)
  // 可以在这里添加额外的处理逻辑
}

/**
 * 清理资源
 */
async function cleanup(): Promise<void> {
  try {
    // 移除事件监听器
    const eventBus = pluginManager.getEventBus()
    if (eventBus) {
      eventBus.off('univer:initialized')
      eventBus.off('workbook:created')
    }
    
    // 停用所有插件
    await pluginManager.deactivateAllPlugins()
    
    console.log('Sugar应用清理完成')
  } catch (error) {
    console.error('Sugar应用清理失败:', error)
  }
}

// 生命周期钩子
onMounted(() => {
  initializeApp()
})

onBeforeUnmount(() => {
  cleanup()
})
</script>

<style scoped>
.sugar-app {
  display: flex;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background-color: #f5f5f5;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0; /* 防止flex子项溢出 */
}

.univer-container {
  flex: 1;
  width: 100%;
  height: 100%;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin: 8px;
}

/* 确保Univer样式正确加载 */
:deep(#univer-sheet-container) {
  width: 100%;
  height: 100%;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .univer-container {
    margin: 4px;
    border-radius: 4px;
  }
}
</style>