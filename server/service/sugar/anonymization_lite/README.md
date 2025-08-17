# Sugar 简化数据匿名化服务 (Lite版)

## 概述

Sugar 简化数据匿名化服务是针对贡献度分析场景优化的轻量级数据隐私保护系统。相比完整版本，lite版专注于基本的匿名化功能，去除了复杂的差分隐私、K-匿名性等高级算法，采用统一的匿名标准。

## 🎯 设计目标

### 简化原则
1. **统一匿名标准** - 不需要设置匿名程度，采用统一配置
2. **专注贡献度** - 数值类型只返回贡献度，不返回基期值和当期值
3. **保留语义映射** - 延用语义映射功能，支持维度下钻和聚合
4. **场景化优化** - 专为"期末比期初"、"6月比5月"、"实际比预算"等对比分析优化

### 适用场景
- 分析期末金额比期初金额增加的原因
- 6月比5月增加的原因  
- 实际比预算低的原因
- 基本只需要考虑在不同维度上的下钻、聚合

## 🏗️ 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    简化匿名化服务架构                        │
├─────────────────────────────────────────────────────────────┤
│  输入层                                                      │
│  ├── 本期数据查询                                            │
│  ├── 基期数据查询                                            │
│  └── 分组维度配置                                            │
├─────────────────────────────────────────────────────────────┤
│  计算层                                                      │
│  ├── 并发数据获取                                            │
│  ├── 贡献度计算（简化）                                      │
│  └── 正负向驱动判断                                          │
├─────────────────────────────────────────────────────────────┤
│  匿名化层                                                    │
│  ├── 维度语义映射                                            │
│  ├── 值代号生成                                              │
│  └── 轻微噪声添加                                            │
├─────────────────────────────────────────────────────────────┤
│  输出层                                                      │
│  ├── AI可读数据                                              │
│  ├── 解码映射表                                              │
│  └── 响应解码                                                │
└─────────────────────────────────────────────────────────────┘
```

## 📦 核心组件

### 1. LiteConfig (简化配置)
```go
type LiteConfig struct {
    UseSemanticMapping bool    // 是否使用语义映射
    NoiseLevel         float64 // 统一噪声级别 (0.0-1.0)
    RandomSeed         int64   // 随机种子
}
```

### 2. ContributionItem (贡献度项，简化版)
```go
type ContributionItem struct {
    DimensionValues     map[string]interface{} // 维度值组合
    ContributionPercent float64                // 贡献度百分比
    IsPositiveDriver    bool                   // 是否为正向驱动因子
    // 移除了 CurrentValue, BaseValue, ChangeValue
}
```

### 3. LiteAnonymizationSession (简化会话)
- 维护映射关系 (ForwardMap/ReverseMap)
- 存储AI可读数据
- 提供解码功能
- 统计信息记录

## 🔧 核心功能

### 1. 数据获取与计算
- **并发获取**: 同时获取本期和基期数据
- **贡献度计算**: 修复原版本的计算公式错误
- **正负向判断**: 完善边界情况处理

### 2. 匿名化处理
- **语义映射**: 
  - 地理维度 → LOC01, LOC02...
  - 产品维度 → PRD_HV01, PRD_ST01...
  - 时间维度 → TIME01, TIME02...
  - 组织维度 → ORG01, ORG02...
- **轻微噪声**: 可配置的轻微数值扰动
- **值分类**: HV(高价值), ST(标准), BS(基础), T1(一线城市)等

### 3. AI交互
- **数据序列化**: 结构化文本格式，AI友好
- **响应解码**: 将匿名代号还原为原始值
- **会话管理**: 维护匿名化会话状态

## 📋 使用指南

### 1. 基础使用

```go
import "github.com/flipped-aurora/gin-vue-admin/server/service/sugar/anonymization_lite"

// 创建服务
config := anonymization_lite.DefaultLiteConfig()
service := anonymization_lite.NewLiteAnonymizationService(config)

// 构建请求
request := &anonymization_lite.AIAnalysisRequest{
    ModelName:    "销售数据模型",
    TargetMetric: "销售金额",
    CurrentPeriodFilters: map[string]interface{}{
        "年份": "2024", "月份": "12",
    },
    BasePeriodFilters: map[string]interface{}{
        "年份": "2024", "月份": "11",
    },
    GroupByDimensions: []string{"区域", "产品类别"},
    UserID: "user123",
}

// 执行匿名化处理
ctx := context.Background()
session, err := service.ProcessAndAnonymize(ctx, request)
if err != nil {
    log.Fatal(err)
}

// 获取AI可读数据
aiData := session.GetAIReadyData()
dataText, _ := session.SerializeToText()

// ... 发送给AI分析 ...

// 解码AI响应
aiResponse := "LOC01的PRD_HV01表现突出，贡献度达到45.2%..."
decodedResponse, _ := session.DecodeAIResponse(aiResponse)
// 结果: "华东区域的A产品表现突出，贡献度达到45.2%..."
```

### 2. 配置选择

```go
// 默认配置 (推荐)
config := anonymization_lite.DefaultLiteConfig()

// 自定义配置
config := &anonymization_lite.LiteConfig{
    UseSemanticMapping: true,  // 启用语义映射
    NoiseLevel:         0.05,  // 5%轻微噪声
    RandomSeed:         12345, // 固定种子
}
```

## 🔄 与完整版对比

| 特性 | 完整版 | Lite版 | 说明 |
|------|--------|--------|------|
| **复杂度** | 高 | 低 | 移除高级算法 |
| **配置** | 多种模式 | 统一标准 | 简化配置选择 |
| **数值返回** | 完整 | 仅贡献度 | 专注核心指标 |
| **算法** | 差分隐私+K匿名+L多样性 | 语义映射+轻微噪声 | 大幅简化 |
| **性能** | 中等 | 高 | 减少计算开销 |
| **维护性** | 复杂 | 简单 | 代码更易维护 |
| **适用场景** | 通用 | 贡献度分析 | 场景化优化 |

## ⚡ 性能优化

### 1. 并发处理
- 本期和基期数据并发获取
- 减少数据库查询等待时间

### 2. 内存优化
- 移除不必要的数据存储
- 简化数据结构

### 3. 算法简化
- 去除复杂的隐私算法
- 专注核心匿名化功能

## 📊 输出示例

### 匿名化数据格式
```
【简化匿名化贡献度分析数据】

数据字段说明：
- 维度代号：表示业务维度（如区域、产品等）
- 值代号：表示具体的维度值
- contribution_percent：贡献度百分比
- is_positive_driver：是否为正向驱动因子

数据内容：
项目 1:
  LOC01: LOC01_HV01
  PRD01: PRD01_HV01
  贡献度: 45.23%
  正向驱动: true

项目 2:
  LOC01: LOC01_ST01
  PRD01: PRD01_ST01
  贡献度: -12.45%
  正向驱动: false
```

## 🔗 集成说明

要在现有系统中使用lite版本，需要：

1. **修改调用代码**: 将原来的 `anonymization` 包调用改为 `anonymization_lite`
2. **更新AI工具**: 调整工具调用参数，移除不需要的字段
3. **验证结果**: 确保匿名化和解码功能正常

## ⚠️ 注意事项

1. **数据范围**: 适用于基本的贡献度分析，不适合需要复杂隐私保护的场景
2. **噪声控制**: 轻微噪声可能不足以满足严格的隐私要求
3. **向后兼容**: 与完整版API不完全兼容，需要代码调整

---

**Sugar 简化数据匿名化服务**专为贡献度分析场景优化，在保证基本数据安全的前提下，大幅简化了使用复杂度和维护成本。