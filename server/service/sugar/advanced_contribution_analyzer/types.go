package advanced_contribution_analyzer

import (
	"fmt"
	"log"
	"math"
)

// DimensionValue 维度值
type DimensionValue struct {
	Dimension string `json:"dimension"` // 维度名称，如"银行"、"币种"
	Value     string `json:"value"`     // 维度值，如"交通银行"、"欧元"
	Label     string `json:"label"`     // 显示标签（可能经过匿名化）
}

// DimensionCombination 维度组合
type DimensionCombination struct {
	Values        []DimensionValue `json:"values"`         // 维度值组合
	Contribution  float64          `json:"contribution"`   // 贡献度（百分比）
	AbsoluteValue float64          `json:"absolute_value"` // 绝对值
	Count         int              `json:"count"`          // 记录数量
}

// String 返回维度组合的字符串表示
func (dc *DimensionCombination) String() string {
	if len(dc.Values) == 0 {
		return "总计"
	}

	var parts []string
	for _, v := range dc.Values {
		parts = append(parts, v.Label)
	}

	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "+"
		}
		result += part
	}
	return result
}

// GetDimensionNames 获取维度名称列表
func (dc *DimensionCombination) GetDimensionNames() []string {
	var names []string
	for _, v := range dc.Values {
		names = append(names, v.Dimension)
	}
	return names
}

// ContainsValue 检查是否包含指定的维度值
func (dc *DimensionCombination) ContainsValue(dimension, value string) bool {
	for _, v := range dc.Values {
		if v.Dimension == dimension && v.Value == value {
			return true
		}
	}
	return false
}

// DimensionAnalysisLevel 维度分析层级
type DimensionAnalysisLevel struct {
	Dimensions      []string                `json:"dimensions"`       // 当前层级的维度列表
	Combinations    []*DimensionCombination `json:"combinations"`     // 该层级的所有组合
	Discrimination  float64                 `json:"discrimination"`   // 区分度
	MaxContribution float64                 `json:"max_contribution"` // 最大贡献度
	MinContribution float64                 `json:"min_contribution"` // 最小贡献度
	VarianceScore   float64                 `json:"variance_score"`   // 方差得分
}

// CalculateDiscrimination 计算区分度
func (dal *DimensionAnalysisLevel) CalculateDiscrimination() {
	if len(dal.Combinations) == 0 {
		dal.Discrimination = 0
		return
	}

	// 计算贡献度的绝对值，确保正确处理负贡献度
	var contributions []float64
	var absContributions []float64
	var sum float64

	for _, combo := range dal.Combinations {
		contrib := combo.Contribution
		absContrib := math.Abs(contrib)
		contributions = append(contributions, contrib)
		absContributions = append(absContributions, absContrib)
		sum += absContrib // 使用绝对值计算均值
	}

	if len(contributions) == 0 {
		dal.Discrimination = 0
		return
	}

	mean := sum / float64(len(contributions))

	// 计算最大最小贡献度差异（使用绝对值）
	dal.MaxContribution = absContributions[0]
	dal.MinContribution = absContributions[0]

	for _, absContrib := range absContributions {
		if absContrib > dal.MaxContribution {
			dal.MaxContribution = absContrib
		}
		if absContrib < dal.MinContribution {
			dal.MinContribution = absContrib
		}
	}

	// 特殊处理：单组合情况
	if len(dal.Combinations) == 1 {
		// 对于单组合，如果贡献度很高（>30%），认为有很高的区分度
		// 这种情况通常出现在最明细层级，有一个主导因子
		if dal.MaxContribution > 30.0 {
			dal.Discrimination = 80.0 + dal.MaxContribution*0.5 // 基础80分 + 贡献度加权
		} else if dal.MaxContribution > 15.0 {
			dal.Discrimination = 60.0 + dal.MaxContribution*0.8 // 基础60分 + 贡献度加权
		} else {
			dal.Discrimination = dal.MaxContribution * 2.0 // 低贡献度时的基础分数
		}
		dal.VarianceScore = 0 // 单组合无方差
		return
	}

	// 多组合情况：使用绝对值计算方差，更好地反映贡献度的分散程度
	var variance float64
	for _, absContrib := range absContributions {
		variance += math.Pow(absContrib-mean, 2)
	}
	variance /= float64(len(contributions))
	dal.VarianceScore = variance

	// 改进的区分度计算算法
	standardDev := math.Sqrt(variance)
	range_ := dal.MaxContribution - dal.MinContribution

	// 计算变异系数（标准差/均值），更好地反映相对差异
	coefficientOfVariation := 0.0
	if mean > 0 {
		coefficientOfVariation = standardDev / mean
	}

	// 优化权重分配：
	// - 变异系数占50%：反映相对差异程度
	// - 极差占30%：反映绝对差异程度
	// - 最大值权重占20%：突出主导因子的重要性
	maxContributionWeight := dal.MaxContribution / 100.0 // 归一化到0-1范围

	dal.Discrimination = coefficientOfVariation*50.0 + range_*0.3 + maxContributionWeight*20.0
}

