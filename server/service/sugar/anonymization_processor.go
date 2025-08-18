package sugar

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sugarRes "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/sugar/anonymization_lite"
	"go.uber.org/zap"
)

// AnonymizationProcessor 匿名化处理器 - 负责数据匿名化和解密
type AnonymizationProcessor struct {
	dataProcessor *DataProcessor
}

// NewAnonymizationProcessor 创建匿名化处理器
func NewAnonymizationProcessor() *AnonymizationProcessor {
	return &AnonymizationProcessor{
		dataProcessor: NewDataProcessor(),
	}
}

// ProcessAnonymizedDataAnalysis 执行匿名化数据分析处理（向后兼容的旧版本方法）
func (ap *AnonymizationProcessor) ProcessAnonymizedDataAnalysis(ctx context.Context, modelName, targetMetric string, currentPeriodFilters, basePeriodFilters map[string]interface{}, groupByDimensions []string, userId string) (*AnonymizationSession, error) {
	global.GVA_LOG.Info("开始匿名化数据处理",
		zap.String("modelName", modelName),
		zap.String("targetMetric", targetMetric),
		zap.Strings("groupByDimensions", groupByDimensions),
		zap.String("userId", userId))

	// 1. 并发获取本期和基期数据
	currentData, baseData, err := ap.dataProcessor.FetchDataConcurrently(ctx, modelName, targetMetric, currentPeriodFilters, basePeriodFilters, groupByDimensions, userId)
	if err != nil {
		return nil, fmt.Errorf("并发获取数据失败: %w", err)
	}

	// 2. 计算贡献度分析
	contributions, err := ap.calculateContributions(currentData, baseData, targetMetric, groupByDimensions)
	if err != nil {
		return nil, fmt.Errorf("计算贡献度失败: %w", err)
	}

	// 3. 创建匿名化会话并进行数据加密
	session, err := ap.createAnonymizedSession(contributions)
	if err != nil {
		return nil, fmt.Errorf("创建匿名化会话失败: %w", err)
	}

	global.GVA_LOG.Info("匿名化数据处理完成",
		zap.Int("contributionCount", len(contributions)),
		zap.Int("aiDataCount", len(session.AIReadyData)),
		zap.Int("mappingCount", len(session.forwardMap)))

	return session, nil
}

// SerializeAnonymizedDataToText 将匿名化数据序列化为文本格式
func (ap *AnonymizationProcessor) SerializeAnonymizedDataToText(data []map[string]interface{}) (string, error) {
	if len(data) == 0 {
		return "", errors.New("匿名化数据为空")
	}

	var builder strings.Builder
	builder.WriteString("【匿名化贡献度分析数据】\n")
	builder.WriteString("说明：以下数据已进行匿名化处理，维度名称和值都已替换为代号\n\n")

	// 添加数据列说明
	builder.WriteString("数据字段说明：\n")
	builder.WriteString("- 维度代号（D01, D02等）：表示敏感业务维度\n")
	builder.WriteString("- 值代号（D01_V01, D01_V02等）：表示具体的维度值\n")
	builder.WriteString("- contribution_percent：贡献度百分比\n")
	builder.WriteString("- is_positive_driver：是否为正向驱动因子\n")
	builder.WriteString("- change_value：变化值\n")
	builder.WriteString("- current_value：本期值\n")
	builder.WriteString("- base_value：基期值\n\n")

	builder.WriteString("数据内容：\n")
	for i, item := range data {
		builder.WriteString(fmt.Sprintf("项目 %d:\n", i+1))

		// 先输出维度信息
		for key, value := range item {
			if strings.HasPrefix(key, "D") && !strings.Contains(key, "_") {
				builder.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
			}
		}

		// 再输出分析数据
		if cp, ok := item["contribution_percent"]; ok {
			builder.WriteString(fmt.Sprintf("  贡献度: %.2f%%\n", cp))
		}
		if ipd, ok := item["is_positive_driver"]; ok {
			builder.WriteString(fmt.Sprintf("  正向驱动: %v\n", ipd))
		}
		if cv, ok := item["change_value"]; ok {
			builder.WriteString(fmt.Sprintf("  变化值: %.2f\n", cv))
		}
		if curr, ok := item["current_value"]; ok {
			builder.WriteString(fmt.Sprintf("  本期值: %.2f\n", curr))
		}
		if base, ok := item["base_value"]; ok {
			builder.WriteString(fmt.Sprintf("  基期值: %.2f\n", base))
		}

		builder.WriteString("\n")
	}

	global.GVA_LOG.Info("匿名化数据序列化完成",
		zap.Int("dataCount", len(data)),
		zap.Int("textLength", len(builder.String())))

	return builder.String(), nil
}

