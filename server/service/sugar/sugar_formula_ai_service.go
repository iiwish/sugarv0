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
	"github.com/flipped-aurora/gin-vue-admin/server/service/sugar/advanced_contribution_analyzer"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

type SugarFormulaAiService struct {
	advancedAnalyzer *advanced_contribution_analyzer.AdvancedContributionService
}

var llmService = system.SysLLMService{}
var executionLogService = SugarExecutionLogService{}

// NewSugarFormulaAiService 创建新的AI服务实例
func NewSugarFormulaAiService() *SugarFormulaAiService {
	// 安全检查：确保全局日志已初始化
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("开始初始化SugarFormulaAiService")
	}

	// 获取默认配置
	defaultConfig := advanced_contribution_analyzer.DefaultAnalysisConfig()
	if defaultConfig == nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("获取默认分析配置失败")
		}
		return &SugarFormulaAiService{advancedAnalyzer: nil}
	}
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("默认分析配置获取成功",
			zap.Float64("discriminationThreshold", defaultConfig.DiscriminationThreshold),
			zap.Int("maxDrillDownLevels", defaultConfig.MaxDrillDownLevels))
	}

	// 创建增强版分析服务
	advancedAnalyzer := advanced_contribution_analyzer.NewAdvancedContributionService(defaultConfig)
	if advancedAnalyzer == nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("创建增强版分析服务失败")
		}
		return &SugarFormulaAiService{advancedAnalyzer: nil}
	}
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("增强版分析服务创建成功")
	}

	// 验证分析器配置
	currentConfig := advancedAnalyzer.GetCurrentConfig()
	if currentConfig == nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("增强版分析器配置为空，将在运行时回退到lite版本")
		}
		return &SugarFormulaAiService{advancedAnalyzer: nil}
	}
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("增强版分析器配置验证成功",
			zap.Float64("discriminationThreshold", currentConfig.DiscriminationThreshold),
			zap.Int("maxDrillDownLevels", currentConfig.MaxDrillDownLevels))
	}

	service := &SugarFormulaAiService{
		advancedAnalyzer: advancedAnalyzer,
	}

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("SugarFormulaAiService初始化完成，增强版分析器状态正常")
	}
	return service
}

// GetSugarFormulaAiService 获取AI服务实例（单例模式）
var sugarFormulaAiServiceInstance *SugarFormulaAiService

func GetSugarFormulaAiService() *SugarFormulaAiService {
	if sugarFormulaAiServiceInstance == nil {
		sugarFormulaAiServiceInstance = NewSugarFormulaAiService()
	}
	return sugarFormulaAiServiceInstance
}

// ResetSugarFormulaAiService 重置AI服务实例（用于调试）
func ResetSugarFormulaAiService() {
	global.GVA_LOG.Info("重置SugarFormulaAiService单例实例")
	sugarFormulaAiServiceInstance = nil
}

// ExecuteAiFetchFormula 执行 AIFETCH 公式（使用模块化架构）
func (s *SugarFormulaAiService) ExecuteAiFetchFormula(ctx context.Context, req *sugarReq.SugarFormulaAiFetchRequest, userId string) (*sugarRes.SugarFormulaAiResponse, error) {
	global.GVA_LOG.Info("开始执行AIFETCH公式",
		zap.String("agentName", req.AgentName),
		zap.String("description", req.Description),
		zap.String("userId", userId))

	// 创建AI获取处理器并执行请求
	processor := NewAiFetchProcessor(s.advancedAnalyzer)
	return processor.ProcessRequest(ctx, req, userId)
}

