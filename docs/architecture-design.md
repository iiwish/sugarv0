# Univer表格应用架构设计方案

## 1. 整体架构概览

### 1.1 设计原则
- **模块化**：每个功能模块独立，便于维护和扩展
- **插件化**：支持动态加载功能面板和自定义公式
- **响应式**：统一的状态管理，组件间高效通信
- **可扩展**：为未来功能预留接口和扩展点

### 1.2 核心架构：组合式函数驱动

本应用采用先进的“组合式函数驱动”架构。该架构的核心思想是将所有复杂的初始化、业务逻辑和跨模块通信从视图层剥离，下沉到独立的、可复用的Vue组合式函数中。

```
 ┌──────────────────┐       ┌───────────────────────┐      ┌────────────────────────┐
 │ App Entry (View) │──────▶│  useApp (Orchestrator)  │─────▶│ usePluginManager (Core)  │
 │ (sugar/index.vue)│       │  (composables/useApp.ts)│      │(composables/usePlugin...)│
 └──────────────────┘       └───────────────────────┘      └────────────────────────┘
           │                         │                                │
           ▼                         ▼                                ▼
 ┌──────────────────┐       ┌────────────────────────┐      ┌────────────────────────┐
 │UI Rendering Only │       │   Coordinates Startup/ │      │ Manages All Plugins    │
 │(Just render HTML)│       │   Shutdown Processes   │      │(Register, Activate, etc)│
 └──────────────────┘       └────────────────────────┘      └────────────────────────┘
                                      │
                                      ▼
                             ┌──────────────────────────┐
                             │ useWorkbookManager (Core)  │
                             │(composables/useWorkbook...)│
                             └──────────────────────────┘
                                      │
                                      ▼
                             ┌──────────────────────────┐
                             │ Manages Workbook Data &  │
                             │ Interacts with UniverAPI │
                             └──────────────────────────┘
```

这个流程确保了视图层 ([`web/src/view/sugar/index.vue`]) 保持极度简洁，仅负责UI渲染和调用 `useApp()`，而所有复杂的生命周期和逻辑管理都由 `useApp.ts` 进行集中协调。

## 2. 核心模块详解

*   **组合式函数 (`web/src/composables/`)**: 存放可复用的Vue Composition API逻辑，这是新架构的核心。
    *   **`useApp.ts`**: ([`web/src/composables/useApp.ts`]) 作为应用的“大脑”，负责协调所有模块的启动和关闭。它调用 `usePluginManager` 和 `useWorkbookManager`，并编排整个初始化流程，最后通过 `onMounted` 钩子自动运行。
    *   **`usePluginManager.ts`**: ([`web/src/composables/usePluginManager.ts`]) 封装了与插件系统相关的所有交互，如插件的注册、激活和停用。
    *   **`useWorkbookManager.ts`**: ([`web/src/composables/useWorkbookManager.ts`]) 专注于工作簿的管理，包括加载数据、创建Univer实例以及与Pinia状态同步。

*   **核心模块 (`web/src/core/`)**: 封装了与Univer表格引擎交互的核心逻辑，包括插件管理、事件总线和状态管理。
    *   `plugin/`: 包含插件管理器 ([`PluginManager.ts`](web/src/core/plugin/PluginManager.ts)) 和插件基类 ([`BasePlugin.ts`](web/src/core/plugin/BasePlugin.ts))，是插件化架构的基石。
    *   `events/`: 提供了全局事件总线 ([`EventBus.ts`](web/src/core/events/EventBus.ts))，用于实现模块间的解耦通信。
    *   `store/`: 使用 Pinia 进行状态管理，例如管理工作簿状态 ([`workbook.ts`](web/src/core/store/modules/workbook.ts))。

*   **插件实现 (`web/src/plugins/`)**: 存放所有可插拔的功能模块。
    *   `univer-core/`: ([`web/src/plugins/univer-core/index.ts`]) 封装了对 Univer 核心API的调用，是所有表格功能的基础。
    *   `custom-formulas/`: ([`web/src/plugins/custom-formulas/index.ts`]) 实现了自定义公式功能，例如金融领域的 [`LMDI` 公式](web/src/plugins/custom-formulas/formulas/financial.ts)。

## 3. 实施计划

本轮重构已完成核心阶段，后续可按计划继续开发功能插件。

### 阶段一：核心重构 (已完成)
1.  **[√]** 创建 `useWorkbookManager` `useApp` 等核心组合式函数。
2.  **[√]** 将 `index.vue` 重构为瘦组件。
3.  **[√]** 建立由 `useApp` 驱动的初始化流程。

### 阶段二：功能插件开发
1.  开发文件树插件
2.  开发AI聊天插件
3.  开发数据透视表插件

## 4. 扩展性考虑

得益于新的组合式函数驱动架构，应用的扩展性得到了极大增强。未来无论是增加新的面板、工具栏按钮还是复杂的后台交互，都可以通过开发新的插件并由 `usePluginManager` 进行管理来实现，而无需改动核心的应用启动流程。