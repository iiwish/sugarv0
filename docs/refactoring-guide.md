# Univer表格应用重构实施指南

## 概述

本文档提供了从当前架构迁移到新架构的详细实施指南。重构将分阶段进行，确保在整个过程中应用保持可用状态。

## 重构目标

1. **模块化架构**：将单体组件拆分为独立的功能模块
2. **插件化系统**：支持动态加载功能面板和自定义公式
3. **统一状态管理**：使用Pinia管理应用状态
4. **类型安全**：完整的TypeScript类型定义
5. **可扩展性**：为未来功能预留接口和扩展点

## 实施阶段

### 阶段一：基础架构重构 (1-2周)

#### 1.1 创建核心类型系统
- [x] 已完成：创建 `web/src/core/types/` 目录结构
- [x] 已完成：定义工作簿、公式、插件、事件相关类型

#### 1.2 实现事件系统
- [x] 已完成：创建 `EventBus` 类
- [x] 已完成：实现发布订阅模式
- [ ] 集成到现有组件中

#### 1.3 建立状态管理
- [x] 已完成：创建Pinia store模块
- [ ] 迁移现有状态到新的store
- [ ] 更新组件以使用新的状态管理

**实施步骤：**

1. **安装依赖**
```bash
cd web
npm install pinia
```

2. **更新main.js**
```javascript
import { createApp } from 'vue'
import App from './App.vue'
import { setupStore } from '@/core/store'

const app = createApp(App)
setupStore(app)
app.mount('#app')
```

3. **迁移现有状态**
   - 将 `web/src/view/main/index.vue` 中的状态迁移到 `useWorkbookStore`
   - 更新组件以使用store中的状态和方法

### 阶段二：插件系统实现 (2-3周)

#### 2.1 实现插件基础架构
- [x] 已完成：创建 `BasePlugin` 基类
- [x] 已完成：实现 `PluginManager`
- [ ] 创建插件注册机制

#### 2.2 重构现有功能为插件
- [ ] 将表格核心功能封装为插件
- [ ] 将公式系统重构为插件
- [ ] 创建基础UI插件

#### 2.3 实现面板插件基类
```typescript
// web/src/core/plugin/BasePanelPlugin.ts
import { BasePlugin } from './BasePlugin'
import type { IPanelPlugin, PanelPosition } from '../types/plugin'

export abstract class BasePanelPlugin extends BasePlugin implements IPanelPlugin {
  abstract readonly panelId: string
  abstract readonly title: string
  abstract readonly position: PanelPosition
  
  abstract createPanel(): any
  abstract destroyPanel(): void
}
```

**实施步骤：**

1. **创建插件上下文**
```typescript
// web/src/core/context.ts
import type { PluginContext } from './types/plugin'
import { globalEventBus } from './events'
import { pinia } from './store'

export function createPluginContext(): PluginContext {
  return {
    app: getCurrentInstance()?.appContext.app,
    store: pinia,
    eventBus: globalEventBus,
    logger: console, // 可以替换为更完善的日志系统
    config: {}
  }
}
```

2. **初始化插件管理器**
```typescript
// web/src/core/index.ts
import { initializePluginManager } from './plugin'
import { createPluginContext } from './context'

export async function initializeCore() {
  const context = createPluginContext()
  const pluginManager = initializePluginManager(context)
  
  // 注册核心插件
  // await pluginManager.register(new WorkbookPlugin())
  // await pluginManager.register(new FormulaPlugin())
  
  return { pluginManager, context }
}
```

### 阶段三：功能插件开发 (3-4周)

