# 项目背景说明 (PROJECT_BACKGROUND.md)

## 1. 项目概述 (Project Overview)

本项是一个名为 **Sugar (数格)** 的智能化在线数据分析与协作平台。项目基于成熟的 [**Gin-Vue-Admin**](https://github.com/flipped-aurora/gin-vue-admin) 框架进行深度定制开发，前端核心交互界面采用了高性能的开源表格引擎 [**Univer**](https://github.com/dream-num/univer)。

项目的核心目标是打造一个集数据连接、语义建模、智能分析与团队协作为一体的在线 SaaS 应用。它旨在让业务用户能像使用电子表格一样方便地进行复杂的数据查询和分析，同时通过集成的 AI 能力和强大的权限系统，满足企业级应用的需求。

## 2. 技术栈 (Technology Stack)

### 2.1. 后端 (Backend)

*   **框架**: [`Gin`](server/go.mod:13) (v1.10.0) - 高性能的 Go Web 框架。
*   **语言**: Go (v1.23)
*   **数据库**: 主要使用 [`MySQL`](server/go.mod:49) (v1.5.7)，但通过 [`GORM`](server/go.mod:53) (v1.25.12) ORM 实现了对 PostgreSQL, SQL Server, SQLite 的兼容支持。
*   **权限控制**: [`Casbin`](server/go.mod:10) (v2.103.0)
*   **缓存**: [`Go-Redis`](server/go.mod:30) (v9.7.0)
*   **配置管理**: [`Viper`](server/go.mod:34) (v1.19.0)
*   **日志**: [`Zap`](server/go.mod:44) (v1.27.0)

### 2.2. 前端 (Frontend)

*   **框架**: [`Vue.js`](web/package.json:48) (v3.5.7)
*   **构建工具**: [`Vite`](web/package.json:76) (v6.2.3)
*   **UI 库**: [`Element Plus`](web/package.json:30) (v2.10.2)
*   **核心表格引擎**: [`Univer`](web/package.json:17) (v0.10.1)
*   **状态管理**: [`Pinia`](web/package.json:39) (v2.2.2)
*   **路由**: [`Vue Router`](web/package.json:52) (v4.4.3)
*   **HTTP客户端**: [`Axios`](web/package.json:26) (v1.8.2)

## 3. 项目结构 (Project Structure)

```
.
├── docs/                      # 文档
│   └── Sugar表结构设计.sql     # 核心数据库 Schema
├── server/                    # 后端 Go 项目
│   ├── api/v1/sugar/          # "Sugar" 业务模块的 API 接口
│   ├── model/sugar/           # "Sugar" 业务模块的 GORM 模型
│   ├── service/sugar/         # "Sugar" 业务模块的服务逻辑
│   └── router/sugar/          # "Sugar" 业务模块的路由
└── web/                       # 前端 Vue 项目
    ├── src/
    |   ├── api/sugar/         # "Sugar" 业务模块的前端 API 接口
    │   ├── view/sugar/        # "Sugar" 业务模块的视图
    │   │   └── index.vue      # 核心 Univer 表格页面
    │   ├── plugins/           # 自定义插件目录
    │   │   ├── custom-formulas/ # 自定义公式
    │   │   └── univer-core/   # univer组件
    │   └── core/              # 前端核心逻辑（根据设计文档）
    └── package.json           # 前端依赖
```

## 4. 核心功能与设计 (Core Features & Design)

### 4.1. 后端核心逻辑

后端基于 `gin-vue-admin` 的代码生成器创建了初始的 CRUD 接口，并在此基础上扩展了复杂业务逻辑。

*   **统一资源管理**: 通过 [`sugar_workspaces`](docs/Sugar表结构设计.sql:47) 表统一管理文件和文件夹，形成层级结构。
*   **多租户与协作**: 以“团队”(`sugar_teams`)为核心进行数据隔离和协作，支持团队成员和多种角色。
*   **语义层 (Semantic Layer)**:
    *   **数据连接器** ([`sugar_db_connections`](docs/Sugar表结构设计.sql:130)): 允许用户配置并安全地存储外部数据库（如 MySQL, PostgreSQL 等）的连接信息。
    *   **语义模型** ([`sugar_semantic_models`](docs/Sugar表结构设计.sql:168)): 项目的核心创新之一。它允许管理员将物理数据表封装成面向业务的、易于理解的“模型”。模型中定义了可查询的维度、可计算的指标以及参数化的筛选条件，前端用户无需编写 SQL 即可通过简单交互进行数据探索。
*   **AI Agent 集成**: [`sugar_agents`](docs/Sugar表结构设计.sql:204) 表定义了可被公式系统调用的 AI Agent。这使得用户可以在表格中通过类似 `=AI_AGENT(...)` 的公式，将单元格数据发送给外部 AI 服务进行处理，实现数据清洗、分析、预测等高级功能。
*   **精细化权限控制**:
    *   通过 `sugar_workspace_permissions` 实现对文件/文件夹的显式授权。
    *   通过 [`sugar_city_permissions`](docs/Sugar表结构设计.sql:242) 和 `permission_key_column` 字段的设计，实现了行级安全 (Row-Level Security)，确保不同权限的用户只能看到其有权访问的数据行。
*   **审计日志**: [`sugar_execution_logs`](docs/Sugar表结构设计.sql:279) 记录所有高成本的后台任务（如数据库查询、AI 调用），用于审计、计费和调试。

### 4.2. 前端核心逻辑

前端的核心是构建一个高度可扩展的、基于 Univer 的表格应用。

*   **Univer 集成**: 在 [`web/src/view/sugar/index.vue`](web/src/view/sugar/index.vue:1-218) 中，项目成功初始化了 Univer 实例。
*   **插件化与模块化**:
    *   前端采用了先进的插件化设计理念。文件树、AI 聊天面板、数据透视表等功能被设计为可独立加载、卸载的插件。
    *   这种设计通过统一的**插件管理器**、**事件总线**和 **Pinia 状态管理**实现，保证了核心系统的整洁和高度可扩展性。
*   **自定义公式**: 项目展示了扩展 Univer 公式系统的能力。在 [`web/src/view/sugar/index.vue`](web/src/view/sugar/index.vue:16) 中，一个名为 `LMDI` 的自定义金融公式被成功注册和使用，验证了公式系统的扩展性。

## 5. 如何理解和贡献

*   **从数据库开始**: 理解项目的最佳起点是 [`docs/Sugar表结构设计.sql`](docs/Sugar表结构设计.sql)。该文件清晰地定义了项目的核心实体和它们之间的关系。
*   **后端先行**: 后端逻辑相对直接，可以从 `server/router/sugar/` 入手，跟踪一个 API 请求从路由、API 处理函数、服务层到数据模型的完整流程。
*   **前端看设计**: 对于前端，建议先阅读 [`docs/architecture-design.md`](docs/architecture-design.md) 来理解其宏大的设计目标和插件化思想，然后再查看 [`web/src/view/sugar/index.vue`](web/src/view/sugar/index.vue) 的具体实现。
*   **关注核心抽象**: 本项目的精髓在于其“语义模型”和“AI Agent”的抽象。理解这两部分的设计意图，是理解项目长远价值的关键。