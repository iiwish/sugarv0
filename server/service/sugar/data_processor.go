package sugar

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	sugarRes "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

// DataProcessor 数据处理器 - 负责数据获取、验证和范围探索
type DataProcessor struct {
	formulaQueryService SugarFormulaQueryService
}

// NewDataProcessor 创建数据处理器
func NewDataProcessor() *DataProcessor {
	return &DataProcessor{
		formulaQueryService: SugarFormulaQueryService{},
	}
}

// GetAgentAndLLMConfig 获取Agent信息和LLM配置
func (dp *DataProcessor) GetAgentAndLLMConfig(ctx context.Context, agentName, userId string) (*sugar.SugarAgents, *system.LLMConfig, error) {
	// 获取Agent信息
	agent, err := dp.getAgentByName(ctx, agentName, userId)
	if err != nil {
		return nil, nil, fmt.Errorf("获取Agent信息失败: %w", err)
	}

	// 获取LLM配置
	var llmConfig *system.LLMConfig
	llmService := system.SysLLMService{}

	if agent.EndpointConfig != "" {
		global.GVA_LOG.Debug("解析Agent的LLM配置", zap.String("endpointConfig", agent.EndpointConfig))
		llmConfig, err = llmService.ParseLLMConfigFromJSON(agent.EndpointConfig)
		if err != nil {
			global.GVA_LOG.Warn("解析Agent LLM配置失败，使用默认LLM配置", zap.Error(err))
			llmConfig = llmService.GetDefaultLLMConfig()
		} else {
			global.GVA_LOG.Info("成功解析Agent LLM配置", zap.String("model", llmConfig.ModelName))
		}
	} else {
		global.GVA_LOG.Info("Agent未配置LLM，使用默认LLM配置")
		llmConfig = llmService.GetDefaultLLMConfig()
	}

	return agent, llmConfig, nil
}

// ValidateDataAvailability 验证数据可用性
func (dp *DataProcessor) ValidateDataAvailability(ctx context.Context, modelName string, groupByDimensions []string, currentPeriodFilters, basePeriodFilters map[string]interface{}, userId string) (*DataValidationResult, error) {
	global.GVA_LOG.Info("开始验证数据可用性",
		zap.String("modelName", modelName),
		zap.Strings("groupByDimensions", groupByDimensions))

	result := &DataValidationResult{
		IsDataAvailable:   false,
		ValidationMessage: "",
		RecordCount:       0,
		MissingDimensions: make([]string, 0),
	}

	// 检查是否为年初年末对比类型的表
	isYearEndComparison := dp.isYearEndComparisonModel(modelName, "")

	if isYearEndComparison {
		return dp.validateYearEndComparisonData(ctx, modelName, groupByDimensions, currentPeriodFilters, userId, result)
	}

	return dp.validateTimeBasedData(ctx, modelName, groupByDimensions, currentPeriodFilters, basePeriodFilters, userId, result)
}

