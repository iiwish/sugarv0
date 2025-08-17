package advanced_contribution_analyzer

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// AdvancedContributionService 增强版贡献度分析服务
type AdvancedContributionService struct {
	analyzer      *AdvancedContributionAnalyzer
	dataOptimizer *DataOptimizer
	config        *AnalysisConfig
}

// NewAdvancedContributionService 创建增强版贡献度分析服务
func NewAdvancedContributionService(config *AnalysisConfig) *AdvancedContributionService {
	if config == nil {
		config = DefaultAnalysisConfig()
	}

	return &AdvancedContributionService{
		analyzer:      NewAdvancedContributionAnalyzer(config),
		dataOptimizer: NewDataOptimizer(config),
		config:        config,
	}
}

// AnalysisRequest 分析请求
type AnalysisRequest struct {
	ModelName            string                  `json:"model_name"`
	Metric               string                  `json:"metric"`
	Dimensions           []string                `json:"dimensions"`
	CurrentPeriodFilters map[string]interface{}  `json:"current_period_filters"`
	BasePeriodFilters    map[string]interface{}  `json:"base_period_filters"`
	IsYearEndComparison  bool                    `json:"is_year_end_comparison"`
	RawContributions     []*DimensionCombination `json:"raw_contributions"`
	TotalChange          float64                 `json:"total_change"`
}

// AnalysisResponse 分析响应
type AnalysisResponse struct {
	Success           bool                    `json:"success"`
	DrillDownResult   *DrillDownResult        `json:"drill_down_result,omitempty"`
	AnalysisMetrics   *AnalysisMetrics        `json:"analysis_metrics,omitempty"`
	DataQualityReport *DataQualityReport      `json:"data_quality_report,omitempty"`
	OptimizedPrompt   *OptimizedPromptRequest `json:"optimized_prompt,omitempty"`
	EnhancedSummary   string                  `json:"enhanced_summary"`
	BusinessInsights  []string                `json:"business_insights"`
	ErrorMessage      string                  `json:"error_message,omitempty"`
	Warnings          []string                `json:"warnings,omitempty"`
}

// PerformAdvancedAnalysis 执行增强版分析
func (acs *AdvancedContributionService) PerformAdvancedAnalysis(ctx context.Context, request *AnalysisRequest) (*AnalysisResponse, error) {
	log.Printf("开始执行增强版贡献度分析: 模型=%s, 指标=%s, 维度=%v",
		request.ModelName, request.Metric, request.Dimensions)

	response := &AnalysisResponse{
		Success:  false,
		Warnings: make([]string, 0),
	}

	// 1. 生成优化的数据获取提示词
	optimizedPrompt := acs.dataOptimizer.GenerateOptimizedPrompt(
		request.ModelName,
		request.Dimensions,
		request.Metric,
		request.CurrentPeriodFilters,
		request.BasePeriodFilters,
		request.IsYearEndComparison,
	)
	response.OptimizedPrompt = optimizedPrompt

	// 2. 构建贡献度数据
	contributionData := &ContributionData{
		DimensionCombinations: request.RawContributions,
		TotalChange:           request.TotalChange,
		AvailableDimensions:   request.Dimensions,
	}

	// 3. 数据质量分析
	qualityReport := acs.dataOptimizer.AnalyzeDataQuality(contributionData)
	response.DataQualityReport = qualityReport

	// 检查数据质量是否足够进行分析
	if qualityReport.QualityScore < 60 {
		response.Warnings = append(response.Warnings,
			fmt.Sprintf("数据质量得分较低(%.1f)，分析结果可能不够准确", qualityReport.QualityScore))
	}

	// 4. 执行智能下钻分析
	drillDownResult, metrics, err := acs.analyzer.AnalyzeWithIntelligentDrillDown(contributionData)
	if err != nil {
		response.ErrorMessage = fmt.Sprintf("智能下钻分析失败: %v", err)
		return response, err
	}

	response.DrillDownResult = drillDownResult
	response.AnalysisMetrics = metrics

	// 5. 生成增强的业务洞察
	businessInsights := acs.generateBusinessInsights(drillDownResult, qualityReport)
	response.BusinessInsights = businessInsights

	// 6. 生成增强摘要
	enhancedSummary := acs.generateEnhancedSummary(drillDownResult, metrics, qualityReport)
	response.EnhancedSummary = enhancedSummary

	response.Success = true

	log.Printf("增强版贡献度分析完成: 分析层级=%d, 最优层级=%d, 处理时间=%dms",
		metrics.AnalyzedLevels, drillDownResult.OptimalLevel+1, metrics.ProcessingTimeMs)

	return response, nil
}

