# Sugar 数据加密改进方案

## 🚨 当前安全漏洞分析

### 严重问题：toolCall 字段完全未加密

**当前泄露的敏感信息**：
```json
{
  "toolCall": "{\"basePeriodFilters\":{\"预实类型\":\"预算数\",\"城市名称\":\"济南市\",\"统计期间\":\"202505\"},\"currentPeriodFilters\":{\"预实类型\":\"实际数\",\"城市名称\":\"济南市\",\"统计期间\":\"202505\"}}"
}
```

**泄露风险**：
- ❌ 直接暴露真实城市名称"济南市"
- ❌ 暴露业务类型"预算数"、"实际数"
- ❌ 暴露时间信息"202505"
- ❌ 暴露完整的查询条件结构

### 次要问题：数值数据保护不足

**当前数据示例**：
```json
{
  "aiDataText": "项目 1: D01: D01_V01, 贡献度: 0.00%, 变化值: 0.26, 本期值: 0.26, 基期值: 0.00"
}
```

**风险点**：
- ⚠️ 变化值、本期值、基期值为真实数值
- ⚠️ 贡献度百分比未添加差分隐私噪声
- ⚠️ 可通过数值组合反推业务规模

## 🛡️ 安全改进方案

### 方案1：完全加密策略（推荐）

#### 1.1 toolCall 字段完全匿名化
```go
// 新增函数：匿名化 toolCall 数据
func (s *AnonymizationService) anonymizeToolCallData(toolCall string, session *AnonymizationSession) (string, error) {
    var toolCallData map[string]interface{}
    if err := json.Unmarshal([]byte(toolCall), &toolCallData); err != nil {
        return "", err
    }
    
    // 递归匿名化所有字段值
    anonymizedData := s.anonymizeMapRecursively(toolCallData, session)
    
    anonymizedBytes, err := json.Marshal(anonymizedData)
    if err != nil {
        return "", err
    }
    
    return string(anonymizedBytes), nil
}

func (s *AnonymizationService) anonymizeMapRecursively(data map[string]interface{}, session *AnonymizationSession) map[string]interface{} {
    result := make(map[string]interface{})
    
    for key, value := range data {
        switch v := value.(type) {
        case string:
            // 匿名化字符串值
            if s.isSensitiveValue(v) {
                result[key] = s.getOrCreateAnonymizedValue(session, key, v, make(map[string]int))
            } else {
                result[key] = v
            }
        case map[string]interface{}:
            // 递归处理嵌套对象
            result[key] = s.anonymizeMapRecursively(v, session)
        default:
            result[key] = value
        }
    }
    
    return result
}
```

#### 1.2 数值数据差分隐私保护
```go
// 修改数值序列化函数，添加差分隐私噪声
func (s *SugarFormulaAiService) serializeAnonymizedDataToTextWithDP(data []map[string]interface{}, config *AdvancedAnonymizationConfig) (string, error) {
    var builder strings.Builder
    builder.WriteString("【匿名化贡献度分析数据】\n")
    builder.WriteString("说明：以下数据已进行差分隐私处理，所有数值都添加了校准噪声\n\n")
    
    noiseGen := NewNoiseGenerator(config)
    
    for i, item := range data {
        builder.WriteString(fmt.Sprintf("项目 %d:\n", i+1))
        
        // 输出匿名化维度
        for key, value := range item {
            if strings.HasPrefix(key, "D") && !strings.Contains(key, "_") {
                builder.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
            }
        }
        
        // 添加差分隐私噪声的数值数据
        if cp, ok := item["contribution_percent"].(float64); ok {
            noisyCP := cp + noiseGen.GenerateLaplaceNoise(config.GlobalSensitivity/config.Epsilon)
            builder.WriteString(fmt.Sprintf("  贡献度: %.2f%%\n", noisyCP))
        }
        
        if cv, ok := item["change_value"].(float64); ok {
            noisyCV := cv + noiseGen.GenerateLaplaceNoise(config.GlobalSensitivity/config.Epsilon)
            builder.WriteString(fmt.Sprintf("  变化值: %.2f\n", noisyCV))
        }
        
        if curr, ok := item["current_value"].(float64); ok {
            noisyCurr := curr + noiseGen.GenerateLaplaceNoise(config.GlobalSensitivity/config.Epsilon)
            builder.WriteString(fmt.Sprintf("  本期值: %.2f\n", noisyCurr))
        }
        
        if base, ok := item["base_value"].(float64); ok {
            noisyBase := base + noiseGen.GenerateLaplaceNoise(config.GlobalSensitivity/config.Epsilon)
            builder.WriteString(fmt.Sprintf("  基期值: %.2f\n", noisyBase))
        }
        
        if ipd, ok := item["is_positive_driver"]; ok {
            builder.WriteString(fmt.Sprintf("  正向驱动: %v\n", ipd))
        }
        
        builder.WriteString("\n")
    }
    
    return builder.String(), nil
}
```

### 方案2：选择性加密策略