// ExploreDataScope 执行数据范围探索
func (dp *DataProcessor) ExploreDataScope(ctx context.Context, modelName string, exploreDimensions []string, sampleFilters map[string]interface{}, userId string) (*DataScopeInfo, error) {
	global.GVA_LOG.Info("开始执行数据范围探索",
		zap.String("modelName", modelName),
		zap.Strings("exploreDimensions", exploreDimensions))

	// 构建探索查询
	returnColumns := exploreDimensions

	exploreReq := &sugarReq.SugarFormulaGetRequest{
		ModelName:     modelName,
		ReturnColumns: returnColumns,
		Filters:       sampleFilters,
	}

	// 执行查询
	exploreData, err := dp.formulaQueryService.ExecuteGetFormula(ctx, exploreReq, userId)
	if err != nil {
		return nil, fmt.Errorf("执行探索查询失败: %w", err)
	}
	if exploreData.Error != "" {
		return nil, fmt.Errorf("探索查询错误: %s", exploreData.Error)
	}

	// 分析数据范围
	scopeInfo := &DataScopeInfo{
		TotalRecords:       len(exploreData.Results),
		DimensionCoverage:  make(map[string][]string),
		SampleData:         make([]map[string]interface{}, 0),
		DataQualityInfo:    make(map[string]interface{}),
		RecommendedFilters: make(map[string]interface{}),
	}

	// 统计各维度的唯一值
	dimensionValues := make(map[string]map[string]bool)
	for _, dim := range exploreDimensions {
		dimensionValues[dim] = make(map[string]bool)
	}

	// 遍历数据，统计维度值
	sampleSize := 10
	for i, row := range exploreData.Results {
		// 保存样本数据
		if i < sampleSize {
			scopeInfo.SampleData = append(scopeInfo.SampleData, row)
		}

		// 统计维度值
		for _, dim := range exploreDimensions {
			if value, exists := row[dim]; exists {
				valueStr := fmt.Sprintf("%v", value)
				if valueStr != "" && valueStr != "<nil>" {
					dimensionValues[dim][valueStr] = true
				}
			}
		}
	}

	// 转换为切片格式
	for dim, valueMap := range dimensionValues {
		var values []string
		for value := range valueMap {
			values = append(values, value)
		}
		scopeInfo.DimensionCoverage[dim] = values
	}

	// 生成数据质量信息
	scopeInfo.DataQualityInfo["completeness"] = dp.calculateDataCompleteness(exploreData.Results, exploreDimensions)
	scopeInfo.DataQualityInfo["distinct_combinations"] = dp.calculateDistinctCombinations(exploreData.Results, exploreDimensions)

	// 生成推荐筛选条件
	scopeInfo.RecommendedFilters = dp.generateRecommendedFilters(scopeInfo.DimensionCoverage)

	return scopeInfo, nil
}

// FormatDataScopeResult 格式化数据范围探索结果
func (dp *DataProcessor) FormatDataScopeResult(scopeInfo *DataScopeInfo) string {
	var builder strings.Builder

	builder.WriteString("📊 数据范围探索结果\n\n")
	builder.WriteString(fmt.Sprintf("📈 数据总览：共找到 %d 条记录\n\n", scopeInfo.TotalRecords))

	// 维度覆盖情况
	builder.WriteString("🔍 维度数据覆盖情况：\n")
	for dim, values := range scopeInfo.DimensionCoverage {
		builder.WriteString(fmt.Sprintf("  • %s: %d个不同值", dim, len(values)))
		if len(values) <= 10 {
			builder.WriteString(fmt.Sprintf(" [%s]", strings.Join(values, ", ")))
		} else {
			builder.WriteString(fmt.Sprintf(" [%s, ...等%d个]", strings.Join(values[:5], ", "), len(values)-5))
		}
		builder.WriteString("\n")
	}

	// 数据质量信息
	if completeness, ok := scopeInfo.DataQualityInfo["completeness"].(map[string]float64); ok {
		builder.WriteString("\n📋 数据完整性：\n")
		for dim, ratio := range completeness {
			builder.WriteString(fmt.Sprintf("  • %s: %.1f%%\n", dim, ratio*100))
		}
	}

	// 推荐筛选条件
	if len(scopeInfo.RecommendedFilters) > 0 {
		builder.WriteString("\n💡 建议的筛选条件：\n")
		for dim, filter := range scopeInfo.RecommendedFilters {
			builder.WriteString(fmt.Sprintf("  • %s: %v\n", dim, filter))
		}
	}

	// 注意事项
	builder.WriteString("\n⚠️  使用建议：\n")
	builder.WriteString("  • 请根据以上数据范围调整您的分析需求\n")
	builder.WriteString("  • 如果某些您关心的维度值不在上述列表中，可能需要调整时间范围或其他筛选条件\n")
	builder.WriteString("  • 建议使用 anonymized_data_analyzer 工具进行深入分析\n")

	return builder.String()
}

