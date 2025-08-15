package anonymization

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	sugarRes "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/sugar"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// AnonymizationService 是一个无状态的服务，负责处理数据匿名化的所有流程。
// 它通过依赖注入接收 SugarFormulaQueryService，并集成了高级匿名化功能。
type AnonymizationService struct {
	SugarQuerySvc      *sugar.SugarFormulaQueryService // 用于调用 ExecuteGetFormula 的服务实例
	AdvancedAnonymizer *AdvancedAnonymizer             // 高级匿名化器
	Config             *AdvancedAnonymizationConfig    // 匿名化配置
}

// NewAnonymizationService 是 AnonymizationService 的构造函数。
func NewAnonymizationService(querySvc *sugar.SugarFormulaQueryService) *AnonymizationService {
	config := DefaultAdvancedConfig()
	return &AnonymizationService{
		SugarQuerySvc:      querySvc,
		AdvancedAnonymizer: NewAdvancedAnonymizer(config),
		Config:             config,
	}
}

// NewAnonymizationServiceWithConfig 使用自定义配置创建匿名化服务
func NewAnonymizationServiceWithConfig(querySvc *sugar.SugarFormulaQueryService, config *AdvancedAnonymizationConfig) *AnonymizationService {
	if config == nil {
		config = DefaultAdvancedConfig()
	}
	return &AnonymizationService{
		SugarQuerySvc:      querySvc,
		AdvancedAnonymizer: NewAdvancedAnonymizer(config),
		Config:             config,
	}
}

// ProcessAndAnonymize 执行数据获取、预处理、计算和匿名化的完整流程（使用基础匿名化）
func (s *AnonymizationService) ProcessAndAnonymize(ctx context.Context, req *AIAnalysisRequest, userId string) (*AnonymizationSession, error) {
	// 1. 参数校验
	if err := req.Validate(); err != nil {
		return nil, err
	}

	global.GVA_LOG.Info("开始处理基础匿名化请求",
		zap.String("modelName", req.ModelName),
		zap.String("targetMetric", req.TargetMetric),
		zap.Strings("groupByDimensions", req.GroupByDimensions),
		zap.String("userId", userId))

	// 2. 并发获取数据
	currentData, baseData, err := s.fetchDataConcurrently(ctx, req, userId)
	if err != nil {
		return nil, NewProcessingError("并发获取数据失败", err)
	}

	// 3. 数据预处理与计算
	contributions, err := s.calculateContributions(currentData, baseData, req)
	if err != nil {
		return nil, NewProcessingError("计算贡献度失败", err)
	}

	// 4. 执行匿名化并创建会话
	session, err := s.createAnonymizedSession(contributions, req)
	if err != nil {
		return nil, NewAnonymizationError("创建匿名化会话失败", err)
	}

	global.GVA_LOG.Info("基础匿名化处理完成",
		zap.Int("contributionCount", len(contributions)),
		zap.Int("aiDataCount", len(session.AIReadyData)),
		zap.String("userId", userId))

	return session, nil
}

// ProcessAndAnonymizeAdvanced 执行高级匿名化处理
func (s *AnonymizationService) ProcessAndAnonymizeAdvanced(ctx context.Context, req *AIAnalysisRequest, userId string) (*AdvancedAnonymizationSession, error) {
	// 1. 参数校验
	if err := req.Validate(); err != nil {
		return nil, err
	}

	global.GVA_LOG.Info("开始处理高级匿名化请求",
		zap.String("modelName", req.ModelName),
		zap.String("targetMetric", req.TargetMetric),
		zap.Strings("groupByDimensions", req.GroupByDimensions),
		zap.String("userId", userId),
		zap.Float64("epsilon", s.Config.Epsilon),
		zap.Int("kAnonymity", s.Config.KAnonymity))

	// 2. 并发获取数据
	currentData, baseData, err := s.fetchDataConcurrently(ctx, req, userId)
	if err != nil {
		return nil, NewProcessingError("并发获取数据失败", err)
	}

	// 3. 数据预处理与计算
	contributions, err := s.calculateContributions(currentData, baseData, req)
	if err != nil {
		return nil, NewProcessingError("计算贡献度失败", err)
	}

	// 4. 执行高级匿名化并创建会话
	session, err := s.AdvancedAnonymizer.CreateAdvancedAnonymizationSession(contributions, req)
	if err != nil {
		return nil, NewAnonymizationError("创建高级匿名化会话失败", err)
	}

	global.GVA_LOG.Info("高级匿名化处理完成",
		zap.Int("contributionCount", len(contributions)),
		zap.Int("aiDataCount", len(session.AIReadyData)),
		zap.Float64("privacyScore", session.privacyMetrics.PrivacyScore),
		zap.Float64("dataUtility", session.qualityMetrics.DataUtility),
		zap.String("userId", userId))

	return session, nil
}

// ProcessWithCustomConfig 使用自定义配置处理匿名化
func (s *AnonymizationService) ProcessWithCustomConfig(ctx context.Context, req *AIAnalysisRequest, config *AdvancedAnonymizationConfig, userId string) (*AdvancedAnonymizationSession, error) {
	// 临时使用自定义配置
	originalConfig := s.Config
	originalAnonymizer := s.AdvancedAnonymizer

	defer func() {
		s.Config = originalConfig
		s.AdvancedAnonymizer = originalAnonymizer
	}()

	// 应用自定义配置
	s.Config = config
	s.AdvancedAnonymizer = NewAdvancedAnonymizer(config)

	return s.ProcessAndAnonymizeAdvanced(ctx, req, userId)
}

