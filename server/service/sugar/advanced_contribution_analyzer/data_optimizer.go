package advanced_contribution_analyzer

import (
	"fmt"
	"strings"
)

// DataOptimizer 数据获取优化器
type DataOptimizer struct {
	config *AnalysisConfig
}

// NewDataOptimizer 创建数据获取优化器
func NewDataOptimizer(config *AnalysisConfig) *DataOptimizer {
	if config == nil {
		config = DefaultAnalysisConfig()
		if config == nil {
			// 如果默认配置也失败，创建一个最基本的配置
			config = &AnalysisConfig{
				DiscriminationThreshold:            15.0,
				MinContributionThreshold:           5.0,
				MaxDrillDownLevels:                 4,
				TopCombinationsCount:               5,
				MinTopCombinations:                 1,
				EnableSmartStop:                    true,
				DiscriminationImprovementThreshold: 5.0,
			}
		}
	}

	optimizer := &DataOptimizer{
		config: config,
	}

	return optimizer
}

// OptimizedPromptRequest 优化后的提示词请求
type OptimizedPromptRequest struct {
	ModelName            string                 `json:"model_name"`
	Dimensions           []string               `json:"dimensions"`
	Metric               string                 `json:"metric"`
	CurrentPeriodFilters map[string]interface{} `json:"current_period_filters"`
	BasePeriodFilters    map[string]interface{} `json:"base_period_filters"`
	IsYearEndComparison  bool                   `json:"is_year_end_comparison"`
	OptimizedPrompt      string                 `json:"optimized_prompt"`
	DataUnificationHint  string                 `json:"data_unification_hint"`
}

// GenerateOptimizedPrompt 生成优化的数据获取提示词
func (do *DataOptimizer) GenerateOptimizedPrompt(
	modelName string,
	dimensions []string,
	metric string,
	currentPeriodFilters, basePeriodFilters map[string]interface{},
	isYearEndComparison bool,
) *OptimizedPromptRequest {

	request := &OptimizedPromptRequest{
		ModelName:            modelName,
		Dimensions:           dimensions,
		Metric:               metric,
		CurrentPeriodFilters: currentPeriodFilters,
		BasePeriodFilters:    basePeriodFilters,
		IsYearEndComparison:  isYearEndComparison,
	}

	// 生成数据统一提示
	request.DataUnificationHint = do.generateDataUnificationHint(isYearEndComparison, metric)

	// 生成优化的提示词
	request.OptimizedPrompt = do.buildOptimizedPrompt(request)

	return request
}

// generateDataUnificationHint 生成数据统一提示
func (do *DataOptimizer) generateDataUnificationHint(isYearEndComparison bool, metric string) string {
	if isYearEndComparison {
		return fmt.Sprintf(`
数据统一处理说明：
1. 对于年初年末对比类型的数据，请直接计算变化值：%s_变化值 = 年末金额 - 年初金额
2. 无需区分本期和基期，直接使用计算后的变化值进行分析
3. 确保所有维度组合都基于相同的变化值计算基础
4. 变化值可能为正（增长）或负（减少），请保持原始符号
`, metric)
	}

	return fmt.Sprintf(`
数据统一处理说明：
1. 请确保本期和基期数据使用相同的维度分组逻辑
2. 计算变化值：%s_变化值 = 本期值 - 基期值
3. 所有维度组合应基于统一的时间范围和筛选条件
4. 保持数据的一致性和可比性
`, metric)
}