#### 2.1 只传递贡献度相关数据
```go
// 新结构：最小化AI输入数据
type MinimalAIData struct {
    AnonymizedContributions []AnonymizedContribution `json:"anonymized_contributions"`
    Summary                AISummary                 `json:"summary"`
    IsEncrypted            bool                      `json:"is_encrypted"`
}

type AnonymizedContribution struct {
    DimensionCode      string  `json:"dimension_code"`       // 如：D01_V01
    ContributionLevel  string  `json:"contribution_level"`   // 如：HIGH, MEDIUM, LOW
    DirectionType      string  `json:"direction_type"`       // 如：POSITIVE, NEGATIVE
    SignificanceLevel  int     `json:"significance_level"`   // 1-5级重要性
}

type AISummary struct {
    TotalItems        int     `json:"total_items"`
    PositiveDrivers   int     `json:"positive_drivers"`
    NegativeDrivers   int     `json:"negative_drivers"`
    DataQualityScore  float64 `json:"data_quality_score"`
}
```

#### 2.2 实现选择性加密服务
```go
func (s *AnonymizationService) CreateMinimalAIData(contributions []ContributionItem, session *AnonymizationSession) (*MinimalAIData, error) {
    var anonymizedContribs []AnonymizedContribution
    
    positiveCount := 0
    negativeCount := 0
    
    for _, contrib := range contributions {
        // 获取匿名化维度代码
        var dimensionCode string
        for dimName, dimValue := range contrib.DimensionValues {
            anonymizedDim := session.forwardMap[dimName]
            anonymizedVal := session.forwardMap[fmt.Sprintf("%s:%v", dimName, dimValue)]
            dimensionCode = anonymizedVal // 使用第一个维度值作为代表
            break
        }
        
        // 将贡献度转换为等级
        contributionLevel := s.categorizeContributionLevel(contrib.ContributionPercent)
        
        // 统计正负向
        directionType := "POSITIVE"
        if contrib.IsPositiveDriver {
            positiveCount++
        } else {
            directionType = "NEGATIVE"
            negativeCount++
        }
        
        // 计算重要性等级
        significanceLevel := s.calculateSignificanceLevel(contrib.ContributionPercent)
        
        anonymizedContribs = append(anonymizedContribs, AnonymizedContribution{
            DimensionCode:     dimensionCode,
            ContributionLevel: contributionLevel,
            DirectionType:     directionType,
            SignificanceLevel: significanceLevel,
        })
    }
    
    summary := AISummary{
        TotalItems:       len(contributions),
        PositiveDrivers:  positiveCount,
        NegativeDrivers:  negativeCount,
        DataQualityScore: s.calculateDataQualityScore(contributions),
    }
    
    return &MinimalAIData{
        AnonymizedContributions: anonymizedContribs,
        Summary:                summary,
        IsEncrypted:            true,
    }, nil
}

func (s *AnonymizationService) categorizeContributionLevel(percent float64) string {
    absPercent := math.Abs(percent)
    if absPercent >= 10 {
        return "HIGH"
    } else if absPercent >= 3 {
        return "MEDIUM"
    } else {
        return "LOW"
    }
}

func (s *AnonymizationService) calculateSignificanceLevel(percent float64) int {
    absPercent := math.Abs(percent)
    if absPercent >= 20 {
        return 5
    } else if absPercent >= 10 {
        return 4
    } else if absPercent >= 5 {
        return 3
    } else if absPercent >= 1 {
        return 2
    } else {
        return 1
    }
}
```

## 🔧 具体修复实现

### 修复1：toolCall 字段加密
```go
// 在 sugar_formula_ai_service.go 中修改
func (s *SugarFormulaAiService) handleSmartAnonymizedAnalyzer(ctx context.Context, toolCall system.OpenAIToolCall, logCtx *ExecutionLogContext, req *sugarReq.SugarFormulaAiFetchRequest, agent *sugar.SugarAgents, llmConfig *system.LLMConfig) (*sugarRes.SugarFormulaAiResponse, error) {
    // ... 现有逻辑 ...
    
    // 更新日志记录匿名化信息 - 修复前
    if logCtx != nil {
        // ❌ 当前代码直接暴露敏感信息
        anonymizedInputData := map[string]interface{}{
            "aiDataText":   aiDataText,
            "toolCall":     toolCall.Function.Arguments, // 敏感信息泄露
            "mappingCount": len(anonymizedResult.forwardMap),
            "isEncrypted":  true,
        }
        
        // ✅ 修复后：匿名化 toolCall
        anonymizedToolCall, err := s.anonymizeToolCallData(toolCall.Function.Arguments, anonymizedResult)
        if err != nil {
            global.GVA_LOG.Warn("toolCall匿名化失败", zap.Error(err))
            anonymizedToolCall = "[已加密]"
        }
        
        anonymizedInputData := map[string]interface{}{
            "aiDataText":        aiDataText,
            "toolCall":          anonymizedToolCall, // 已匿名化
            "mappingCount":      len(anonymizedResult.forwardMap),
            "isEncrypted":       true,
            "validationEnabled": enableDataValidation,
        }
        
        _ = executionLogService.UpdateExecutionLogWithAnonymization(ctx, logCtx, anonymizedInputData, nil)
    }
    
    // ... 其余逻辑 ...
}
```

