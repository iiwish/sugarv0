package anonymization_lite

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// LiteAnonymizationService 简化匿名化服务
// 专注于匿名化处理，不涉及数据获取等业务逻辑
type LiteAnonymizationService struct {
	config *LiteConfig
}

// NewLiteAnonymizationService 创建新的简化匿名化服务
func NewLiteAnonymizationService(config *LiteConfig) *LiteAnonymizationService {
	if config == nil {
		config = DefaultLiteConfig()
	}
	return &LiteAnonymizationService{
		config: config,
	}
}

// ProcessContributionData 处理贡献度数据并进行匿名化
// 输入：贡献度分析结果（已计算好的数据）
// 输出：匿名化后的AI可读文本
func (s *LiteAnonymizationService) ProcessContributionData(contributions []ContributionItem) (*LiteAnonymizationSession, error) {
	global.GVA_LOG.Info("开始处理匿名化贡献度数据", zap.Int("itemCount", len(contributions)))

	if len(contributions) == 0 {
		return nil, fmt.Errorf("贡献度数据为空")
	}

	// 创建匿名化会话
	session := &LiteAnonymizationSession{
		ForwardMap:  make(map[string]string),
		ReverseMap:  make(map[string]string),
		AIReadyData: make([]map[string]interface{}, 0),
		Config:      s.config,
	}

	// 维度计数器，用于生成唯一代号
	dimensionCounters := make(map[string]int)
	valueCounters := make(map[string]int)

	// 处理每个贡献项
	for i, contribution := range contributions {
		aiItem := make(map[string]interface{})

		// 处理维度值的匿名化
		for dimName, dimValue := range contribution.DimensionValues {
			anonymizedDimName := s.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
			anonymizedDimValue := s.getOrCreateAnonymizedValue(session, dimName, fmt.Sprintf("%v", dimValue), valueCounters)

			aiItem[anonymizedDimName] = anonymizedDimValue
		}

		// 添加简化的数值数据（只有贡献度）
		aiItem["contribution_percent"] = s.anonymizeNumericValue(contribution.ContributionPercent)
		aiItem["is_positive_driver"] = contribution.IsPositiveDriver

		session.AIReadyData = append(session.AIReadyData, aiItem)

		// 记录进度
		if i%10 == 0 || i == len(contributions)-1 {
			global.GVA_LOG.Debug("匿名化进度",
				zap.Int("processed", i+1),
				zap.Int("total", len(contributions)),
				zap.Int("currentMappings", len(session.ForwardMap)))
		}
	}

	global.GVA_LOG.Info("匿名化处理完成",
		zap.Int("forwardMapSize", len(session.ForwardMap)),
		zap.Int("reverseMapSize", len(session.ReverseMap)),
		zap.Int("aiDataSize", len(session.AIReadyData)))

	return session, nil
}

// generateSemanticDimensionCode 生成语义化的维度代号
func (s *LiteAnonymizationService) generateSemanticDimensionCode(dimName string, counters map[string]int) string {
	dimNameLower := strings.ToLower(dimName)

	// 根据维度名称的语义生成有意义的代号
	var prefix string
	switch {
	case strings.Contains(dimNameLower, "地区") || strings.Contains(dimNameLower, "区域") || strings.Contains(dimNameLower, "城市"):
		prefix = "LOC" // Location
	case strings.Contains(dimNameLower, "产品") || strings.Contains(dimNameLower, "商品"):
		prefix = "PRD" // Product
	case strings.Contains(dimNameLower, "时间") || strings.Contains(dimNameLower, "日期") || strings.Contains(dimNameLower, "月") || strings.Contains(dimNameLower, "年"):
		prefix = "TIME" // Time
	case strings.Contains(dimNameLower, "部门") || strings.Contains(dimNameLower, "组织"):
		prefix = "ORG" // Organization
	case strings.Contains(dimNameLower, "客户") || strings.Contains(dimNameLower, "用户"):
		prefix = "CUST" // Customer
	default:
		prefix = "DIM" // Generic dimension
	}

	counters[prefix]++
	return fmt.Sprintf("%s%02d", prefix, counters[prefix])
}

// generateSemanticValueCode 生成语义化的值代号
func (s *LiteAnonymizationService) generateSemanticValueCode(dimCode, value string, sequence int) string {
	valueLower := strings.ToLower(value)

	// 根据值的内容生成有意义的后缀
	var suffix string
	switch {
	case strings.Contains(valueLower, "高") || strings.Contains(valueLower, "大") || strings.Contains(valueLower, "特级"):
		suffix = "HV" // High Value
	case strings.Contains(valueLower, "中") || strings.Contains(valueLower, "标准"):
		suffix = "ST" // Standard
	case strings.Contains(valueLower, "低") || strings.Contains(valueLower, "小") || strings.Contains(valueLower, "基础"):
		suffix = "BS" // Basic
	case strings.Contains(valueLower, "一线") || strings.Contains(valueLower, "核心"):
		suffix = "T1" // Tier 1
	case strings.Contains(valueLower, "二线"):
		suffix = "T2" // Tier 2
	case strings.Contains(valueLower, "三线"):
		suffix = "T3" // Tier 3
	default:
		suffix = "V" // Generic value
	}

	return fmt.Sprintf("%s_%s%02d", dimCode, suffix, sequence)
}

// anonymizeNumericValue 对数值进行轻微匿名化处理
func (s *LiteAnonymizationService) anonymizeNumericValue(value float64) float64 {
	// 使用配置的噪声级别
	maxPerturbation := s.config.NoiseLevel * 100 // 转换为百分比
	perturbation := (rand.Float64() - 0.5) * 2 * maxPerturbation
	anonymizedValue := value + perturbation

	// 确保百分比在合理范围内
	if anonymizedValue > 100.0 {
		anonymizedValue = 100.0
	} else if anonymizedValue < -100.0 {
		anonymizedValue = -100.0
	}

	// 保留2位小数
	return math.Round(anonymizedValue*100) / 100
}

// ProcessAndSerialize 处理贡献度数据并序列化为AI可读文本
func (s *LiteAnonymizationService) ProcessAndSerialize(contributions []ContributionItem) (string, *LiteAnonymizationSession, error) {
	session, err := s.ProcessContributionData(contributions)
	if err != nil {
		return "", nil, err
	}

	text, err := session.SerializeToText()
	if err != nil {
		return "", nil, err
	}

	return text, session, nil
}

// DecodeResponse 解码AI响应（使用会话方法）
func (s *LiteAnonymizationService) DecodeResponse(session *LiteAnonymizationSession, aiResponse string) (string, error) {
	return session.DecodeAIResponse(aiResponse)
}
