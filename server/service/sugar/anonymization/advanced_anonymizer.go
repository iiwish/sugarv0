package anonymization

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// AdvancedAnonymizer 高级匿名化器
type AdvancedAnonymizer struct {
	config         *AdvancedAnonymizationConfig
	analyzer       *DataAnalyzer
	noiseGenerator *NoiseGenerator
	semanticMapper *SemanticMapper
	mutex          sync.RWMutex
}

// NewAdvancedAnonymizer 创建高级匿名化器
func NewAdvancedAnonymizer(config *AdvancedAnonymizationConfig) *AdvancedAnonymizer {
	if config == nil {
		config = DefaultAdvancedConfig()
	}

	return &AdvancedAnonymizer{
		config:         config,
		analyzer:       NewDataAnalyzer(config),
		noiseGenerator: NewNoiseGenerator(config),
		semanticMapper: NewSemanticMapper(config),
	}
}

// AdvancedAnonymizationSession 高级匿名化会话
type AdvancedAnonymizationSession struct {
	*AnonymizationSession

	// 高级特性
	config           *AdvancedAnonymizationConfig
	characteristics  *DataCharacteristics
	privacyMetrics   *PrivacyMetrics
	qualityMetrics   *DataQualityMetrics
	semanticMappings map[string]*SemanticMapping

	// 差分隐私状态
	epsilonBudgetUsed float64
	noiseParameters   map[string]float64

	// K-匿名性相关
	equivalenceClasses map[string][]int
	anonymityLevel     int
}

// CreateAdvancedAnonymizationSession 创建高级匿名化会话
func (aa *AdvancedAnonymizer) CreateAdvancedAnonymizationSession(contributions []ContributionItem, req *AIAnalysisRequest) (*AdvancedAnonymizationSession, error) {
	aa.mutex.Lock()
	defer aa.mutex.Unlock()

	// 1. 数据特征分析
	characteristics := aa.analyzer.AnalyzeDataCharacteristics(contributions)

	global.GVA_LOG.Info("开始高级匿名化处理",
		zap.Int("recordCount", characteristics.TotalRecords),
		zap.Int("dimensionCount", characteristics.DimensionCount),
		zap.Float64("epsilon", aa.config.Epsilon))

	// 2. 创建基础会话并追踪隐私预算
	baseSession, epsilonUsed, err := aa.createBaseSessionWithBudget(contributions, req)
	if err != nil {
		return nil, NewAnonymizationError("创建基础会话失败", err)
	}

	// 3. 创建高级会话
	advancedSession := &AdvancedAnonymizationSession{
		AnonymizationSession: baseSession,
		config:               aa.config,
		characteristics:      characteristics,
		semanticMappings:     make(map[string]*SemanticMapping),
		noiseParameters:      make(map[string]float64),
		equivalenceClasses:   make(map[string][]int),
		epsilonBudgetUsed:    epsilonUsed, // 正确设置隐私预算使用量
	}

	// 4. 应用高级匿名化算法
	if err := aa.applyAdvancedAnonymization(advancedSession, contributions, req); err != nil {
		return nil, NewAnonymizationError("应用高级匿名化失败", err)
	}

	// 5. 计算隐私和质量指标
	advancedSession.privacyMetrics = aa.calculatePrivacyMetrics(advancedSession)
	advancedSession.qualityMetrics = aa.calculateQualityMetrics(advancedSession, contributions)

	// 6. 验证隐私预算未超支
	if advancedSession.epsilonBudgetUsed > aa.config.Epsilon {
		global.GVA_LOG.Warn("隐私预算超支",
			zap.Float64("used", advancedSession.epsilonBudgetUsed),
			zap.Float64("budget", aa.config.Epsilon))
		return nil, NewAnonymizationError(fmt.Sprintf("隐私预算超支: 使用%.3f, 预算%.3f",
			advancedSession.epsilonBudgetUsed, aa.config.Epsilon), nil)
	}

	global.GVA_LOG.Info("高级匿名化会话创建完成",
		zap.Float64("epsilonUsed", advancedSession.epsilonBudgetUsed),
		zap.Float64("epsilonBudget", aa.config.Epsilon),
		zap.Float64("epsilonRemaining", aa.config.Epsilon-advancedSession.epsilonBudgetUsed),
		zap.Int("anonymityLevel", advancedSession.anonymityLevel),
		zap.Float64("dataUtility", advancedSession.qualityMetrics.DataUtility))

	return advancedSession, nil
}

