package advanced_contribution_analyzer

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// AdvancedContributionAnalyzer 增强版贡献度分析器
type AdvancedContributionAnalyzer struct {
	config *AnalysisConfig
}

// NewAdvancedContributionAnalyzer 创建新的增强版贡献度分析器
func NewAdvancedContributionAnalyzer(config *AnalysisConfig) *AdvancedContributionAnalyzer {
	if config == nil {
		config = DefaultAnalysisConfig()
	}
	return &AdvancedContributionAnalyzer{
		config: config,
	}
}

// AnalyzeWithIntelligentDrillDown 执行智能下钻分析
func (aca *AdvancedContributionAnalyzer) AnalyzeWithIntelligentDrillDown(data *ContributionData) (*DrillDownResult, *AnalysisMetrics, error) {
	startTime := time.Now()

	// 验证输入数据
	validation := data.Validate()
	if !validation.IsValid {
		return nil, nil, fmt.Errorf("数据验证失败: %s", validation.ErrorMessage)
	}

	result := &DrillDownResult{
		Levels: make([]*DimensionAnalysisLevel, 0),
	}

	metrics := &AnalysisMetrics{
		TotalCombinations: len(data.DimensionCombinations),
	}

	// 按维度数量对组合进行分组
	dimensionGroups := aca.groupByDimensionCount(data.DimensionCombinations)

	// 执行逐层分析
	var previousDiscrimination float64 = 0
	stopReason := "达到最大层级"

	for level := 1; level <= aca.config.MaxDrillDownLevels; level++ {
		combinations, exists := dimensionGroups[level]
		if !exists || len(combinations) == 0 {
			stopReason = "没有更多维度组合"
			break
		}

		// 创建当前层级分析
		analysisLevel := aca.createAnalysisLevel(combinations, data.AvailableDimensions, level)
		result.Levels = append(result.Levels, analysisLevel)

		// 检查是否应该停止下钻
		shouldStop, reason := aca.shouldStopDrillDown(analysisLevel, previousDiscrimination, level)
		if shouldStop {
			stopReason = reason
			break
		}

		previousDiscrimination = analysisLevel.Discrimination
	}

	// 确定最优层级
	result.OptimalLevel = aca.findOptimalLevel(result.Levels)

	// 提取顶级贡献组合
	if result.OptimalLevel >= 0 && result.OptimalLevel < len(result.Levels) {
		result.TopCombinations = aca.extractTopCombinations(result.Levels[result.OptimalLevel])
	}

	// 生成分析摘要和下钻路径
	result.AnalysisSummary = aca.generateAnalysisSummary(result)
	result.DrillDownPath = aca.generateDrillDownPath(result)

	// 完善分析指标
	metrics.AnalyzedLevels = len(result.Levels)
	metrics.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	metrics.StopReason = stopReason
	if len(result.Levels) > 0 && result.OptimalLevel >= 0 {
		metrics.OptimalDiscrimination = result.Levels[result.OptimalLevel].Discrimination
	}

	return result, metrics, nil
}

// groupByDimensionCount 按维度数量对组合进行分组
func (aca *AdvancedContributionAnalyzer) groupByDimensionCount(combinations []*DimensionCombination) map[int][]*DimensionCombination {
	groups := make(map[int][]*DimensionCombination)

	for _, combo := range combinations {
		dimensionCount := len(combo.Values)
		if dimensionCount > 0 { // 排除总计行
			groups[dimensionCount] = append(groups[dimensionCount], combo)
		}
	}

	// 对每组内的组合按贡献度排序
	for _, group := range groups {
		sort.Slice(group, func(i, j int) bool {
			return math.Abs(group[i].Contribution) > math.Abs(group[j].Contribution)
		})
	}

	return groups
}

// createAnalysisLevel 创建分析层级
func (aca *AdvancedContributionAnalyzer) createAnalysisLevel(combinations []*DimensionCombination, availableDimensions []string, level int) *DimensionAnalysisLevel {
	// 过滤低贡献度组合
	filteredCombinations := aca.filterByContribution(combinations)

	// 确定当前层级的维度
	dimensions := aca.extractDimensionsFromLevel(filteredCombinations, availableDimensions, level)

	analysisLevel := &DimensionAnalysisLevel{
		Dimensions:   dimensions,
		Combinations: filteredCombinations,
	}

	// 计算区分度
	analysisLevel.CalculateDiscrimination()

	return analysisLevel
}

