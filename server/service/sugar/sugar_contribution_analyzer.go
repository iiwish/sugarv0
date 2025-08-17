package sugar

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// SugarContributionAnalyzer 贡献度分析器服务
type SugarContributionAnalyzer struct{}

// DimensionAggregation 维度聚合结果
type DimensionAggregation struct {
	DimensionName        string           `json:"dimension_name"`        // 维度名称
	DimensionCombination []string         `json:"dimension_combination"` // 维度组合
	AggregatedItems      []AggregatedItem `json:"aggregated_items"`      // 聚合后的项目
	TotalVariance        float64          `json:"total_variance"`        // 总方差
	SignificanceScore    float64          `json:"significance_score"`    // 显著性得分
	Summary              string           `json:"summary"`               // 聚合摘要
}

// AggregatedItem 聚合后的贡献项
type AggregatedItem struct {
	DimensionValues     map[string]interface{} `json:"dimension_values"`     // 维度值
	ContributionPercent float64                `json:"contribution_percent"` // 聚合后的贡献度
	IsPositiveDriver    bool                   `json:"is_positive_driver"`   // 是否为正向驱动
	ItemCount           int                    `json:"item_count"`           // 包含的原始项目数
	MaxContribution     float64                `json:"max_contribution"`     // 最大单项贡献度
	MinContribution     float64                `json:"min_contribution"`     // 最小单项贡献度
}

// AnalysisInsight 分析洞察
type AnalysisInsight struct {
	InsightType         string             `json:"insight_type"`         // 洞察类型：trend/outlier/pattern
	Title               string             `json:"title"`                // 洞察标题
	Description         string             `json:"description"`          // 详细描述
	SignificanceLevel   string             `json:"significance_level"`   // 显著性级别：high/medium/low
	AffectedDimensions  []string           `json:"affected_dimensions"`  // 影响的维度
	KeyMetrics          map[string]float64 `json:"key_metrics"`          // 关键指标
	BusinessImplication string             `json:"business_implication"` // 业务含义
}

// ContributionAnalysisResult 贡献度分析结果
type ContributionAnalysisResult struct {
	OriginalItemCount       int                     `json:"original_item_count"`      // 原始项目数量
	BestAggregation         *DimensionAggregation   `json:"best_aggregation"`         // 最佳聚合结果
	AlternativeAggregations []*DimensionAggregation `json:"alternative_aggregations"` // 备选聚合结果
	KeyInsights             []AnalysisInsight       `json:"key_insights"`             // 关键洞察
	RecommendedPrompt       string                  `json:"recommended_prompt"`       // 推荐的AI提示词
	AnalysisSummary         string                  `json:"analysis_summary"`         // 分析摘要
}

// AnalyzeContributions 分析贡献度数据，找出最佳聚合维度
func (s *SugarContributionAnalyzer) AnalyzeContributions(ctx context.Context, contributions []ContributionItem, groupByDimensions []string) (*ContributionAnalysisResult, error) {
	global.GVA_LOG.Info("开始贡献度聚合分析",
		zap.Int("contributionCount", len(contributions)),
		zap.Strings("groupByDimensions", groupByDimensions))

	if len(contributions) == 0 {
		return nil, fmt.Errorf("贡献度数据为空")
	}

	result := &ContributionAnalysisResult{
		OriginalItemCount:       len(contributions),
		AlternativeAggregations: make([]*DimensionAggregation, 0),
		KeyInsights:             make([]AnalysisInsight, 0),
	}

	// 1. 生成所有可能的维度组合
	dimensionCombinations := s.generateDimensionCombinations(groupByDimensions)
	global.GVA_LOG.Info("生成维度组合", zap.Int("combinationCount", len(dimensionCombinations)))

	// 2. 对每个维度组合进行聚合分析
	var aggregations []*DimensionAggregation
	for _, combination := range dimensionCombinations {
		aggregation, err := s.aggregateByDimensions(contributions, combination)
		if err != nil {
			global.GVA_LOG.Warn("维度聚合失败", zap.Strings("combination", combination), zap.Error(err))
			continue
		}
		aggregations = append(aggregations, aggregation)
	}

	if len(aggregations) == 0 {
		return nil, fmt.Errorf("所有维度聚合都失败")
	}

	// 3. 计算每个聚合的显著性得分
	for _, agg := range aggregations {
		agg.SignificanceScore = s.calculateSignificanceScore(agg)
	}

	// 4. 按显著性得分排序
	sort.Slice(aggregations, func(i, j int) bool {
		return aggregations[i].SignificanceScore > aggregations[j].SignificanceScore
	})

	// 5. 选择最佳聚合
	result.BestAggregation = aggregations[0]
	if len(aggregations) > 1 {
		result.AlternativeAggregations = aggregations[1:]
		// 只保留前3个备选方案
		if len(result.AlternativeAggregations) > 3 {
			result.AlternativeAggregations = result.AlternativeAggregations[:3]
		}
	}

	// 6. 生成关键洞察
	result.KeyInsights = s.generateInsights(result.BestAggregation, contributions)

	// 7. 生成推荐的AI提示词
	result.RecommendedPrompt = s.generateRecommendedPrompt(result.BestAggregation, result.KeyInsights)

	// 8. 生成分析摘要
	result.AnalysisSummary = s.generateAnalysisSummary(result)

	global.GVA_LOG.Info("贡献度聚合分析完成",
		zap.String("bestDimension", result.BestAggregation.DimensionName),
		zap.Float64("bestScore", result.BestAggregation.SignificanceScore),
		zap.Int("insightCount", len(result.KeyInsights)))

	return result, nil
}