// createBaseSessionWithBudget 创建基础匿名化会话并返回隐私预算使用量
func (aa *AdvancedAnonymizer) createBaseSessionWithBudget(contributions []ContributionItem, req *AIAnalysisRequest) (*AnonymizationSession, float64, error) {
	session := &AnonymizationSession{
		forwardMap:  make(map[string]string),
		reverseMap:  make(map[string]string),
		AIReadyData: make([]map[string]interface{}, 0),
	}

	// 维度和值计数器
	dimensionCounters := make(map[string]int)
	valueCounters := make(map[string]int)

	// 追踪隐私预算使用
	var totalEpsilonUsed float64

	// 处理每个贡献项
	for _, contribution := range contributions {
		aiItem := make(map[string]interface{})

		// 处理维度值的匿名化
		for dimName, dimValue := range contribution.DimensionValues {
			var anonymizedDimName, anonymizedDimValue string

			if aa.config.UseSemanticMapping {
				// 使用语义映射
				anonymizedDimName = aa.semanticMapper.GetSemanticDimensionName(session, dimName, dimensionCounters)
				anonymizedDimValue = aa.semanticMapper.GetSemanticDimensionValue(session, dimName, fmt.Sprintf("%v", dimValue), valueCounters)
			} else {
				// 使用传统映射
				anonymizedDimName = aa.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
				anonymizedDimValue = aa.getOrCreateAnonymizedValue(session, dimName, fmt.Sprintf("%v", dimValue), valueCounters)
			}

			aiItem[anonymizedDimName] = anonymizedDimValue
		}

		// 添加经过差分隐私处理的数值数据
		contribWithNoise, epsilonUsed1 := aa.addDifferentialPrivacyNoiseWithBudget(contribution.ContributionPercent, "contribution")
		changeWithNoise, epsilonUsed2 := aa.addDifferentialPrivacyNoiseWithBudget(contribution.ChangeValue, "change")
		currentWithNoise, epsilonUsed3 := aa.addDifferentialPrivacyNoiseWithBudget(contribution.CurrentValue, "current")
		baseWithNoise, epsilonUsed4 := aa.addDifferentialPrivacyNoiseWithBudget(contribution.BaseValue, "base")

		aiItem["contribution_percent"] = contribWithNoise
		aiItem["change_value"] = changeWithNoise
		aiItem["current_value"] = currentWithNoise
		aiItem["base_value"] = baseWithNoise
		aiItem["is_positive_driver"] = contribution.IsPositiveDriver

		// 累计隐私预算使用
		totalEpsilonUsed += epsilonUsed1 + epsilonUsed2 + epsilonUsed3 + epsilonUsed4

		session.AIReadyData = append(session.AIReadyData, aiItem)
	}

	global.GVA_LOG.Info("基础会话创建完成",
		zap.Float64("totalEpsilonUsed", totalEpsilonUsed),
		zap.Float64("epsilonBudget", aa.config.Epsilon),
		zap.Float64("budgetRemaining", aa.config.Epsilon-totalEpsilonUsed))

	return session, totalEpsilonUsed, nil
}

