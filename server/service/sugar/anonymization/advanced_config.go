package anonymization

import (
	"math"
	"math/rand"
	"time"
)

// AdvancedAnonymizationConfig 高级匿名化配置
type AdvancedAnonymizationConfig struct {
	// 差分隐私参数
	Epsilon           float64 `json:"epsilon"`           // 隐私预算 (0.1-1.0, 越小隐私保护越强)
	Delta             float64 `json:"delta"`             // 失败概率 (通常设为 1/n^2)
	GlobalSensitivity float64 `json:"globalSensitivity"` // 全局敏感度

	// K-匿名性参数
	KAnonymity int `json:"kAnonymity"` // K值，每个等价类至少有K个记录
	LDiversity int `json:"lDiversity"` // L值，敏感属性至少有L个不同值

	// 数据质量保护参数
	PreserveTrends bool    `json:"preserveTrends"` // 是否保留趋势特征
	PreserveCorr   bool    `json:"preserveCorr"`   // 是否保留相关性
	NoiseVariance  float64 `json:"noiseVariance"`  // 噪声方差调节参数

	// 语义映射参数
	UseSemanticMapping bool `json:"useSemanticMapping"` // 是否使用语义映射
	PreserveBusiness   bool `json:"preserveBusiness"`   // 是否保留业务逻辑特征

	// 自适应参数
	AdaptiveNoise       bool `json:"adaptiveNoise"`       // 自适应噪声强度
	SensitivityAnalysis bool `json:"sensitivityAnalysis"` // 敏感性分析

	// 随机种子
	RandomSeed int64 `json:"randomSeed"`
}

// DefaultAdvancedConfig 返回默认的高级匿名化配置
func DefaultAdvancedConfig() *AdvancedAnonymizationConfig {
	return &AdvancedAnonymizationConfig{
		Epsilon:             0.5,  // 中等隐私保护
		Delta:               1e-5, // 标准失败概率
		GlobalSensitivity:   1.0,  // 标准敏感度
		KAnonymity:          3,    // 至少3个相似记录
		LDiversity:          2,    // 至少2个不同敏感值
		PreserveTrends:      true,
		PreserveCorr:        true,
		NoiseVariance:       0.1,
		UseSemanticMapping:  true,
		PreserveBusiness:    true,
		AdaptiveNoise:       true,
		SensitivityAnalysis: true,
		RandomSeed:          time.Now().UnixNano(),
	}
}

// HighPrivacyConfig 返回高隐私保护配置
func HighPrivacyConfig() *AdvancedAnonymizationConfig {
	config := DefaultAdvancedConfig()
	config.Epsilon = 0.1              // 强隐私保护
	config.KAnonymity = 5             // 更高的K值
	config.LDiversity = 3             // 更高的L值
	config.NoiseVariance = 0.2        // 更大的噪声
	config.UseSemanticMapping = false // 禁用语义映射以增强隐私
	return config
}

// BalancedConfig 返回平衡的配置（隐私与数据质量平衡）
func BalancedConfig() *AdvancedAnonymizationConfig {
	return DefaultAdvancedConfig() // 默认配置就是平衡配置
}

// HighQualityConfig 返回高数据质量配置（适用于对数据分析精度要求高的场景）
func HighQualityConfig() *AdvancedAnonymizationConfig {
	config := DefaultAdvancedConfig()
	config.Epsilon = 1.0         // 较低的隐私保护
	config.KAnonymity = 2        // 较低的K值
	config.LDiversity = 2        // 标准L值
	config.NoiseVariance = 0.05  // 较小的噪声
	config.PreserveTrends = true // 强化趋势保护
	config.PreserveCorr = true   // 强化相关性保护
	return config
}

// DataAnalyzer 数据特征分析器
type DataAnalyzer struct {
	config *AdvancedAnonymizationConfig
	rand   *rand.Rand
}

// NewDataAnalyzer 创建数据分析器
func NewDataAnalyzer(config *AdvancedAnonymizationConfig) *DataAnalyzer {
	return &DataAnalyzer{
		config: config,
		rand:   rand.New(rand.NewSource(config.RandomSeed)),
	}
}

// AnalyzeDataCharacteristics 分析数据特征
func (da *DataAnalyzer) AnalyzeDataCharacteristics(contributions []ContributionItem) *DataCharacteristics {
	if len(contributions) == 0 {
		return &DataCharacteristics{}
	}

	characteristics := &DataCharacteristics{
		TotalRecords:        len(contributions),
		DimensionCount:      0,
		ValueDistribution:   make(map[string][]float64),
		StatisticalFeatures: &StatisticalFeatures{},
	}

	// 计算统计特征
	var values []float64
	var contributions_percents []float64
	dimensionSet := make(map[string]bool)

	for _, contrib := range contributions {
		values = append(values, contrib.ChangeValue)
		contributions_percents = append(contributions_percents, contrib.ContributionPercent)

		// 统计维度数量
		for dim := range contrib.DimensionValues {
			dimensionSet[dim] = true
		}
	}

	characteristics.DimensionCount = len(dimensionSet)
	characteristics.StatisticalFeatures = da.calculateStatisticalFeatures(values, contributions_percents)

	return characteristics
}

// calculateStatisticalFeatures 计算统计特征
func (da *DataAnalyzer) calculateStatisticalFeatures(values, contributions []float64) *StatisticalFeatures {
	if len(values) == 0 {
		return &StatisticalFeatures{}
	}

	features := &StatisticalFeatures{}

	// 计算均值
	var sum, contribSum float64
	for i, v := range values {
		sum += v
		if i < len(contributions) {
			contribSum += contributions[i]
		}
	}
	features.Mean = sum / float64(len(values))
	features.ContributionMean = contribSum / float64(len(contributions))

	// 计算方差和标准差
	var variance, contribVariance float64
	for i, v := range values {
		diff := v - features.Mean
		variance += diff * diff

		if i < len(contributions) {
			contribDiff := contributions[i] - features.ContributionMean
			contribVariance += contribDiff * contribDiff
		}
	}
	features.Variance = variance / float64(len(values))
	features.StdDev = math.Sqrt(features.Variance)
	features.ContributionVariance = contribVariance / float64(len(contributions))
	features.ContributionStdDev = math.Sqrt(features.ContributionVariance)

	// 计算范围
	if len(values) > 0 {
		min, max := values[0], values[0]
		contribMin, contribMax := contributions[0], contributions[0]

		for i, v := range values {
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}

			if i < len(contributions) {
				if contributions[i] < contribMin {
					contribMin = contributions[i]
				}
				if contributions[i] > contribMax {
					contribMax = contributions[i]
				}
			}
		}

		features.Range = max - min
		features.ContributionRange = contribMax - contribMin
	}

	return features
}

// DataCharacteristics 数据特征
type DataCharacteristics struct {
	TotalRecords        int                  `json:"totalRecords"`
	DimensionCount      int                  `json:"dimensionCount"`
	ValueDistribution   map[string][]float64 `json:"valueDistribution"`
	StatisticalFeatures *StatisticalFeatures `json:"statisticalFeatures"`
}

// StatisticalFeatures 统计特征
type StatisticalFeatures struct {
	Mean                 float64 `json:"mean"`
	Variance             float64 `json:"variance"`
	StdDev               float64 `json:"stdDev"`
	Range                float64 `json:"range"`
	ContributionMean     float64 `json:"contributionMean"`
	ContributionVariance float64 `json:"contributionVariance"`
	ContributionStdDev   float64 `json:"contributionStdDev"`
	ContributionRange    float64 `json:"contributionRange"`
}