// filterByContribution 按贡献度过滤组合
func (aca *AdvancedContributionAnalyzer) filterByContribution(combinations []*DimensionCombination) []*DimensionCombination {
	var filtered []*DimensionCombination

	for _, combo := range combinations {
		if math.Abs(combo.Contribution) >= aca.config.MinContributionThreshold {
			filtered = append(filtered, combo)
		}
	}

	return filtered
}

// extractDimensionsFromLevel 从层级中提取维度信息
func (aca *AdvancedContributionAnalyzer) extractDimensionsFromLevel(combinations []*DimensionCombination, availableDimensions []string, level int) []string {
	dimensionSet := make(map[string]bool)

	for _, combo := range combinations {
		for _, value := range combo.Values {
			dimensionSet[value.Dimension] = true
		}
	}

	var dimensions []string
	for dim := range dimensionSet {
		dimensions = append(dimensions, dim)
	}

	// 按照可用维度的顺序排序
	sort.Slice(dimensions, func(i, j int) bool {
		iIndex := aca.findDimensionIndex(dimensions[i], availableDimensions)
		jIndex := aca.findDimensionIndex(dimensions[j], availableDimensions)
		return iIndex < jIndex
	})

	return dimensions
}

// findDimensionIndex 查找维度在可用维度列表中的索引
func (aca *AdvancedContributionAnalyzer) findDimensionIndex(dimension string, availableDimensions []string) int {
	for i, dim := range availableDimensions {
		if dim == dimension {
			return i
		}
	}
	return len(availableDimensions) // 如果没找到，放到最后
}

// shouldStopDrillDown 判断是否应该停止下钻
func (aca *AdvancedContributionAnalyzer) shouldStopDrillDown(currentLevel *DimensionAnalysisLevel, previousDiscrimination float64, level int) (bool, string) {
	// 检查区分度阈值
	if currentLevel.Discrimination < aca.config.DiscriminationThreshold {
		return true, fmt.Sprintf("区分度%.2f%%低于阈值%.2f%%", currentLevel.Discrimination, aca.config.DiscriminationThreshold)
	}

	// 检查智能停止条件
	if aca.config.EnableSmartStop && level > 1 && previousDiscrimination > 0 {
		improvement := currentLevel.Discrimination - previousDiscrimination
		if improvement < aca.config.DiscriminationImprovementThreshold {
			return true, fmt.Sprintf("区分度改善%.2f%%低于阈值%.2f%%", improvement, aca.config.DiscriminationImprovementThreshold)
		}
	}

	// 检查组合数量
	if len(currentLevel.Combinations) <= 1 {
		return true, "有效组合数量不足"
	}

	return false, ""
}

// findOptimalLevel 找到最优分析层级
func (aca *AdvancedContributionAnalyzer) findOptimalLevel(levels []*DimensionAnalysisLevel) int {
	if len(levels) == 0 {
		return -1
	}

	maxDiscrimination := -1.0
	optimalLevel := 0

	for i, level := range levels {
		// 综合考虑区分度和组合数量
		score := level.Discrimination

		// 对组合数量进行加权：组合数量适中时得分更高
		combinationCount := float64(len(level.Combinations))
		if combinationCount >= 3 && combinationCount <= 8 {
			score *= 1.2 // 组合数量适中时加权
		} else if combinationCount > 8 {
			score *= 0.9 // 组合过多时减权
		}

		if score > maxDiscrimination {
			maxDiscrimination = score
			optimalLevel = i
		}
	}

	return optimalLevel
}