// createBaseSession 创建基础匿名化会话
func (aa *AdvancedAnonymizer) createBaseSession(contributions []ContributionItem, req *AIAnalysisRequest) (*AnonymizationSession, error) {
	session := &AnonymizationSession{
		forwardMap:  make(map[string]string),
		reverseMap:  make(map[string]string),
		AIReadyData: make([]map[string]interface{}, 0),
	}

	// 维度和值计数器
	dimensionCounters := make(map[string]int)
	valueCounters := make(map[string]int)

	// 追踪隐私预算使用
	var totalEpsilonUsed float64

	// 处理每个贡献项
	for _, contribution := range contributions {
		aiItem := make(map[string]interface{})

		// 处理维度值的匿名化
		for dimName, dimValue := range contribution.DimensionValues {
			var anonymizedDimName, anonymizedDimValue string

			if aa.config.UseSemanticMapping {
				// 使用语义映射
				anonymizedDimName = aa.semanticMapper.GetSemanticDimensionName(session, dimName, dimensionCounters)
				anonymizedDimValue = aa.semanticMapper.GetSemanticDimensionValue(session, dimName, fmt.Sprintf("%v", dimValue), valueCounters)
			} else {
				// 使用传统映射
				anonymizedDimName = aa.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
				anonymizedDimValue = aa.getOrCreateAnonymizedValue(session, dimName, fmt.Sprintf("%v", dimValue), valueCounters)
			}

			aiItem[anonymizedDimName] = anonymizedDimValue
		}

		// 添加经过差分隐私处理的数值数据
		contribWithNoise, epsilonUsed1 := aa.addDifferentialPrivacyNoiseWithBudget(contribution.ContributionPercent, "contribution")
		changeWithNoise, epsilonUsed2 := aa.addDifferentialPrivacyNoiseWithBudget(contribution.ChangeValue, "change")
		currentWithNoise, epsilonUsed3 := aa.addDifferentialPrivacyNoiseWithBudget(contribution.CurrentValue, "current")
		baseWithNoise, epsilonUsed4 := aa.addDifferentialPrivacyNoiseWithBudget(contribution.BaseValue, "base")

		aiItem["contribution_percent"] = contribWithNoise
		aiItem["change_value"] = changeWithNoise
		aiItem["current_value"] = currentWithNoise
		aiItem["base_value"] = baseWithNoise
		aiItem["is_positive_driver"] = contribution.IsPositiveDriver

		// 累计隐私预算使用
		totalEpsilonUsed += epsilonUsed1 + epsilonUsed2 + epsilonUsed3 + epsilonUsed4

		session.AIReadyData = append(session.AIReadyData, aiItem)
	}

	global.GVA_LOG.Info("基础会话创建完成",
		zap.Float64("totalEpsilonUsed", totalEpsilonUsed),
		zap.Float64("epsilonBudget", aa.config.Epsilon),
		zap.Float64("budgetRemaining", aa.config.Epsilon-totalEpsilonUsed))

	return session, nil
}

// addDifferentialPrivacyNoiseWithBudget 添加差分隐私噪声并返回预算使用量
func (aa *AdvancedAnonymizer) addDifferentialPrivacyNoiseWithBudget(value float64, metricType string) (float64, float64) {
	if aa.config.Epsilon <= 0 {
		return value, 0.0
	}

	// 计算每次查询消耗的隐私预算
	budgetPerQuery := aa.config.Epsilon / 4.0 // 四个数值字段平均分配预算

	// 使用拉普拉斯机制添加噪声
	sensitivity := aa.config.GlobalSensitivity
	scale := sensitivity / budgetPerQuery

	noise := aa.noiseGenerator.GenerateLaplaceNoise(scale)
	noisyValue := value + noise

	// 记录噪声参数
	aa.noiseGenerator.RecordNoiseParameters(metricType, noise, scale)

	return noisyValue, budgetPerQuery
}

// applyAdvancedAnonymization 应用高级匿名化算法
func (aa *AdvancedAnonymizer) applyAdvancedAnonymization(session *AdvancedAnonymizationSession, contributions []ContributionItem, req *AIAnalysisRequest) error {
	// 1. 应用K-匿名性
	if err := aa.applyKAnonymity(session, contributions); err != nil {
		return err
	}

	// 2. 应用L-多样性
	if err := aa.applyLDiversity(session, contributions); err != nil {
		return err
	}

	// 3. 保留数据特征
	if err := aa.preserveDataCharacteristics(session, contributions); err != nil {
		return err
	}

	// 4. 应用自适应噪声
	if aa.config.AdaptiveNoise {
		if err := aa.applyAdaptiveNoise(session, contributions); err != nil {
			return err
		}
	}

	return nil
}