// DecodeAIResponse 解码AI响应中的匿名代号（使用lite版本会话）
func (ap *AnonymizationProcessor) DecodeAIResponse(session *anonymization_lite.LiteAnonymizationSession, aiResponse string) (string, error) {
	if aiResponse == "" {
		return "", nil
	}

	if session == nil {
		global.GVA_LOG.Warn("匿名化会话为空，直接返回原始响应")
		return aiResponse, nil
	}

	global.GVA_LOG.Info("开始lite版本AI响应解码",
		zap.Int("originalLength", len(aiResponse)),
		zap.Int("mappingCount", len(session.ReverseMap)))

	// 使用lite版本的解码功能
	decodedResult, err := session.DecodeAIResponse(aiResponse)
	if err != nil {
		global.GVA_LOG.Error("lite版本解码失败", zap.Error(err))
		return aiResponse, err
	}

	global.GVA_LOG.Info("lite版本AI响应解码完成",
		zap.Int("originalLength", len(aiResponse)),
		zap.Int("decodedLength", len(decodedResult)))

	return decodedResult, nil
}

// DecodeAIResponseLegacy 解码AI响应中的匿名代号（使用旧版本会话，向后兼容）
func (ap *AnonymizationProcessor) DecodeAIResponseLegacy(session *AnonymizationSession, aiText string) (string, error) {
	if session == nil {
		return "", errors.New("匿名化会话为空")
	}

	if aiText == "" {
		global.GVA_LOG.Warn("AI响应为空，无需解码")
		return "", nil
	}

	global.GVA_LOG.Info("开始解码AI响应",
		zap.Int("originalLength", len(aiText)),
		zap.Int("mappingCount", len(session.reverseMap)))

	// 获取所有需要替换的代号，按长度降序排序以避免部分替换问题
	var codes []string
	for code := range session.reverseMap {
		codes = append(codes, code)
	}

	// 按字符串长度降序排序，确保长代号先被替换
	for i := 0; i < len(codes); i++ {
		for j := i + 1; j < len(codes); j++ {
			if len(codes[i]) < len(codes[j]) {
				codes[i], codes[j] = codes[j], codes[i]
			}
		}
	}

	// 执行替换
	decodedText := aiText
	replacementCount := 0
	replacementDetails := make(map[string]string)

	for _, code := range codes {
		originalValue := session.reverseMap[code]
		if strings.Contains(decodedText, code) {
			oldText := decodedText
			decodedText = strings.ReplaceAll(decodedText, code, originalValue)

			// 统计实际替换次数
			occurrences := strings.Count(oldText, code)
			if occurrences > 0 {
				replacementCount += occurrences
				replacementDetails[code] = originalValue

				global.GVA_LOG.Debug("执行代号替换",
					zap.String("code", code),
					zap.String("originalValue", originalValue),
					zap.Int("occurrences", occurrences))
			}
		}
	}

	// 验证解码结果
	if replacementCount == 0 {
		global.GVA_LOG.Warn("未发现需要解码的匿名代号", zap.String("aiText", aiText))
	}

	global.GVA_LOG.Info("AI响应解码完成",
		zap.Int("totalCodes", len(codes)),
		zap.Int("foundCodes", len(replacementDetails)),
		zap.Int("totalReplacements", replacementCount),
		zap.Int("originalLength", len(aiText)),
		zap.Int("decodedLength", len(decodedText)))

	return decodedText, nil
}

// 私有方法

// calculateContributions 计算贡献度分析（向后兼容的旧版本方法）
func (ap *AnonymizationProcessor) calculateContributions(currentData, baseData *sugarRes.SugarFormulaGetResponse, targetMetric string, groupByDimensions []string) ([]ContributionItem, error) {
	// 将数据按维度组合进行分组
	currentGroups := ap.groupDataByDimensions(currentData.Results, groupByDimensions, targetMetric)
	baseGroups := ap.groupDataByDimensions(baseData.Results, groupByDimensions, targetMetric)

	// 计算每个维度组合的贡献度
	var contributions []ContributionItem
	var totalChange float64

	// 获取所有唯一的维度组合
	allKeys := ap.getAllUniqueKeys(currentGroups, baseGroups)

	// 第一轮：计算变化值和总变化
	for _, key := range allKeys {
		currentValue := currentGroups[key]
		baseValue := baseGroups[key]
		changeValue := currentValue - baseValue

		totalChange += changeValue

		// 解析维度值
		dimensionValues := ap.parseDimensionKey(key, groupByDimensions)

		contributions = append(contributions, ContributionItem{
			DimensionValues: dimensionValues,
			CurrentValue:    currentValue,
			BaseValue:       baseValue,
			ChangeValue:     changeValue,
		})
	}

	// 第二轮：计算贡献度百分比和正负向判断
	for i := range contributions {
		if totalChange != 0 {
			contributions[i].ContributionPercent = (contributions[i].ChangeValue / totalChange) * 100
		} else {
			contributions[i].ContributionPercent = 0
		}

		// 判断是否为正向驱动因子
		contributions[i].IsPositiveDriver = (contributions[i].ChangeValue * totalChange) >= 0
	}

	global.GVA_LOG.Info("贡献度计算完成",
		zap.Float64("totalChange", totalChange),
		zap.Int("contributionCount", len(contributions)))

	return contributions, nil
}

