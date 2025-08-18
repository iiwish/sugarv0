package sugar

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sugarRes "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/sugar/advanced_contribution_analyzer"
	"github.com/flipped-aurora/gin-vue-admin/server/service/sugar/anonymization_lite"
	"go.uber.org/zap"
)

// ContributionAnalyzer 贡献度分析器 - 负责贡献度计算和智能分析
type ContributionAnalyzer struct {
	advancedAnalyzer *advanced_contribution_analyzer.AdvancedContributionService // 增强版分析器
	dataProcessor    *DataProcessor
}

// NewContributionAnalyzer 创建贡献度分析器
func NewContributionAnalyzer(advancedAnalyzer *advanced_contribution_analyzer.AdvancedContributionService) *ContributionAnalyzer {
	return &ContributionAnalyzer{
		advancedAnalyzer: advancedAnalyzer,
		dataProcessor:    NewDataProcessor(),
	}
}

// PerformAnalysis 执行贡献度分析（优先使用增强版分析器）
func (ca *ContributionAnalyzer) PerformAnalysis(ctx context.Context, modelName, targetMetric string, currentPeriodFilters, basePeriodFilters map[string]interface{}, groupByDimensions []string, userId string) (string, *anonymization_lite.LiteAnonymizationSession, bool, error) {
	global.GVA_LOG.Info("开始执行贡献度分析",
		zap.String("modelName", modelName),
		zap.String("targetMetric", targetMetric),
		zap.Strings("groupByDimensions", groupByDimensions))

	// 检查增强版分析器状态
	if ca.advancedAnalyzer == nil {
		global.GVA_LOG.Warn("增强版分析器为nil，直接使用lite版本")
	} else {
		global.GVA_LOG.Info("增强版分析器可用，开始使用增强版分析")
		aiDataText, session, err := ca.processAdvancedAnalysis(ctx, ca.advancedAnalyzer, modelName, targetMetric, currentPeriodFilters, basePeriodFilters, groupByDimensions, userId)
		if err != nil {
			global.GVA_LOG.Warn("增强版分析器处理失败，回退到lite版本", zap.Error(err))
		} else {
			global.GVA_LOG.Info("增强版分析器处理成功")
			return aiDataText, session, true, nil
		}
	}

	// 回退到lite版本分析
	global.GVA_LOG.Info("使用lite版本进行贡献度分析")
	aiDataText, session, err := ca.processLiteAnalysis(ctx, modelName, targetMetric, currentPeriodFilters, basePeriodFilters, groupByDimensions, userId)
	return aiDataText, session, false, err
}

// processAdvancedAnalysis 使用增强版分析器进行智能分析
func (ca *ContributionAnalyzer) processAdvancedAnalysis(ctx context.Context, advancedService *advanced_contribution_analyzer.AdvancedContributionService, modelName, targetMetric string, currentPeriodFilters, basePeriodFilters map[string]interface{}, groupByDimensions []string, userId string) (string, *anonymization_lite.LiteAnonymizationSession, error) {
	global.GVA_LOG.Info("使用增强版分析器进行智能分析")

	// 验证增强版分析器服务
	if advancedService == nil {
		return "", nil, errors.New("增强版分析器服务不可用")
	}

	// 验证增强版分析器配置
	currentConfig := advancedService.GetCurrentConfig()
	if currentConfig == nil {
		return "", nil, errors.New("增强版分析器配置不可用")
	}

	// 1. 获取基础数据
	currentData, baseData, err := ca.dataProcessor.FetchDataConcurrently(ctx, modelName, targetMetric, currentPeriodFilters, basePeriodFilters, groupByDimensions, userId)
	if err != nil {
		return "", nil, fmt.Errorf("获取数据失败: %w", err)
	}

	// 2. 计算基础贡献度
	contributions, err := ca.calculateContributions(currentData, baseData, targetMetric, groupByDimensions)
	if err != nil {
		return "", nil, fmt.Errorf("计算贡献度失败: %w", err)
	}

	// 3. 构建增强版分析请求
	analysisRequest := &advanced_contribution_analyzer.AnalysisRequest{
		ModelName:            modelName,
		Metric:               targetMetric,
		Dimensions:           groupByDimensions,
		CurrentPeriodFilters: currentPeriodFilters,
		BasePeriodFilters:    basePeriodFilters,
		IsYearEndComparison:  ca.dataProcessor.isYearEndComparisonModel(modelName, targetMetric),
		RawContributions:     ca.convertToAdvancedContributions(contributions),
		TotalChange:          ca.calculateTotalChange(contributions),
	}

	// 4. 执行增强版分析
	analysisResponse, err := advancedService.PerformAdvancedAnalysis(ctx, analysisRequest)
	if err != nil {
		return "", nil, fmt.Errorf("增强版分析失败: %w", err)
	}

	if !analysisResponse.Success {
		return "", nil, errors.New(analysisResponse.ErrorMessage)
	}

	global.GVA_LOG.Info("增强版智能分析完成",
		zap.Int("analyzedLevels", analysisResponse.AnalysisMetrics.AnalyzedLevels),
		zap.Int("optimalLevelIndex", analysisResponse.DrillDownResult.OptimalLevel))

	// 5. 转换优化后的数据为匿名化格式
	contributionData := ca.convertDrillDownToContributionData(analysisResponse.DrillDownResult.TopCombinations)
	if len(contributionData) == 0 {
		return "", nil, errors.New("没有有效的优化贡献度数据")
	}

	// 6. 使用 anonymization_lite 进行匿名化处理
	config := anonymization_lite.DefaultLiteConfig()
	liteService := anonymization_lite.NewLiteAnonymizationService(config)

	aiDataText, session, err := liteService.ProcessAndSerialize(contributionData)
	if err != nil {
		return "", nil, fmt.Errorf("增强版匿名化处理失败: %w", err)
	}

	// 7. 构建增强的AI数据文本
	enhancedAiDataText := ca.buildAdvancedAnalysisText(analysisResponse, aiDataText)

	global.GVA_LOG.Info("增强版分析处理完成",
		zap.Int("originalDataCount", len(contributionData)),
		zap.Int("processedDataCount", len(session.AIReadyData)),
		zap.Int("textLength", len(enhancedAiDataText)))

	return enhancedAiDataText, session, nil
}