// applyKAnonymity 应用K-匿名性
func (aa *AdvancedAnonymizer) applyKAnonymity(session *AdvancedAnonymizationSession, contributions []ContributionItem) error {
	// 构建等价类
	equivalenceClasses := make(map[string][]int)

	for i, contrib := range contributions {
		// 构建准标识符组合
		quasiId := aa.buildQuasiIdentifier(contrib.DimensionValues)
		equivalenceClasses[quasiId] = append(equivalenceClasses[quasiId], i)
	}

	// 检查K-匿名性
	minClassSize := math.MaxInt32
	for quasiId, indices := range equivalenceClasses {
		if len(indices) < aa.config.KAnonymity {
			// 需要进行泛化或抑制
			if err := aa.generalizeEquivalenceClass(session, quasiId, indices, contributions); err != nil {
				return err
			}
		}
		if len(indices) < minClassSize {
			minClassSize = len(indices)
		}
	}

	session.equivalenceClasses = equivalenceClasses
	session.anonymityLevel = minClassSize

	global.GVA_LOG.Info("K-匿名性处理完成",
		zap.Int("kValue", aa.config.KAnonymity),
		zap.Int("achievedLevel", minClassSize),
		zap.Int("equivalenceClasses", len(equivalenceClasses)))

	return nil
}

// applyLDiversity 应用L-多样性
func (aa *AdvancedAnonymizer) applyLDiversity(session *AdvancedAnonymizationSession, contributions []ContributionItem) error {
	// 对每个等价类检查L-多样性
	for quasiId, indices := range session.equivalenceClasses {
		sensitiveValues := make(map[string]bool)

		// 收集敏感属性值（这里以贡献度符号作为敏感属性）
		for _, idx := range indices {
			if idx < len(contributions) {
				contrib := contributions[idx]
				sensitiveValue := "positive"
				if !contrib.IsPositiveDriver {
					sensitiveValue = "negative"
				}
				sensitiveValues[sensitiveValue] = true
			}
		}

		// 检查是否满足L-多样性
		if len(sensitiveValues) < aa.config.LDiversity {
			// 需要增加多样性或进行进一步泛化
			if err := aa.enhanceDiversity(session, quasiId, indices, contributions); err != nil {
				return err
			}
		}
	}

	global.GVA_LOG.Info("L-多样性处理完成", zap.Int("lValue", aa.config.LDiversity))
	return nil
}

// preserveDataCharacteristics 保留数据特征
func (aa *AdvancedAnonymizer) preserveDataCharacteristics(session *AdvancedAnonymizationSession, contributions []ContributionItem) error {
	if !aa.config.PreserveTrends && !aa.config.PreserveCorr {
		return nil
	}

	// 计算原始数据的统计特征
	originalFeatures := session.characteristics.StatisticalFeatures

	// 调整匿名化后的数据以保留特征
	for i, aiItem := range session.AIReadyData {
		if aa.config.PreserveTrends {
			// 保留趋势特征
			if contribPercent, ok := aiItem["contribution_percent"].(float64); ok {
				// 根据原始分布调整值
				adjustedValue := aa.adjustValueForTrendPreservation(contribPercent, originalFeatures)
				aiItem["contribution_percent"] = adjustedValue
			}
		}

		if aa.config.PreserveCorr && i < len(contributions) {
			// 保留相关性特征
			if err := aa.adjustForCorrelationPreservation(aiItem, contributions[i], originalFeatures); err != nil {
				return err
			}
		}

		session.AIReadyData[i] = aiItem
	}

	global.GVA_LOG.Info("数据特征保留处理完成",
		zap.Bool("preserveTrends", aa.config.PreserveTrends),
		zap.Bool("preserveCorr", aa.config.PreserveCorr))

	return nil
}

// addDifferentialPrivacyNoise 添加差分隐私噪声
func (aa *AdvancedAnonymizer) addDifferentialPrivacyNoise(value float64, metricType string) float64 {
	if aa.config.Epsilon <= 0 {
		return value
	}

	// 计算每次查询消耗的隐私预算（简化处理，每个字段消耗相等预算）
	budgetPerQuery := aa.config.Epsilon / 4.0 // 四个数值字段平均分配预算

	// 使用拉普拉斯机制添加噪声
	sensitivity := aa.config.GlobalSensitivity
	scale := sensitivity / budgetPerQuery

	noise := aa.noiseGenerator.GenerateLaplaceNoise(scale)
	noisyValue := value + noise

	// 记录噪声参数
	aa.noiseGenerator.RecordNoiseParameters(metricType, noise, scale)

	global.GVA_LOG.Debug("差分隐私噪声添加",
		zap.String("metricType", metricType),
		zap.Float64("originalValue", value),
		zap.Float64("noise", noise),
		zap.Float64("noisyValue", noisyValue),
		zap.Float64("budgetUsed", budgetPerQuery))

	return noisyValue
}