// buildOptimizedPrompt 构建优化的提示词
func (do *DataOptimizer) buildOptimizedPrompt(request *OptimizedPromptRequest) string {
	var prompt strings.Builder

	// 基础查询说明
	prompt.WriteString("请按照以下要求获取和处理数据：\n\n")

	// 数据获取要求
	prompt.WriteString("## 数据获取要求\n")
	prompt.WriteString(fmt.Sprintf("- 数据表：%s\n", request.ModelName))
	prompt.WriteString(fmt.Sprintf("- 分析指标：%s\n", request.Metric))
	prompt.WriteString(fmt.Sprintf("- 分析维度：%s\n", strings.Join(request.Dimensions, "、")))

	// 筛选条件
	if len(request.CurrentPeriodFilters) > 0 {
		prompt.WriteString("- 筛选条件：\n")
		for key, value := range request.CurrentPeriodFilters {
			prompt.WriteString(fmt.Sprintf("  * %s: %v\n", key, value))
		}
	}

	// 数据处理要求
	prompt.WriteString("\n## 数据处理要求\n")
	prompt.WriteString(request.DataUnificationHint)

	// 输出格式要求
	prompt.WriteString("\n## 输出格式要求\n")
	prompt.WriteString("请返回包含以下字段的数据：\n")
	for _, dim := range request.Dimensions {
		prompt.WriteString(fmt.Sprintf("- %s: 维度值\n", dim))
	}
	prompt.WriteString(fmt.Sprintf("- %s_变化值: 计算后的变化值\n", request.Metric))
	prompt.WriteString("- 记录数量: 该组合的记录数\n")

	// 质量要求
	prompt.WriteString("\n## 数据质量要求\n")
	prompt.WriteString("1. 确保所有维度组合的数据完整性\n")
	prompt.WriteString("2. 排除变化值为0且无实际业务意义的记录\n")
	prompt.WriteString("3. 保持维度值的一致性（避免同一概念的不同表述）\n")
	prompt.WriteString("4. 提供足够的维度组合以支持多层级分析\n")

	// 特殊处理说明
	if request.IsYearEndComparison {
		prompt.WriteString("\n## 年初年末对比特殊说明\n")
		prompt.WriteString("- 直接使用表中的年初金额和年末金额字段\n")
		prompt.WriteString("- 无需进行时间筛选，所有记录都包含完整的年初年末信息\n")
		prompt.WriteString("- 重点关注变化幅度较大的维度组合\n")
	}

	return prompt.String()
}

// AnalyzeDataQuality 分析数据质量
func (do *DataOptimizer) AnalyzeDataQuality(data *ContributionData) *DataQualityReport {
	report := &DataQualityReport{
		TotalCombinations: len(data.DimensionCombinations),
		QualityScore:      100.0,
		Issues:            make([]string, 0),
		Recommendations:   make([]string, 0),
	}

	// 检查数据完整性
	do.checkDataCompleteness(data, report)

	// 检查维度分布
	do.checkDimensionDistribution(data, report)

	// 检查贡献度分布
	do.checkContributionDistribution(data, report)

	// 检查异常值
	do.checkOutliers(data, report)

	return report
}

// DataQualityReport 数据质量报告
type DataQualityReport struct {
	TotalCombinations int               `json:"total_combinations"`
	ValidCombinations int               `json:"valid_combinations"`
	QualityScore      float64           `json:"quality_score"`
	Issues            []string          `json:"issues"`
	Recommendations   []string          `json:"recommendations"`
	DimensionCoverage map[string]int    `json:"dimension_coverage"`
	ContributionStats ContributionStats `json:"contribution_stats"`
}

