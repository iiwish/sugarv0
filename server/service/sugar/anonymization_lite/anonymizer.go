package anonymization_lite

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// getOrCreateAnonymizedDimension 获取或创建维度名的匿名化代号（简化版）
func (s *LiteAnonymizationService) getOrCreateAnonymizedDimension(session *LiteAnonymizationSession, dimName string, counters map[string]int) string {
	// 检查是否已经存在匿名化代号
	if anonymized, exists := session.ForwardMap[dimName]; exists {
		return anonymized
	}

	// 使用简单的DIM编号，但在描述中包含原始维度名称
	counters["dimension"]++
	anonymized := fmt.Sprintf("DIM%02d", counters["dimension"])

	// 存储映射关系
	session.ForwardMap[dimName] = anonymized
	session.ReverseMap[anonymized] = dimName

	// 存储维度语义信息，直接使用原始维度名称作为描述
	session.DimensionSemantics[anonymized] = &DimensionSemanticInfo{
		AnonymizedName: anonymized,
		OriginalName:   dimName,
		SemanticType:   "业务维度",
		Description:    dimName, // 直接使用原始维度名称
	}

	global.GVA_LOG.Debug("创建维度匿名映射",
		zap.String("original", dimName),
		zap.String("anonymized", anonymized),
		zap.String("description", dimName))

	return anonymized
}

// getOrCreateAnonymizedValue 获取或创建维度值的匿名化代号（简化版）
func (s *LiteAnonymizationService) getOrCreateAnonymizedValue(session *LiteAnonymizationSession, dimName, dimValue string, counters map[string]int) string {
	// 构建完整的键（维度名+值）
	fullKey := fmt.Sprintf("%s:%s", dimName, dimValue)

	// 检查是否已经存在匿名化代号
	if anonymized, exists := session.ForwardMap[fullKey]; exists {
		return anonymized
	}

	// 获取维度的匿名化代号
	anonymizedDim := session.ForwardMap[dimName]
	if anonymizedDim == "" {
		// 如果维度还没有匿名化，先创建维度代号
		dimensionCounters := make(map[string]int)
		anonymizedDim = s.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
	}

	// 生成简单的值代号
	dimKey := fmt.Sprintf("value_%s", dimName)
	counters[dimKey]++
	anonymized := fmt.Sprintf("%s_V%02d", anonymizedDim, counters[dimKey])

	// 存储映射关系
	session.ForwardMap[fullKey] = anonymized
	session.ReverseMap[anonymized] = dimValue

	global.GVA_LOG.Debug("创建值匿名映射",
		zap.String("dimension", dimName),
		zap.String("originalValue", dimValue),
		zap.String("anonymizedValue", anonymized))

	return anonymized
}

// DecodeAIResponse 解码AI响应中的匿名代号（简化版）
func (session *LiteAnonymizationSession) DecodeAIResponse(aiText string) (string, error) {
	if session == nil {
		return "", NewLiteAnonymizationError("会话为空", "SESSION_NULL")
	}

	if aiText == "" {
		return "", nil
	}

	global.GVA_LOG.Info("开始解码AI响应",
		zap.Int("originalLength", len(aiText)),
		zap.Int("mappingCount", len(session.ReverseMap)))

	// 获取所有需要替换的代号，按长度降序排序以避免部分替换问题
	var codes []string
	for code := range session.ReverseMap {
		codes = append(codes, code)
	}

	// 按字符串长度降序排序
	sort.Slice(codes, func(i, j int) bool {
		return len(codes[i]) > len(codes[j])
	})

	// 执行替换
	decodedText := aiText
	replacementCount := 0

	for _, code := range codes {
		originalValue := session.ReverseMap[code]
		if strings.Contains(decodedText, code) {
			decodedText = strings.ReplaceAll(decodedText, code, originalValue)
			replacementCount++
		}
	}

	global.GVA_LOG.Info("AI响应解码完成",
		zap.Int("totalCodes", len(codes)),
		zap.Int("replacementCount", replacementCount),
		zap.Int("originalLength", len(aiText)),
		zap.Int("decodedLength", len(decodedText)))

	return decodedText, nil
}

// GetAIReadyData 获取准备发送给AI的匿名化数据
func (session *LiteAnonymizationSession) GetAIReadyData() []map[string]interface{} {
	if session == nil {
		return nil
	}
	return session.AIReadyData
}

// GetMappingStats 获取映射统计信息
func (session *LiteAnonymizationSession) GetMappingStats() map[string]interface{} {
	if session == nil {
		return nil
	}

	stats := map[string]interface{}{
		"total_mappings":     len(session.ForwardMap),
		"ai_data_count":      len(session.AIReadyData),
		"contribution_count": session.ContributionCount,
		"mapping_count":      session.MappingCount,
		"user_id":            session.UserID,
		"created_at":         session.CreatedAt,
		"config":             session.Config,
	}

	// 统计维度和值的数量
	dimensionCount := 0
	valueCount := 0
	for key := range session.ForwardMap {
		if strings.Contains(key, ":") {
			valueCount++
		} else {
			dimensionCount++
		}
	}

	stats["dimension_count"] = dimensionCount
	stats["value_count"] = valueCount

	return stats
}