// Helper methods

func (aa *AdvancedAnonymizer) buildQuasiIdentifier(dimensionValues map[string]interface{}) string {
	var parts []string
	for dim, value := range dimensionValues {
		parts = append(parts, fmt.Sprintf("%s:%v", dim, value))
	}
	sort.Strings(parts)
	return strings.Join(parts, "|")
}

func (aa *AdvancedAnonymizer) generalizeEquivalenceClass(session *AdvancedAnonymizationSession, quasiId string, indices []int, contributions []ContributionItem) error {
	// 简化实现：对小于K的等价类进行抑制
	// 在实际应用中，可以实现更复杂的泛化策略
	global.GVA_LOG.Warn("等价类大小不足，进行抑制处理",
		zap.String("quasiId", quasiId),
		zap.Int("size", len(indices)),
		zap.Int("required", aa.config.KAnonymity))

	// 标记这些记录为抑制状态
	for _, idx := range indices {
		if idx < len(session.AIReadyData) {
			session.AIReadyData[idx]["_suppressed"] = true
		}
	}

	return nil
}

func (aa *AdvancedAnonymizer) enhanceDiversity(session *AdvancedAnonymizationSession, quasiId string, indices []int, contributions []ContributionItem) error {
	// 简化实现：通过添加噪声来增加敏感属性的多样性
	global.GVA_LOG.Warn("等价类多样性不足，进行增强处理",
		zap.String("quasiId", quasiId),
		zap.Int("indices", len(indices)))

	// 为这些记录添加额外的多样性噪声
	for _, idx := range indices {
		if idx < len(session.AIReadyData) {
			if contribPercent, ok := session.AIReadyData[idx]["contribution_percent"].(float64); ok {
				diversityNoise := aa.noiseGenerator.GenerateGaussianNoise(0, aa.config.NoiseVariance)
				session.AIReadyData[idx]["contribution_percent"] = contribPercent + diversityNoise
			}
		}
	}

	return nil
}

func (aa *AdvancedAnonymizer) adjustValueForTrendPreservation(value float64, features *StatisticalFeatures) float64 {
	// 保持值在合理范围内，同时保留相对趋势
	if features.ContributionStdDev > 0 {
		// 标准化后再反标准化，保留分布特征
		normalized := (value - features.ContributionMean) / features.ContributionStdDev
		// 添加小量噪声但保持趋势
		trendNoise := aa.noiseGenerator.GenerateGaussianNoise(0, 0.01)
		return features.ContributionMean + (normalized+trendNoise)*features.ContributionStdDev
	}
	return value
}

func (aa *AdvancedAnonymizer) adjustForCorrelationPreservation(aiItem map[string]interface{}, contribution ContributionItem, features *StatisticalFeatures) error {
	// 保留贡献度与变化值之间的相关性
	if contribPercent, ok := aiItem["contribution_percent"].(float64); ok {
		if changeValue, ok := aiItem["change_value"].(float64); ok {
			// 更智能的相关性保护逻辑
			originalContrib := contribution.ContributionPercent
			originalChange := contribution.ChangeValue

			// 计算原始相关性强度
			originalCorrelation := originalContrib * originalChange
			currentCorrelation := contribPercent * changeValue

			// 如果相关性方向发生改变且改变程度较大，进行调整
			correlationThreshold := 0.1 // 相关性变化的容忍阈值
			if (originalCorrelation > 0 && currentCorrelation < -correlationThreshold) ||
				(originalCorrelation < 0 && currentCorrelation > correlationThreshold) {

				// 调整变化值以保持相关性方向
				if originalCorrelation >= 0 {
					// 原本正相关，确保调整后也是正相关
					if contribPercent > 0 {
						aiItem["change_value"] = math.Abs(changeValue) * 0.8 // 稍微减弱以避免过度调整
					} else {
						aiItem["change_value"] = -math.Abs(changeValue) * 0.8
					}
				} else {
					// 原本负相关，确保调整后也是负相关
					if contribPercent > 0 {
						aiItem["change_value"] = -math.Abs(changeValue) * 0.8
					} else {
						aiItem["change_value"] = math.Abs(changeValue) * 0.8
					}
				}

				global.GVA_LOG.Debug("相关性调整",
					zap.Float64("originalContrib", originalContrib),
					zap.Float64("originalChange", originalChange),
					zap.Float64("adjustedContrib", contribPercent),
					zap.Float64("adjustedChange", aiItem["change_value"].(float64)))
			}
		}
	}
	return nil
}