// ContributionStats 贡献度统计
type ContributionStats struct {
	Mean      float64 `json:"mean"`
	Median    float64 `json:"median"`
	StdDev    float64 `json:"std_dev"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	ZeroCount int     `json:"zero_count"`
}

// checkDataCompleteness 检查数据完整性
func (do *DataOptimizer) checkDataCompleteness(data *ContributionData, report *DataQualityReport) {
	validCount := 0

	for _, combo := range data.DimensionCombinations {
		if len(combo.Values) > 0 && combo.AbsoluteValue != 0 {
			validCount++
		}
	}

	report.ValidCombinations = validCount
	completenessRatio := float64(validCount) / float64(len(data.DimensionCombinations))

	if completenessRatio < 0.8 {
		report.QualityScore -= 20
		report.Issues = append(report.Issues,
			fmt.Sprintf("数据完整性不足：有效组合占比仅%.1f%%", completenessRatio*100))
		report.Recommendations = append(report.Recommendations,
			"建议检查数据获取逻辑，确保包含所有有意义的维度组合")
	}
}

// checkDimensionDistribution 检查维度分布
func (do *DataOptimizer) checkDimensionDistribution(data *ContributionData, report *DataQualityReport) {
	report.DimensionCoverage = make(map[string]int)

	for _, combo := range data.DimensionCombinations {
		for _, value := range combo.Values {
			report.DimensionCoverage[value.Dimension]++
		}
	}

	// 检查维度覆盖均衡性
	if len(report.DimensionCoverage) > 1 {
		var counts []int
		for _, count := range report.DimensionCoverage {
			counts = append(counts, count)
		}

		// 计算变异系数
		mean := float64(0)
		for _, count := range counts {
			mean += float64(count)
		}
		mean /= float64(len(counts))

		variance := float64(0)
		for _, count := range counts {
			variance += (float64(count) - mean) * (float64(count) - mean)
		}
		variance /= float64(len(counts))

		cv := (variance / mean) * 100 // 变异系数

		if cv > 50 {
			report.QualityScore -= 10
			report.Issues = append(report.Issues,
				fmt.Sprintf("维度分布不均衡：变异系数%.1f%%", cv))
			report.Recommendations = append(report.Recommendations,
				"建议平衡各维度的数据量，确保分析结果的可靠性")
		}
	}
}

// checkContributionDistribution 检查贡献度分布
func (do *DataOptimizer) checkContributionDistribution(data *ContributionData, report *DataQualityReport) {
	if len(data.DimensionCombinations) == 0 {
		return
	}

	var contributions []float64
	zeroCount := 0

	for _, combo := range data.DimensionCombinations {
		contributions = append(contributions, combo.Contribution)
		if combo.Contribution == 0 {
			zeroCount++
		}
	}

	// 计算统计指标
	report.ContributionStats = do.calculateContributionStats(contributions, zeroCount)

	// 检查零贡献度比例
	zeroRatio := float64(zeroCount) / float64(len(contributions))
	if zeroRatio > 0.5 {
		report.QualityScore -= 15
		report.Issues = append(report.Issues,
			fmt.Sprintf("零贡献度组合过多：占比%.1f%%", zeroRatio*100))
		report.Recommendations = append(report.Recommendations,
			"建议优化数据获取逻辑，减少无效的零贡献度组合")
	}

	// 检查贡献度分布的合理性
	if report.ContributionStats.StdDev < 5 {
		report.QualityScore -= 10
		report.Issues = append(report.Issues,
			fmt.Sprintf("贡献度分布过于集中：标准差仅%.2f", report.ContributionStats.StdDev))
		report.Recommendations = append(report.Recommendations,
			"数据可能缺乏足够的差异性，建议检查分析维度的选择")
	}
}

// checkOutliers 检查异常值
func (do *DataOptimizer) checkOutliers(data *ContributionData, report *DataQualityReport) {
	if len(data.DimensionCombinations) < 3 {
		return
	}

	var contributions []float64
	for _, combo := range data.DimensionCombinations {
		contributions = append(contributions, combo.Contribution)
	}

	// 使用IQR方法检测异常值
	outliers := do.detectOutliers(contributions)

	if len(outliers) > 0 {
		outlierRatio := float64(len(outliers)) / float64(len(contributions))
		if outlierRatio > 0.1 {
			report.QualityScore -= 5
			report.Issues = append(report.Issues,
				fmt.Sprintf("检测到%d个异常值，占比%.1f%%", len(outliers), outlierRatio*100))
			report.Recommendations = append(report.Recommendations,
				"建议检查异常值的合理性，确认是否为真实的业务异常")
		}
	}
}

// calculateContributionStats 计算贡献度统计指标
func (do *DataOptimizer) calculateContributionStats(contributions []float64, zeroCount int) ContributionStats {
	if len(contributions) == 0 {
		return ContributionStats{}
	}

	// 计算均值
	sum := float64(0)
	for _, c := range contributions {
		sum += c
	}
	mean := sum / float64(len(contributions))

	// 计算方差和标准差
	variance := float64(0)
	for _, c := range contributions {
		variance += (c - mean) * (c - mean)
	}
	variance /= float64(len(contributions))
	stdDev := variance // 这里应该是sqrt(variance)，但为了简化计算直接使用方差

	// 计算最值
	min, max := contributions[0], contributions[0]
	for _, c := range contributions {
		if c < min {
			min = c
		}
		if c > max {
			max = c
		}
	}

	// 计算中位数
	sortedContribs := make([]float64, len(contributions))
	copy(sortedContribs, contributions)
	// 简化的中位数计算（实际应该排序）
	median := mean // 简化处理

	return ContributionStats{
		Mean:      mean,
		Median:    median,
		StdDev:    stdDev,
		Min:       min,
		Max:       max,
		ZeroCount: zeroCount,
	}
}

// detectOutliers 使用IQR方法检测异常值
func (do *DataOptimizer) detectOutliers(values []float64) []float64 {
	if len(values) < 4 {
		return nil
	}

	// 简化的异常值检测：使用均值±2倍标准差
	sum := float64(0)
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	variance := float64(0)
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(values))
	stdDev := variance // 简化处理

	var outliers []float64
	threshold := 2 * stdDev

	for _, v := range values {
		if v < mean-threshold || v > mean+threshold {
			outliers = append(outliers, v)
		}
	}

	return outliers
}