// createAnonymizedSession 创建匿名化会话（向后兼容的旧版本方法）
func (ap *AnonymizationProcessor) createAnonymizedSession(contributions []ContributionItem) (*AnonymizationSession, error) {
	session := &AnonymizationSession{
		forwardMap:  make(map[string]string),
		reverseMap:  make(map[string]string),
		AIReadyData: make([]map[string]interface{}, 0),
	}

	// 维度计数器，用于生成唯一代号
	dimensionCounters := make(map[string]int)
	valueCounters := make(map[string]int)

	global.GVA_LOG.Info("开始创建匿名化会话", zap.Int("contributionCount", len(contributions)))

	// 处理每个贡献项
	for i, contribution := range contributions {
		aiItem := make(map[string]interface{})

		// 处理维度值的匿名化
		for dimName, dimValue := range contribution.DimensionValues {
			anonymizedDimName := ap.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
			anonymizedDimValue := ap.getOrCreateAnonymizedValue(session, dimName, fmt.Sprintf("%v", dimValue), valueCounters)

			aiItem[anonymizedDimName] = anonymizedDimValue
		}

		// 添加经过脱敏处理的数值数据
		aiItem["contribution_percent"] = ap.anonymizeNumericValue(contribution.ContributionPercent, "contribution")
		aiItem["is_positive_driver"] = contribution.IsPositiveDriver
		aiItem["change_value"] = ap.anonymizeNumericValue(contribution.ChangeValue, "change")
		aiItem["current_value"] = ap.anonymizeNumericValue(contribution.CurrentValue, "current")
		aiItem["base_value"] = ap.anonymizeNumericValue(contribution.BaseValue, "base")

		session.AIReadyData = append(session.AIReadyData, aiItem)

		// 记录匿名化进度
		if i%10 == 0 || i == len(contributions)-1 {
			global.GVA_LOG.Debug("匿名化进度",
				zap.Int("processed", i+1),
				zap.Int("total", len(contributions)),
				zap.Int("currentMappings", len(session.forwardMap)))
		}
	}

	global.GVA_LOG.Info("匿名化会话创建完成",
		zap.Int("forwardMapSize", len(session.forwardMap)),
		zap.Int("reverseMapSize", len(session.reverseMap)),
		zap.Int("aiDataSize", len(session.AIReadyData)))

	return session, nil
}

// getOrCreateAnonymizedDimension 获取或创建维度名的匿名化代号
func (ap *AnonymizationProcessor) getOrCreateAnonymizedDimension(session *AnonymizationSession, dimName string, counters map[string]int) string {
	// 检查是否已经存在匿名化代号
	if anonymized, exists := session.forwardMap[dimName]; exists {
		return anonymized
	}

	// 生成新的维度代号
	counters["dimension"]++
	anonymized := fmt.Sprintf("D%02d", counters["dimension"])

	// 存储映射关系
	session.forwardMap[dimName] = anonymized
	session.reverseMap[anonymized] = dimName

	return anonymized
}

// getOrCreateAnonymizedValue 获取或创建维度值的匿名化代号
func (ap *AnonymizationProcessor) getOrCreateAnonymizedValue(session *AnonymizationSession, dimName, dimValue string, counters map[string]int) string {
	// 构建完整的键（维度名+值）
	fullKey := fmt.Sprintf("%s:%s", dimName, dimValue)

	// 检查是否已经存在匿名化代号
	if anonymized, exists := session.forwardMap[fullKey]; exists {
		return anonymized
	}

	// 获取维度的匿名化代号
	anonymizedDim := session.forwardMap[dimName]
	if anonymizedDim == "" {
		// 如果维度还没有匿名化，先创建维度代号
		dimensionCounters := make(map[string]int)
		anonymizedDim = ap.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
	}

	// 生成新的值代号
	dimKey := fmt.Sprintf("value_%s", dimName)
	counters[dimKey]++
	anonymized := fmt.Sprintf("%s_V%02d", anonymizedDim, counters[dimKey])

	// 存储映射关系
	session.forwardMap[fullKey] = anonymized
	session.reverseMap[anonymized] = dimValue

	return anonymized
}

