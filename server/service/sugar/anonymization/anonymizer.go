package anonymization

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// createAnonymizedSession 创建匿名化会话
func (s *AnonymizationService) createAnonymizedSession(contributions []ContributionItem, req *AIAnalysisRequest) (*AnonymizationSession, error) {
	session := &AnonymizationSession{
		forwardMap:  make(map[string]string),
		reverseMap:  make(map[string]string),
		AIReadyData: make([]map[string]interface{}, 0),
	}

	// 维度计数器，用于生成唯一代号
	dimensionCounters := make(map[string]int)
	valueCounters := make(map[string]int)

	// 处理每个贡献项
	for _, contribution := range contributions {
		aiItem := make(map[string]interface{})

		// 处理维度值的匿名化
		for dimName, dimValue := range contribution.DimensionValues {
			anonymizedDimName := s.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
			anonymizedDimValue := s.getOrCreateAnonymizedValue(session, dimName, fmt.Sprintf("%v", dimValue), valueCounters)

			aiItem[anonymizedDimName] = anonymizedDimValue
		}

		// 添加经过脱敏处理的数值数据
		aiItem["contribution_percent"] = s.anonymizeNumericValue(contribution.ContributionPercent, "contribution")
		aiItem["change_value"] = s.anonymizeNumericValue(contribution.ChangeValue, "change")
		aiItem["current_value"] = s.anonymizeNumericValue(contribution.CurrentValue, "current")
		aiItem["base_value"] = s.anonymizeNumericValue(contribution.BaseValue, "base")
		aiItem["is_positive_driver"] = contribution.IsPositiveDriver

		session.AIReadyData = append(session.AIReadyData, aiItem)
	}

	global.GVA_LOG.Info("匿名化会话创建完成",
		zap.Int("forwardMapSize", len(session.forwardMap)),
		zap.Int("aiDataSize", len(session.AIReadyData)))

	return session, nil
}