// extractTopCombinations 提取顶级贡献组合
func (aca *AdvancedContributionAnalyzer) extractTopCombinations(level *DimensionAnalysisLevel) []*DimensionCombination {
	combinations := make([]*DimensionCombination, len(level.Combinations))
	copy(combinations, level.Combinations)

	// 按贡献度绝对值排序
	sort.Slice(combinations, func(i, j int) bool {
		return math.Abs(combinations[i].Contribution) > math.Abs(combinations[j].Contribution)
	})

	// 返回前N个组合
	count := aca.config.TopCombinationsCount
	if count > len(combinations) {
		count = len(combinations)
	}

	return combinations[:count]
}

// generateAnalysisSummary 生成分析摘要
func (aca *AdvancedContributionAnalyzer) generateAnalysisSummary(result *DrillDownResult) string {
	if len(result.TopCombinations) == 0 {
		return "未发现显著的贡献度差异"
	}

	var summary strings.Builder

	// 描述最显著的贡献组合
	topCombo := result.TopCombinations[0]
	summary.WriteString(fmt.Sprintf("最显著贡献来自%s，贡献度为%.1f%%",
		topCombo.String(), topCombo.Contribution))

	// 如果有多个显著组合，描述前几个
	if len(result.TopCombinations) > 1 {
		summary.WriteString("；其他显著贡献包括：")
		for i := 1; i < len(result.TopCombinations) && i < 3; i++ {
			combo := result.TopCombinations[i]
			summary.WriteString(fmt.Sprintf("%s(%.1f%%)", combo.String(), combo.Contribution))
			if i < len(result.TopCombinations)-1 && i < 2 {
				summary.WriteString("、")
			}
		}
	}

	// 添加分析层级信息
	if result.OptimalLevel >= 0 && result.OptimalLevel < len(result.Levels) {
		optimalLevel := result.Levels[result.OptimalLevel]
		summary.WriteString(fmt.Sprintf("。分析在%d维度层级达到最优区分度%.1f%%",
			len(optimalLevel.Dimensions), optimalLevel.Discrimination))
	}

	return summary.String()
}

// generateDrillDownPath 生成下钻路径
func (aca *AdvancedContributionAnalyzer) generateDrillDownPath(result *DrillDownResult) []string {
	var path []string

	for i, level := range result.Levels {
		pathItem := fmt.Sprintf("L%d: %s (区分度%.1f%%)",
			i+1, strings.Join(level.Dimensions, "+"), level.Discrimination)
		path = append(path, pathItem)

		if i == result.OptimalLevel {
			path[len(path)-1] += " [最优]"
		}
	}

	return path
}

// GetDimensionPriorityOrder 获取维度优先级排序
func (aca *AdvancedContributionAnalyzer) GetDimensionPriorityOrder(data *ContributionData) ([]string, error) {
	// 计算每个维度的单独贡献度方差
	dimensionVariances := make(map[string]float64)

	for _, dimension := range data.AvailableDimensions {
		variance := aca.calculateDimensionVariance(data.DimensionCombinations, dimension)
		dimensionVariances[dimension] = variance
	}

	// 按方差排序维度
	type dimensionScore struct {
		dimension string
		variance  float64
	}

	var scores []dimensionScore
	for dim, variance := range dimensionVariances {
		scores = append(scores, dimensionScore{dimension: dim, variance: variance})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].variance > scores[j].variance
	})

	var priorityOrder []string
	for _, score := range scores {
		priorityOrder = append(priorityOrder, score.dimension)
	}

	return priorityOrder, nil
}

// calculateDimensionVariance 计算单个维度的贡献度方差
func (aca *AdvancedContributionAnalyzer) calculateDimensionVariance(combinations []*DimensionCombination, dimension string) float64 {
	var contributions []float64

	// 收集该维度的所有单维度组合的贡献度
	for _, combo := range combinations {
		if len(combo.Values) == 1 && combo.Values[0].Dimension == dimension {
			contributions = append(contributions, combo.Contribution)
		}
	}

	if len(contributions) <= 1 {
		return 0
	}

	// 计算方差
	var sum float64
	for _, contrib := range contributions {
		sum += contrib
	}
	mean := sum / float64(len(contributions))

	var variance float64
	for _, contrib := range contributions {
		variance += math.Pow(contrib-mean, 2)
	}
	variance /= float64(len(contributions))

	return variance
}