// ExecuteAiExplainFormula 执行 AIEXPLAIN 公式
func (s *SugarFormulaAiService) ExecuteAiExplainFormula(ctx context.Context, req *sugarReq.SugarFormulaAiExplainRangeRequest, userId string) (*sugarRes.SugarFormulaAiResponse, error) {
	global.GVA_LOG.Info("开始执行AIEXPLAIN公式",
		zap.String("description", req.Description),
		zap.String("userId", userId),
		zap.Int("dataSourceRows", len(req.DataSource)))

	// 创建执行日志
	explainReq := &sugarReq.SugarFormulaAiFetchRequest{
		AgentName:   "AiExplain",
		Description: req.Description,
		DataRange:   "",
	}
	explainAgentId := "AiExplain"
	logCtx, err := executionLogService.CreateExecutionLog(ctx, explainReq, userId, &explainAgentId)
	if err != nil {
		global.GVA_LOG.Error("创建AIEXPLAIN执行日志失败", zap.Error(err))
	}

	// 序列化数据
	dataText, err := s.serializeDataToText(req.DataSource)
	if err != nil {
		global.GVA_LOG.Error("数据序列化失败", zap.Error(err))
		if logCtx != nil {
			errorMsg := "数据序列化失败: " + err.Error()
			_ = executionLogService.FinishExecutionLog(ctx, logCtx, "", "failed", &errorMsg)
		}
		return sugarRes.NewAiErrorResponse("数据序列化失败: " + err.Error()), nil
	}

	// 获取LLM配置
	var llmConfig *system.LLMConfig
	var systemPrompt string

	agent, err := s.getAiExplainPrompt()
	if err != nil {
		global.GVA_LOG.Warn("获取Agent失败，使用默认配置", zap.Error(err))
		llmConfig = llmService.GetDefaultLLMConfig()
		systemPrompt = "你是一个专业的数据分析师，请根据用户提供的数据和需求进行分析。"
	} else {
		if agent.EndpointConfig != "" {
			llmConfig, err = llmService.ParseLLMConfigFromJSON(agent.EndpointConfig)
			if err != nil {
				global.GVA_LOG.Warn("解析Agent LLM配置失败，使用默认配置", zap.Error(err))
				llmConfig = llmService.GetDefaultLLMConfig()
			}
		} else {
			llmConfig = llmService.GetDefaultLLMConfig()
		}
		systemPrompt = s.buildSystemPrompt(agent, userId)
	}

	// 记录日志
	if logCtx != nil {
		executionLogService.RecordAISystemPrompt(ctx, logCtx, systemPrompt)
		_ = executionLogService.RecordLLMConfig(ctx, logCtx, llmConfig)
	}

	// 构建用户消息
	userMessage := fmt.Sprintf("请分析以下数据：\n\n%s\n\n分析要求：%s", dataText, req.Description)

	// 记录用户消息
	if logCtx != nil {
		executionLogService.RecordAIUserMessage(ctx, logCtx, userMessage)
	}

	// 构建消息列表
	messages := []system.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	// 调用LLM
	global.GVA_LOG.Info("开始调用LLM进行AIEXPLAIN分析", zap.String("model", llmConfig.ModelName))
	response, err := llmService.ChatSimple(ctx, *llmConfig, messages)
	if err != nil {
		global.GVA_LOG.Error("AIEXPLAIN调用失败", zap.Error(err))
		if logCtx != nil {
			errorMsg := "AI分析失败: " + err.Error()
			_ = executionLogService.FinishExecutionLog(ctx, logCtx, "", "failed", &errorMsg)
		}
		return sugarRes.NewAiErrorResponse("AI分析失败: " + err.Error()), nil
	}

	global.GVA_LOG.Info("AIEXPLAIN分析完成", zap.String("responseLength", fmt.Sprintf("%d", len(response))))

	// 记录响应
	if logCtx != nil {
		modelName := llmConfig.ModelName
		executionLogService.RecordLLMResponse(ctx, logCtx, response, &modelName, nil)
		_ = executionLogService.UpdateExecutionLogWithAIInteractionsToDatabase(ctx, logCtx)
		_ = executionLogService.FinishExecutionLog(ctx, logCtx, response, "success", nil)
	}

	return sugarRes.NewAiSuccessResponseWithText(response), nil
}

// getAgentByName 根据名称获取Agent
func (s *SugarFormulaAiService) getAgentByName(ctx context.Context, agentName, userId string) (*sugar.SugarAgents, error) {
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

// getAiExplainPrompt 获取AIEXPLAIN的提示词配置
func (s *SugarFormulaAiService) getAiExplainPrompt() (*sugar.SugarAgents, error) {
	var agent sugar.SugarAgents
	err := global.GVA_DB.Where("id = 'AiExplain'").First(&agent).Error
	if err != nil {
		return nil, errors.New("Agent不存在: 'AiExplain'")
	}
	return &agent, nil
}

// buildSystemPrompt 构建系统提示词
func (s *SugarFormulaAiService) buildSystemPrompt(agent *sugar.SugarAgents, userId string) string {
	enhancedPrompt := fmt.Sprintf(`你是一个专业的数据分析助手，专门负责调用数据分析工具。

📋 重要工作流程指导：
1. **使用智能匿名化分析工具**：对于贡献度分析需求，请使用 smart_anonymized_analyzer 工具
2. **精确匹配原则**：生成的筛选条件必须与用户问题中的具体实体对应
3. **数据验证策略**：工具会自动验证数据可用性
4. **结果可信度评估**：基于实际数据的完整性和代表性评估结论的可信度

🔧 工具使用指南：
- **推荐工具**：smart_anonymized_analyzer - 完整的智能匿名化分析流程
- 当前用户ID为 %s，调用工具时必须传递此用户ID
- 启用数据验证（enableDataValidation: true）以确保数据质量

💡 智能分析策略：
- 优先分析数据中贡献度最高的维度组合
- 对异常值和趋势变化提供深入洞察
- 结合业务常识给出可操作的建议
- 明确说明分析的局限性和数据范围`, userId)

	return enhancedPrompt
}

// serializeDataToText 将二维数组数据序列化为文本格式
func (s *SugarFormulaAiService) serializeDataToText(data [][]interface{}) (string, error) {
	if len(data) == 0 {
		return "", errors.New("数据为空")
	}

	var builder strings.Builder
	if len(data) > 0 {
		for i, cell := range data[0] {
			if i > 0 {
				builder.WriteString("\t")
			}
			builder.WriteString(fmt.Sprintf("%v", cell))
		}
		builder.WriteString("\n")
	}
	for i := 1; i < len(data); i++ {
		for j, cell := range data[i] {
			if j > 0 {
				builder.WriteString("\t")
			}
			builder.WriteString(fmt.Sprintf("%v", cell))
		}
		builder.WriteString("\n")
	}
	return builder.String(), nil
}