func (aa *AdvancedAnonymizer) applyAdaptiveNoise(session *AdvancedAnonymizationSession, contributions []ContributionItem) error {
	// 根据数据重要性自适应调整噪声强度
	for i, aiItem := range session.AIReadyData {
		if i < len(contributions) {
			contrib := contributions[i]

			// 计算重要性分数（基于贡献度的绝对值）
			importanceScore := math.Abs(contrib.ContributionPercent) / 100.0

			// 重要的数据添加较少噪声，不重要的数据添加较多噪声
			// 使用反比例关系：重要性越高，噪声越小
			noiseReductionFactor := 1.0 - (importanceScore * 0.5) // 最多减少50%的噪声
			if noiseReductionFactor < 0.1 {                       // 确保至少保留10%的噪声
				noiseReductionFactor = 0.1
			}

			adaptiveNoiseScale := aa.config.NoiseVariance * noiseReductionFactor
			adaptiveNoise := aa.noiseGenerator.GenerateGaussianNoise(0, adaptiveNoiseScale)

			if contribPercent, ok := aiItem["contribution_percent"].(float64); ok {
				originalValue := contribPercent
				aiItem["contribution_percent"] = contribPercent + adaptiveNoise

				global.GVA_LOG.Debug("自适应噪声应用",
					zap.Int("index", i),
					zap.Float64("importanceScore", importanceScore),
					zap.Float64("noiseReductionFactor", noiseReductionFactor),
					zap.Float64("originalValue", originalValue),
					zap.Float64("noise", adaptiveNoise),
					zap.Float64("finalValue", aiItem["contribution_percent"].(float64)))
			}
		}
	}

	global.GVA_LOG.Info("自适应噪声处理完成")
	return nil
}

func (aa *AdvancedAnonymizer) getOrCreateAnonymizedDimension(session *AnonymizationSession, dimName string, counters map[string]int) string {
	if anonymized, exists := session.forwardMap[dimName]; exists {
		return anonymized
	}

	counters["dimension"]++
	anonymized := fmt.Sprintf("D%02d", counters["dimension"])

	session.forwardMap[dimName] = anonymized
	session.reverseMap[anonymized] = dimName

	return anonymized
}

func (aa *AdvancedAnonymizer) getOrCreateAnonymizedValue(session *AnonymizationSession, dimName, dimValue string, counters map[string]int) string {
	fullKey := fmt.Sprintf("%s:%s", dimName, dimValue)

	if anonymized, exists := session.forwardMap[fullKey]; exists {
		return anonymized
	}

	anonymizedDim := session.forwardMap[dimName]
	if anonymizedDim == "" {
		dimensionCounters := make(map[string]int)
		anonymizedDim = aa.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
	}

	dimKey := fmt.Sprintf("value_%s", dimName)
	counters[dimKey]++
	anonymized := fmt.Sprintf("%s_V%02d", anonymizedDim, counters[dimKey])

	session.forwardMap[fullKey] = anonymized
	session.reverseMap[anonymized] = dimValue

	return anonymized
}