// FetchDataConcurrently 并发获取本期和基期数据
func (dp *DataProcessor) FetchDataConcurrently(ctx context.Context, modelName, targetMetric string, currentPeriodFilters, basePeriodFilters map[string]interface{}, groupByDimensions []string, userId string) (*sugarRes.SugarFormulaGetResponse, *sugarRes.SugarFormulaGetResponse, error) {
	// 检查是否为年初年末对比类型的表
	isYearEndComparison := dp.isYearEndComparisonModel(modelName, targetMetric)

	if isYearEndComparison {
		return dp.fetchYearEndComparisonData(ctx, modelName, targetMetric, currentPeriodFilters, groupByDimensions, userId)
	}

	return dp.fetchTimeBasedData(ctx, modelName, targetMetric, currentPeriodFilters, basePeriodFilters, groupByDimensions, userId)
}

// 私有方法

// getAgentByName 根据名称获取Agent
func (dp *DataProcessor) getAgentByName(ctx context.Context, agentName, userId string) (*sugar.SugarAgents, error) {
	var agent sugar.SugarAgents

	// 获取用户所属团队
	var teamIds []string
	err := global.GVA_DB.Table("sugar_team_members").Where("user_id = ?", userId).Pluck("team_id", &teamIds).Error
	if err != nil {
		return nil, errors.New("获取用户团队信息失败")
	}
	if len(teamIds) == 0 {
		return nil, errors.New("用户未加入任何团队")
	}

	// 获取团队共享表信息
	var teamAgentIds []string
	err = global.GVA_DB.Table("sugar_agent_shares").Where("team_id in ? AND deleted_at is null", teamIds).Pluck("agent_id", &teamAgentIds).Error
	if err != nil {
		return nil, errors.New("获取用户团队Agent信息失败")
	}
	if len(teamAgentIds) == 0 {
		return nil, errors.New("用户团队没有Agent权限")
	}

	// 查找Agent
	err = global.GVA_DB.Where("name = ? AND team_id IN ?", agentName, teamIds).First(&agent).Error
	if err != nil {
		return nil, errors.New("Agent不存在或无权访问: " + agentName)
	}

	return &agent, nil
}

// isYearEndComparisonModel 判断是否为年初年末对比类型的模型
func (dp *DataProcessor) isYearEndComparisonModel(modelName, targetMetric string) bool {
	modelNameLower := strings.ToLower(modelName)
	targetMetricLower := strings.ToLower(targetMetric)

	yearEndModels := []string{"货币资金", "cash", "应收账款", "receivable", "存货", "inventory",
		"固定资产", "fixed", "无形资产", "intangible", "应付账款", "payable",
		"短期借款", "short_term", "长期借款", "long_term", "实收资本", "capital", "未分配利润", "retained"}

	yearEndMetrics := []string{"年末金额", "ending_balance", "年初金额", "beginning_balance"}

	for _, keyword := range yearEndModels {
		if strings.Contains(modelNameLower, keyword) {
			for _, metric := range yearEndMetrics {
				if strings.Contains(targetMetricLower, metric) {
					return true
				}
			}
		}
	}

	return false
}