// fetchDataConcurrently 并发获取本期和基期数据
func (s *AnonymizationService) fetchDataConcurrently(ctx context.Context, req *AIAnalysisRequest, userId string) (*sugarRes.SugarFormulaGetResponse, *sugarRes.SugarFormulaGetResponse, error) {
	// 构建返回列：目标指标 + 分组维度
	returnColumns := append([]string{req.TargetMetric}, req.GroupByDimensions...)

	// 使用 errgroup 进行并发处理
	g, gCtx := errgroup.WithContext(ctx)

	var currentData, baseData *sugarRes.SugarFormulaGetResponse
	var currentErr, baseErr error

	// 并发获取本期数据
	g.Go(func() error {
		currentReq := &sugarReq.SugarFormulaGetRequest{
			ModelName:     req.ModelName,
			ReturnColumns: returnColumns,
			Filters:       req.CurrentPeriodFilters,
		}
		currentData, currentErr = s.SugarQuerySvc.ExecuteGetFormula(gCtx, currentReq, userId)
		if currentErr != nil {
			return fmt.Errorf("获取本期数据失败: %w", currentErr)
		}
		if currentData.Error != "" {
			return fmt.Errorf("本期数据查询错误: %s", currentData.Error)
		}
		return nil
	})

	// 并发获取基期数据
	g.Go(func() error {
		baseReq := &sugarReq.SugarFormulaGetRequest{
			ModelName:     req.ModelName,
			ReturnColumns: returnColumns,
			Filters:       req.BasePeriodFilters,
		}
		baseData, baseErr = s.SugarQuerySvc.ExecuteGetFormula(gCtx, baseReq, userId)
		if baseErr != nil {
			return fmt.Errorf("获取基期数据失败: %w", baseErr)
		}
		if baseData.Error != "" {
			return fmt.Errorf("基期数据查询错误: %s", baseData.Error)
		}
		return nil
	})

	// 等待所有goroutine完成
	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	global.GVA_LOG.Info("数据获取完成",
		zap.Int("currentDataCount", len(currentData.Results)),
		zap.Int("baseDataCount", len(baseData.Results)))

	return currentData, baseData, nil
}

// calculateContributions 计算贡献度分析
func (s *AnonymizationService) calculateContributions(currentData, baseData *sugarRes.SugarFormulaGetResponse, req *AIAnalysisRequest) ([]ContributionItem, error) {
	// 将数据按维度组合进行分组
	currentGroups := s.groupDataByDimensions(currentData.Results, req.GroupByDimensions, req.TargetMetric)
	baseGroups := s.groupDataByDimensions(baseData.Results, req.GroupByDimensions, req.TargetMetric)

	// 计算每个维度组合的贡献度
	var contributions []ContributionItem
	var totalChange float64

	// 获取所有唯一的维度组合
	allKeys := s.getAllUniqueKeys(currentGroups, baseGroups)

	// 第一轮：计算变化值和总变化
	for _, key := range allKeys {
		currentValue := currentGroups[key]
		baseValue := baseGroups[key]
		changeValue := currentValue - baseValue

		totalChange += changeValue

		// 解析维度值
		dimensionValues := s.parseDimensionKey(key, req.GroupByDimensions)

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
			contributions[i].ContributionPercent = (contributions[i].ChangeValue / math.Abs(totalChange)) * 100
		} else {
			contributions[i].ContributionPercent = 0
		}

		// 判断是否为正向驱动因子
		// 简化逻辑：如果变化值与总变化同向，则为正向驱动
		contributions[i].IsPositiveDriver = (contributions[i].ChangeValue * totalChange) >= 0
	}

	// 按贡献度绝对值排序
	sort.Slice(contributions, func(i, j int) bool {
		return math.Abs(contributions[i].ContributionPercent) > math.Abs(contributions[j].ContributionPercent)
	})

	global.GVA_LOG.Info("贡献度计算完成",
		zap.Float64("totalChange", totalChange),
		zap.Int("contributionCount", len(contributions)))

	return contributions, nil
}

// groupDataByDimensions 按维度组合对数据进行分组聚合
func (s *AnonymizationService) groupDataByDimensions(data []map[string]interface{}, dimensions []string, targetMetric string) map[string]float64 {
	groups := make(map[string]float64)

	for _, row := range data {
		// 构建维度组合的键
		key := s.buildDimensionKey(row, dimensions)

		// 获取目标指标值
		value := s.extractFloatValue(row[targetMetric])

		// 累加到对应的组
		groups[key] += value
	}

	return groups
}

// buildDimensionKey 构建维度组合的键
func (s *AnonymizationService) buildDimensionKey(row map[string]interface{}, dimensions []string) string {
	var keyParts []string
	for _, dim := range dimensions {
		value := fmt.Sprintf("%v", row[dim])
		keyParts = append(keyParts, fmt.Sprintf("%s:%s", dim, value))
	}
	return strings.Join(keyParts, "|")
}

// parseDimensionKey 解析维度键回到维度值映射
func (s *AnonymizationService) parseDimensionKey(key string, dimensions []string) map[string]interface{} {
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
func (s *AnonymizationService) getAllUniqueKeys(groups1, groups2 map[string]float64) []string {
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
func (s *AnonymizationService) extractFloatValue(value interface{}) float64 {
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
		// 尝试解析字符串为数字
		var result float64
		if n, err := fmt.Sscanf(v, "%f", &result); err == nil && n == 1 {
			return result
		}
		return 0.0
	default:
		return 0.0
	}
}