// processLiteAnalysis 使用lite版本进行分析
func (ca *ContributionAnalyzer) processLiteAnalysis(ctx context.Context, modelName, targetMetric string, currentPeriodFilters, basePeriodFilters map[string]interface{}, groupByDimensions []string, userId string) (string, *anonymization_lite.LiteAnonymizationSession, error) {
	global.GVA_LOG.Info("使用lite版本进行贡献度分析")

	// 1. 获取数据
	currentData, baseData, err := ca.dataProcessor.FetchDataConcurrently(ctx, modelName, targetMetric, currentPeriodFilters, basePeriodFilters, groupByDimensions, userId)
	if err != nil {
		return "", nil, fmt.Errorf("获取数据失败: %w", err)
	}

	// 2. 转换为lite版本需要的格式
	contributionData := ca.convertToContributionData(currentData, baseData, targetMetric, groupByDimensions)
	if len(contributionData) == 0 {
		return "", nil, errors.New("没有有效的贡献度数据")
	}

	// 3. 使用 anonymization_lite 进行匿名化处理
	config := anonymization_lite.DefaultLiteConfig()
	liteService := anonymization_lite.NewLiteAnonymizationService(config)

	aiDataText, session, err := liteService.ProcessAndSerialize(contributionData)
	if err != nil {
		return "", nil, fmt.Errorf("lite版本匿名化处理失败: %w", err)
	}

	global.GVA_LOG.Info("lite版本分析处理完成",
		zap.Int("originalDataCount", len(contributionData)),
		zap.Int("processedDataCount", len(session.AIReadyData)),
		zap.Int("textLength", len(aiDataText)))

	return aiDataText, session, nil
}

// calculateContributions 计算贡献度分析
func (ca *ContributionAnalyzer) calculateContributions(currentData, baseData *sugarRes.SugarFormulaGetResponse, targetMetric string, groupByDimensions []string) ([]ContributionItem, error) {
	// 将数据按维度组合进行分组
	currentGroups := ca.groupDataByDimensions(currentData.Results, groupByDimensions, targetMetric)
	baseGroups := ca.groupDataByDimensions(baseData.Results, groupByDimensions, targetMetric)

	// 计算每个维度组合的贡献度
	var contributions []ContributionItem
	var totalChange float64

	// 获取所有唯一的维度组合
	allKeys := ca.getAllUniqueKeys(currentGroups, baseGroups)

	// 第一轮：计算变化值和总变化
	for _, key := range allKeys {
		currentValue := currentGroups[key]
		baseValue := baseGroups[key]
		changeValue := currentValue - baseValue

		totalChange += changeValue

		// 解析维度值
		dimensionValues := ca.parseDimensionKey(key, groupByDimensions)

		contributions = append(contributions, ContributionItem{
			DimensionValues: dimensionValues,
			CurrentValue:    currentValue,
			BaseValue:       baseValue,
			ChangeValue:     changeValue,
		})
	}

	// 第二轮：计算贡献度百分比和正负向判断
	for i := range contributions {
		if totalChange != 0 {
			contributions[i].ContributionPercent = (contributions[i].ChangeValue / totalChange) * 100
		} else {
			contributions[i].ContributionPercent = 0
		}

		// 判断是否为正向驱动因子
		contributions[i].IsPositiveDriver = (contributions[i].ChangeValue * totalChange) >= 0
	}

	global.GVA_LOG.Info("贡献度计算完成",
		zap.Float64("totalChange", totalChange),
		zap.Int("contributionCount", len(contributions)))

	return contributions, nil
}