// generateDimensionCombinations 生成所有可能的维度组合
func (s *SugarContributionAnalyzer) generateDimensionCombinations(dimensions []string) [][]string {
	var combinations [][]string

	// 单个维度
	for _, dim := range dimensions {
		combinations = append(combinations, []string{dim})
	}

	// 两个维度的组合
	for i := 0; i < len(dimensions); i++ {
		for j := i + 1; j < len(dimensions); j++ {
			combinations = append(combinations, []string{dimensions[i], dimensions[j]})
		}
	}

	// 三个维度的组合（如果维度数量足够）
	if len(dimensions) >= 3 {
		for i := 0; i < len(dimensions); i++ {
			for j := i + 1; j < len(dimensions); j++ {
				for k := j + 1; k < len(dimensions); k++ {
					combinations = append(combinations, []string{dimensions[i], dimensions[j], dimensions[k]})
				}
			}
		}
	}

	// 全维度组合（作为基准）
	if len(dimensions) > 1 {
		combinations = append(combinations, dimensions)
	}

	return combinations
}

// aggregateByDimensions 按指定维度组合进行聚合
func (s *SugarContributionAnalyzer) aggregateByDimensions(contributions []ContributionItem, dimensions []string) (*DimensionAggregation, error) {
	if len(dimensions) == 0 {
		return nil, fmt.Errorf("维度组合为空")
	}

	// 按维度组合分组
	groups := make(map[string][]ContributionItem)
	for _, contrib := range contributions {
		key := s.buildAggregationKey(contrib.DimensionValues, dimensions)
		groups[key] = append(groups[key], contrib)
	}

	// 聚合每个组
	var aggregatedItems []AggregatedItem
	var contributionValues []float64

	for key, items := range groups {
		if len(items) == 0 {
			continue
		}

		// 计算聚合指标
		totalContribution := 0.0
		maxContrib := items[0].ContributionPercent
		minContrib := items[0].ContributionPercent
		positiveCount := 0

		for _, item := range items {
			totalContribution += item.ContributionPercent
			if item.ContributionPercent > maxContrib {
				maxContrib = item.ContributionPercent
			}
			if item.ContributionPercent < minContrib {
				minContrib = item.ContributionPercent
			}
			if item.IsPositiveDriver {
				positiveCount++
			}
		}

		// 解析维度值
		dimensionValues := s.parseAggregationKey(key, dimensions)

		aggregatedItem := AggregatedItem{
			DimensionValues:     dimensionValues,
			ContributionPercent: totalContribution,
			IsPositiveDriver:    positiveCount > len(items)/2, // 多数决定
			ItemCount:           len(items),
			MaxContribution:     maxContrib,
			MinContribution:     minContrib,
		}

		aggregatedItems = append(aggregatedItems, aggregatedItem)
		contributionValues = append(contributionValues, totalContribution)
	}

	// 计算方差
	totalVariance := s.calculateVariance(contributionValues)

	// 生成维度名称
	dimensionName := strings.Join(dimensions, "+")

	// 生成摘要
	summary := s.generateAggregationSummary(aggregatedItems, dimensions)

	aggregation := &DimensionAggregation{
		DimensionName:        dimensionName,
		DimensionCombination: dimensions,
		AggregatedItems:      aggregatedItems,
		TotalVariance:        totalVariance,
		Summary:              summary,
	}

	return aggregation, nil
}