// validateYearEndComparisonData 验证年初年末对比数据的可用性
func (dp *DataProcessor) validateYearEndComparisonData(ctx context.Context, modelName string, groupByDimensions []string, filters map[string]interface{}, userId string, result *DataValidationResult) (*DataValidationResult, error) {
	returnColumns := append([]string{"年末金额", "年初金额"}, groupByDimensions...)
	cleanedFilters := dp.filterWildcardConditions(filters)

	validateReq := &sugarReq.SugarFormulaGetRequest{
		ModelName:     modelName,
		ReturnColumns: returnColumns,
		Filters:       cleanedFilters,
	}

	validateData, err := dp.formulaQueryService.ExecuteGetFormula(ctx, validateReq, userId)
	if err != nil {
		return nil, fmt.Errorf("执行年初年末验证查询失败: %w", err)
	}
	if validateData.Error != "" {
		return nil, fmt.Errorf("年初年末验证查询错误: %s", validateData.Error)
	}

	result.RecordCount = len(validateData.Results)

	// 验证数据质量
	validRecordCount := 0
	for _, record := range validateData.Results {
		beginningBalance := dp.extractFloatValue(record["年初金额"])
		endingBalance := dp.extractFloatValue(record["年末金额"])

		if beginningBalance != 0 || endingBalance != 0 {
			validRecordCount++
		}
	}

	// 判断数据可用性
	if result.RecordCount == 0 {
		result.IsDataAvailable = false
		result.ValidationMessage = "根据您提供的筛选条件，未找到匹配的年初年末对比数据记录。建议检查筛选条件是否正确。"
	} else if validRecordCount == 0 {
		result.IsDataAvailable = false
		result.ValidationMessage = fmt.Sprintf("找到%d条记录，但年初和年末金额字段均为空。请检查数据完整性。", result.RecordCount)
	} else if validRecordCount < 3 {
		result.IsDataAvailable = false
		result.ValidationMessage = fmt.Sprintf("找到%d条记录，但只有%d条有效记录。数据量过少，无法进行可靠的贡献度分析。建议调整筛选条件。", result.RecordCount, validRecordCount)
	} else {
		result.IsDataAvailable = true
		result.ValidationMessage = fmt.Sprintf("数据验证通过：找到%d条记录，其中%d条有效记录，可以进行年初年末对比分析。", result.RecordCount, validRecordCount)
	}

	return result, nil
}

// validateTimeBasedData 验证基于时间维度的数据可用性
func (dp *DataProcessor) validateTimeBasedData(ctx context.Context, modelName string, groupByDimensions []string, currentPeriodFilters, basePeriodFilters map[string]interface{}, userId string, result *DataValidationResult) (*DataValidationResult, error) {
	returnColumns := groupByDimensions[:1]
	if len(returnColumns) == 0 {
		returnColumns = []string{"*"}
	}

	cleanedCurrentFilters := dp.filterWildcardConditions(currentPeriodFilters)
	cleanedBaseFilters := dp.filterWildcardConditions(basePeriodFilters)

	validateReq := &sugarReq.SugarFormulaGetRequest{
		ModelName:     modelName,
		ReturnColumns: returnColumns,
		Filters:       cleanedCurrentFilters,
	}

	validateData, err := dp.formulaQueryService.ExecuteGetFormula(ctx, validateReq, userId)
	if err != nil {
		return nil, fmt.Errorf("执行验证查询失败: %w", err)
	}
	if validateData.Error != "" {
		return nil, fmt.Errorf("验证查询错误: %s", validateData.Error)
	}

	result.RecordCount = len(validateData.Results)

	// 验证基期数据
	baseValidateReq := &sugarReq.SugarFormulaGetRequest{
		ModelName:     modelName,
		ReturnColumns: returnColumns,
		Filters:       cleanedBaseFilters,
	}

	baseValidateData, err := dp.formulaQueryService.ExecuteGetFormula(ctx, baseValidateReq, userId)
	baseRecordCount := 0
	if err == nil && baseValidateData.Error == "" {
		baseRecordCount = len(baseValidateData.Results)
	}

	// 判断数据可用性
	if result.RecordCount == 0 {
		result.IsDataAvailable = false
		result.ValidationMessage = "根据您提供的本期筛选条件，未找到匹配的数据记录。建议检查时间范围、地区名称等筛选条件是否正确。"
	} else if baseRecordCount == 0 {
		result.IsDataAvailable = false
		result.ValidationMessage = fmt.Sprintf("本期找到%d条记录，但基期未找到匹配的数据记录。建议检查基期的筛选条件是否正确。", result.RecordCount)
	} else if result.RecordCount < 3 || baseRecordCount < 3 {
		result.IsDataAvailable = false
		result.ValidationMessage = fmt.Sprintf("本期找到%d条记录，基期找到%d条记录。数据量过少，无法进行可靠的贡献度分析。建议扩大时间范围或调整筛选条件。", result.RecordCount, baseRecordCount)
	} else {
		result.IsDataAvailable = true
		result.ValidationMessage = fmt.Sprintf("数据验证通过：本期找到%d条记录，基期找到%d条记录，可以进行贡献度分析。", result.RecordCount, baseRecordCount)
	}

	return result, nil
}