### 修复2：数值数据差分隐私保护
```go
// 修改 serializeAnonymizedDataToText 函数
func (s *SugarFormulaAiService) serializeAnonymizedDataToText(data []map[string]interface{}) (string, error) {
    if len(data) == 0 {
        return "", errors.New("匿名化数据为空")
    }
    
    // 获取差分隐私配置
    config := DefaultAdvancedConfig()
    noiseGenerator := NewNoiseGenerator(config)
    
    var builder strings.Builder
    builder.WriteString("【匿名化贡献度分析数据】\n")
    builder.WriteString("说明：以下数据已进行匿名化和差分隐私处理，所有维度名称、值和数值都已加密\n\n")
    
    builder.WriteString("数据字段说明：\n")
    builder.WriteString("- 维度代号（D01, D02等）：表示加密后的业务维度\n")
    builder.WriteString("- 值代号（D01_V01, D01_V02等）：表示加密后的维度值\n")
    builder.WriteString("- contribution_percent：添加噪声后的贡献度百分比\n")
    builder.WriteString("- is_positive_driver：是否为正向驱动因子\n")
    builder.WriteString("- 所有数值都添加了差分隐私噪声保护\n\n")
    
    builder.WriteString("数据内容：\n")
    for i, item := range data {
        builder.WriteString(fmt.Sprintf("项目 %d:\n", i+1))
        
        // 输出维度信息
        for key, value := range item {
            if strings.HasPrefix(key, "D") && !strings.Contains(key, "_") {
                builder.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
            }
        }
        
        // 输出添加噪声的分析数据
        if cp, ok := item["contribution_percent"].(float64); ok {
            // 添加拉普拉斯噪声
            noise := noiseGenerator.GenerateLaplaceNoise(config.GlobalSensitivity / config.Epsilon)
            noisyCP := cp + noise
            builder.WriteString(fmt.Sprintf("  贡献度: %.2f%%\n", noisyCP))
        }
        
        if ipd, ok := item["is_positive_driver"]; ok {
            builder.WriteString(fmt.Sprintf("  正向驱动: %v\n", ipd))
        }
        
        // 对数值字段添加差分隐私噪声
        if cv, ok := item["change_value"].(float64); ok {
            noise := noiseGenerator.GenerateLaplaceNoise(config.GlobalSensitivity / config.Epsilon)
            noisyCV := cv + noise
            builder.WriteString(fmt.Sprintf("  变化值: %.2f\n", noisyCV))
        }
        
        if curr, ok := item["current_value"].(float64); ok {
            noise := noiseGenerator.GenerateLaplaceNoise(config.GlobalSensitivity / config.Epsilon)
            noisyCurr := curr + noise
            builder.WriteString(fmt.Sprintf("  本期值: %.2f\n", noisyCurr))
        }
        
        if base, ok := item["base_value"].(float64); ok {
            noise := noiseGenerator.GenerateLaplaceNoise(config.GlobalSensitivity / config.Epsilon)
            noisyBase := base + noise
            builder.WriteString(fmt.Sprintf("  基期值: %.2f\n", noisyBase))
        }
        
        builder.WriteString("\n")
    }
    
    return builder.String(), nil
}
```

## 📊 安全性评估

### 修复前后对比

| 安全项 | 修复前 | 修复后 |
|--------|---------|---------|
| **toolCall字段** | ❌ 完全暴露 | ✅ 完全匿名化 |
| **城市名称** | ❌ 明文"济南市" | ✅ 代号"LOC01" |
| **业务类型** | ❌ 明文"预算数" | ✅ 代号"TYPE01" |
| **时间信息** | ❌ 明文"202505" | ✅ 代号"TIME01" |
| **数值数据** | ⚠️ 真实数值 | ✅ 差分隐私保护 |
| **贡献度** | ⚠️ 精确百分比 | ✅ 添加噪声 |

### 隐私保护等级

- **修复前**: Level 2 (基础匿名化)
- **修复后**: Level 5 (完全差分隐私保护)

## 🚀 实施建议

### 优先级排序
1. **紧急修复**: toolCall 字段匿名化 (安全漏洞)
2. **重要修复**: 数值数据差分隐私保护
3. **优化改进**: 实施选择性加密策略

### 实施步骤
1. 修复 `sugar_formula_ai_service.go` 中的 toolCall 处理
2. 升级 `serializeAnonymizedDataToText` 函数
3. 添加敏感值检测和匿名化函数
4. 完善差分隐私噪声生成
5. 进行全面测试验证

---

**重要提醒**: 当前的数据泄露问题属于**严重安全漏洞**，建议立即修复，避免敏感业务信息被AI服务提供商获取。

**版本**: v1.0  
**更新时间**: 2025-08-15  
**安全等级**: 🔴 严重  
**负责人**: Principal Software Engineer Roo