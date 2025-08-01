-- SQL Schema For Project "Sugar (数-)" - V2 Enhanced Version
-- Design by: Gemini
-- Dialect: MySQL 8+

-- =================================================================
-- Section 1: Unified Team and Permission Management
-- 统一的团队与权限管理
-- =================================================================

-- 团队表: 统一管理团队和个人空间
CREATE TABLE `sugar_teams` (
    `id` CHAR(36) NOT NULL,
    `team_name` VARCHAR(100) NOT NULL,
    `owner_id` VARCHAR(20) NOT NULL COMMENT '团队创建者/个人空间的所有者',
    `is_personal` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为个人空间团队 (true代表个人空间)',
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(20) NULL,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_sugar_teams_owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='统一管理团队和个人空间，个人空间是is_personal=true的特殊团队';

-- 团队成员表: (结构不变)
CREATE TABLE `sugar_team_members` (
    `id` BIGINT AUTO_INCREMENT NOT NULL,
    `team_id` CHAR(36) NOT NULL,
    `user_id` VARCHAR(20) NOT NULL,
    `role` ENUM('owner', 'admin', 'editor', 'viewer') NOT NULL,
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(20) NULL,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_team_user` (`team_id`, `user_id`),
    INDEX `idx_sugar_team_members_user_id` (`user_id`),
    FOREIGN KEY (`team_id`) REFERENCES `sugar_teams`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户与团队的多对多关系表';

-- =================================================================
-- Section 2: Resource and Content Management (Simplified)
-- 资源与内容管理 (简化后)
-- =================================================================

-- 工作区对象表: (简化归属)
CREATE TABLE `sugar_workspaces` (
    `id` CHAR(36) NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    `type` ENUM('folder', 'file') NOT NULL,
    `parent_id` CHAR(36) NULL,
    `team_id` CHAR(36) NOT NULL COMMENT '资源统一归属于团队',
    `content` JSON NULL,
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(20) NULL,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_parent_name_team` (`parent_id`, `team_id`, `name`),
    INDEX `idx_sugar_workspaces_team_id` (`team_id`),
    FOREIGN KEY (`parent_id`) REFERENCES `sugar_workspaces`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`team_id`) REFERENCES `sugar_teams`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='统一存储文件和文件夹，归属逻辑简化为仅关联team_id';


-- 文件历史版本表
CREATE TABLE `sugar_file_versions` (
    `id` BIGINT AUTO_INCREMENT NOT NULL,
    `file_id` CHAR(36) NOT NULL,
    `version_number` INTEGER NOT NULL,
    `content` JSON NOT NULL,
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_file_version` (`file_id`, `version_number`),
    FOREIGN KEY (`file_id`) REFERENCES `sugar_workspaces`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='存储文件的历史版本，用于版本回溯';


-- =================================================================
-- Section 3: Sharing and Permissions
-- 共享与权限控制
-- =================================================================

-- 权限授予表: 精细化控制用户/团队对文件/文件夹的权限
CREATE TABLE `sugar_workspace_permissions` (
    `id` BIGINT AUTO_INCREMENT NOT NULL,
    `workspace_id` CHAR(36) NOT NULL,
    `grantee_type` ENUM('user', 'team') NOT NULL COMMENT '授权对象类型',
    `grantee_id` VARCHAR(36) NOT NULL COMMENT 'user_id 或 team_id',
    `permission_level` ENUM('editor', 'commenter', 'viewer') NOT NULL COMMENT '权限级别',
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(20) NULL,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_workspace_grantee` (`workspace_id`, `grantee_type`, `grantee_id`),
    INDEX `idx_sugar_workspace_permissions_grantee` (`grantee_type`, `grantee_id`),
    FOREIGN KEY (`workspace_id`) REFERENCES `sugar_workspaces`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='精细化控制用户/团队对文件/文件夹的权限';


-- 公开分享链接表
CREATE TABLE `sugar_share_links` (
    `id` BIGINT AUTO_INCREMENT NOT NULL,
    `workspace_id` CHAR(36) NOT NULL,
    `token` VARCHAR(50) NOT NULL,
    `permission_level` ENUM('editor', 'viewer') NOT NULL,
    `password_hash` VARCHAR(255) NULL,
    `expires_at` TIMESTAMP NULL DEFAULT NULL,
    `is_active` BOOLEAN NOT NULL DEFAULT true,
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(20) NULL,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_token` (`token`),
    INDEX `idx_sugar_share_links_workspace_id` (`workspace_id`),
    FOREIGN KEY (`workspace_id`) REFERENCES `sugar_workspaces`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='存储通过链接分享的配置';


-- =================================================================
-- Section 3: Semantic Layer and Data Connectors
-- 语义层与数据连接器
-- =================================================================

-- 数据库连接配置表: (增加本库支持)
CREATE TABLE `sugar_db_connections` (
    `id` CHAR(36) NOT NULL,
    `name` VARCHAR(100) NOT NULL,
    `team_id` CHAR(36) NOT NULL COMMENT '所有者团队ID',
    `db_type` VARCHAR(20) NOT NULL,
    `is_internal` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为Sugar内部数据库 (true代表本库)',
    `host` VARCHAR(255) NULL COMMENT '内部数据库可为null',
    `port` INTEGER NULL COMMENT '内部数据库可为null',
    `username` VARCHAR(100) NULL COMMENT '内部数据库可为null',
    `encrypted_password` TEXT NULL COMMENT '内部数据库可为null',
    `database_name` VARCHAR(100) NULL,
    `ssl_config` JSON NULL,
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(20) NULL,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_team_connection_name` (`team_id`, `name`),
    FOREIGN KEY (`team_id`) REFERENCES `sugar_teams`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='存储外部及内部数据库的连接信息';

-- (新增) 数据库连接共享表
CREATE TABLE `sugar_db_connection_shares` (
    `id` BIGINT AUTO_INCREMENT NOT NULL,
    `connection_id` CHAR(36) NOT NULL COMMENT '被共享的数据库连接ID',
    `team_id` CHAR(36) NOT NULL COMMENT '被授予使用权的团队ID',
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_connection_team` (`connection_id`, `team_id`),
    FOREIGN KEY (`connection_id`) REFERENCES `sugar_db_connections`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`team_id`) REFERENCES `sugar_teams`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='记录数据库连接的共享关系';


-- 语义模型表: (本次优化的核心)
CREATE TABLE `sugar_semantic_models` (
    `id` CHAR(36) NOT NULL,
    `name` VARCHAR(100) NOT NULL COMMENT '模型的业务名称, 如“季度销售报告”',
    `description` TEXT NULL,
    `team_id` CHAR(36) NOT NULL COMMENT '所有者团队ID',
    `connection_id` CHAR(36) NOT NULL COMMENT '关联的数据库连接',
    `source_table_name` VARCHAR(255) NOT NULL COMMENT '源数据库中的真实表名',
    `parameter_config` JSON NOT NULL COMMENT '查询参数配置, 定义用户可用的筛选条件',
    `returnable_columns_config` JSON NOT NULL COMMENT '可返回字段配置, 定义用户可获取的数据列',
    `permission_key_column` VARCHAR(255) NULL COMMENT '用于行级权限判断的字段名, 如 city_code',
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(20) NULL,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_team_model_name` (`team_id`, `name`),
    FOREIGN KEY (`team_id`) REFERENCES `sugar_teams`(`id`),
    FOREIGN KEY (`connection_id`) REFERENCES `sugar_db_connections`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='核心语义层，将物理表封装为业务对象，简化查询';

-- (新增) 语义模型共享表
CREATE TABLE `sugar_semantic_model_shares` (
    `id` BIGINT AUTO_INCREMENT NOT NULL,
    `model_id` CHAR(36) NOT NULL COMMENT '被共享的语义模型ID',
    `team_id` CHAR(36) NOT NULL COMMENT '被授予使用权的团队ID',
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_model_team` (`model_id`, `team_id`),
    FOREIGN KEY (`model_id`) REFERENCES `sugar_semantic_models`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`team_id`) REFERENCES `sugar_teams`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='记录语义模型的共享关系';

-- AI Agent 定义表
CREATE TABLE `sugar_agents` (
    `id` CHAR(36) NOT NULL,
    `name` VARCHAR(100) NOT NULL,
    `description` TEXT NULL,
    `agent_type` ENUM('system', 'custom') NOT NULL COMMENT '系统预置, 团队自定义',
    `team_id` CHAR(36) NOT NULL COMMENT '所有者团队ID',
    `endpoint_config` JSON NOT NULL COMMENT '定义 Agent 的调用方式, 如 API URL, headers 等',
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(20) NULL,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_agent_name` (`name`),
    INDEX `idx_sugar_agents_team_id` (`team_id`),
    FOREIGN KEY (`team_id`) REFERENCES `sugar_teams`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='定义可被 CALL_AGENT 公式调用的 AI Agent';

-- (新增) AI Agent 共享表
CREATE TABLE `sugar_agent_shares` (
    `id` BIGINT AUTO_INCREMENT NOT NULL,
    `agent_id` CHAR(36) NOT NULL COMMENT '被共享的AI Agent ID',
    `team_id` CHAR(36) NOT NULL COMMENT '被授予使用权的团队ID',
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_agent_team` (`agent_id`, `team_id`),
    FOREIGN KEY (`agent_id`) REFERENCES `sugar_agents`(`id`) ON DELETE CASCADE,
    FOREIGN KEY (`team_id`) REFERENCES `sugar_teams`(`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='记录AI Agent的共享关系';

-- =================================================================
-- Section 4: Row-Level Security Configuration
-- 行级权限配置
-- =================================================================

-- 用户城市权限映射表
CREATE TABLE `sugar_city_permissions` (
    `id` BIGINT AUTO_INCREMENT NOT NULL,
    `user_id` VARCHAR(20) NOT NULL,
    `city_code` VARCHAR(50) NOT NULL COMMENT '城市编码',
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(20) NULL,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_city` (`user_id`, `city_code`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_city_code` (`city_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='存储用户有权访问的城市数据';


-- 行级权限豁免表
CREATE TABLE `sugar_row_level_overrides` (
    `id` BIGINT AUTO_INCREMENT NOT NULL,
    `user_id` VARCHAR(20) NOT NULL,
    `description` VARCHAR(255) NULL COMMENT '配置原因说明',
    `created_by` VARCHAR(20) NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_by` VARCHAR(20) NULL,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='配置可忽略行级权限的用户 (如管理员)';


-- =================================================================
-- Section 5: Logging
-- 日志记录
-- =================================================================

-- 任务执行日志表
CREATE TABLE `sugar_execution_logs` (
    `id` BIGINT AUTO_INCREMENT NOT NULL,
    `log_type` ENUM('db_query', 'ai_agent') NOT NULL,
    `workspace_id` CHAR(36) NULL,
    `user_id` VARCHAR(20) NULL,
    `connection_id` CHAR(36) NULL,
    `agent_id` CHAR(36) NULL,
    `input_payload` JSON NULL,
    `status` ENUM('pending', 'success', 'failed', 'timeout') NOT NULL,
    `result_summary` TEXT NULL,
    `duration_ms` INTEGER NULL,
    `executed_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    INDEX `idx_sugar_execution_logs_user_id` (`user_id`),
    INDEX `idx_sugar_execution_logs_workspace_id` (`workspace_id`),
    INDEX `idx_sugar_execution_logs_type_status` (`log_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='记录所有高成本的后台任务执行，用于审计、计费和调试';