// convertToContributionData 转换数据为anonymization_lite包需要的ContributionItem格式
func (ca *ContributionAnalyzer) convertToContributionData(currentData, baseData *sugarRes.SugarFormulaGetResponse, targetMetric string, groupByDimensions []string) []anonymization_lite.ContributionItem {
	// 将数据按维度组合进行分组
	currentGroups := ca.groupDataByDimensions(currentData.Results, groupByDimensions, targetMetric)
	baseGroups := ca.groupDataByDimensions(baseData.Results, groupByDimensions, targetMetric)

	var contributions []anonymization_lite.ContributionItem
	var totalChange float64

	// 获取所有唯一的维度组合
	allKeys := ca.getAllUniqueKeys(currentGroups, baseGroups)

	// 第一轮：计算变化值和总变化
	changeValues := make(map[string]float64)
	baseValues := make(map[string]float64)
	currentValues := make(map[string]float64)

	for _, key := range allKeys {
		currentValue := currentGroups[key]
		baseValue := baseGroups[key]
		changeValue := currentValue - baseValue

		totalChange += changeValue
		changeValues[key] = changeValue
		baseValues[key] = baseValue
		currentValues[key] = currentValue
	}

	// 第二轮：计算贡献度百分比和增强信息
	var contributionPercents []float64
	for _, key := range allKeys {
		changeValue := changeValues[key]
		baseValue := baseValues[key]
		currentValue := currentValues[key]

		// 解析维度值
		dimensionValues := ca.parseDimensionKey(key, groupByDimensions)

		contributionPercent := 0.0
		if totalChange != 0 {
			contributionPercent = (changeValue / totalChange) * 100
		}
		contributionPercents = append(contributionPercents, math.Abs(contributionPercent))

		// 判断是否为正向驱动因子
		isPositiveDriver := (changeValue * totalChange) >= 0

		// 计算变化率百分比（避免泄露绝对值）
		changeRatePercent := 0.0
		if baseValue != 0 {
			changeRatePercent = (changeValue / baseValue) * 100
		} else if currentValue != 0 {
			// 基期为0但当期有值，视为100%增长
			changeRatePercent = 100.0
		}

		// 确定趋势方向
		trendDirection := "持平"
		if changeValue > 0 {
			trendDirection = "增长"
		} else if changeValue < 0 {
			trendDirection = "下降"
		}

		// 计算影响程度（基于贡献度绝对值）
		impactLevel := "低"
		absContribution := math.Abs(contributionPercent)
		if absContribution >= 10.0 {
			impactLevel = "高"
		} else if absContribution >= 3.0 {
			impactLevel = "中"
		}

		contributions = append(contributions, anonymization_lite.ContributionItem{
			DimensionValues:     dimensionValues,
			ContributionPercent: contributionPercent,
			IsPositiveDriver:    isPositiveDriver,
			ChangeRatePercent:   changeRatePercent,
			TrendDirection:      trendDirection,
			ImpactLevel:         impactLevel,
			RelativeImportance:  0, // 将在第三轮计算
		})
	}

	// 第三轮：计算相对重要性（基于贡献度绝对值的排名百分位）
	ca.calculateRelativeImportance(contributions, contributionPercents)

	global.GVA_LOG.Info("数据转换完成",
		zap.Float64("totalChange", totalChange),
		zap.Int("contributionCount", len(contributions)))

	return contributions
}