// calculatePrivacyMetrics 计算隐私指标
func (aa *AdvancedAnonymizer) calculatePrivacyMetrics(session *AdvancedAnonymizationSession) *PrivacyMetrics {
	return &PrivacyMetrics{
		EpsilonUsed:     session.epsilonBudgetUsed,
		KAnonymityLevel: session.anonymityLevel,
		LDiversityLevel: aa.config.LDiversity,
		NoiseVariance:   aa.config.NoiseVariance,
		PrivacyScore:    aa.calculatePrivacyScore(session),
	}
}

// calculateQualityMetrics 计算数据质量指标
func (aa *AdvancedAnonymizer) calculateQualityMetrics(session *AdvancedAnonymizationSession, original []ContributionItem) *DataQualityMetrics {
	// 简化的数据质量计算
	dataUtility := 1.0 - (session.epsilonBudgetUsed/aa.config.Epsilon)*0.3

	return &DataQualityMetrics{
		DataUtility:       dataUtility,
		InformationLoss:   1.0 - dataUtility,
		StatisticalError:  aa.calculateStatisticalError(session, original),
		TrendPreservation: aa.calculateTrendPreservation(session, original),
	}
}

func (aa *AdvancedAnonymizer) calculatePrivacyScore(session *AdvancedAnonymizationSession) float64 {
	// 综合隐私分数计算
	epsilonScore := (1.0 - aa.config.Epsilon) * 0.4                     // ε越小隐私越好
	kScore := math.Min(float64(session.anonymityLevel)/10.0, 1.0) * 0.3 // K值贡献
	lScore := math.Min(float64(aa.config.LDiversity)/5.0, 1.0) * 0.3    // L值贡献

	return epsilonScore + kScore + lScore
}

func (aa *AdvancedAnonymizer) calculateStatisticalError(session *AdvancedAnonymizationSession, original []ContributionItem) float64 {
	// 计算统计误差
	if len(original) == 0 || len(session.AIReadyData) == 0 {
		return 0.0
	}

	var totalError float64
	count := 0

	for i, aiItem := range session.AIReadyData {
		if i < len(original) {
			if anonValue, ok := aiItem["contribution_percent"].(float64); ok {
				originalValue := original[i].ContributionPercent
				error := math.Abs(anonValue - originalValue)
				totalError += error
				count++
			}
		}
	}

	if count > 0 {
		return totalError / float64(count)
	}
	return 0.0
}

func (aa *AdvancedAnonymizer) calculateTrendPreservation(session *AdvancedAnonymizationSession, original []ContributionItem) float64 {
	// 计算趋势保留度
	if len(original) == 0 || len(session.AIReadyData) == 0 {
		return 1.0
	}

	correctTrends := 0
	totalComparisons := 0

	// 比较相邻记录的趋势是否保持
	for i := 0; i < len(original)-1 && i < len(session.AIReadyData)-1; i++ {
		origTrend := original[i+1].ContributionPercent - original[i].ContributionPercent

		if anonValue1, ok := session.AIReadyData[i]["contribution_percent"].(float64); ok {
			if anonValue2, ok := session.AIReadyData[i+1]["contribution_percent"].(float64); ok {
				anonTrend := anonValue2 - anonValue1

				// 检查趋势方向是否一致
				if (origTrend > 0 && anonTrend > 0) || (origTrend < 0 && anonTrend < 0) || (origTrend == 0 && math.Abs(anonTrend) < 0.01) {
					correctTrends++
				}
				totalComparisons++
			}
		}
	}

	if totalComparisons > 0 {
		return float64(correctTrends) / float64(totalComparisons)
	}
	return 1.0
}

// PrivacyMetrics 隐私指标
type PrivacyMetrics struct {
	EpsilonUsed     float64 `json:"epsilonUsed"`
	KAnonymityLevel int     `json:"kAnonymityLevel"`
	LDiversityLevel int     `json:"lDiversityLevel"`
	NoiseVariance   float64 `json:"noiseVariance"`
	PrivacyScore    float64 `json:"privacyScore"`
}

// DataQualityMetrics 数据质量指标
type DataQualityMetrics struct {
	DataUtility       float64 `json:"dataUtility"`
	InformationLoss   float64 `json:"informationLoss"`
	StatisticalError  float64 `json:"statisticalError"`
	TrendPreservation float64 `json:"trendPreservation"`
}