// buildAggregationKey 构建聚合键
func (s *SugarContributionAnalyzer) buildAggregationKey(dimensionValues map[string]interface{}, dimensions []string) string {
	var keyParts []string
	for _, dim := range dimensions {
		value := fmt.Sprintf("%v", dimensionValues[dim])
		keyParts = append(keyParts, fmt.Sprintf("%s:%s", dim, value))
	}
	return strings.Join(keyParts, "|")
}

// parseAggregationKey 解析聚合键
func (s *SugarContributionAnalyzer) parseAggregationKey(key string, dimensions []string) map[string]interface{} {
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

// calculateVariance 计算方差
func (s *SugarContributionAnalyzer) calculateVariance(values []float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}

	// 计算平均值
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// 计算方差
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))

	return variance
}

// calculateSignificanceScore 计算显著性得分
func (s *SugarContributionAnalyzer) calculateSignificanceScore(agg *DimensionAggregation) float64 {
	if len(agg.AggregatedItems) == 0 {
		return 0.0
	}

	// 基础得分：方差（数据分散程度）
	varianceScore := math.Min(agg.TotalVariance/100.0, 10.0) // 标准化到0-10

	// 聚合效果得分：聚合后项目数量的合理性
	aggregationScore := 0.0
	itemCount := len(agg.AggregatedItems)
	if itemCount >= 3 && itemCount <= 10 {
		aggregationScore = 10.0 // 理想的聚合数量
	} else if itemCount >= 2 && itemCount <= 15 {
		aggregationScore = 7.0 // 可接受的聚合数量
	} else if itemCount > 15 {
		aggregationScore = 3.0 // 聚合效果不明显
	} else {
		aggregationScore = 1.0 // 过度聚合
	}

	// 贡献度分布得分：检查是否有明显的主导项
	contributionDistributionScore := s.calculateContributionDistributionScore(agg.AggregatedItems)

	// 维度复杂度惩罚：维度组合越复杂，得分越低
	complexityPenalty := math.Max(0, float64(len(agg.DimensionCombination)-1)*2.0)

	// 综合得分
	totalScore := varianceScore + aggregationScore + contributionDistributionScore - complexityPenalty

	return math.Max(0, totalScore)
}

// calculateContributionDistributionScore 计算贡献度分布得分
func (s *SugarContributionAnalyzer) calculateContributionDistributionScore(items []AggregatedItem) float64 {
	if len(items) == 0 {
		return 0.0
	}

	// 找出最大贡献度
	maxContrib := 0.0
	for _, item := range items {
		if math.Abs(item.ContributionPercent) > math.Abs(maxContrib) {
			maxContrib = item.ContributionPercent
		}
	}

	// 计算主导性：最大贡献度占总贡献度的比例
	totalAbsContrib := 0.0
	for _, item := range items {
		totalAbsContrib += math.Abs(item.ContributionPercent)
	}

	if totalAbsContrib == 0 {
		return 0.0
	}

	dominanceRatio := math.Abs(maxContrib) / totalAbsContrib

	// 理想的主导性在30%-70%之间
	if dominanceRatio >= 0.3 && dominanceRatio <= 0.7 {
		return 8.0
	} else if dominanceRatio >= 0.2 && dominanceRatio <= 0.8 {
		return 6.0
	} else {
		return 3.0
	}
}

// generateAggregationSummary 生成聚合摘要
func (s *SugarContributionAnalyzer) generateAggregationSummary(items []AggregatedItem, dimensions []string) string {
	if len(items) == 0 {
		return "无有效聚合结果"
	}

	// 找出贡献度最大的项
	var maxItem *AggregatedItem
	maxAbsContrib := 0.0
	for i := range items {
		if math.Abs(items[i].ContributionPercent) > maxAbsContrib {
			maxAbsContrib = math.Abs(items[i].ContributionPercent)
			maxItem = &items[i]
		}
	}

	if maxItem == nil {
		return "无显著贡献项"
	}

	// 构建摘要
	dimensionDesc := strings.Join(dimensions, "、")

	var valueDesc []string
	for _, dim := range dimensions {
		if value, exists := maxItem.DimensionValues[dim]; exists {
			valueDesc = append(valueDesc, fmt.Sprintf("%v", value))
		}
	}
	valueDescStr := strings.Join(valueDesc, "、")

	direction := "增长"
	if maxItem.ContributionPercent < 0 {
		direction = "下降"
	}

	summary := fmt.Sprintf("按%s聚合，%s的%s贡献最显著(%.1f%%)",
		dimensionDesc, valueDescStr, direction, math.Abs(maxItem.ContributionPercent))

	return summary
}