#### 3.1 文件树插件
```typescript
// web/src/plugins/file-tree/FileTreePlugin.ts
import { BasePanelPlugin } from '@/core/plugin/BasePanelPlugin'
import { PanelPosition } from '@/core/types/plugin'
import FileTreePanel from './components/FileTreePanel.vue'

export class FileTreePlugin extends BasePanelPlugin {
  readonly panelId = 'file-tree'
  readonly title = '文件树'
  readonly position = PanelPosition.LEFT

  protected async onInstall() {
    // 注册API路由
    // 初始化文件系统监听
  }

  protected async onActivate() {
    // 激活文件监听
  }

  protected async onDeactivate() {
    // 停止文件监听
  }

  protected async onUninstall() {
    // 清理资源
  }

  createPanel() {
    return FileTreePanel
  }

  destroyPanel() {
    // 清理面板资源
  }
}
```

#### 3.2 AI聊天插件
```typescript
// web/src/plugins/ai-chat/AiChatPlugin.ts
import { BasePanelPlugin } from '@/core/plugin/BasePanelPlugin'
import { PanelPosition } from '@/core/types/plugin'
import AiChatPanel from './components/AiChatPanel.vue'

export class AiChatPlugin extends BasePanelPlugin {
  readonly panelId = 'ai-chat'
  readonly title = 'AI助手'
  readonly position = PanelPosition.RIGHT

  // 实现插件生命周期方法...
}
```

#### 3.3 数据透视表插件
```typescript
// web/src/plugins/pivot-table/PivotTablePlugin.ts
import { BasePanelPlugin } from '@/core/plugin/BasePanelPlugin'
import { PanelPosition } from '@/core/types/plugin'
import PivotTablePanel from './components/PivotTablePanel.vue'

export class PivotTablePlugin extends BasePanelPlugin {
  readonly panelId = 'pivot-table'
  readonly title = '数据透视表'
  readonly position = PanelPosition.RIGHT

  // 实现插件生命周期方法...
}
```

### 阶段四：公式系统重构 (2-3周)

#### 4.1 重构现有LMDI公式
```typescript
// web/src/core/univer/formula/functions/lmdi.ts
import { BaseFunction } from '@univerjs/preset-sheets-core'
import type { ICustomFormula } from '../../types/formula'
import { FormulaCategory } from '../../types/formula'

export class LmdiFormula extends BaseFunction implements ICustomFormula {
  readonly name = 'LMDI'
  readonly category = FormulaCategory.FINANCIAL

  calculate(...args: BaseValueObject[]): BaseValueObject {
    // 迁移现有的LMDI计算逻辑
  }

  getInfo(): FormulaInfo {
    // 返回公式信息
  }
}
```

#### 4.2 创建公式注册系统
```typescript
// web/src/core/univer/formula/registry.ts
import type { FormulaRegistration } from '../types/formula'

export class FormulaRegistry {
  private formulas = new Map<string, FormulaRegistration>()

  register(registration: FormulaRegistration) {
    this.formulas.set(registration.name, registration)
  }

  unregister(name: string) {
    this.formulas.delete(name)
  }

  getAll() {
    return Array.from(this.formulas.values())
  }
}
```

### 阶段五：UI重构和集成 (2-3周)

#### 5.1 创建新的主布局
```vue
<!-- web/src/layouts/MainLayout.vue -->
<template>
  <div class="main-layout">
    <Header />
    <div class="layout-content">
      <Sidebar v-if="!ui.sidebar.collapsed" />
      <div class="content-area">
        <div class="panels-container">
          <PanelContainer 
            v-for="panel in leftPanels" 
            :key="panel.id"
            :panel="panel"
            position="left"
          />
        </div>
        <div class="main-content">
          <router-view />
        </div>
        <div class="panels-container">
          <PanelContainer 
            v-for="panel in rightPanels" 
            :key="panel.id"
            :panel="panel"
            position="right"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useUIStore } from '@/core/store'

const ui = useUIStore()

const leftPanels = computed(() => 
  ui.panelsByPosition.left
)

const rightPanels = computed(() => 
  ui.panelsByPosition.right
)
</script>
```