// convertToAdvancedContributions 转换为增强版分析器需要的数据格式
func (ca *ContributionAnalyzer) convertToAdvancedContributions(contributions []ContributionItem) []*advanced_contribution_analyzer.DimensionCombination {
	var advancedContributions []*advanced_contribution_analyzer.DimensionCombination

	for i, contrib := range contributions {
		if len(contrib.DimensionValues) == 0 {
			global.GVA_LOG.Warn("跳过无效的贡献项：维度值为空", zap.Int("itemIndex", i))
			continue
		}

		// 构建维度值列表
		var values []advanced_contribution_analyzer.DimensionValue
		for dimName, dimValue := range contrib.DimensionValues {
			dimNameStr := strings.TrimSpace(dimName)
			dimValueStr := strings.TrimSpace(fmt.Sprintf("%v", dimValue))

			if dimNameStr == "" || dimValueStr == "" || dimValueStr == "<nil>" {
				continue
			}

			value := advanced_contribution_analyzer.DimensionValue{
				Dimension: dimNameStr,
				Value:     dimValueStr,
				Label:     dimValueStr,
			}
			values = append(values, value)
		}

		if len(values) == 0 {
			continue
		}

		advancedContrib := &advanced_contribution_analyzer.DimensionCombination{
			Values:        values,
			Contribution:  contrib.ContributionPercent,
			AbsoluteValue: math.Abs(contrib.ChangeValue),
			Count:         1,
		}

		advancedContributions = append(advancedContributions, advancedContrib)
	}

	global.GVA_LOG.Info("数据格式转换完成",
		zap.Int("originalCount", len(contributions)),
		zap.Int("convertedCount", len(advancedContributions)))

	return advancedContributions
}

// convertDrillDownToContributionData 将下钻结果转换为 anonymization_lite 包需要的格式
func (ca *ContributionAnalyzer) convertDrillDownToContributionData(topCombinations []*advanced_contribution_analyzer.DimensionCombination) []anonymization_lite.ContributionItem {
	var contributionData []anonymization_lite.ContributionItem
	var contributionPercents []float64

	for _, item := range topCombinations {
		// 重建维度值映射
		dimensionValues := make(map[string]interface{})
		for _, value := range item.Values {
			dimensionValues[value.Dimension] = value.Value
		}

		// 确定趋势方向
		trendDirection := "持平"
		if item.Contribution > 0 {
			trendDirection = "增长"
		} else if item.Contribution < 0 {
			trendDirection = "下降"
		}

		// 计算影响程度（基于贡献度绝对值）
		impactLevel := "低"
		absContribution := math.Abs(item.Contribution)
		if absContribution >= 10.0 {
			impactLevel = "高"
		} else if absContribution >= 3.0 {
			impactLevel = "中"
		}

		// 计算变化率百分比（基于绝对值的估算，避免泄露具体数值）
		// 这里使用贡献度作为变化率的近似值，实际应用中可以根据业务逻辑调整
		changeRatePercent := item.Contribution

		contributionItem := anonymization_lite.ContributionItem{
			DimensionValues:     dimensionValues,
			ContributionPercent: item.Contribution,
			IsPositiveDriver:    item.Contribution >= 0,
			ChangeRatePercent:   changeRatePercent,
			TrendDirection:      trendDirection,
			ImpactLevel:         impactLevel,
			RelativeImportance:  0, // 将在后续计算
		}
		contributionData = append(contributionData, contributionItem)
		contributionPercents = append(contributionPercents, absContribution)
	}

	// 计算相对重要性
	ca.calculateRelativeImportance(contributionData, contributionPercents)

	return contributionData
}

// calculateTotalChange 计算总变化值
func (ca *ContributionAnalyzer) calculateTotalChange(contributions []ContributionItem) float64 {
	var totalChange float64
	for _, contrib := range contributions {
		totalChange += contrib.ChangeValue
	}
	return totalChange
}