// fetchYearEndComparisonData 获取年初年末对比数据
func (dp *DataProcessor) fetchYearEndComparisonData(ctx context.Context, modelName, targetMetric string, filters map[string]interface{}, groupByDimensions []string, userId string) (*sugarRes.SugarFormulaGetResponse, *sugarRes.SugarFormulaGetResponse, error) {
	returnColumns := append([]string{"年末金额", "年初金额"}, groupByDimensions...)
	cleanedFilters := dp.filterWildcardConditions(filters)

	req := &sugarReq.SugarFormulaGetRequest{
		ModelName:     modelName,
		ReturnColumns: returnColumns,
		Filters:       cleanedFilters,
	}

	fullData, err := dp.formulaQueryService.ExecuteGetFormula(ctx, req, userId)
	if err != nil {
		return nil, nil, fmt.Errorf("获取年初年末数据失败: %w", err)
	}
	if fullData.Error != "" {
		return nil, nil, fmt.Errorf("年初年末数据查询错误: %s", fullData.Error)
	}

	// 构造当前期数据（年末金额）和基期数据（年初金额）
	currentData := &sugarRes.SugarFormulaGetResponse{
		Results: make([]map[string]interface{}, len(fullData.Results)),
		Error:   "",
	}

	baseData := &sugarRes.SugarFormulaGetResponse{
		Results: make([]map[string]interface{}, len(fullData.Results)),
		Error:   "",
	}

	// 转换数据格式
	for i, row := range fullData.Results {
		// 当前期数据：使用年末金额作为目标指标值
		currentRow := make(map[string]interface{})
		for key, value := range row {
			if key == "年末金额" {
				currentRow[targetMetric] = value
			} else if key != "年初金额" {
				currentRow[key] = value
			}
		}
		currentData.Results[i] = currentRow

		// 基期数据：使用年初金额作为目标指标值
		baseRow := make(map[string]interface{})
		for key, value := range row {
			if key == "年初金额" {
				baseRow[targetMetric] = value
			} else if key != "年末金额" {
				baseRow[key] = value
			}
		}
		baseData.Results[i] = baseRow
	}

	return currentData, baseData, nil
}