#### 5.2 重构主页面组件
```vue
<!-- web/src/view/main/index.vue -->
<template>
  <div class="univer-container">
    <div id="univer-sheet-container" ref="containerRef"></div>
  </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref } from 'vue'
import { useWorkbookStore } from '@/core/store'
import { useUniver } from '@/composables/useUniver'

const containerRef = ref<HTMLElement>()
const workbookStore = useWorkbookStore()

const { initializeUniver, destroyUniver } = useUniver()

onMounted(async () => {
  if (containerRef.value) {
    await initializeUniver(containerRef.value)
  }
})

onBeforeUnmount(() => {
  destroyUniver()
})
</script>
```

#### 5.3 创建组合式函数
```typescript
// web/src/composables/useUniver.ts
import { ref } from 'vue'
import { createUniver, LocaleType } from '@univerjs/presets'
import { UniverSheetsCorePreset } from '@univerjs/preset-sheets-core'
import { useWorkbookStore } from '@/core/store'

export function useUniver() {
  const univerAPI = ref(null)
  const workbookStore = useWorkbookStore()

  const initializeUniver = async (container: HTMLElement) => {
    const { univerAPI: api } = createUniver({
      locale: LocaleType.ZH_CN,
      presets: [
        UniverSheetsCorePreset({
          container: container
        })
      ]
    })

    univerAPI.value = api
    
    // 注册自定义公式
    await registerCustomFormulas(api)
    
    // 创建示例工作簿
    const workbook = api.createWorkbook(getExampleWorkbookData())
    workbookStore.setCurrentWorkbook(workbook)
  }

  const destroyUniver = () => {
    if (univerAPI.value) {
      univerAPI.value.dispose()
      univerAPI.value = null
    }
  }

  return {
    univerAPI,
    initializeUniver,
    destroyUniver
  }
}
```

## 迁移检查清单

### 代码迁移
- [ ] 将现有状态迁移到Pinia store
- [ ] 重构组件以使用新的状态管理
- [ ] 将公式注册逻辑移到专门的注册器
- [ ] 创建插件实例并注册到插件管理器
- [ ] 更新路由配置以使用新的布局

### 测试验证
- [ ] 验证表格基本功能正常
- [ ] 验证LMDI公式计算正确
- [ ] 验证状态管理工作正常
- [ ] 验证事件系统通信正常
- [ ] 验证插件加载和卸载

### 性能优化
- [ ] 实现组件懒加载
- [ ] 优化状态更新频率
- [ ] 实现虚拟滚动（如需要）
- [ ] 优化事件监听器管理

## 风险控制

### 回滚策略
1. **保留原有代码**：在重构过程中保留原有代码作为备份
2. **分支管理**：使用Git分支进行重构，确保可以随时回滚
3. **渐进式迁移**：逐步迁移功能，确保每个阶段都可以独立工作

### 测试策略
1. **单元测试**：为核心模块编写单元测试
2. **集成测试**：测试模块间的集成
3. **端到端测试**：测试完整的用户流程
4. **性能测试**：确保重构后性能不下降

## 后续扩展

### 新功能开发
1. **插件市场**：开发插件市场，支持第三方插件
2. **协作功能**：实现多人协作编辑
3. **云端同步**：支持云端数据同步
4. **移动端适配**：适配移动端设备

### 技术升级
1. **微前端**：考虑微前端架构
2. **WebAssembly**：使用WASM优化计算密集型操作
3. **PWA**：支持离线使用
4. **国际化**：完善多语言支持

## 总结

这个重构方案提供了一个清晰的路径，从当前的单体架构迁移到模块化、插件化的新架构。通过分阶段实施，可以确保在整个重构过程中应用保持稳定可用，同时为未来的功能扩展奠定坚实的基础。

重构完成后，您将拥有：
- 高度模块化和可扩展的架构
- 强大的插件系统
- 统一的状态管理
- 完整的类型安全
- 优秀的开发体验

**我已经对代码进行了逻辑审查。请您手动进行全面的测试以确保其行为符合预期。**