// getOrCreateAnonymizedDimension 获取或创建维度名的匿名化代号
func (s *AnonymizationService) getOrCreateAnonymizedDimension(session *AnonymizationSession, dimName string, counters map[string]int) string {
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
func (s *AnonymizationService) getOrCreateAnonymizedValue(session *AnonymizationSession, dimName, dimValue string, counters map[string]int) string {
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
		anonymizedDim = s.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
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

// DecodeAIResponse 解码AI响应中的匿名代号
func (session *AnonymizationSession) DecodeAIResponse(aiText string) (string, error) {
	if session == nil {
		return "", NewAnonymizationError("会话为空", nil)
	}

	if aiText == "" {
		return "", nil
	}

	// 获取所有需要替换的代号，按长度降序排序以避免部分替换问题
	var codes []string
	for code := range session.reverseMap {
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
		originalValue := session.reverseMap[code]
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

// DecodeAIResponse 解码高级匿名化会话中的AI响应
func (session *AdvancedAnonymizationSession) DecodeAIResponse(aiText string) (string, error) {
	if session == nil || session.AnonymizationSession == nil {
		return "", NewAnonymizationError("高级会话为空", nil)
	}

	// 使用基础会话的解码功能
	decodedText, err := session.AnonymizationSession.DecodeAIResponse(aiText)
	if err != nil {
		return "", err
	}

	global.GVA_LOG.Info("高级AI响应解码完成",
		zap.Float64("privacyScore", session.privacyMetrics.PrivacyScore),
		zap.Float64("dataUtility", session.qualityMetrics.DataUtility),
		zap.Int("originalLength", len(aiText)),
		zap.Int("decodedLength", len(decodedText)))

	return decodedText, nil
}

// GetAIReadyData 获取高级匿名化的AI就绪数据
func (session *AdvancedAnonymizationSession) GetAIReadyData() []map[string]interface{} {
	if session == nil || session.AnonymizationSession == nil {
		return nil
	}
	return session.AnonymizationSession.GetAIReadyData()
}

// GetPrivacyMetrics 获取隐私指标
func (session *AdvancedAnonymizationSession) GetPrivacyMetrics() *PrivacyMetrics {
	if session == nil {
		return nil
	}
	return session.privacyMetrics
}

// GetQualityMetrics 获取数据质量指标
func (session *AdvancedAnonymizationSession) GetQualityMetrics() *DataQualityMetrics {
	if session == nil {
		return nil
	}
	return session.qualityMetrics
}

// GetAdvancedMappingStats 获取高级映射统计信息
func (session *AdvancedAnonymizationSession) GetAdvancedMappingStats() map[string]interface{} {
	if session == nil || session.AnonymizationSession == nil {
		return nil
	}

	// 获取基础统计信息
	baseStats := session.AnonymizationSession.GetMappingStats()

	// 添加高级统计信息
	advancedStats := map[string]interface{}{
		"base_stats":          baseStats,
		"privacy_metrics":     session.privacyMetrics,
		"quality_metrics":     session.qualityMetrics,
		"config":              session.config,
		"equivalence_classes": len(session.equivalenceClasses),
		"anonymity_level":     session.anonymityLevel,
		"epsilon_used":        session.epsilonBudgetUsed,
		"semantic_mappings":   len(session.semanticMappings),
	}

	return advancedStats
}

// ValidateAdvancedSession 验证高级会话的完整性
func (session *AdvancedAnonymizationSession) ValidateAdvancedSession() error {
	if session == nil {
		return NewAnonymizationError("高级会话为空", nil)
	}

	// 验证基础会话
	if err := session.AnonymizationSession.ValidateSession(); err != nil {
		return err
	}

	// 验证隐私指标
	if session.privacyMetrics == nil {
		return NewAnonymizationError("隐私指标缺失", nil)
	}

	// 验证质量指标
	if session.qualityMetrics == nil {
		return NewAnonymizationError("质量指标缺失", nil)
	}

	// 验证配置
	if session.config == nil {
		return NewAnonymizationError("配置缺失", nil)
	}

	// 验证K-匿名性要求
	if session.anonymityLevel < session.config.KAnonymity {
		return NewAnonymizationError(fmt.Sprintf("K-匿名性不满足要求，当前: %d，要求: %d",
			session.anonymityLevel, session.config.KAnonymity), nil)
	}

	// 验证隐私预算
	if session.epsilonBudgetUsed > session.config.Epsilon {
		return NewAnonymizationError(fmt.Sprintf("隐私预算超支，使用: %.3f，限制: %.3f",
			session.epsilonBudgetUsed, session.config.Epsilon), nil)
	}

	global.GVA_LOG.Info("高级会话验证通过",
		zap.Int("anonymityLevel", session.anonymityLevel),
		zap.Float64("epsilonUsed", session.epsilonBudgetUsed),
		zap.Float64("privacyScore", session.privacyMetrics.PrivacyScore))

	return nil
}

// GetAIReadyData 获取准备发送给AI的匿名化数据
func (session *AnonymizationSession) GetAIReadyData() []map[string]interface{} {
	if session == nil {
		return nil
	}
	return session.AIReadyData
}

// GetMappingStats 获取映射统计信息（用于调试和监控）
func (session *AnonymizationSession) GetMappingStats() map[string]interface{} {
	if session == nil {
		return nil
	}

	stats := map[string]interface{}{
		"total_mappings":  len(session.forwardMap),
		"ai_data_count":   len(session.AIReadyData),
		"dimension_count": 0,
		"value_count":     0,
	}

	// 统计维度和值的数量
	for key := range session.forwardMap {
		if strings.Contains(key, ":") {
			stats["value_count"] = stats["value_count"].(int) + 1
		} else {
			stats["dimension_count"] = stats["dimension_count"].(int) + 1
		}
	}

	return stats
}

// ValidateSession 验证会话的完整性
func (session *AnonymizationSession) ValidateSession() error {
	if session == nil {
		return NewAnonymizationError("会话为空", nil)
	}

	if len(session.forwardMap) != len(session.reverseMap) {
		return NewAnonymizationError("正向和反向映射表大小不一致", nil)
	}

	// 验证映射的一致性
	for forward, reverse := range session.forwardMap {
		if session.reverseMap[reverse] != forward {
			return NewAnonymizationError(fmt.Sprintf("映射不一致: %s -> %s", forward, reverse), nil)
		}
	}

	return nil
}

// ThreadSafeSession 线程安全的会话包装器
type ThreadSafeSession struct {
	session *AnonymizationSession
	mutex   sync.RWMutex
}

// NewThreadSafeSession 创建线程安全的会话
func NewThreadSafeSession(session *AnonymizationSession) *ThreadSafeSession {
	return &ThreadSafeSession{
		session: session,
	}
}

// DecodeAIResponse 线程安全的解码方法
func (ts *ThreadSafeSession) DecodeAIResponse(aiText string) (string, error) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.session.DecodeAIResponse(aiText)
}

// GetAIReadyData 线程安全的获取AI数据方法
func (ts *ThreadSafeSession) GetAIReadyData() []map[string]interface{} {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.session.GetAIReadyData()
}

// init 包初始化，设置随机种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// anonymizeNumericValue 对数值进行基础脱敏处理
func (s *AnonymizationService) anonymizeNumericValue(value float64, valueType string) float64 {
	// 基础脱敏策略：
	// 1. 对于小数值（绝对值 < 1000），保留相对精度但添加小幅扰动
	// 2. 对于大数值，使用数量级保持和舍入策略
	// 3. 对于百分比类型，确保保持在合理范围内

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

	// 安全处理除零错误
	var perturbationPercent float64
	if value != 0 {
		perturbationPercent = math.Abs((anonymizedValue-value)/value) * 100
	}

	global.GVA_LOG.Debug("数值脱敏处理",
		zap.String("valueType", valueType),
		zap.Float64("originalValue", value),
		zap.Float64("anonymizedValue", anonymizedValue),
		zap.Float64("perturbationPercent", perturbationPercent))

	return anonymizedValue
}