// generateInsights 生成关键洞察
func (s *SugarContributionAnalyzer) generateInsights(agg *DimensionAggregation, originalContributions []ContributionItem) []AnalysisInsight {
	var insights []AnalysisInsight

	if len(agg.AggregatedItems) == 0 {
		return insights
	}

	// 1. 主导因子洞察
	dominantInsight := s.generateDominantFactorInsight(agg)
	if dominantInsight != nil {
		insights = append(insights, *dominantInsight)
	}

	// 2. 趋势洞察
	trendInsight := s.generateTrendInsight(agg)
	if trendInsight != nil {
		insights = append(insights, *trendInsight)
	}

	// 3. 异常值洞察
	outlierInsight := s.generateOutlierInsight(agg)
	if outlierInsight != nil {
		insights = append(insights, *outlierInsight)
	}

	return insights
}

// generateDominantFactorInsight 生成主导因子洞察
func (s *SugarContributionAnalyzer) generateDominantFactorInsight(agg *DimensionAggregation) *AnalysisInsight {
	if len(agg.AggregatedItems) == 0 {
		return nil
	}

	// 找出贡献度最大的项
	var maxItem *AggregatedItem
	maxAbsContrib := 0.0
	for i := range agg.AggregatedItems {
		if math.Abs(agg.AggregatedItems[i].ContributionPercent) > maxAbsContrib {
			maxAbsContrib = math.Abs(agg.AggregatedItems[i].ContributionPercent)
			maxItem = &agg.AggregatedItems[i]
		}
	}

	if maxItem == nil || maxAbsContrib < 10.0 {
		return nil
	}

	// 计算主导性
	totalAbsContrib := 0.0
	for _, item := range agg.AggregatedItems {
		totalAbsContrib += math.Abs(item.ContributionPercent)
	}

	dominanceRatio := maxAbsContrib / totalAbsContrib

	significanceLevel := "medium"
	if dominanceRatio > 0.6 {
		significanceLevel = "high"
	} else if dominanceRatio < 0.3 {
		significanceLevel = "low"
	}

	direction := "正向推动"
	if maxItem.ContributionPercent < 0 {
		direction = "负向拖累"
	}

	// 构建维度描述
	var dimDesc []string
	for _, dim := range agg.DimensionCombination {
		if value, exists := maxItem.DimensionValues[dim]; exists {
			dimDesc = append(dimDesc, fmt.Sprintf("%v", value))
		}
	}
	dimDescStr := strings.Join(dimDesc, "、")

	insight := &AnalysisInsight{
		InsightType:        "dominant_factor",
		Title:              "主导因子识别",
		Description:        fmt.Sprintf("%s是主要的%s因子，贡献度为%.1f%%，占总影响的%.1f%%", dimDescStr, direction, maxItem.ContributionPercent, dominanceRatio*100),
		SignificanceLevel:  significanceLevel,
		AffectedDimensions: agg.DimensionCombination,
		KeyMetrics: map[string]float64{
			"contribution_percent": maxItem.ContributionPercent,
			"dominance_ratio":      dominanceRatio,
			"item_count":           float64(maxItem.ItemCount),
		},
		BusinessImplication: s.generateBusinessImplication("dominant_factor", maxItem, agg.DimensionCombination),
	}

	return insight
}