// buildAdvancedAnalysisText 构建增强版分析文本
func (ca *ContributionAnalyzer) buildAdvancedAnalysisText(analysisResponse *advanced_contribution_analyzer.AnalysisResponse, aiDataText string) string {
	var builder strings.Builder

	// 添加智能分析摘要
	builder.WriteString("【增强版智能贡献度分析结果】\n")
	builder.WriteString("说明：以下数据已经过智能下钻分析，基于区分度计算优化维度选择\n\n")

	// 添加分析指标信息
	if analysisResponse.AnalysisMetrics != nil {
		builder.WriteString(fmt.Sprintf("🎯 **智能分析指标**:\n"))
		builder.WriteString(fmt.Sprintf("- 分析层级数: %d\n", analysisResponse.AnalysisMetrics.AnalyzedLevels))
		builder.WriteString(fmt.Sprintf("- 最优区分度: %.2f\n", analysisResponse.AnalysisMetrics.OptimalDiscrimination))
		builder.WriteString(fmt.Sprintf("- 处理时间: %dms\n", analysisResponse.AnalysisMetrics.ProcessingTimeMs))
		if analysisResponse.AnalysisMetrics.StopReason != "" {
			builder.WriteString(fmt.Sprintf("- 停止原因: %s\n", analysisResponse.AnalysisMetrics.StopReason))
		}
		builder.WriteString("\n")
	}

	// 添加数据质量信息
	if analysisResponse.DataQualityReport != nil {
		builder.WriteString(fmt.Sprintf("📊 **数据质量评估**: %.1f分", analysisResponse.DataQualityReport.QualityScore))
		if analysisResponse.DataQualityReport.QualityScore >= 90 {
			builder.WriteString(" (优秀)\n")
		} else if analysisResponse.DataQualityReport.QualityScore >= 70 {
			builder.WriteString(" (良好)\n")
		} else {
			builder.WriteString(" (有待改善)\n")
		}
		builder.WriteString("\n")
	}

	// 分隔线
	builder.WriteString("=" + strings.Repeat("=", 60) + "\n\n")

	// 添加匿名化数据
	builder.WriteString(aiDataText)

	return builder.String()
}

// 辅助方法

// groupDataByDimensions 按维度组合对数据进行分组聚合
func (ca *ContributionAnalyzer) groupDataByDimensions(data []map[string]interface{}, dimensions []string, targetMetric string) map[string]float64 {
	groups := make(map[string]float64)

	for _, row := range data {
		// 构建维度组合的键
		key := ca.buildDimensionKey(row, dimensions)

		// 获取目标指标值
		value := ca.extractFloatValue(row[targetMetric])

		// 累加到对应的组
		groups[key] += value
	}

	return groups
}

// buildDimensionKey 构建维度组合的键
func (ca *ContributionAnalyzer) buildDimensionKey(row map[string]interface{}, dimensions []string) string {
	var keyParts []string
	for _, dim := range dimensions {
		value := fmt.Sprintf("%v", row[dim])
		keyParts = append(keyParts, fmt.Sprintf("%s:%s", dim, value))
	}
	return strings.Join(keyParts, "|")
}

// parseDimensionKey 解析维度键回到维度值映射
func (ca *ContributionAnalyzer) parseDimensionKey(key string, dimensions []string) map[string]interface{} {
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

// getAllUniqueKeys 获取所有唯一的维度组合键
func (ca *ContributionAnalyzer) getAllUniqueKeys(groups1, groups2 map[string]float64) []string {
	keySet := make(map[string]bool)

	for key := range groups1 {
		keySet[key] = true
	}
	for key := range groups2 {
		keySet[key] = true
	}

	var keys []string
	for key := range keySet {
		keys = append(keys, key)
	}

	return keys
}

// extractFloatValue 从interface{}中提取float64值
func (ca *ContributionAnalyzer) extractFloatValue(value interface{}) float64 {
	if value == nil {
		return 0.0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		var result float64
		if n, err := fmt.Sscanf(v, "%f", &result); err == nil && n == 1 {
			return result
		}
		return 0.0
	default:
		return 0.0
	}
}

// calculateRelativeImportance 计算相对重要性（基于贡献度绝对值的排名百分位）
func (ca *ContributionAnalyzer) calculateRelativeImportance(contributions []anonymization_lite.ContributionItem, contributionPercents []float64) {
	if len(contributions) == 0 || len(contributionPercents) == 0 {
		return
	}

	// 对贡献度绝对值进行排序，获取排名
	type indexedContribution struct {
		index           int
		absContribution float64
	}

	var indexed []indexedContribution
	for i, absContrib := range contributionPercents {
		indexed = append(indexed, indexedContribution{
			index:           i,
			absContribution: absContrib,
		})
	}

	// 按贡献度绝对值降序排序
	sort.Slice(indexed, func(i, j int) bool {
		return indexed[i].absContribution > indexed[j].absContribution
	})

	// 计算每个项目的相对重要性（百分位排名）
	totalCount := len(indexed)
	for rank, item := range indexed {
		// 百分位排名：排名越靠前，重要性越高
		// 第1名 = 100分，最后一名 = 0分
		relativeImportance := float64(totalCount-rank) / float64(totalCount) * 100.0
		contributions[item.index].RelativeImportance = relativeImportance
	}
}
