package advanced_contribution_analyzer

import (
	"fmt"
	"log"
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
	log.Printf("NewAdvancedContributionAnalyzer: 开始创建分析器")

	if config == nil {
		log.Printf("NewAdvancedContributionAnalyzer: 输入配置为nil，使用默认配置")
		config = DefaultAnalysisConfig()
		if config == nil {
			log.Printf("NewAdvancedContributionAnalyzer: 默认配置也为nil，创建失败")
			return nil
		}
	}

	log.Printf("NewAdvancedContributionAnalyzer: 配置验证通过 - DiscriminationThreshold=%.2f", config.DiscriminationThreshold)

	analyzer := &AdvancedContributionAnalyzer{
		config: config,
	}

	log.Printf("NewAdvancedContributionAnalyzer: 分析器创建成功，地址=%p", analyzer)
	return analyzer
}

// AnalyzeWithIntelligentDrillDown 执行智能下钻分析
func (aca *AdvancedContributionAnalyzer) AnalyzeWithIntelligentDrillDown(data *ContributionData) (*DrillDownResult, *AnalysisMetrics, error) {
	startTime := time.Now()

	log.Printf("开始智能下钻分析: 输入组合数=%d, 可用维度=%v",
		len(data.DimensionCombinations), data.AvailableDimensions)

	// 验证输入数据
	validation := data.Validate()
	if !validation.IsValid {
		log.Printf("数据验证失败: %s", validation.ErrorMessage)
		return nil, nil, fmt.Errorf("数据验证失败: %s", validation.ErrorMessage)
	}
	log.Printf("数据验证通过")

	result := &DrillDownResult{
		Levels: make([]*DimensionAnalysisLevel, 0),
	}

	metrics := &AnalysisMetrics{
		TotalCombinations: len(data.DimensionCombinations),
	}

	// 生成所有层级的维度组合数据
	allLevelCombinations := aca.generateAllLevelCombinations(data)
	log.Printf("维度聚合完成: 生成了%d个层级的数据", len(allLevelCombinations))

	// 按维度数量对组合进行分组
	dimensionGroups := allLevelCombinations

	// 记录每个分组的详情
	for level, combinations := range dimensionGroups {
		log.Printf("  层级%d: %d个组合", level, len(combinations))
		if len(combinations) > 0 && len(combinations[0].Values) > 0 {
			log.Printf("    样本组合: %d个维度值", len(combinations[0].Values))
		}
	}

	// 获取实际存在的层级范围
	var minLevel, maxLevel int = math.MaxInt32, 0
	for level := range dimensionGroups {
		if level < minLevel {
			minLevel = level
		}
		if level > maxLevel {
			maxLevel = level
		}
	}

	log.Printf("实际数据层级范围: %d-%d", minLevel, maxLevel)

	// 执行逐层分析 - 从实际存在数据的层级开始
	var previousDiscrimination float64 = 0
	stopReason := "达到最大层级"
	levelIndex := 0 // 用于result.Levels的索引

	for level := minLevel; level <= maxLevel && level <= aca.config.MaxDrillDownLevels; level++ {
		log.Printf("开始分析层级%d", level)

		combinations, exists := dimensionGroups[level]
		if !exists || len(combinations) == 0 {
			log.Printf("层级%d没有组合数据，跳过", level)
			continue
		}

		log.Printf("层级%d找到%d个组合", level, len(combinations))

		// 创建当前层级分析
		analysisLevel := aca.createAnalysisLevel(combinations, data.AvailableDimensions, level)
		result.Levels = append(result.Levels, analysisLevel)

		log.Printf("层级%d分析完成: 区分度=%.2f, 有效组合数=%d",
			level, analysisLevel.Discrimination, len(analysisLevel.Combinations))

		// 检查是否应该停止下钻
		shouldStop, reason := aca.shouldStopDrillDown(analysisLevel, previousDiscrimination, levelIndex+1)
		if shouldStop {
			log.Printf("层级%d停止下钻: %s", level, reason)
			stopReason = reason
			break
		}

		previousDiscrimination = analysisLevel.Discrimination
		levelIndex++
	}

	log.Printf("下钻分析完成: 共分析%d个层级", len(result.Levels))

	// 确定最优层级
	result.OptimalLevel = aca.findOptimalLevel(result.Levels)
	log.Printf("最优层级确定: %d", result.OptimalLevel)

	// 提取顶级贡献组合
	if result.OptimalLevel >= 0 && result.OptimalLevel < len(result.Levels) {
		result.TopCombinations = aca.extractTopCombinations(result.Levels[result.OptimalLevel])
		log.Printf("提取顶级组合: %d个", len(result.TopCombinations))
	} else {
		log.Printf("警告: 无法提取顶级组合，最优层级=%d, 总层级数=%d",
			result.OptimalLevel, len(result.Levels))
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

	log.Printf("智能下钻分析完成: 层级数=%d, 最优层级=%d, 处理时间=%dms, 停止原因=%s",
		metrics.AnalyzedLevels, result.OptimalLevel, metrics.ProcessingTimeMs, stopReason)

	return result, metrics, nil
}

// groupByDimensionCount 按维度数量对组合进行分组
func (aca *AdvancedContributionAnalyzer) groupByDimensionCount(combinations []*DimensionCombination) map[int][]*DimensionCombination {
	groups := make(map[int][]*DimensionCombination)

	log.Printf("开始按维度数量分组: 输入组合数=%d", len(combinations))

	for i, combo := range combinations {
		if combo == nil {
			log.Printf("警告: 组合%d为nil，跳过", i)
			continue
		}

		dimensionCount := len(combo.Values)
		log.Printf("组合%d: 维度数=%d, 贡献度=%.2f", i, dimensionCount, combo.Contribution)

		if dimensionCount > 0 { // 排除总计行
			groups[dimensionCount] = append(groups[dimensionCount], combo)
		} else {
			log.Printf("警告: 组合%d维度数为0，跳过", i)
		}
	}

	log.Printf("分组完成: 共%d个分组", len(groups))

	// 对每组内的组合按贡献度排序
	for level, group := range groups {
		log.Printf("对层级%d的%d个组合进行排序", level, len(group))
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

	// 优化组合数量检查：对于单组合的高贡献度情况，不应该停止下钻
	if len(currentLevel.Combinations) == 0 {
		return true, "没有有效组合"
	}

	// 如果只有一个组合，但贡献度很高（>30%），不应该停止，让它成为候选结果
	if len(currentLevel.Combinations) == 1 {
		if currentLevel.MaxContribution > 30.0 {
			log.Printf("检测到单组合高贡献度(%.2f%%)，继续分析以便选择", currentLevel.MaxContribution)
			return false, "" // 不停止，让这个层级参与最优层级选择
		} else {
			return true, "单组合贡献度不足"
		}
	}

	// 修改智能停止条件：在可能存在高贡献度组合的情况下，不应该因为区分度改善不足而停止
	if aca.config.EnableSmartStop && level > 1 && previousDiscrimination > 0 {
		improvement := currentLevel.Discrimination - previousDiscrimination

		// 检查当前层级是否有高贡献度组合的潜力
		// 如果当前层级的最大贡献度已经很高（>25%），说明下一层级可能有更高的贡献度
		hasHighContributionPotential := currentLevel.MaxContribution > 25.0

		if improvement < aca.config.DiscriminationImprovementThreshold {
			if hasHighContributionPotential {
				log.Printf("检测到高贡献度潜力(%.2f%%)，忽略区分度改善阈值，继续下钻", currentLevel.MaxContribution)
				return false, "" // 继续下钻以寻找更高贡献度的组合
			} else {
				return true, fmt.Sprintf("区分度改善%.2f%%低于阈值%.2f%%", improvement, aca.config.DiscriminationImprovementThreshold)
			}
		}
	}

	return false, ""
}

// findOptimalLevel 找到最优分析层级
func (aca *AdvancedContributionAnalyzer) findOptimalLevel(levels []*DimensionAnalysisLevel) int {
	if len(levels) == 0 {
		return -1
	}

	// 优先考虑最明细层级（最后一个层级）的高贡献度组合
	// 如果最明细层级有显著的主导因子，直接选择它
	if len(levels) > 0 {
		lastLevel := levels[len(levels)-1]
		lastLevelIndex := len(levels) - 1
		// 检查最明细层级是否有高贡献度的主导因子
		if len(lastLevel.Combinations) > 0 && lastLevel.MaxContribution > 30.0 {
			// 如果最明细层级有超过30%贡献度的组合，且区分度足够高，优先选择
			if lastLevel.Discrimination > 60.0 {
				log.Printf("检测到最明细层级有高贡献度主导因子(%.2f%%)，优先选择最明细层级%d（维度数=%d）",
					lastLevel.MaxContribution, lastLevelIndex+1, len(lastLevel.Dimensions))
				return lastLevelIndex
			}
		}
	}

	// 优化层级选择逻辑：优先选择多维度组合层级
	maxScore := -1.0
	optimalLevel := 0

	for i, level := range levels {
		// 基础分数：区分度
		score := level.Discrimination

		// 维度数量加权：优先选择多维度组合
		dimensionCount := len(level.Dimensions)
		if dimensionCount >= 2 {
			// 多维度组合获得显著加权，确保能够显示完整的维度信息
			score *= 1.5
			log.Printf("层级%d为多维度组合(维度数=%d)，获得加权，调整后分数=%.2f",
				i, dimensionCount, score)
		}

		// 对组合数量进行加权：组合数量适中时得分更高
		combinationCount := float64(len(level.Combinations))
		if combinationCount >= 3 && combinationCount <= 15 {
			score *= 1.2 // 组合数量适中时加权
		} else if combinationCount > 15 {
			score *= 0.9 // 组合过多时减权
		}

		// 对最明细层级给予额外加权，鼓励选择更详细的分析
		if i == len(levels)-1 && level.MaxContribution > 20.0 {
			score *= 1.1 // 最明细层级且有较高贡献度时加权
		}

		// 特别处理：如果有多个层级，且当前层级是多维度的，进一步加权
		if len(levels) > 1 && dimensionCount >= 2 && level.MaxContribution > 10.0 {
			score *= 1.2
			log.Printf("层级%d为多维度且有较高贡献度(%.2f%%)，进一步加权，最终分数=%.2f",
				i, level.MaxContribution, score)
		}

		if score > maxScore {
			maxScore = score
			optimalLevel = i
		}
	}

	log.Printf("最优层级选择完成：选择层级索引%d，维度数=%d，最终分数=%.2f",
		optimalLevel, len(levels[optimalLevel].Dimensions), maxScore)

	return optimalLevel
}

// extractTopCombinations 提取顶级贡献组合
func (aca *AdvancedContributionAnalyzer) extractTopCombinations(level *DimensionAnalysisLevel) []*DimensionCombination {
	if len(level.Combinations) == 0 {
		log.Printf("警告: 层级没有组合数据，返回空结果")
		return []*DimensionCombination{}
	}

	// 复制并排序，避免修改原始数据
	combinations := make([]*DimensionCombination, len(level.Combinations))
	copy(combinations, level.Combinations)
	sort.Slice(combinations, func(i, j int) bool {
		return math.Abs(combinations[i].Contribution) > math.Abs(combinations[j].Contribution)
	})

	// 首先过滤出大于1%贡献度的组合
	var significantCombinations []*DimensionCombination
	for _, combo := range combinations {
		if math.Abs(combo.Contribution) >= aca.config.MinContributionThreshold {
			significantCombinations = append(significantCombinations, combo)
		}
	}

	// 确定返回数量：结合两种限制条件
	topNCount := aca.config.TopCombinationsCount
	significantCount := len(significantCombinations)

	// 核心逻辑：如果大于1%的组合很多，也是最多返回15条数据
	finalCount := topNCount // 默认使用TopN限制（15条）

	// 如果显著组合数量少于TopN，则返回显著组合数量
	if significantCount < topNCount {
		finalCount = significantCount
	}

	// 确保至少返回MinTopCombinations个（如果有足够数据）
	if finalCount < aca.config.MinTopCombinations && len(combinations) >= aca.config.MinTopCombinations {
		finalCount = aca.config.MinTopCombinations
	}

	// 确保不超过实际数据量
	if finalCount > len(combinations) {
		finalCount = len(combinations)
	}

	// 严格限制：无论如何都不能超过TopN配置的数量
	if finalCount > topNCount {
		finalCount = topNCount
	}

	result := combinations[:finalCount]

	log.Printf("提取顶级组合完成: TopN=%d, 显著组合=%d, 最终提取=%d",
		aca.config.TopCombinationsCount, significantCount, len(result))

	return result
}

// generateAnalysisSummary 生成分析摘要（简化版，移除自然语言描述）
func (aca *AdvancedContributionAnalyzer) generateAnalysisSummary(result *DrillDownResult) string {
	// 根据需求，移除自然语言描述，只保留技术指标
	if len(result.TopCombinations) == 0 {
		return "未发现显著的贡献度差异"
	}

	var summary strings.Builder

	// 添加分析层级信息（技术指标）
	if result.OptimalLevel >= 0 && result.OptimalLevel < len(result.Levels) {
		optimalLevel := result.Levels[result.OptimalLevel]
		summary.WriteString(fmt.Sprintf("分析在第%d层级达到最优区分度%.1f%%",
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

// generateAllLevelCombinations 生成所有层级的维度组合数据
func (aca *AdvancedContributionAnalyzer) generateAllLevelCombinations(data *ContributionData) map[int][]*DimensionCombination {
	log.Printf("开始生成所有层级的维度组合数据")

	// 首先按现有维度数量分组
	originalGroups := aca.groupByDimensionCount(data.DimensionCombinations)

	// 找到最大维度数
	maxDimensions := 0
	for level := range originalGroups {
		if level > maxDimensions {
			maxDimensions = level
		}
	}

	log.Printf("原始数据最大维度数: %d", maxDimensions)

	// 如果只有一个层级，需要生成其他层级的聚合数据
	if len(originalGroups) == 1 {
		log.Printf("检测到单一层级数据，开始生成聚合层级")
		return aca.generateAggregatedLevels(data, maxDimensions)
	}

	// 如果已有多个层级，直接返回
	log.Printf("检测到多层级数据，直接使用现有分组")
	return originalGroups
}

// generateAggregatedLevels 从最明细数据生成聚合层级
func (aca *AdvancedContributionAnalyzer) generateAggregatedLevels(data *ContributionData, maxDimensions int) map[int][]*DimensionCombination {
	result := make(map[int][]*DimensionCombination)

	// 保留原始最明细数据
	originalGroups := aca.groupByDimensionCount(data.DimensionCombinations)
	for level, combinations := range originalGroups {
		result[level] = combinations
	}

	// 从最明细数据生成各个聚合层级
	for targetLevel := 1; targetLevel < maxDimensions; targetLevel++ {
		log.Printf("生成%d维度层级的聚合数据", targetLevel)
		aggregatedCombinations := aca.aggregateToLevel(data.DimensionCombinations, targetLevel, data.AvailableDimensions)
		if len(aggregatedCombinations) > 0 {
			result[targetLevel] = aggregatedCombinations
			log.Printf("生成了%d个%d维度组合", len(aggregatedCombinations), targetLevel)
		}
	}

	return result
}

// aggregateToLevel 将明细数据聚合到指定维度层级
func (aca *AdvancedContributionAnalyzer) aggregateToLevel(detailCombinations []*DimensionCombination, targetLevel int, availableDimensions []string) []*DimensionCombination {
	if targetLevel <= 0 || targetLevel > len(availableDimensions) {
		return nil
	}

	// 生成目标层级的所有维度组合
	dimensionCombinations := aca.generateDimensionCombinations(availableDimensions, targetLevel)
	log.Printf("生成了%d个%d维度的组合模式", len(dimensionCombinations), targetLevel)

	// 按目标维度分组聚合所有数据
	aggregationMap := make(map[string]*DimensionCombination)

	for _, detail := range detailCombinations {
		// 对每个维度组合模式，检查明细数据是否匹配
		for _, dimCombo := range dimensionCombinations {
			// 检查明细数据是否包含这个维度组合的所有维度
			if aca.detailContainsAllDimensions(detail, dimCombo) {
				// 提取目标维度的值
				key, values := aca.extractTargetDimensionValues(detail, dimCombo)
				if key == "" {
					continue
				}

				if existing, exists := aggregationMap[key]; exists {
					// 累加贡献度和绝对值
					existing.Contribution += detail.Contribution
					existing.AbsoluteValue += detail.AbsoluteValue
					existing.Count += detail.Count
				} else {
					// 创建新的聚合组合
					aggregationMap[key] = &DimensionCombination{
						Values:        values,
						Contribution:  detail.Contribution,
						AbsoluteValue: detail.AbsoluteValue,
						Count:         detail.Count,
					}
				}
				break // 找到匹配的维度组合后跳出，避免重复计算
			}
		}
	}

	// 将聚合结果转换为切片
	var result []*DimensionCombination
	for _, combo := range aggregationMap {
		result = append(result, combo)
	}

	log.Printf("聚合完成，生成了%d个%d维度的组合", len(result), targetLevel)

	// 按贡献度绝对值排序
	sort.Slice(result, func(i, j int) bool {
		return math.Abs(result[i].Contribution) > math.Abs(result[j].Contribution)
	})

	return result
}

// detailContainsAllDimensions 检查明细数据是否包含指定维度组合的所有维度
func (aca *AdvancedContributionAnalyzer) detailContainsAllDimensions(detail *DimensionCombination, targetDimensions []string) bool {
	detailDimensions := make(map[string]bool)
	for _, value := range detail.Values {
		detailDimensions[value.Dimension] = true
	}

	for _, targetDim := range targetDimensions {
		if !detailDimensions[targetDim] {
			return false
		}
	}

	return true
}

// generateDimensionCombinations 生成指定层级的所有维度组合
func (aca *AdvancedContributionAnalyzer) generateDimensionCombinations(dimensions []string, level int) [][]string {
	if level <= 0 || level > len(dimensions) {
		return nil
	}

	var result [][]string
	aca.generateCombinationsRecursive(dimensions, level, 0, []string{}, &result)
	return result
}

// generateCombinationsRecursive 递归生成维度组合
func (aca *AdvancedContributionAnalyzer) generateCombinationsRecursive(dimensions []string, targetSize, startIndex int, current []string, result *[][]string) {
	if len(current) == targetSize {
		// 复制当前组合
		combo := make([]string, len(current))
		copy(combo, current)
		*result = append(*result, combo)
		return
	}

	for i := startIndex; i < len(dimensions); i++ {
		current = append(current, dimensions[i])
		aca.generateCombinationsRecursive(dimensions, targetSize, i+1, current, result)
		current = current[:len(current)-1] // 回溯
	}
}

// extractTargetDimensionValues 从明细组合中提取目标维度的值
func (aca *AdvancedContributionAnalyzer) extractTargetDimensionValues(detail *DimensionCombination, targetDimensions []string) (string, []DimensionValue) {
	var values []DimensionValue
	var keyParts []string

	// 为每个目标维度查找对应的值
	for _, targetDim := range targetDimensions {
		found := false
		for _, value := range detail.Values {
			if value.Dimension == targetDim {
				values = append(values, value)
				keyParts = append(keyParts, fmt.Sprintf("%s:%s", value.Dimension, value.Value))
				found = true
				break
			}
		}
		if !found {
			// 如果明细数据中没有目标维度，跳过这个组合
			return "", nil
		}
	}

	if len(values) == 0 {
		return "", nil
	}

	key := strings.Join(keyParts, "|")
	return key, values
}