// fetchTimeBasedData 获取基于时间维度的数据
func (dp *DataProcessor) fetchTimeBasedData(ctx context.Context, modelName, targetMetric string, currentPeriodFilters, basePeriodFilters map[string]interface{}, groupByDimensions []string, userId string) (*sugarRes.SugarFormulaGetResponse, *sugarRes.SugarFormulaGetResponse, error) {
	returnColumns := append([]string{targetMetric}, groupByDimensions...)
	cleanedCurrentFilters := dp.filterWildcardConditions(currentPeriodFilters)
	cleanedBaseFilters := dp.filterWildcardConditions(basePeriodFilters)

	type dataResult struct {
		data *sugarRes.SugarFormulaGetResponse
		err  error
	}

	currentCh := make(chan dataResult, 1)
	baseCh := make(chan dataResult, 1)

	// 并发获取本期数据
	go func() {
		currentReq := &sugarReq.SugarFormulaGetRequest{
			ModelName:     modelName,
			ReturnColumns: returnColumns,
			Filters:       cleanedCurrentFilters,
		}
		currentData, err := dp.formulaQueryService.ExecuteGetFormula(ctx, currentReq, userId)
		if err != nil {
			currentCh <- dataResult{nil, fmt.Errorf("获取本期数据失败: %w", err)}
			return
		}
		if currentData.Error != "" {
			currentCh <- dataResult{nil, fmt.Errorf("本期数据查询错误: %s", currentData.Error)}
			return
		}
		currentCh <- dataResult{currentData, nil}
	}()

	// 并发获取基期数据
	go func() {
		baseReq := &sugarReq.SugarFormulaGetRequest{
			ModelName:     modelName,
			ReturnColumns: returnColumns,
			Filters:       cleanedBaseFilters,
		}
		baseData, err := dp.formulaQueryService.ExecuteGetFormula(ctx, baseReq, userId)
		if err != nil {
			baseCh <- dataResult{nil, fmt.Errorf("获取基期数据失败: %w", err)}
			return
		}
		if baseData.Error != "" {
			baseCh <- dataResult{nil, fmt.Errorf("基期数据查询错误: %s", baseData.Error)}
			return
		}
		baseCh <- dataResult{baseData, nil}
	}()

	// 等待两个goroutine完成
	currentResult := <-currentCh
	baseResult := <-baseCh

	if currentResult.err != nil {
		return nil, nil, currentResult.err
	}
	if baseResult.err != nil {
		return nil, nil, baseResult.err
	}

	return currentResult.data, baseResult.data, nil
}

// filterWildcardConditions 过滤掉通配符条件
func (dp *DataProcessor) filterWildcardConditions(filters map[string]interface{}) map[string]interface{} {
	if filters == nil {
		return make(map[string]interface{})
	}

	cleanedFilters := make(map[string]interface{})
	wildcardPatterns := []string{"*", "%", "all", "全部", "所有"}

	for key, value := range filters {
		valueStr := fmt.Sprintf("%v", value)
		isWildcard := false

		for _, pattern := range wildcardPatterns {
			if valueStr == pattern {
				isWildcard = true
				break
			}
		}

		if !isWildcard && valueStr != "" && valueStr != "<nil>" {
			cleanedFilters[key] = value
		}
	}

	return cleanedFilters
}

// extractFloatValue 从interface{}中提取float64值
func (dp *DataProcessor) extractFloatValue(value interface{}) float64 {
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

// calculateDataCompleteness 计算数据完整性
func (dp *DataProcessor) calculateDataCompleteness(data []map[string]interface{}, dimensions []string) map[string]float64 {
	completeness := make(map[string]float64)
	total := len(data)

	if total == 0 {
		return completeness
	}

	for _, dim := range dimensions {
		nonNullCount := 0
		for _, row := range data {
			if value, exists := row[dim]; exists {
				valueStr := fmt.Sprintf("%v", value)
				if valueStr != "" && valueStr != "<nil>" {
					nonNullCount++
				}
			}
		}
		completeness[dim] = float64(nonNullCount) / float64(total)
	}

	return completeness
}

// calculateDistinctCombinations 计算不同维度组合的数量
func (dp *DataProcessor) calculateDistinctCombinations(data []map[string]interface{}, dimensions []string) int {
	combinations := make(map[string]bool)

	for _, row := range data {
		var keyParts []string
		for _, dim := range dimensions {
			value := fmt.Sprintf("%v", row[dim])
			keyParts = append(keyParts, value)
		}
		key := strings.Join(keyParts, "|")
		combinations[key] = true
	}

	return len(combinations)
}

// generateRecommendedFilters 生成推荐的筛选条件
func (dp *DataProcessor) generateRecommendedFilters(dimensionCoverage map[string][]string) map[string]interface{} {
	recommended := make(map[string]interface{})

	for dim, values := range dimensionCoverage {
		if len(values) <= 5 {
			recommended[dim] = values
		} else {
			recommended[dim] = fmt.Sprintf("建议从以下值中选择: %s", strings.Join(values[:3], ", "))
		}
	}

	return recommended
}