// generateTrendInsight 生成趋势洞察
func (s *SugarContributionAnalyzer) generateTrendInsight(agg *DimensionAggregation) *AnalysisInsight {
	if len(agg.AggregatedItems) < 2 {
		return nil
	}

	positiveCount := 0
	negativeCount := 0
	totalPositiveContrib := 0.0
	totalNegativeContrib := 0.0

	for _, item := range agg.AggregatedItems {
		if item.ContributionPercent > 0 {
			positiveCount++
			totalPositiveContrib += item.ContributionPercent
		} else if item.ContributionPercent < 0 {
			negativeCount++
			totalNegativeContrib += math.Abs(item.ContributionPercent)
		}
	}

	// 判断整体趋势
	var description string
	var significanceLevel string

	if positiveCount > negativeCount*2 {
		description = fmt.Sprintf("整体呈现正向趋势，%d个项目贡献正向增长(%.1f%%)，%d个项目产生负向影响(%.1f%%)",
			positiveCount, totalPositiveContrib, negativeCount, totalNegativeContrib)
		significanceLevel = "high"
	} else if negativeCount > positiveCount*2 {
		description = fmt.Sprintf("整体呈现负向趋势，%d个项目产生负向影响(%.1f%%)，%d个项目贡献正向增长(%.1f%%)",
			negativeCount, totalNegativeContrib, positiveCount, totalPositiveContrib)
		significanceLevel = "high"
	} else {
		description = fmt.Sprintf("呈现混合趋势，正向项目%d个(%.1f%%)，负向项目%d个(%.1f%%)，影响相对均衡",
			positiveCount, totalPositiveContrib, negativeCount, totalNegativeContrib)
		significanceLevel = "medium"
	}

	insight := &AnalysisInsight{
		InsightType:        "trend",
		Title:              "整体趋势分析",
		Description:        description,
		SignificanceLevel:  significanceLevel,
		AffectedDimensions: agg.DimensionCombination,
		KeyMetrics: map[string]float64{
			"positive_count":        float64(positiveCount),
			"negative_count":        float64(negativeCount),
			"positive_contribution": totalPositiveContrib,
			"negative_contribution": totalNegativeContrib,
		},
		BusinessImplication: s.generateBusinessImplication("trend", nil, agg.DimensionCombination),
	}

	return insight
}

// generateOutlierInsight 生成异常值洞察
func (s *SugarContributionAnalyzer) generateOutlierInsight(agg *DimensionAggregation) *AnalysisInsight {
	if len(agg.AggregatedItems) < 3 {
		return nil
	}

	// 计算贡献度的统计信息
	var contributions []float64
	for _, item := range agg.AggregatedItems {
		contributions = append(contributions, math.Abs(item.ContributionPercent))
	}

	// 计算平均值和标准差
	sum := 0.0
	for _, c := range contributions {
		sum += c
	}
	mean := sum / float64(len(contributions))

	variance := 0.0
	for _, c := range contributions {
		diff := c - mean
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(contributions)))

	// 找出异常值（超过2个标准差）
	threshold := mean + 2*stdDev
	var outliers []AggregatedItem

	for _, item := range agg.AggregatedItems {
		if math.Abs(item.ContributionPercent) > threshold {
			outliers = append(outliers, item)
		}
	}

	if len(outliers) == 0 {
		return nil
	}

	// 构建异常值描述
	var outlierDescs []string
	for _, outlier := range outliers {
		var dimDesc []string
		for _, dim := range agg.DimensionCombination {
			if value, exists := outlier.DimensionValues[dim]; exists {
				dimDesc = append(dimDesc, fmt.Sprintf("%v", value))
			}
		}
		dimDescStr := strings.Join(dimDesc, "、")
		outlierDescs = append(outlierDescs, fmt.Sprintf("%s(%.1f%%)", dimDescStr, outlier.ContributionPercent))
	}

	description := fmt.Sprintf("发现%d个异常贡献项：%s，其贡献度显著超出平均水平(%.1f%%±%.1f%%)",
		len(outliers), strings.Join(outlierDescs, "、"), mean, stdDev)

	insight := &AnalysisInsight{
		InsightType:        "outlier",
		Title:              "异常值识别",
		Description:        description,
		SignificanceLevel:  "high",
		AffectedDimensions: agg.DimensionCombination,
		KeyMetrics: map[string]float64{
			"outlier_count":     float64(len(outliers)),
			"mean_contribution": mean,
			"std_deviation":     stdDev,
			"threshold":         threshold,
		},
		BusinessImplication: s.generateBusinessImplication("outlier", nil, agg.DimensionCombination),
	}

	return insight
}

// generateBusinessImplication 生成业务含义
func (s *SugarContributionAnalyzer) generateBusinessImplication(insightType string, item *AggregatedItem, dimensions []string) string {
	switch insightType {
	case "dominant_factor":
		if item != nil {
			if item.ContributionPercent > 0 {
				return "建议重点关注和复制该项目的成功经验，扩大其正向影响"
			} else {
				return "需要重点关注该项目的问题，分析根本原因并制定改进措施"
			}
		}
		return "建议深入分析主导因子的影响机制"
	case "trend":
		return "根据整体趋势制定相应的业务策略，加强正向因子，改善负向因子"
	case "outlier":
		return "深入分析异常项目的特殊情况，识别潜在的机会或风险"
	default:
		return "建议结合具体业务场景进行深入分析"
	}
}