// ValidateSession 验证会话的完整性（简化版）
func (session *LiteAnonymizationSession) ValidateSession() error {
	if session == nil {
		return NewLiteAnonymizationError("会话为空", "SESSION_NULL")
	}

	if len(session.ForwardMap) != len(session.ReverseMap) {
		return NewLiteAnonymizationError("正向和反向映射表大小不一致", "MAPPING_INCONSISTENT")
	}

	// 验证映射的一致性
	for forward, reverse := range session.ForwardMap {
		if session.ReverseMap[reverse] != forward {
			return NewLiteAnonymizationError(fmt.Sprintf("映射不一致: %s -> %s", forward, reverse), "MAPPING_INCONSISTENT")
		}
	}

	global.GVA_LOG.Info("会话验证通过",
		zap.Int("mappingCount", len(session.ForwardMap)),
		zap.Int("contributionCount", session.ContributionCount))

	return nil
}

// SerializeToText 将匿名化数据序列化为文本格式（简化版）
func (session *LiteAnonymizationSession) SerializeToText() (string, error) {
	if session == nil || len(session.AIReadyData) == 0 {
		return "", NewLiteAnonymizationError("匿名化数据为空", "DATA_EMPTY")
	}

	var builder strings.Builder
	builder.WriteString("【简化匿名化贡献度分析数据】\n")
	builder.WriteString("说明：以下数据已进行匿名化处理，专注于贡献度分析\n\n")

	// 添加维度代号说明
	if len(session.DimensionSemantics) > 0 {
		builder.WriteString("维度代号说明：\n")
		for anonymizedName, semanticInfo := range session.DimensionSemantics {
			builder.WriteString(fmt.Sprintf("- %s-%s\n",
				anonymizedName, semanticInfo.Description))
		}
		builder.WriteString("\n")
	}

	// 添加数据列说明
	builder.WriteString("数据字段说明：\n")
	builder.WriteString("- 维度代号：表示业务维度，具体含义见上方维度说明\n")
	builder.WriteString("- 值代号：表示具体的维度值\n")
	builder.WriteString("- contribution_percent：贡献度百分比\n")
	builder.WriteString("- is_positive_driver：是否为正向驱动因子\n")
	builder.WriteString("- change_rate_percent：变化率百分比\n")
	builder.WriteString("- trend_direction：趋势方向（增长/下降/持平）\n")
	builder.WriteString("- impact_level：影响程度（高/中/低）\n")
	builder.WriteString("- relative_importance：相对重要性（0-100分）\n\n")

	builder.WriteString("数据内容：\n")
	for i, item := range session.AIReadyData {
		builder.WriteString(fmt.Sprintf("项目 %d:\n", i+1))

		// 先输出维度信息
		for key, value := range item {
			if strings.HasPrefix(key, "DIM") {
				// 添加维度名称提示
				dimensionHint := ""
				if semanticInfo, exists := session.DimensionSemantics[key]; exists {
					dimensionHint = fmt.Sprintf("(%s)", semanticInfo.Description)
				}
				builder.WriteString(fmt.Sprintf("  %s%s: %v\n", key, dimensionHint, value))
			}
		}

		// 再输出分析数据
		if cp, ok := item["contribution_percent"]; ok {
			builder.WriteString(fmt.Sprintf("  贡献度: %.2f%%\n", cp))
		}
		if ipd, ok := item["is_positive_driver"]; ok {
			builder.WriteString(fmt.Sprintf("  正向驱动: %v\n", ipd))
		}
		if crp, ok := item["change_rate_percent"]; ok {
			builder.WriteString(fmt.Sprintf("  变化率: %.2f%%\n", crp))
		}
		if td, ok := item["trend_direction"]; ok {
			builder.WriteString(fmt.Sprintf("  趋势方向: %v\n", td))
		}
		if il, ok := item["impact_level"]; ok {
			builder.WriteString(fmt.Sprintf("  影响程度: %v\n", il))
		}
		if ri, ok := item["relative_importance"]; ok {
			builder.WriteString(fmt.Sprintf("  相对重要性: %.1f分\n", ri))
		}

		builder.WriteString("\n")
	}

	global.GVA_LOG.Info("匿名化数据序列化完成",
		zap.Int("dataCount", len(session.AIReadyData)),
		zap.Int("dimensionSemantics", len(session.DimensionSemantics)),
		zap.Int("textLength", len(builder.String())))

	return builder.String(), nil
}