// DrillDownResult 下钻分析结果
type DrillDownResult struct {
	Levels          []*DimensionAnalysisLevel `json:"levels"`           // 各层级分析结果
	OptimalLevel    int                       `json:"optimal_level"`    // 最优层级索引
	TopCombinations []*DimensionCombination   `json:"top_combinations"` // 顶级贡献组合
	AnalysisSummary string                    `json:"analysis_summary"` // 分析摘要
	DrillDownPath   []string                  `json:"drill_down_path"`  // 下钻路径
}

// AnalysisConfig 分析配置
type AnalysisConfig struct {
	// 区分度阈值：超过此值继续下钻
	DiscriminationThreshold float64 `json:"discrimination_threshold"`

	// 最小贡献度阈值：低于此值的组合将被过滤
	MinContributionThreshold float64 `json:"min_contribution_threshold"`

	// 最大下钻层级
	MaxDrillDownLevels int `json:"max_drill_down_levels"`

	// 每层级保留的顶级组合数量
	TopCombinationsCount int `json:"top_combinations_count"`

	// 至少返回的顶级组合数量
	MinTopCombinations int `json:"min_top_combinations"`

	// 是否启用智能停止（基于区分度变化率）
	EnableSmartStop bool `json:"enable_smart_stop"`

	// 区分度改善阈值：如果新层级的区分度改善小于此值，则停止下钻
	DiscriminationImprovementThreshold float64 `json:"discrimination_improvement_threshold"`
}

// DefaultAnalysisConfig 默认分析配置
func DefaultAnalysisConfig() *AnalysisConfig {
	log.Printf("DefaultAnalysisConfig: 开始创建默认分析配置")

	config := &AnalysisConfig{
		DiscriminationThreshold:            15.0, // 区分度阈值15%
		MinContributionThreshold:           1.0,  // 最小贡献度1%
		MaxDrillDownLevels:                 4,    // 最多4层下钻
		TopCombinationsCount:               15,   // 最多保留前15个组合
		MinTopCombinations:                 1,    // 至少返回1个组合
		EnableSmartStop:                    true, // 启用智能停止
		DiscriminationImprovementThreshold: 5.0,  // 区分度改善阈值5%
	}

	log.Printf("DefaultAnalysisConfig: 默认配置创建成功 - DiscriminationThreshold=%.2f, MaxDrillDownLevels=%d, EnableSmartStop=%t",
		config.DiscriminationThreshold, config.MaxDrillDownLevels, config.EnableSmartStop)

	return config
}

// ContributionData 贡献度数据
type ContributionData struct {
	DimensionCombinations []*DimensionCombination `json:"dimension_combinations"`
	TotalChange           float64                 `json:"total_change"`
	AvailableDimensions   []string                `json:"available_dimensions"`
}

// AnalysisMetrics 分析指标
type AnalysisMetrics struct {
	TotalCombinations     int     `json:"total_combinations"`     // 总组合数
	AnalyzedLevels        int     `json:"analyzed_levels"`        // 分析层级数
	OptimalDiscrimination float64 `json:"optimal_discrimination"` // 最优区分度
	ProcessingTimeMs      int64   `json:"processing_time_ms"`     // 处理时间（毫秒）
	StopReason            string  `json:"stop_reason"`            // 停止原因
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	ErrorMessage string   `json:"error_message,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

// Validate 验证贡献度数据
func (cd *ContributionData) Validate() *ValidationResult {
	result := &ValidationResult{IsValid: true}

	if len(cd.DimensionCombinations) == 0 {
		result.IsValid = false
		result.ErrorMessage = "没有可分析的维度组合数据"
		return result
	}

	if cd.TotalChange == 0 {
		result.IsValid = false
		result.ErrorMessage = "总变化值为0，无法计算贡献度"
		return result
	}

	if len(cd.AvailableDimensions) == 0 {
		result.IsValid = false
		result.ErrorMessage = "没有可用的维度信息"
		return result
	}

	// 检查贡献度总和是否合理
	var totalContribution float64
	for _, combo := range cd.DimensionCombinations {
		totalContribution += combo.Contribution
	}

	if math.Abs(totalContribution-100.0) > 1.0 { // 允许1%的误差
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("贡献度总和为%.2f%%，可能存在数据不一致", totalContribution))
	}

	return result
}
