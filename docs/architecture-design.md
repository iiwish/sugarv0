# Univer表格应用架构设计方案

## 1. 整体架构概览

### 1.1 设计原则
- **模块化**：每个功能模块独立，便于维护和扩展
- **插件化**：支持动态加载功能面板和自定义公式
- **响应式**：统一的状态管理，组件间高效通信
- **可扩展**：为未来功能预留接口和扩展点

### 1.2 核心架构层次
```
┌─────────────────────────────────────────────────────────────┐
│                        应用层 (App Layer)                    │
├─────────────────────────────────────────────────────────────┤
│                      插件层 (Plugin Layer)                   │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │  文件树面板  │ │  AI聊天面板  │ │ 数据透视表   │ │  更多...  │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
├─────────────────────────────────────────────────────────────┤
│                      核心层 (Core Layer)                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │  表格引擎    │ │  公式系统    │ │  状态管理    │ │  事件总线 │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
├─────────────────────────────────────────────────────────────┤
│                      基础层 (Base Layer)                     │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────┐ │
│  │   Univer    │ │     Vue3     │ │    Pinia     │ │  Utils   │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 2. 目录结构设计

```
web/src/
├── core/                           # 核心模块
│   ├── univer/                     # Univer相关核心功能
│   │   ├── engine/                 # 表格引擎
│   │   │   ├── index.ts           # 引擎主入口
│   │   │   ├── config.ts          # 引擎配置
│   │   │   └── lifecycle.ts       # 生命周期管理
│   │   ├── formula/                # 公式系统
│   │   │   ├── index.ts           # 公式系统主入口
│   │   │   ├── registry.ts        # 公式注册器
│   │   │   ├── base/              # 基础公式类
│   │   │   │   ├── BaseFormula.ts
│   │   │   │   └── types.ts
│   │   │   └── functions/         # 具体公式实现
│   │   │       ├── lmdi.ts
│   │   │       ├── financial/     # 金融类公式
│   │   │       ├── statistical/   # 统计类公式
│   │   │       └── custom/        # 自定义公式
│   │   ├── events/                 # 事件系统
│   │   │   ├── index.ts
│   │   │   ├── EventBus.ts
│   │   │   └── types.ts
│   │   └── types/                  # 类型定义
│   │       ├── index.ts
│   │       ├── workbook.ts
│   │       └── formula.ts
│   ├── plugin/                     # 插件系统
│   │   ├── index.ts               # 插件管理器
│   │   ├── PluginManager.ts       # 插件管理器实现
│   │   ├── BasePlugin.ts          # 插件基类
│   │   └── types.ts               # 插件类型定义
│   └── store/                      # 状态管理
│       ├── index.ts
│       ├── modules/
│       │   ├── workbook.ts        # 工作簿状态
│       │   ├── ui.ts              # UI状态
│       │   ├── formula.ts         # 公式状态
│       │   └── plugin.ts          # 插件状态
│       └── types.ts
├── plugins/                        # 插件实现
│   ├── file-tree/                  # 文件树插件
│   │   ├── index.ts
│   │   ├── FileTreePlugin.ts
│   │   ├── components/
│   │   │   ├── FileTree.vue
│   │   │   ├── FileNode.vue
│   │   │   └── FileActions.vue
│   │   ├── store/
│   │   │   └── fileTree.ts
│   │   └── api/
│   │       └── fileTree.js
│   ├── ai-chat/                    # AI聊天插件
│   │   ├── index.ts
│   │   ├── AiChatPlugin.ts
│   │   ├── components/
│   │   │   ├── ChatPanel.vue
│   │   │   ├── MessageList.vue
│   │   │   └── InputBox.vue
│   │   ├── store/
│   │   │   └── aiChat.ts
│   │   └── api/
│   │       └── aiChat.js
│   └── pivot-table/                # 数据透视表插件
│       ├── index.ts
│       ├── PivotTablePlugin.ts
│       ├── components/
│       │   ├── PivotPanel.vue
│       │   ├── FieldList.vue
│       │   └── PivotConfig.vue
│       ├── store/
│       │   └── pivotTable.ts
│       └── api/
│           └── pivotTable.js
├── layouts/                        # 布局组件
│   ├── MainLayout.vue             # 主布局
│   ├── components/
│   │   ├── Header.vue
│   │   ├── Sidebar.vue
│   │   ├── ContentArea.vue
│   │   └── PanelContainer.vue
│   └── types.ts
├── view/
│   └── main/
│       └── index.vue              # 简化的主页面
└── composables/                    # 组合式函数
    ├── useUniver.ts               # Univer相关逻辑
    ├── usePlugin.ts               # 插件相关逻辑
    └── useWorkbook.ts             # 工作簿相关逻辑
```

## 3. 核心模块设计

### 3.1 表格引擎 (Engine)
负责Univer实例的创建、配置和生命周期管理。

### 3.2 公式系统 (Formula System)
- **注册器模式**：统一管理公式注册和卸载
- **插件化公式**：支持动态加载公式模块
- **类型安全**：完整的TypeScript类型定义

### 3.3 插件系统 (Plugin System)
- **生命周期管理**：install、activate、deactivate、uninstall
- **依赖管理**：插件间依赖关系处理
- **热插拔**：运行时动态加载/卸载插件

### 3.4 状态管理 (State Management)
- **模块化Store**：按功能域划分状态模块
- **响应式通信**：插件间通过状态变化通信
- **持久化**：关键状态的本地存储

### 3.5 事件系统 (Event System)
- **发布订阅模式**：解耦组件间通信
- **类型安全事件**：TypeScript事件类型定义
- **事件命名空间**：避免事件名冲突

## 4. 插件开发规范

### 4.1 插件基类
```typescript
abstract class BasePlugin {
  abstract name: string
  abstract version: string
  abstract dependencies?: string[]
  
  abstract install(): Promise<void>
  abstract activate(): Promise<void>
  abstract deactivate(): Promise<void>
  abstract uninstall(): Promise<void>
}
```

### 4.2 插件注册
```typescript
// 插件自动注册机制
export default {
  name: 'file-tree',
  version: '1.0.0',
  plugin: FileTreePlugin,
  dependencies: ['workbook']
}
```

## 5. 状态管理方案

### 5.1 状态模块划分
- **workbook**: 工作簿数据、当前选中等
- **ui**: 界面状态、面板显示/隐藏等
- **formula**: 公式注册状态、计算结果缓存等
- **plugin**: 插件状态、配置等

### 5.2 跨插件通信
通过Pinia状态变化和事件总线实现插件间通信。

## 6. 实施计划

### 阶段一：核心重构
1. 重构表格引擎封装
2. 实现公式注册系统
3. 建立基础状态管理

### 阶段二：插件系统
1. 实现插件管理器
2. 重构现有功能为插件
3. 建立插件开发规范

### 阶段三：功能插件
1. 开发文件树插件
2. 开发AI聊天插件
3. 开发数据透视表插件

### 阶段四：优化完善
1. 性能优化
2. 错误处理完善
3. 文档和测试补充

## 7. 技术选型

- **前端框架**: Vue 3 + Composition API
- **状态管理**: Pinia
- **表格引擎**: Univer
- **构建工具**: Vite
- **类型检查**: TypeScript
- **UI组件**: Element Plus (继承现有)

## 8. 扩展性考虑

### 8.1 公式扩展
- 支持公式分类管理
- 支持公式版本控制
- 支持公式权限控制

### 8.2 插件扩展
- 支持第三方插件
- 支持插件市场
- 支持插件配置界面

### 8.3 主题扩展
- 支持多主题切换
- 支持自定义主题
- 支持插件主题适配