// generateRecommendedPrompt 生成推荐的AI提示词
func (s *SugarContributionAnalyzer) generateRecommendedPrompt(agg *DimensionAggregation, insights []AnalysisInsight) string {
	var builder strings.Builder

	// 基础分析指导
	builder.WriteString("基于聚合分析结果，请重点关注以下发现：\n\n")

	// 添加最佳聚合维度信息
	builder.WriteString(fmt.Sprintf("**主要分析维度**: %s\n", agg.DimensionName))
	builder.WriteString(fmt.Sprintf("**聚合摘要**: %s\n\n", agg.Summary))

	// 添加关键洞察
	if len(insights) > 0 {
		builder.WriteString("**关键洞察**:\n")
		for i, insight := range insights {
			if i < 3 { // 只显示前3个最重要的洞察
				builder.WriteString(fmt.Sprintf("%d. %s: %s\n", i+1, insight.Title, insight.Description))
			}
		}
		builder.WriteString("\n")
	}

	// 添加分析指导
	builder.WriteString("**分析要求**:\n")
	builder.WriteString("1. 用一句话总结最显著的变化，重点突出主导因子\n")
	builder.WriteString("2. 采用业务语言描述，避免使用技术术语\n")
	builder.WriteString("3. 如果有明显的正负向趋势，请在开头概括整体方向\n")
	builder.WriteString("4. 重点关注贡献度最高的维度组合\n\n")

	// 添加输出格式示例
	builder.WriteString("**输出格式参考**:\n")
	if len(agg.AggregatedItems) > 0 {
		// 找出最显著的项目作为示例
		var maxItem *AggregatedItem
		maxAbsContrib := 0.0
		for i := range agg.AggregatedItems {
			if math.Abs(agg.AggregatedItems[i].ContributionPercent) > maxAbsContrib {
				maxAbsContrib = math.Abs(agg.AggregatedItems[i].ContributionPercent)
				maxItem = &agg.AggregatedItems[i]
			}
		}

		if maxItem != nil {
			direction := "增长"
			if maxItem.ContributionPercent < 0 {
				direction = "下降"
			}

			var valueDesc []string
			for _, dim := range agg.DimensionCombination {
				if value, exists := maxItem.DimensionValues[dim]; exists {
					valueDesc = append(valueDesc, fmt.Sprintf("%v", value))
				}
			}
			valueDescStr := strings.Join(valueDesc, "")

			builder.WriteString(fmt.Sprintf("- 示例: \"%s%s%s最显著\"\n", valueDescStr, direction, "贡献"))
		}
	}

	return builder.String()
}

// generateAnalysisSummary 生成分析摘要
func (s *SugarContributionAnalyzer) generateAnalysisSummary(result *ContributionAnalysisResult) string {
	var builder strings.Builder

	builder.WriteString("📊 **贡献度聚合分析摘要**\n\n")

	// 基本信息
	builder.WriteString(fmt.Sprintf("- **原始数据项**: %d个明细项目\n", result.OriginalItemCount))
	builder.WriteString(fmt.Sprintf("- **最佳聚合维度**: %s\n", result.BestAggregation.DimensionName))
	builder.WriteString(fmt.Sprintf("- **聚合后项目数**: %d个关键项目\n", len(result.BestAggregation.AggregatedItems)))
	builder.WriteString(fmt.Sprintf("- **显著性得分**: %.1f分\n", result.BestAggregation.SignificanceScore))

	// 关键发现
	if len(result.KeyInsights) > 0 {
		builder.WriteString("\n🔍 **关键发现**:\n")
		for i, insight := range result.KeyInsights {
			if i < 2 { // 只显示前2个最重要的发现
				builder.WriteString(fmt.Sprintf("- %s\n", insight.Description))
			}
		}
	}

	// 推荐行动
	builder.WriteString("\n💡 **推荐行动**:\n")
	builder.WriteString("- 重点关注聚合后的关键驱动因子\n")
	builder.WriteString("- 基于主导因子制定针对性的业务策略\n")
	builder.WriteString("- 持续监控异常项目的变化趋势\n")

	return builder.String()
}