// generateBusinessInsights 生成业务洞察
func (acs *AdvancedContributionService) generateBusinessInsights(result *DrillDownResult, qualityReport *DataQualityReport) []string {
	var insights []string

	if len(result.TopCombinations) == 0 {
		insights = append(insights, "当前数据未显示明显的贡献度差异，建议检查分析维度或时间范围")
		return insights
	}

	// 主要贡献者洞察
	topCombo := result.TopCombinations[0]
	if len(topCombo.Values) > 1 {
		// 多维度组合
		var dimensionParts []string
		for _, value := range topCombo.Values {
			dimensionParts = append(dimensionParts, value.Label)
		}
		insights = append(insights, fmt.Sprintf("最显著的变化来自%s的组合，贡献度达到%.1f%%，表明这一特定组合在业务变化中起到关键作用",
			strings.Join(dimensionParts, "与"), topCombo.Contribution))
	} else if len(topCombo.Values) == 1 {
		// 单维度
		insights = append(insights, fmt.Sprintf("%s维度中的%s表现最为突出，贡献度为%.1f%%，是推动整体变化的主要因素",
			topCombo.Values[0].Dimension, topCombo.Values[0].Label, topCombo.Contribution))
	}

	// 分析层级洞察
	if result.OptimalLevel >= 0 && result.OptimalLevel < len(result.Levels) {
		optimalLevel := result.Levels[result.OptimalLevel]
		if len(optimalLevel.Dimensions) > 1 {
			insights = append(insights, fmt.Sprintf("通过%s的组合分析能够获得最佳的业务洞察，区分度达到%.1f%%",
				strings.Join(optimalLevel.Dimensions, "与"), optimalLevel.Discrimination))
		}
	}

	// 对比分析洞察
	if len(result.TopCombinations) > 1 {
		secondCombo := result.TopCombinations[1]
		contributionGap := topCombo.Contribution - secondCombo.Contribution
		if contributionGap > 10 {
			insights = append(insights, fmt.Sprintf("领先组合的贡献度比第二位高出%.1f个百分点，显示出明显的集中性特征", contributionGap))
		} else {
			insights = append(insights, "前几位贡献者的差距较小，变化相对分散，需要关注多个重点领域")
		}
	}

	// 数据质量相关洞察
	if qualityReport.QualityScore < 80 {
		insights = append(insights, "建议优化数据获取策略以提高分析精度，当前数据可能存在完整性或一致性问题")
	}

	// 业务建议
	if topCombo.Contribution > 50 {
		insights = append(insights, "单一组合贡献度超过50%，建议重点关注该领域的风险管控和持续优化")
	} else if topCombo.Contribution < 20 {
		insights = append(insights, "变化较为分散，建议采用多元化的管理策略，关注各个维度的协调发展")
	}

	return insights
}

// generateEnhancedSummary 生成增强摘要
func (acs *AdvancedContributionService) generateEnhancedSummary(result *DrillDownResult, metrics *AnalysisMetrics, qualityReport *DataQualityReport) string {
	var summary strings.Builder

	// 基础分析结果
	summary.WriteString(result.AnalysisSummary)

	// 添加分析深度信息
	summary.WriteString(fmt.Sprintf("。本次分析共检查了%d个层级，", metrics.AnalyzedLevels))

	if metrics.AnalyzedLevels > 1 {
		summary.WriteString(fmt.Sprintf("通过智能下钻算法确定第%d层级为最优分析深度", result.OptimalLevel+1))
	} else {
		summary.WriteString("在单层级分析中获得结果")
	}

	// 添加数据质量信息
	if qualityReport.QualityScore >= 90 {
		summary.WriteString("，数据质量优秀")
	} else if qualityReport.QualityScore >= 70 {
		summary.WriteString("，数据质量良好")
	} else {
		summary.WriteString("，数据质量有待改善")
	}

	// 添加停止原因
	if metrics.StopReason != "" {
		summary.WriteString(fmt.Sprintf("。分析停止原因：%s", metrics.StopReason))
	}

	return summary.String()
}

// GetOptimizedPromptForDataFetch 获取优化的数据获取提示词
func (acs *AdvancedContributionService) GetOptimizedPromptForDataFetch(
	modelName string,
	dimensions []string,
	metric string,
	currentPeriodFilters, basePeriodFilters map[string]interface{},
	isYearEndComparison bool,
) *OptimizedPromptRequest {
	return acs.dataOptimizer.GenerateOptimizedPrompt(
		modelName, dimensions, metric,
		currentPeriodFilters, basePeriodFilters,
		isYearEndComparison,
	)
}

// ValidateAnalysisRequest 验证分析请求
func (acs *AdvancedContributionService) ValidateAnalysisRequest(request *AnalysisRequest) error {
	if request.ModelName == "" {
		return fmt.Errorf("模型名称不能为空")
	}

	if request.Metric == "" {
		return fmt.Errorf("分析指标不能为空")
	}

	if len(request.Dimensions) == 0 {
		return fmt.Errorf("分析维度不能为空")
	}

	if len(request.RawContributions) == 0 {
		return fmt.Errorf("贡献度数据不能为空")
	}

	if request.TotalChange == 0 {
		return fmt.Errorf("总变化值不能为0")
	}

	return nil
}

// GetDimensionPriorityRecommendation 获取维度优先级建议
func (acs *AdvancedContributionService) GetDimensionPriorityRecommendation(contributionData *ContributionData) ([]string, error) {
	return acs.analyzer.GetDimensionPriorityOrder(contributionData)
}

// UpdateAnalysisConfig 更新分析配置
func (acs *AdvancedContributionService) UpdateAnalysisConfig(config *AnalysisConfig) {
	acs.config = config
	acs.analyzer = NewAdvancedContributionAnalyzer(config)
	acs.dataOptimizer = NewDataOptimizer(config)
}

// GetCurrentConfig 获取当前配置
func (acs *AdvancedContributionService) GetCurrentConfig() *AnalysisConfig {
	return acs.config
}
