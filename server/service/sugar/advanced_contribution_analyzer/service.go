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
	log.Printf("NewAdvancedContributionService: 开始创建增强版贡献度分析服务")

	if config == nil {
		log.Printf("NewAdvancedContributionService: 输入配置为nil，使用默认配置")
		config = DefaultAnalysisConfig()
		if config == nil {
			log.Printf("NewAdvancedContributionService: 默认配置也为nil，创建失败")
			return nil
		}
	}
	log.Printf("NewAdvancedContributionService: 配置验证通过 - DiscriminationThreshold=%.2f, MaxDrillDownLevels=%d",
		config.DiscriminationThreshold, config.MaxDrillDownLevels)

	log.Printf("NewAdvancedContributionService: 正在创建分析器...")
	analyzer := NewAdvancedContributionAnalyzer(config)
	if analyzer == nil {
		log.Printf("NewAdvancedContributionService: 创建分析器失败")
		return nil
	}
	log.Printf("NewAdvancedContributionService: 分析器创建成功")

	log.Printf("NewAdvancedContributionService: 正在创建数据优化器...")
	dataOptimizer := NewDataOptimizer(config)
	if dataOptimizer == nil {
		log.Printf("NewAdvancedContributionService: 创建数据优化器失败")
		return nil
	}
	log.Printf("NewAdvancedContributionService: 数据优化器创建成功")

	service := &AdvancedContributionService{
		analyzer:      analyzer,
		dataOptimizer: dataOptimizer,
		config:        config,
	}

	log.Printf("NewAdvancedContributionService: 服务创建成功，地址=%p", service)
	return service
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

	log.Printf("增强版贡献度分析完成: 分析层级=%d, 最优层级索引=%d, 处理时间=%dms",
		metrics.AnalyzedLevels, drillDownResult.OptimalLevel, metrics.ProcessingTimeMs)

	return response, nil
}

// generateBusinessInsights 生成业务洞察（简化版，移除自然语言描述）
func (acs *AdvancedContributionService) generateBusinessInsights(result *DrillDownResult, qualityReport *DataQualityReport) []string {
	// 根据需求，移除自然语言描述，只保留技术指标
	// 业务洞察将通过排序后的数据列表体现，不需要额外的文本描述
	return []string{}
}

// generateEnhancedSummary 生成增强摘要（简化版，移除自然语言描述）
func (acs *AdvancedContributionService) generateEnhancedSummary(result *DrillDownResult, metrics *AnalysisMetrics, qualityReport *DataQualityReport) string {
	// 根据需求，移除自然语言描述，只保留技术指标摘要
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("本次分析共检查了%d个层级，", metrics.AnalyzedLevels))

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
	if acs == nil {
		return nil
	}
	return acs.config
}