// anonymizeNumericValue 对数值进行基础脱敏处理
func (ap *AnonymizationProcessor) anonymizeNumericValue(value float64, valueType string) float64 {
	absValue := math.Abs(value)
	var anonymizedValue float64

	switch valueType {
	case "contribution":
		// 贡献度百分比：添加小幅随机扰动（±5%以内）
		maxPerturbation := 5.0
		perturbation := (rand.Float64() - 0.5) * 2 * maxPerturbation
		anonymizedValue = value + perturbation

		// 确保百分比在合理范围内
		if anonymizedValue > 100.0 {
			anonymizedValue = 100.0
		} else if anonymizedValue < -100.0 {
			anonymizedValue = -100.0
		}

	case "current", "base":
		// 本期值和基期值：根据数值大小应用不同脱敏策略
		if absValue < 1000 {
			// 小数值：添加5-15%的相对扰动
			perturbationRatio := 0.05 + rand.Float64()*0.10 // 5%-15%
			perturbation := absValue * perturbationRatio * (rand.Float64() - 0.5) * 2
			anonymizedValue = value + perturbation
		} else {
			// 大数值：保持数量级，添加一定扰动后舍入
			magnitude := math.Pow(10, math.Floor(math.Log10(absValue)))
			perturbationRatio := 0.10 + rand.Float64()*0.20 // 10%-30%
			perturbation := absValue * perturbationRatio * (rand.Float64() - 0.5) * 2
			anonymizedValue = value + perturbation

			// 根据数量级进行适当舍入
			if magnitude >= 1000 {
				roundTo := magnitude / 100 // 舍入到百位
				anonymizedValue = math.Round(anonymizedValue/roundTo) * roundTo
			}
		}

	case "change":
		// 变化值：保持符号一致性，但添加扰动
		if absValue < 100 {
			// 小变化值：添加10-25%扰动
			perturbationRatio := 0.10 + rand.Float64()*0.15
			perturbation := absValue * perturbationRatio * (rand.Float64() - 0.5) * 2
			anonymizedValue = value + perturbation
		} else {
			// 大变化值：添加15-35%扰动并舍入
			perturbationRatio := 0.15 + rand.Float64()*0.20
			perturbation := absValue * perturbationRatio * (rand.Float64() - 0.5) * 2
			anonymizedValue = value + perturbation

			// 舍入处理
			if absValue >= 1000 {
				anonymizedValue = math.Round(anonymizedValue/10) * 10
			}
		}

	default:
		// 默认策略：添加10%扰动
		perturbationRatio := 0.10
		perturbation := absValue * perturbationRatio * (rand.Float64() - 0.5) * 2
		anonymizedValue = value + perturbation
	}

	// 保留合理的精度（最多2位小数）
	anonymizedValue = math.Round(anonymizedValue*100) / 100

	return anonymizedValue
}

// 辅助方法

// groupDataByDimensions 按维度组合对数据进行分组聚合
func (ap *AnonymizationProcessor) groupDataByDimensions(data []map[string]interface{}, dimensions []string, targetMetric string) map[string]float64 {
	groups := make(map[string]float64)

	for _, row := range data {
		// 构建维度组合的键
		key := ap.buildDimensionKey(row, dimensions)

		// 获取目标指标值
		value := ap.extractFloatValue(row[targetMetric])

		// 累加到对应的组
		groups[key] += value
	}

	return groups
}

// buildDimensionKey 构建维度组合的键
func (ap *AnonymizationProcessor) buildDimensionKey(row map[string]interface{}, dimensions []string) string {
	var keyParts []string
	for _, dim := range dimensions {
		value := fmt.Sprintf("%v", row[dim])
		keyParts = append(keyParts, fmt.Sprintf("%s:%s", dim, value))
	}
	return strings.Join(keyParts, "|")
}

// parseDimensionKey 解析维度键回到维度值映射
func (ap *AnonymizationProcessor) parseDimensionKey(key string, dimensions []string) map[string]interface{} {
	result := make(map[string]interface{})
	parts := strings.Split(key, "|")

	for _, part := range parts {
		if colonIndex := strings.Index(part, ":"); colonIndex > 0 {
			dimName := part[:colonIndex]
			dimValue := part[colonIndex+1:]
			result[dimName] = dimValue
		}
	}

	return result
}

// getAllUniqueKeys 获取所有唯一的维度组合键
func (ap *AnonymizationProcessor) getAllUniqueKeys(groups1, groups2 map[string]float64) []string {
	keySet := make(map[string]bool)

	for key := range groups1 {
		keySet[key] = true
	}
	for key := range groups2 {
		keySet[key] = true
	}

	var keys []string
	for key := range keySet {
		keys = append(keys, key)
	}

	return keys
}

// extractFloatValue 从interface{}中提取float64值
func (ap *AnonymizationProcessor) extractFloatValue(value interface{}) float64 {
	if value == nil {
		return 0.0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		var result float64
		if n, err := fmt.Sscanf(v, "%f", &result); err == nil && n == 1 {
			return result
		}
		return 0.0
	default:
		return 0.0
	}
}
