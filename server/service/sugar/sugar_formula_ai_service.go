package sugar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	sugarRes "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

type SugarFormulaAiService struct{}

var llmService = system.SysLLMService{}
var executionLogService = SugarExecutionLogService{}

// init 包初始化，设置随机种子
func init() {
	// 在 Go 1.20+ 中，rand 包会自动使用安全的种子
	// 这里不需要手动设置种子
}

// AnonymizationSession 匿名化会话，为单次请求保存状态
type AnonymizationSession struct {
	// forwardMap 用于编码： "华东区域" -> "D01_V01"
	forwardMap map[string]string
	// reverseMap 用于解码： "D01_V01" -> "华东区域"
	reverseMap map[string]string
	// AIReadyData 是准备好发送给AI的、完全匿名化的数据
	AIReadyData []map[string]interface{}
}

// ToolCallResponse 用于解析LLM返回的工具调用指令
type ToolCallResponse struct {
	Type    string                  `json:"type"`
	Content []system.OpenAIToolCall `json:"content"`
}

// ExecuteAiFetchFormula 执行 AIFETCH 公式（使用OpenAI工具调用模式）
func (s *SugarFormulaAiService) ExecuteAiFetchFormula(ctx context.Context, req *sugarReq.SugarFormulaAiFetchRequest, userId string) (*sugarRes.SugarFormulaAiResponse, error) {
	global.GVA_LOG.Info("开始执行AIFETCH公式",
		zap.String("agentName", req.AgentName),
		zap.String("description", req.Description),
		zap.String("userId", userId))

	// 1. 获取Agent信息
	agent, err := s.getAgentByName(ctx, req.AgentName, userId)
	if err != nil {
		global.GVA_LOG.Error("获取Agent信息失败", zap.Error(err), zap.String("agentName", req.AgentName))
		return sugarRes.NewAiErrorResponse(err.Error()), nil
	}
	global.GVA_LOG.Info("成功获取Agent信息", zap.String("agentId", s.safeString(agent.Id)), zap.String("agentName", s.safeString(agent.Name)))

	// 2. 创建执行日志
	logCtx, err := executionLogService.CreateExecutionLog(ctx, req, userId, agent.Id)
	if err != nil {
		global.GVA_LOG.Error("创建执行日志失败", zap.Error(err))
		// 即使日志创建失败，也继续执行主要逻辑，但记录错误
	}

	// 2. 获取LLM配置
	var llmConfig *system.LLMConfig
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

	// 3. 构建系统提示词
	systemPrompt := s.buildSystemPrompt(agent, userId)
	global.GVA_LOG.Debug("构建系统提示词", zap.String("systemPrompt", systemPrompt))

	// 记录系统提示词到日志
	if logCtx != nil {
		executionLogService.RecordAISystemPrompt(ctx, logCtx, systemPrompt)
	}

	// 4. 构建用户消息
	userMessage := s.buildUserMessage(req.Description, agent.Semantic, req.DataRange)
	global.GVA_LOG.Debug("构建用户消息", zap.String("userMessage", userMessage))

	// 记录用户消息到日志
	if logCtx != nil {
		executionLogService.RecordAIUserMessage(ctx, logCtx, userMessage)
	}

	// 5. 准备工具定义（智能匿名化数据分析工具）
	tools := []system.ToolDefinition{
		{
			Name:        "smart_anonymized_analyzer",
			Description: "智能匿名化数据分析工具，自动进行数据范围探索和匿名化分析的完整流程。该工具会先验证数据可用性，然后进行匿名化贡献度分析，确保数据安全和分析准确性。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"modelName": map[string]interface{}{
						"type":        "string",
						"description": "要分析的语义模型名称。",
					},
					"targetMetric": map[string]interface{}{
						"type":        "string",
						"description": "需要分析的核心指标列名，例如 '销售金额'、'利润' 等。",
					},
					"currentPeriodFilters": map[string]interface{}{
						"type":        "object",
						"description": "获取本期数据的筛选条件，格式为 {\"列名\": \"筛选值\"}。",
					},
					"basePeriodFilters": map[string]interface{}{
						"type":        "object",
						"description": "获取基期（如上期、预算）数据的筛选条件，格式为 {\"列名\": \"筛选值\"}。",
					},
					"groupByDimensions": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "进行分组和归因分析的维度列名列表，如 ['区域', '产品类别']。",
					},
					"userId": map[string]interface{}{
						"type":        "string",
						"description": "发起请求的用户ID，工具内部需要此参数进行鉴权。",
					},
					"enableDataValidation": map[string]interface{}{
						"type":        "boolean",
						"description": "是否启用数据范围验证，默认为true。启用后会先验证筛选条件的有效性。",
					},
				},
				"required": []string{"modelName", "targetMetric", "currentPeriodFilters", "basePeriodFilters", "groupByDimensions", "userId"},
			},
		},
	}

	// 6. 构建消息列表
	messages := []system.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	// 记录LLM配置到日志
	if logCtx != nil {
		_ = executionLogService.RecordLLMConfig(ctx, logCtx, llmConfig)
	}

	// 7. 调用LLM，传入工具定义
	global.GVA_LOG.Info("开始调用LLM",
		zap.String("model", llmConfig.ModelName),
		zap.Int("toolsCount", len(tools)),
		zap.Int("messagesCount", len(messages)))

	llmResponse, err := llmService.ChatWithTools(ctx, *llmConfig, messages, tools)
	if err != nil {
		global.GVA_LOG.Error("AIFETCH LLM调用失败", zap.Error(err), zap.String("agent", req.AgentName))
		// 记录错误日志
		if logCtx != nil {
			errorMsg := "AI分析失败: " + err.Error()
			_ = executionLogService.FinishExecutionLog(ctx, logCtx, "", "failed", &errorMsg)
		}
		return sugarRes.NewAiErrorResponse("AI分析失败: " + err.Error()), nil
	}

	global.GVA_LOG.Info("LLM调用成功", zap.String("responseLength", fmt.Sprintf("%d", len(llmResponse))))
	global.GVA_LOG.Debug("LLM原始响应", zap.String("llmResponse", llmResponse))

	// 记录LLM响应到日志
	if logCtx != nil {
		modelName := llmConfig.ModelName
		executionLogService.RecordLLMResponse(ctx, logCtx, llmResponse, &modelName, nil)
	}

	// 8. 解析响应并处理可能的工具调用
	result, err := s.processAiFetchResponse(ctx, llmResponse, userId, req, agent, llmConfig, logCtx)
	if err != nil {
		// 记录错误日志
		if logCtx != nil {
			errorMsg := err.Error()
			_ = executionLogService.FinishExecutionLog(ctx, logCtx, "", "failed", &errorMsg)
		}
		return result, err
	}

	// 记录成功日志
	if logCtx != nil && result != nil {
		// 更新AI交互信息到数据库（现在数据库字段已添加）
		_ = executionLogService.UpdateExecutionLogWithAIInteractionsToDatabase(ctx, logCtx)

		finalResult := ""
		if result.Text != "" {
			finalResult = result.Text
		}
		_ = executionLogService.FinishExecutionLog(ctx, logCtx, finalResult, "success", nil)
	}

	return result, nil
}

// processAiFetchResponse 处理AIFETCH的响应，可能包含工具调用
func (s *SugarFormulaAiService) processAiFetchResponse(ctx context.Context, llmResponse string, userId string, req *sugarReq.SugarFormulaAiFetchRequest, agent *sugar.SugarAgents, llmConfig *system.LLMConfig, logCtx *ExecutionLogContext) (*sugarRes.SugarFormulaAiResponse, error) {
	global.GVA_LOG.Info("开始处理AIFETCH响应", zap.String("userId", userId))

	var toolCallResp ToolCallResponse
	err := json.Unmarshal([]byte(llmResponse), &toolCallResp)

	// 如果解析失败或不是工具调用，则认为是普通文本响应
	if err != nil || toolCallResp.Type != "tool_call" {
		if err != nil {
			global.GVA_LOG.Debug("响应不是JSON格式，作为普通文本处理", zap.Error(err))
		} else {
			global.GVA_LOG.Debug("响应类型不是工具调用，作为普通文本处理", zap.String("type", toolCallResp.Type))
		}
		return sugarRes.NewAiSuccessResponseWithText(llmResponse), nil
	}

	// 处理工具调用
	if len(toolCallResp.Content) > 0 {
		global.GVA_LOG.Info("检测到工具调用", zap.Int("toolCallCount", len(toolCallResp.Content)))

		toolCall := toolCallResp.Content[0]
		global.GVA_LOG.Info("处理工具调用",
			zap.String("functionName", toolCall.Function.Name),
			zap.String("arguments", toolCall.Function.Arguments))

		if toolCall.Function.Name == "smart_anonymized_analyzer" {
			// 处理智能匿名化分析工具调用
			return s.handleSmartAnonymizedAnalyzer(ctx, toolCall, logCtx, req, agent, llmConfig)
		} else if toolCall.Function.Name == "data_scope_explorer" {
			// 处理数据范围探索工具调用（保留向后兼容）
			return s.handleDataScopeExplorer(ctx, toolCall, logCtx)
		} else if toolCall.Function.Name == "anonymized_data_analyzer" {
			// 记录工具调用开始时间
			toolCallStartTime := time.Now()

			// 解析匿名化数据分析工具调用参数
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				global.GVA_LOG.Error("解析匿名化工具调用参数失败", zap.Error(err), zap.String("arguments", toolCall.Function.Arguments))

				// 记录工具调用错误
				if logCtx != nil {
					durationMs := int(time.Since(toolCallStartTime).Milliseconds())
					errorMsg := "解析工具调用参数失败: " + err.Error()
					executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, nil, &errorMsg, durationMs, false)
				}

				return sugarRes.NewAiErrorResponse("解析工具调用参数失败: " + err.Error()), nil
			}

			// 提取参数
			modelName, _ := args["modelName"].(string)
			targetMetric, _ := args["targetMetric"].(string)
			currentPeriodFilters, _ := args["currentPeriodFilters"].(map[string]interface{})
			basePeriodFilters, _ := args["basePeriodFilters"].(map[string]interface{})
			groupByDimensionsInterface, _ := args["groupByDimensions"].([]interface{})

			var groupByDimensions []string
			for _, item := range groupByDimensionsInterface {
				if str, ok := item.(string); ok {
					groupByDimensions = append(groupByDimensions, str)
				}
			}

			global.GVA_LOG.Info("提取匿名化工具调用参数",
				zap.String("modelName", modelName),
				zap.String("targetMetric", targetMetric),
				zap.Strings("groupByDimensions", groupByDimensions),
				zap.Any("currentPeriodFilters", currentPeriodFilters),
				zap.Any("basePeriodFilters", basePeriodFilters))

			// 执行匿名化数据处理
			anonymizedResult, err := s.processAnonymizedDataAnalysis(ctx, modelName, targetMetric, currentPeriodFilters, basePeriodFilters, groupByDimensions, userId)
			if err != nil {
				global.GVA_LOG.Error("匿名化数据处理失败", zap.Error(err))

				// 记录工具调用错误
				if logCtx != nil {
					durationMs := int(time.Since(toolCallStartTime).Milliseconds())
					errorMsg := "匿名化数据处理失败: " + err.Error()
					executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, nil, &errorMsg, durationMs, true)
				}

				return sugarRes.NewAiErrorResponse("匿名化数据处理失败: " + err.Error()), nil
			}

			// 将匿名化数据转换为AI可读格式
			aiDataText, err := s.serializeAnonymizedDataToText(anonymizedResult.AIReadyData)
			if err != nil {
				return sugarRes.NewAiErrorResponse("匿名化数据序列化失败: " + err.Error()), nil
			}

			global.GVA_LOG.Info("数据已完成匿名化处理，准备发送给AI",
				zap.Int("anonymizedDataLength", len(aiDataText)),
				zap.Int("mappingCount", len(anonymizedResult.forwardMap)),
				zap.String("dataPreview", func() string {
					if len(aiDataText) > 200 {
						return aiDataText[:200] + "..."
					}
					return aiDataText
				}()))

			// 更新日志记录匿名化信息
			if logCtx != nil {
				// 记录匿名化输入数据
				anonymizedInputData := map[string]interface{}{
					"aiDataText":   aiDataText,
					"toolCall":     toolCall.Function.Arguments,
					"mappingCount": len(anonymizedResult.forwardMap),
					"isEncrypted":  true, // 标记数据已加密
				}
				_ = executionLogService.UpdateExecutionLogWithAnonymization(ctx, logCtx, anonymizedInputData, nil)
			}

			// 进行AI分析（使用匿名化数据）
			global.GVA_LOG.Info("开始向AI发送匿名化数据进行分析")
			analysisResult, err := s.performDataAnalysis(ctx, aiDataText, req.Description, agent, llmConfig)
			if err != nil {
				return sugarRes.NewAiErrorResponse("AI数据分析失败: " + err.Error()), nil
			}

			global.GVA_LOG.Info("AI分析完成，准备解密响应",
				zap.Int("responseLength", len(analysisResult)),
				zap.String("responsePreview", func() string {
					if len(analysisResult) > 200 {
						return analysisResult[:200] + "..."
					}
					return analysisResult
				}()))

			// 更新匿名化输出
			if logCtx != nil {
				_ = executionLogService.UpdateExecutionLogWithAnonymization(ctx, logCtx, nil, &analysisResult)
			}

			// 解密AI分析结果
			decodedResult, err := s.decodeAIResponse(anonymizedResult, analysisResult)
			if err != nil {
				global.GVA_LOG.Error("AI结果解密失败", zap.Error(err))

				// 记录工具调用错误
				if logCtx != nil {
					durationMs := int(time.Since(toolCallStartTime).Milliseconds())
					errorMsg := "AI结果解密失败: " + err.Error()
					executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, nil, &errorMsg, durationMs, true)
				}

				return sugarRes.NewAiErrorResponse("AI结果解密失败: " + err.Error()), nil
			}

			global.GVA_LOG.Info("AI响应解密完成，返回最终结果",
				zap.Int("decodedLength", len(decodedResult)),
				zap.String("decodedPreview", func() string {
					if len(decodedResult) > 200 {
						return decodedResult[:200] + "..."
					}
					return decodedResult
				}()))

			// 记录工具调用成功
			if logCtx != nil {
				durationMs := int(time.Since(toolCallStartTime).Milliseconds())
				toolResult := map[string]interface{}{
					"decoded_result":        decodedResult,
					"anonymized_data_count": len(anonymizedResult.AIReadyData),
				}
				executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, toolResult, nil, durationMs, true)
			}

			return sugarRes.NewAiSuccessResponseWithText(decodedResult), nil
		}
	}

	// 如果没有可处理的工具调用，返回原始响应
	return sugarRes.NewAiSuccessResponseWithText(llmResponse), nil
}

// convertMapsToSlice 将map切片转换为二维接口切片
func (s *SugarFormulaAiService) convertMapsToSlice(maps []map[string]interface{}, columns []string) [][]interface{} {
	if len(maps) == 0 {
		return [][]interface{}{}
	}

	// 创建结果切片，第一行为表头
	result := make([][]interface{}, len(maps)+1)
	header := make([]interface{}, len(columns))
	for i, colName := range columns {
		header[i] = colName
	}
	result[0] = header

	// 填充数据行
	for i, rowMap := range maps {
		row := make([]interface{}, len(columns))
		for j, colName := range columns {
			row[j] = rowMap[colName]
		}
		result[i+1] = row
	}

	return result
}

// ExecuteAiExplainFormula 执行 AIEXPLAIN 公式（使用OpenAI兼容接口）
func (s *SugarFormulaAiService) ExecuteAiExplainFormula(ctx context.Context, req *sugarReq.SugarFormulaAiExplainRangeRequest, userId string) (*sugarRes.SugarFormulaAiResponse, error) {
	global.GVA_LOG.Info("开始执行AIEXPLAIN公式",
		zap.String("description", req.Description),
		zap.String("userId", userId),
		zap.Int("dataSourceRows", len(req.DataSource)))

	// 创建AIEXPLAIN的执行日志 (使用特殊的Agent ID)
	explainReq := &sugarReq.SugarFormulaAiFetchRequest{
		AgentName:   "AiExplain",
		Description: req.Description,
		DataRange:   "",
	}
	explainAgentId := "AiExplain"
	logCtx, err := executionLogService.CreateExecutionLog(ctx, explainReq, userId, &explainAgentId)
	if err != nil {
		global.GVA_LOG.Error("创建AIEXPLAIN执行日志失败", zap.Error(err))
		// 即使日志创建失败，也继续执行主要逻辑
	}

	// 1. 序列化数据为可读格式
	dataText, err := s.serializeDataToText(req.DataSource)
	if err != nil {
		global.GVA_LOG.Error("数据序列化失败", zap.Error(err))
		if logCtx != nil {
			errorMsg := "数据序列化失败: " + err.Error()
			_ = executionLogService.FinishExecutionLog(ctx, logCtx, "", "failed", &errorMsg)
		}
		return sugarRes.NewAiErrorResponse("数据序列化失败: " + err.Error()), nil
	}
	global.GVA_LOG.Debug("数据序列化成功", zap.String("dataTextLength", fmt.Sprintf("%d", len(dataText))))

	// 2. 获取LLM配置
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

	// 记录系统提示词和LLM配置
	if logCtx != nil {
		executionLogService.RecordAISystemPrompt(ctx, logCtx, systemPrompt)
		_ = executionLogService.RecordLLMConfig(ctx, logCtx, llmConfig)
	}

	// 3. 构建用户消息
	userMessage := fmt.Sprintf("请分析以下数据：\n\n%s\n\n分析要求：%s", dataText, req.Description)
	global.GVA_LOG.Debug("构建用户消息", zap.String("userMessage", userMessage))

	// 记录用户消息
	if logCtx != nil {
		executionLogService.RecordAIUserMessage(ctx, logCtx, userMessage)
	}

	// 4. 构建消息列表
	messages := []system.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	// 5. 直接调用OpenAI兼容接口（不带工具）
	global.GVA_LOG.Info("开始调用LLM进行AIEXPLAIN分析", zap.String("model", llmConfig.ModelName))
	response, err := llmService.ChatSimple(ctx, *llmConfig, messages)
	if err != nil {
		global.GVA_LOG.Error("AIEXPLAIN OpenAI调用失败", zap.Error(err))
		if logCtx != nil {
			errorMsg := "AI分析失败: " + err.Error()
			_ = executionLogService.FinishExecutionLog(ctx, logCtx, "", "failed", &errorMsg)
		}
		return sugarRes.NewAiErrorResponse("AI分析失败: " + err.Error()), nil
	}

	global.GVA_LOG.Info("AIEXPLAIN分析完成", zap.String("responseLength", fmt.Sprintf("%d", len(response))))
	global.GVA_LOG.Debug("AIEXPLAIN响应内容", zap.String("response", response))

	// 记录LLM响应
	if logCtx != nil {
		modelName := llmConfig.ModelName
		executionLogService.RecordLLMResponse(ctx, logCtx, response, &modelName, nil)

		// 更新AI交互信息到数据库（现在数据库字段已添加）
		_ = executionLogService.UpdateExecutionLogWithAIInteractionsToDatabase(ctx, logCtx)

		// 记录成功日志
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

// get AIEXPLAIN prompt
func (s *SugarFormulaAiService) getAiExplainPrompt() (*sugar.SugarAgents, error) {
	var agent sugar.SugarAgents

	// 查找Agent
	err := global.GVA_DB.Where(" id = 'AiExplain' ").First(&agent).Error
	if err != nil {
		return nil, errors.New("Agent不存在: 'AiExplain' ")
	}

	return &agent, nil
}

// buildSystemPrompt 构建智能系统提示词
func (s *SugarFormulaAiService) buildSystemPrompt(agent *sugar.SugarAgents, userId string) string {
	basePrompt := ""
	if agent.Prompt != nil && *agent.Prompt != "" {
		basePrompt = *agent.Prompt
	} else {
		basePrompt = "你是一个专业的数据分析师，请根据用户的需求进行数据分析。"
	}

	// 构建智能化的系统提示词
	enhancedPrompt := fmt.Sprintf(`%s

📋 重要工作流程指导：
1. **使用智能匿名化分析工具**：对于贡献度分析需求，请使用 smart_anonymized_analyzer 工具，它会自动完成数据验证和匿名化分析的完整流程
2. **精确匹配原则**：生成的筛选条件必须与用户问题中的具体实体对应，避免过于宽泛或不存在的条件
3. **数据验证策略**：工具会自动验证数据可用性，如果数据不足会给出明确提示
4. **结果可信度评估**：基于实际数据的完整性和代表性评估结论的可信度

🔧 工具使用指南：
- **推荐工具**：smart_anonymized_analyzer - 完整的智能匿名化分析流程
- **备用工具**：data_scope_explorer（仅数据探索）、anonymized_data_analyzer（传统匿名化）
- 当前用户ID为 %s，调用工具时必须传递此用户ID
- 启用数据验证（enableDataValidation: true）以确保数据质量

💡 智能分析策略：
- 优先分析数据中贡献度最高的维度组合
- 对异常值和趋势变化提供深入洞察
- 结合业务常识给出可操作的建议
- 明确说明分析的局限性和数据范围`, basePrompt, userId)

	return enhancedPrompt
}

// buildUserMessage 构建用户消息
func (s *SugarFormulaAiService) buildUserMessage(description string, semantic *string, dataRange string) string {
	message := description

	// 如果有语义模型标识，从数据库获取详细信息
	if semantic != nil && *semantic != "" {
		semanticInfo, err := s.getSemanticModelInfo(*semantic)
		if err != nil {
			global.GVA_LOG.Warn("获取语义模型信息失败", zap.String("semantic", *semantic), zap.Error(err))
			message += fmt.Sprintf("\n\n可用的数据模型信息：\n%s", *semantic)
		} else {
			message += fmt.Sprintf("\n\n可用的数据模型信息：\n%s", semanticInfo)
		}
	}

	// 如果提供了DataRange数据，将其包含在提示词中
	if dataRange != "" {
		global.GVA_LOG.Info("包含DataRange数据到提示词中", zap.Int("dataRangeLength", len(dataRange)))

		message += fmt.Sprintf("\n\n相关数据范围：\n%s", dataRange)
	}

	return message
}

// getSemanticModelInfo 根据语义模型名称或ID获取详细信息
func (s *SugarFormulaAiService) getSemanticModelInfo(semantic string) (string, error) {
	var models []sugar.SugarSemanticModels

	err := global.GVA_DB.Where("(name = ? OR id = ?) AND deleted_at IS NULL", semantic, semantic).Find(&models).Error
	if err != nil {
		return "", fmt.Errorf("查询语义模型失败: %w", err)
	}
	if len(models) == 0 {
		return "", fmt.Errorf("未找到语义模型: %s", semantic)
	}
	var builder strings.Builder
	for i, model := range models {
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(fmt.Sprintf("模型名称: %s\n", s.safeString(model.Name)))
		if model.Description != nil && *model.Description != "" {
			builder.WriteString(fmt.Sprintf("模型描述: %s\n", *model.Description))
		}
		if model.SourceTableName != nil && *model.SourceTableName != "" {
			builder.WriteString(fmt.Sprintf("数据表: %s\n", *model.SourceTableName))
		}
		if len(model.ParameterConfig) > 0 {
			paramInfo, err := s.parseParameterConfig(model.ParameterConfig)
			if err != nil {
				global.GVA_LOG.Warn("解析参数配置失败", zap.Error(err))
			} else {
				builder.WriteString(fmt.Sprintf("可用筛选条件:\n%s\n", paramInfo))
			}
		}
		if len(model.ReturnableColumnsConfig) > 0 {
			columnInfo, err := s.parseReturnableColumnsConfig(model.ReturnableColumnsConfig)
			if err != nil {
				global.GVA_LOG.Warn("解析返回字段配置失败", zap.Error(err))
			} else {
				builder.WriteString(fmt.Sprintf("可返回字段:\n%s", columnInfo))
			}
		}
	}
	return builder.String(), nil
}

// parseParameterConfig 解析参数配置JSON
func (s *SugarFormulaAiService) parseParameterConfig(configJSON []byte) (string, error) {
	var config map[string]map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return "", fmt.Errorf("解析参数配置JSON失败: %w", err)
	}
	var builder strings.Builder
	for paramName, paramConfig := range config {
		builder.WriteString(fmt.Sprintf("  - %s: ", paramName))
		if desc, ok := paramConfig["description"].(string); ok {
			builder.WriteString(desc)
		}
		if paramType, ok := paramConfig["type"].(string); ok {
			builder.WriteString(fmt.Sprintf(" (类型: %s)", paramType))
		}
		// if column, ok := paramConfig["column"].(string); ok {
		// 	builder.WriteString(fmt.Sprintf(" [字段: %s]", column))
		// }
		if operator, ok := paramConfig["operator"].(string); ok {
			builder.WriteString(fmt.Sprintf(" [操作符: %s]", operator))
		}
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

// parseReturnableColumnsConfig 解析返回字段配置JSON
func (s *SugarFormulaAiService) parseReturnableColumnsConfig(configJSON []byte) (string, error) {
	var config map[string]map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return "", fmt.Errorf("解析返回字段配置JSON失败: %w", err)
	}
	var builder strings.Builder
	for columnName, columnConfig := range config {
		builder.WriteString(fmt.Sprintf("  - %s: ", columnName))
		if desc, ok := columnConfig["description"].(string); ok {
			builder.WriteString(desc)
		}
		if columnType, ok := columnConfig["type"].(string); ok {
			builder.WriteString(fmt.Sprintf(" (类型: %s)", columnType))
		}
		// if column, ok := columnConfig["column"].(string); ok {
		// 	builder.WriteString(fmt.Sprintf(" [字段: %s]", column))
		// }
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

// safeString 安全地获取字符串指针的值
func (s *SugarFormulaAiService) safeString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
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

// performDataAnalysis 对获取的数据进行AI分析（上下文感知版本）
func (s *SugarFormulaAiService) performDataAnalysis(ctx context.Context, dataText string, userDescription string, agent *sugar.SugarAgents, llmConfig *system.LLMConfig) (string, error) {
	global.GVA_LOG.Info("开始执行上下文感知数据分析",
		zap.String("userDescription", userDescription),
		zap.String("dataLength", fmt.Sprintf("%d", len(dataText))))

	// 构建上下文感知的分析提示词
	systemPrompt := s.buildContextAwareAnalysisPrompt(agent)
	global.GVA_LOG.Debug("构建上下文感知分析系统提示词", zap.String("systemPrompt", systemPrompt))

	// 构建增强的用户消息，包含数据范围说明和分析要求
	userMessage := s.buildEnhancedAnalysisMessage(dataText, userDescription)
	global.GVA_LOG.Debug("构建增强分析用户消息", zap.String("userMessage", userMessage))

	// 构建消息列表
	messages := []system.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	// 调用LLM进行分析
	global.GVA_LOG.Info("开始调用LLM进行上下文感知数据分析", zap.String("model", llmConfig.ModelName))
	response, err := llmService.ChatSimple(ctx, *llmConfig, messages)
	if err != nil {
		global.GVA_LOG.Error("AI数据分析调用失败", zap.Error(err))
		return "", fmt.Errorf("AI分析失败: %w", err)
	}

	// 对分析结果进行质量评估
	qualityScore := s.evaluateAnalysisQuality(response, dataText, userDescription)
	global.GVA_LOG.Info("数据分析完成",
		zap.String("responseLength", fmt.Sprintf("%d", len(response))),
		zap.Float64("qualityScore", qualityScore))

	global.GVA_LOG.Debug("数据分析响应", zap.String("response", response))
	return response, nil
}

// buildContextAwareAnalysisPrompt 构建上下文感知的分析提示词
func (s *SugarFormulaAiService) buildContextAwareAnalysisPrompt(agent *sugar.SugarAgents) string {
	basePrompt := ""
	if agent.Prompt != nil && *agent.Prompt != "" {
		basePrompt = *agent.Prompt
	} else {
		basePrompt = "你是一个专业的数据分析师，擅长从匿名化数据中挖掘商业洞察。"
	}

	enhancedPrompt := fmt.Sprintf(`%s

🎯 上下文感知分析指导：

📊 **数据理解要求**：
1. **匿名化数据解读**：数据中的维度代号（如D01、D02）和值代号（如D01_V01）都是匿名化处理的敏感业务维度
2. **数据完整性评估**：在分析前要评估数据的代表性和完整性，明确指出数据范围的局限性
3. **统计意义判断**：基于数据量和分布情况判断分析结论的统计显著性

🔍 **深度分析策略**：
1. **贡献度优先排序**：重点分析贡献度绝对值最大的维度组合，识别主要驱动因子
2. **正负向分类**：区分正向驱动因子和负向拖累因子，分别给出针对性建议
3. **异常值识别**：识别贡献度异常高或异常低的维度组合，探索潜在原因
4. **趋势模式分析**：从变化值和贡献度中识别业务趋势和模式

📈 **结论输出要求**：
1. **分层次结论**：从整体趋势到细分维度，层层递进给出洞察
2. **量化表述**：用具体的百分比和数值支撑每个结论
3. **可信度说明**：明确说明每个结论的可信度和数据支撑强度
4. **行动建议**：基于分析结果给出具体的业务行动建议

⚠️ **重要注意事项**：
- 由于数据已匿名化，不要尝试推测具体的业务实体名称
- 关注数据模式和相对关系，而非绝对值
- 如发现数据异常或不完整，要明确指出并说明对结论的影响`, basePrompt)

	return enhancedPrompt
}

// buildEnhancedAnalysisMessage 构建增强的分析消息
func (s *SugarFormulaAiService) buildEnhancedAnalysisMessage(dataText string, userDescription string) string {
	var builder strings.Builder

	builder.WriteString("请对以下匿名化贡献度数据进行深度分析：\n\n")
	builder.WriteString(dataText)
	builder.WriteString("\n")

	// 分析数据基本特征
	dataStats := s.analyzeDataCharacteristics(dataText)
	builder.WriteString(fmt.Sprintf("📋 数据基本特征：\n"))
	builder.WriteString(fmt.Sprintf("- 数据项总数：%d\n", dataStats["itemCount"]))
	builder.WriteString(fmt.Sprintf("- 维度组合数：%d\n", dataStats["dimensionCount"]))
	builder.WriteString(fmt.Sprintf("- 正向驱动因子数：%d\n", dataStats["positiveDrivers"]))
	builder.WriteString(fmt.Sprintf("- 负向拖累因子数：%d\n", dataStats["negativeDrivers"]))

	if avgContribution, ok := dataStats["avgContribution"].(float64); ok {
		builder.WriteString(fmt.Sprintf("- 平均贡献度：%.2f%%\n", avgContribution))
	}

	builder.WriteString("\n🎯 用户分析需求：\n")
	builder.WriteString(userDescription)
	builder.WriteString("\n\n")

	builder.WriteString("📊 请按以下结构进行分析：\n")
	builder.WriteString("1. **整体趋势分析**：总体变化方向和主要特征\n")
	builder.WriteString("2. **关键驱动因子**：贡献度最高的前3-5个因子及其影响\n")
	builder.WriteString("3. **异常点识别**：值得关注的异常表现和可能原因\n")
	builder.WriteString("4. **业务洞察**：基于数据模式的商业洞察和建议\n")
	builder.WriteString("5. **结论可信度**：分析结论的可靠性评估\n")

	return builder.String()
}

// analyzeDataCharacteristics 分析数据基本特征
func (s *SugarFormulaAiService) analyzeDataCharacteristics(dataText string) map[string]interface{} {
	stats := make(map[string]interface{})

	// 简单的文本分析来提取基本统计信息
	lines := strings.Split(dataText, "\n")
	itemCount := 0
	dimensionCount := 0
	positiveDrivers := 0
	negativeDrivers := 0
	contributionSum := 0.0
	contributionCount := 0

	dimensionSet := make(map[string]bool)

	for _, line := range lines {
		// 统计项目数
		if strings.Contains(line, "项目") && strings.Contains(line, ":") {
			itemCount++
		}

		// 统计维度
		if strings.Contains(line, "D") && strings.Contains(line, "_V") {
			// 提取维度代号
			if strings.HasPrefix(strings.TrimSpace(line), "D") {
				parts := strings.Split(strings.TrimSpace(line), ":")
				if len(parts) > 0 {
					dimCode := strings.Split(parts[0], "_")[0]
					dimensionSet[dimCode] = true
				}
			}
		}

		// 统计正负向驱动因子
		if strings.Contains(line, "正向驱动: true") {
			positiveDrivers++
		} else if strings.Contains(line, "正向驱动: false") {
			negativeDrivers++
		}

		// 统计贡献度
		if strings.Contains(line, "贡献度:") {
			// 提取贡献度数值
			parts := strings.Split(line, "贡献度:")
			if len(parts) > 1 {
				contributionStr := strings.TrimSpace(strings.Replace(parts[1], "%", "", -1))
				if contribution := s.parseFloatFromString(contributionStr); contribution != 0 {
					contributionSum += contribution
					contributionCount++
				}
			}
		}
	}

	dimensionCount = len(dimensionSet)

	stats["itemCount"] = itemCount
	stats["dimensionCount"] = dimensionCount
	stats["positiveDrivers"] = positiveDrivers
	stats["negativeDrivers"] = negativeDrivers

	if contributionCount > 0 {
		stats["avgContribution"] = contributionSum / float64(contributionCount)
	} else {
		stats["avgContribution"] = 0.0
	}

	return stats
}

// parseFloatFromString 从字符串中解析浮点数
func (s *SugarFormulaAiService) parseFloatFromString(str string) float64 {
	// 移除所有非数字字符（除了小数点和负号）
	cleanStr := ""
	for _, char := range str {
		if (char >= '0' && char <= '9') || char == '.' || char == '-' {
			cleanStr += string(char)
		}
	}

	if cleanStr == "" {
		return 0.0
	}

	var result float64
	if n, err := fmt.Sscanf(cleanStr, "%f", &result); err == nil && n == 1 {
		return result
	}
	return 0.0
}

// evaluateAnalysisQuality 评估分析结果质量
func (s *SugarFormulaAiService) evaluateAnalysisQuality(response string, dataText string, userDescription string) float64 {
	qualityScore := 0.0
	maxScore := 100.0

	// 1. 结构完整性 (30分)
	structureScore := 0.0
	if strings.Contains(response, "整体趋势") || strings.Contains(response, "总体") {
		structureScore += 10.0
	}
	if strings.Contains(response, "驱动因子") || strings.Contains(response, "关键") {
		structureScore += 10.0
	}
	if strings.Contains(response, "建议") || strings.Contains(response, "洞察") {
		structureScore += 10.0
	}
	qualityScore += structureScore

	// 2. 数据引用度 (25分)
	dataReferenceScore := 0.0
	// 检查是否引用了具体的代号
	if strings.Contains(response, "D01") || strings.Contains(response, "D02") {
		dataReferenceScore += 10.0
	}
	// 检查是否引用了具体的百分比
	if strings.Contains(response, "%") {
		dataReferenceScore += 10.0
	}
	// 检查是否有量化描述
	if strings.Contains(response, "贡献度") {
		dataReferenceScore += 5.0
	}
	qualityScore += dataReferenceScore

	// 3. 逻辑连贯性 (20分)
	logicalScore := 0.0
	responseLen := len(response)
	if responseLen > 200 {
		logicalScore += 10.0
	}
	if responseLen > 500 {
		logicalScore += 10.0
	}
	qualityScore += logicalScore

	// 4. 问题相关性 (25分)
	relevanceScore := 0.0
	userWords := strings.Fields(strings.ToLower(userDescription))
	responseLower := strings.ToLower(response)

	matchedWords := 0
	for _, word := range userWords {
		if len(word) > 2 && strings.Contains(responseLower, word) {
			matchedWords++
		}
	}

	if len(userWords) > 0 {
		relevanceRatio := float64(matchedWords) / float64(len(userWords))
		relevanceScore = relevanceRatio * 25.0
	}
	qualityScore += relevanceScore

	// 计算最终得分
	finalScore := (qualityScore / maxScore) * 100.0
	if finalScore > 100.0 {
		finalScore = 100.0
	}

	global.GVA_LOG.Debug("分析质量评估详情",
		zap.Float64("structureScore", structureScore),
		zap.Float64("dataReferenceScore", dataReferenceScore),
		zap.Float64("logicalScore", logicalScore),
		zap.Float64("relevanceScore", relevanceScore),
		zap.Float64("finalScore", finalScore))

	return finalScore
}

// processAnonymizedDataAnalysis 执行匿名化数据分析处理
func (s *SugarFormulaAiService) processAnonymizedDataAnalysis(ctx context.Context, modelName, targetMetric string, currentPeriodFilters, basePeriodFilters map[string]interface{}, groupByDimensions []string, userId string) (*AnonymizationSession, error) {
	global.GVA_LOG.Info("开始匿名化数据处理",
		zap.String("modelName", modelName),
		zap.String("targetMetric", targetMetric),
		zap.Strings("groupByDimensions", groupByDimensions),
		zap.String("userId", userId))

	// 1. 并发获取本期和基期数据
	currentData, baseData, err := s.fetchDataConcurrently(ctx, modelName, targetMetric, currentPeriodFilters, basePeriodFilters, groupByDimensions, userId)
	if err != nil {
		return nil, fmt.Errorf("并发获取数据失败: %w", err)
	}

	// 2. 计算贡献度分析
	contributions, err := s.calculateContributions(currentData, baseData, targetMetric, groupByDimensions)
	if err != nil {
		return nil, fmt.Errorf("计算贡献度失败: %w", err)
	}

	// 3. 创建匿名化会话并进行数据加密
	session, err := s.createAnonymizedSession(contributions)
	if err != nil {
		return nil, fmt.Errorf("创建匿名化会话失败: %w", err)
	}

	global.GVA_LOG.Info("匿名化数据处理完成",
		zap.Int("contributionCount", len(contributions)),
		zap.Int("aiDataCount", len(session.AIReadyData)),
		zap.Int("mappingCount", len(session.forwardMap)))

	return session, nil
}

// fetchDataConcurrently 并发获取本期和基期数据
func (s *SugarFormulaAiService) fetchDataConcurrently(ctx context.Context, modelName, targetMetric string, currentPeriodFilters, basePeriodFilters map[string]interface{}, groupByDimensions []string, userId string) (*sugarRes.SugarFormulaGetResponse, *sugarRes.SugarFormulaGetResponse, error) {
	// 构建返回列：目标指标 + 分组维度
	returnColumns := append([]string{targetMetric}, groupByDimensions...)

	// 使用通道进行并发处理
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
			Filters:       currentPeriodFilters,
		}
		formulaQueryService := SugarFormulaQueryService{}
		currentData, err := formulaQueryService.ExecuteGetFormula(ctx, currentReq, userId)
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
			Filters:       basePeriodFilters,
		}
		formulaQueryService := SugarFormulaQueryService{}
		baseData, err := formulaQueryService.ExecuteGetFormula(ctx, baseReq, userId)
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

	global.GVA_LOG.Info("数据获取完成",
		zap.Int("currentDataCount", len(currentResult.data.Results)),
		zap.Int("baseDataCount", len(baseResult.data.Results)))

	return currentResult.data, baseResult.data, nil
}

// calculateContributions 计算贡献度分析
func (s *SugarFormulaAiService) calculateContributions(currentData, baseData *sugarRes.SugarFormulaGetResponse, targetMetric string, groupByDimensions []string) ([]ContributionItem, error) {
	// 将数据按维度组合进行分组
	currentGroups := s.groupDataByDimensions(currentData.Results, groupByDimensions, targetMetric)
	baseGroups := s.groupDataByDimensions(baseData.Results, groupByDimensions, targetMetric)

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
		dimensionValues := s.parseDimensionKey(key, groupByDimensions)

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

// ContributionItem 表示单个维度组合的贡献度分析结果
type ContributionItem struct {
	DimensionValues     map[string]interface{} // 维度值组合，如 {"区域": "华东", "产品": "A产品"}
	CurrentValue        float64                // 本期值
	BaseValue           float64                // 基期值
	ChangeValue         float64                // 变化值 (本期值 - 基期值)
	ContributionPercent float64                // 贡献度百分比
	IsPositiveDriver    bool                   // 是否为正向驱动因子
}

// groupDataByDimensions 按维度组合对数据进行分组聚合
func (s *SugarFormulaAiService) groupDataByDimensions(data []map[string]interface{}, dimensions []string, targetMetric string) map[string]float64 {
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
func (s *SugarFormulaAiService) buildDimensionKey(row map[string]interface{}, dimensions []string) string {
	var keyParts []string
	for _, dim := range dimensions {
		value := fmt.Sprintf("%v", row[dim])
		keyParts = append(keyParts, fmt.Sprintf("%s:%s", dim, value))
	}
	return strings.Join(keyParts, "|")
}

// parseDimensionKey 解析维度键回到维度值映射
func (s *SugarFormulaAiService) parseDimensionKey(key string, dimensions []string) map[string]interface{} {
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
func (s *SugarFormulaAiService) getAllUniqueKeys(groups1, groups2 map[string]float64) []string {
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
func (s *SugarFormulaAiService) extractFloatValue(value interface{}) float64 {
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

// createAnonymizedSession 创建匿名化会话
func (s *SugarFormulaAiService) createAnonymizedSession(contributions []ContributionItem) (*AnonymizationSession, error) {
	session := &AnonymizationSession{
		forwardMap:  make(map[string]string),
		reverseMap:  make(map[string]string),
		AIReadyData: make([]map[string]interface{}, 0),
	}

	// 维度计数器，用于生成唯一代号
	dimensionCounters := make(map[string]int)
	valueCounters := make(map[string]int)

	global.GVA_LOG.Info("开始创建匿名化会话", zap.Int("contributionCount", len(contributions)))

	// 处理每个贡献项
	for i, contribution := range contributions {
		aiItem := make(map[string]interface{})

		// 处理维度值的匿名化
		for dimName, dimValue := range contribution.DimensionValues {
			anonymizedDimName := s.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
			anonymizedDimValue := s.getOrCreateAnonymizedValue(session, dimName, fmt.Sprintf("%v", dimValue), valueCounters)

			aiItem[anonymizedDimName] = anonymizedDimValue
		}

		// 添加经过脱敏处理的数值数据
		aiItem["contribution_percent"] = s.anonymizeNumericValue(contribution.ContributionPercent, "contribution")
		aiItem["is_positive_driver"] = contribution.IsPositiveDriver
		aiItem["change_value"] = s.anonymizeNumericValue(contribution.ChangeValue, "change")
		aiItem["current_value"] = s.anonymizeNumericValue(contribution.CurrentValue, "current")
		aiItem["base_value"] = s.anonymizeNumericValue(contribution.BaseValue, "base")

		session.AIReadyData = append(session.AIReadyData, aiItem)

		// 记录匿名化进度
		if i%10 == 0 || i == len(contributions)-1 {
			global.GVA_LOG.Debug("匿名化进度",
				zap.Int("processed", i+1),
				zap.Int("total", len(contributions)),
				zap.Int("currentMappings", len(session.forwardMap)))
		}
	}

	global.GVA_LOG.Info("匿名化会话创建完成",
		zap.Int("forwardMapSize", len(session.forwardMap)),
		zap.Int("reverseMapSize", len(session.reverseMap)),
		zap.Int("aiDataSize", len(session.AIReadyData)))

	return session, nil
}

// getOrCreateAnonymizedDimension 获取或创建维度名的匿名化代号
func (s *SugarFormulaAiService) getOrCreateAnonymizedDimension(session *AnonymizationSession, dimName string, counters map[string]int) string {
	// 检查是否已经存在匿名化代号
	if anonymized, exists := session.forwardMap[dimName]; exists {
		return anonymized
	}

	// 生成新的维度代号
	counters["dimension"]++
	anonymized := fmt.Sprintf("D%02d", counters["dimension"])

	// 存储映射关系
	session.forwardMap[dimName] = anonymized
	session.reverseMap[anonymized] = dimName

	return anonymized
}

// getOrCreateAnonymizedValue 获取或创建维度值的匿名化代号
func (s *SugarFormulaAiService) getOrCreateAnonymizedValue(session *AnonymizationSession, dimName, dimValue string, counters map[string]int) string {
	// 构建完整的键（维度名+值）
	fullKey := fmt.Sprintf("%s:%s", dimName, dimValue)

	// 检查是否已经存在匿名化代号
	if anonymized, exists := session.forwardMap[fullKey]; exists {
		return anonymized
	}

	// 获取维度的匿名化代号
	anonymizedDim := session.forwardMap[dimName]
	if anonymizedDim == "" {
		// 如果维度还没有匿名化，先创建维度代号
		dimensionCounters := make(map[string]int)
		anonymizedDim = s.getOrCreateAnonymizedDimension(session, dimName, dimensionCounters)
	}

	// 生成新的值代号
	dimKey := fmt.Sprintf("value_%s", dimName)
	counters[dimKey]++
	anonymized := fmt.Sprintf("%s_V%02d", anonymizedDim, counters[dimKey])

	// 存储映射关系
	session.forwardMap[fullKey] = anonymized
	session.reverseMap[anonymized] = dimValue

	return anonymized
}

// serializeAnonymizedDataToText 将匿名化数据序列化为文本格式
func (s *SugarFormulaAiService) serializeAnonymizedDataToText(data []map[string]interface{}) (string, error) {
	if len(data) == 0 {
		return "", errors.New("匿名化数据为空")
	}

	var builder strings.Builder
	builder.WriteString("【匿名化贡献度分析数据】\n")
	builder.WriteString("说明：以下数据已进行匿名化处理，维度名称和值都已替换为代号\n\n")

	// 添加数据列说明
	builder.WriteString("数据字段说明：\n")
	builder.WriteString("- 维度代号（D01, D02等）：表示敏感业务维度\n")
	builder.WriteString("- 值代号（D01_V01, D01_V02等）：表示具体的维度值\n")
	builder.WriteString("- contribution_percent：贡献度百分比\n")
	builder.WriteString("- is_positive_driver：是否为正向驱动因子\n")
	builder.WriteString("- change_value：变化值\n")
	builder.WriteString("- current_value：本期值\n")
	builder.WriteString("- base_value：基期值\n\n")

	builder.WriteString("数据内容：\n")
	for i, item := range data {
		builder.WriteString(fmt.Sprintf("项目 %d:\n", i+1))

		// 先输出维度信息
		for key, value := range item {
			if strings.HasPrefix(key, "D") && !strings.Contains(key, "_") {
				builder.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
			}
		}

		// 再输出分析数据
		if cp, ok := item["contribution_percent"]; ok {
			builder.WriteString(fmt.Sprintf("  贡献度: %.2f%%\n", cp))
		}
		if ipd, ok := item["is_positive_driver"]; ok {
			builder.WriteString(fmt.Sprintf("  正向驱动: %v\n", ipd))
		}
		if cv, ok := item["change_value"]; ok {
			builder.WriteString(fmt.Sprintf("  变化值: %.2f\n", cv))
		}
		if curr, ok := item["current_value"]; ok {
			builder.WriteString(fmt.Sprintf("  本期值: %.2f\n", curr))
		}
		if base, ok := item["base_value"]; ok {
			builder.WriteString(fmt.Sprintf("  基期值: %.2f\n", base))
		}

		builder.WriteString("\n")
	}

	global.GVA_LOG.Info("匿名化数据序列化完成",
		zap.Int("dataCount", len(data)),
		zap.Int("textLength", len(builder.String())))

	return builder.String(), nil
}

// decodeAIResponse 解码AI响应中的匿名代号
func (s *SugarFormulaAiService) decodeAIResponse(session *AnonymizationSession, aiText string) (string, error) {
	if session == nil {
		return "", errors.New("匿名化会话为空")
	}

	if aiText == "" {
		global.GVA_LOG.Warn("AI响应为空，无需解码")
		return "", nil
	}

	global.GVA_LOG.Info("开始解码AI响应",
		zap.Int("originalLength", len(aiText)),
		zap.Int("mappingCount", len(session.reverseMap)))

	// 获取所有需要替换的代号，按长度降序排序以避免部分替换问题
	var codes []string
	for code := range session.reverseMap {
		codes = append(codes, code)
	}

	// 按字符串长度降序排序，确保长代号先被替换
	for i := 0; i < len(codes); i++ {
		for j := i + 1; j < len(codes); j++ {
			if len(codes[i]) < len(codes[j]) {
				codes[i], codes[j] = codes[j], codes[i]
			}
		}
	}

	// 执行替换
	decodedText := aiText
	replacementCount := 0
	replacementDetails := make(map[string]string)

	for _, code := range codes {
		originalValue := session.reverseMap[code]
		if strings.Contains(decodedText, code) {
			oldText := decodedText
			decodedText = strings.ReplaceAll(decodedText, code, originalValue)

			// 统计实际替换次数
			occurrences := strings.Count(oldText, code)
			if occurrences > 0 {
				replacementCount += occurrences
				replacementDetails[code] = originalValue

				global.GVA_LOG.Debug("执行代号替换",
					zap.String("code", code),
					zap.String("originalValue", originalValue),
					zap.Int("occurrences", occurrences))
			}
		}
	}

	// 验证解码结果
	if replacementCount == 0 {
		global.GVA_LOG.Warn("未发现需要解码的匿名代号", zap.String("aiText", aiText))
	}

	global.GVA_LOG.Info("AI响应解码完成",
		zap.Int("totalCodes", len(codes)),
		zap.Int("foundCodes", len(replacementDetails)),
		zap.Int("totalReplacements", replacementCount),
		zap.Int("originalLength", len(aiText)),
		zap.Int("decodedLength", len(decodedText)))

	return decodedText, nil
}

// handleDataScopeExplorer 处理数据范围探索工具调用
func (s *SugarFormulaAiService) handleDataScopeExplorer(ctx context.Context, toolCall system.OpenAIToolCall, logCtx *ExecutionLogContext) (*sugarRes.SugarFormulaAiResponse, error) {
	global.GVA_LOG.Info("开始处理数据范围探索工具调用")

	// 记录工具调用开始时间
	toolCallStartTime := time.Now()

	// 解析工具调用参数
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		global.GVA_LOG.Error("解析数据范围探索工具参数失败", zap.Error(err))

		if logCtx != nil {
			durationMs := int(time.Since(toolCallStartTime).Milliseconds())
			errorMsg := "解析工具调用参数失败: " + err.Error()
			executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, nil, &errorMsg, durationMs, false)
		}

		return sugarRes.NewAiErrorResponse("解析工具调用参数失败: " + err.Error()), nil
	}

	// 提取参数
	modelName, _ := args["modelName"].(string)
	exploreDimensionsInterface, _ := args["exploreDimensions"].([]interface{})
	sampleFilters, _ := args["sampleFilters"].(map[string]interface{})
	userId, _ := args["userId"].(string)

	var exploreDimensions []string
	for _, item := range exploreDimensionsInterface {
		if str, ok := item.(string); ok {
			exploreDimensions = append(exploreDimensions, str)
		}
	}

	global.GVA_LOG.Info("提取数据范围探索参数",
		zap.String("modelName", modelName),
		zap.Strings("exploreDimensions", exploreDimensions),
		zap.Any("sampleFilters", sampleFilters),
		zap.String("userId", userId))

	// 执行数据范围探索
	scopeInfo, err := s.exploreDataScope(ctx, modelName, exploreDimensions, sampleFilters, userId)
	if err != nil {
		global.GVA_LOG.Error("数据范围探索失败", zap.Error(err))

		if logCtx != nil {
			durationMs := int(time.Since(toolCallStartTime).Milliseconds())
			errorMsg := "数据范围探索失败: " + err.Error()
			executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, nil, &errorMsg, durationMs, true)
		}

		return sugarRes.NewAiErrorResponse("数据范围探索失败: " + err.Error()), nil
	}

	// 格式化探索结果
	resultText := s.formatDataScopeResult(scopeInfo)

	// 记录工具调用成功
	if logCtx != nil {
		durationMs := int(time.Since(toolCallStartTime).Milliseconds())
		toolResult := map[string]interface{}{
			"scope_info":          scopeInfo,
			"explored_dimensions": len(exploreDimensions),
			"total_records":       scopeInfo.TotalRecords,
			"dimension_coverage":  scopeInfo.DimensionCoverage,
		}
		executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, toolResult, nil, durationMs, true)
	}

	global.GVA_LOG.Info("数据范围探索完成",
		zap.Int("totalRecords", scopeInfo.TotalRecords),
		zap.Int("dimensionCount", len(scopeInfo.DimensionCoverage)))

	return sugarRes.NewAiSuccessResponseWithText(resultText), nil
}

// DataScopeInfo 数据范围信息结构
type DataScopeInfo struct {
	TotalRecords       int                      `json:"total_records"`       // 总记录数
	DimensionCoverage  map[string][]string      `json:"dimension_coverage"`  // 各维度的可用值列表
	SampleData         []map[string]interface{} `json:"sample_data"`         // 样本数据
	DataQualityInfo    map[string]interface{}   `json:"data_quality_info"`   // 数据质量信息
	RecommendedFilters map[string]interface{}   `json:"recommended_filters"` // 推荐的筛选条件
}

// exploreDataScope 执行数据范围探索
func (s *SugarFormulaAiService) exploreDataScope(ctx context.Context, modelName string, exploreDimensions []string, sampleFilters map[string]interface{}, userId string) (*DataScopeInfo, error) {
	global.GVA_LOG.Info("开始执行数据范围探索",
		zap.String("modelName", modelName),
		zap.Strings("exploreDimensions", exploreDimensions))

	// 构建探索查询 - 获取维度值的唯一值
	returnColumns := exploreDimensions

	// 创建查询请求
	exploreReq := &sugarReq.SugarFormulaGetRequest{
		ModelName:     modelName,
		ReturnColumns: returnColumns,
		Filters:       sampleFilters,
	}

	// 执行查询
	formulaQueryService := SugarFormulaQueryService{}
	exploreData, err := formulaQueryService.ExecuteGetFormula(ctx, exploreReq, userId)
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
	sampleSize := 10 // 保留前10条作为样本
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
	scopeInfo.DataQualityInfo["completeness"] = s.calculateDataCompleteness(exploreData.Results, exploreDimensions)
	scopeInfo.DataQualityInfo["distinct_combinations"] = s.calculateDistinctCombinations(exploreData.Results, exploreDimensions)

	// 生成推荐筛选条件
	scopeInfo.RecommendedFilters = s.generateRecommendedFilters(scopeInfo.DimensionCoverage)

	global.GVA_LOG.Info("数据范围探索分析完成",
		zap.Int("totalRecords", scopeInfo.TotalRecords),
		zap.Int("dimensionCount", len(scopeInfo.DimensionCoverage)))

	return scopeInfo, nil
}

// formatDataScopeResult 格式化数据范围探索结果
func (s *SugarFormulaAiService) formatDataScopeResult(scopeInfo *DataScopeInfo) string {
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

// calculateDataCompleteness 计算数据完整性
func (s *SugarFormulaAiService) calculateDataCompleteness(data []map[string]interface{}, dimensions []string) map[string]float64 {
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
func (s *SugarFormulaAiService) calculateDistinctCombinations(data []map[string]interface{}, dimensions []string) int {
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
func (s *SugarFormulaAiService) generateRecommendedFilters(dimensionCoverage map[string][]string) map[string]interface{} {
	recommended := make(map[string]interface{})

	for dim, values := range dimensionCoverage {
		// 如果维度值较少，推荐具体值
		if len(values) <= 5 {
			recommended[dim] = values
		} else {
			// 如果维度值较多，推荐使用前几个常见值
			recommended[dim] = fmt.Sprintf("建议从以下值中选择: %s", strings.Join(values[:3], ", "))
		}
	}

	return recommended
}

// handleSmartAnonymizedAnalyzer 处理智能匿名化分析工具调用
func (s *SugarFormulaAiService) handleSmartAnonymizedAnalyzer(ctx context.Context, toolCall system.OpenAIToolCall, logCtx *ExecutionLogContext, req *sugarReq.SugarFormulaAiFetchRequest, agent *sugar.SugarAgents, llmConfig *system.LLMConfig) (*sugarRes.SugarFormulaAiResponse, error) {
	global.GVA_LOG.Info("开始处理智能匿名化分析工具调用")

	// 记录工具调用开始时间
	toolCallStartTime := time.Now()

	// 解析工具调用参数
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		global.GVA_LOG.Error("解析智能匿名化工具参数失败", zap.Error(err))

		if logCtx != nil {
			durationMs := int(time.Since(toolCallStartTime).Milliseconds())
			errorMsg := "解析工具调用参数失败: " + err.Error()
			executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, nil, &errorMsg, durationMs, false)
		}

		return sugarRes.NewAiErrorResponse("解析工具调用参数失败: " + err.Error()), nil
	}

	// 提取参数
	modelName, _ := args["modelName"].(string)
	targetMetric, _ := args["targetMetric"].(string)
	currentPeriodFilters, _ := args["currentPeriodFilters"].(map[string]interface{})
	basePeriodFilters, _ := args["basePeriodFilters"].(map[string]interface{})
	groupByDimensionsInterface, _ := args["groupByDimensions"].([]interface{})
	userId, _ := args["userId"].(string)
	enableDataValidation, _ := args["enableDataValidation"].(bool)

	// 默认启用数据验证
	if enableDataValidation {
		enableDataValidation = true
	}

	var groupByDimensions []string
	for _, item := range groupByDimensionsInterface {
		if str, ok := item.(string); ok {
			groupByDimensions = append(groupByDimensions, str)
		}
	}

	global.GVA_LOG.Info("提取智能匿名化工具参数",
		zap.String("modelName", modelName),
		zap.String("targetMetric", targetMetric),
		zap.Strings("groupByDimensions", groupByDimensions),
		zap.Bool("enableDataValidation", enableDataValidation),
		zap.Any("currentPeriodFilters", currentPeriodFilters),
		zap.Any("basePeriodFilters", basePeriodFilters))

	// 第一步：数据范围验证（仅用于验证，不暴露原始数据）
	var validationMessage string
	if enableDataValidation {
		validationResult, err := s.validateDataAvailability(ctx, modelName, groupByDimensions, currentPeriodFilters, basePeriodFilters, userId)
		if err != nil {
			global.GVA_LOG.Error("数据可用性验证失败", zap.Error(err))

			if logCtx != nil {
				durationMs := int(time.Since(toolCallStartTime).Milliseconds())
				errorMsg := "数据可用性验证失败: " + err.Error()
				executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, nil, &errorMsg, durationMs, true)
			}

			return sugarRes.NewAiErrorResponse("数据可用性验证失败: " + err.Error()), nil
		}

		// 如果数据不可用，返回建议
		if !validationResult.IsDataAvailable {
			validationMessage = fmt.Sprintf("⚠️ 数据验证提示：%s\n\n", validationResult.ValidationMessage)
		} else {
			validationMessage = "✅ 数据验证通过，开始进行匿名化分析。\n\n"
		}
	}

	// 第二步：执行匿名化数据处理
	anonymizedResult, err := s.processAnonymizedDataAnalysis(ctx, modelName, targetMetric, currentPeriodFilters, basePeriodFilters, groupByDimensions, userId)
	if err != nil {
		global.GVA_LOG.Error("匿名化数据处理失败", zap.Error(err))

		if logCtx != nil {
			durationMs := int(time.Since(toolCallStartTime).Milliseconds())
			errorMsg := "匿名化数据处理失败: " + err.Error()
			executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, nil, &errorMsg, durationMs, true)
		}

		return sugarRes.NewAiErrorResponse("匿名化数据处理失败: " + err.Error()), nil
	}

	// 第三步：将匿名化数据转换为AI可读格式
	aiDataText, err := s.serializeAnonymizedDataToText(anonymizedResult.AIReadyData)
	if err != nil {
		return sugarRes.NewAiErrorResponse("匿名化数据序列化失败: " + err.Error()), nil
	}

	global.GVA_LOG.Info("数据已完成匿名化处理，准备发送给AI",
		zap.Int("anonymizedDataLength", len(aiDataText)),
		zap.Int("mappingCount", len(anonymizedResult.forwardMap)))

	// 更新日志记录匿名化信息
	if logCtx != nil {
		anonymizedInputData := map[string]interface{}{
			"aiDataText":        aiDataText,
			"toolCall":          toolCall.Function.Arguments,
			"mappingCount":      len(anonymizedResult.forwardMap),
			"isEncrypted":       true,
			"validationEnabled": enableDataValidation,
		}
		_ = executionLogService.UpdateExecutionLogWithAnonymization(ctx, logCtx, anonymizedInputData, nil)
	}

	// 第四步：进行AI分析（使用匿名化数据）
	global.GVA_LOG.Info("开始向AI发送匿名化数据进行分析")
	analysisResult, err := s.performDataAnalysis(ctx, aiDataText, req.Description, agent, llmConfig)
	if err != nil {
		return sugarRes.NewAiErrorResponse("AI数据分析失败: " + err.Error()), nil
	}

	global.GVA_LOG.Info("AI分析完成，准备解密响应",
		zap.Int("responseLength", len(analysisResult)))

	// 更新匿名化输出
	if logCtx != nil {
		_ = executionLogService.UpdateExecutionLogWithAnonymization(ctx, logCtx, nil, &analysisResult)
	}

	// 第五步：解密AI分析结果
	decodedResult, err := s.decodeAIResponse(anonymizedResult, analysisResult)
	if err != nil {
		global.GVA_LOG.Error("AI结果解密失败", zap.Error(err))

		if logCtx != nil {
			durationMs := int(time.Since(toolCallStartTime).Milliseconds())
			errorMsg := "AI结果解密失败: " + err.Error()
			executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, nil, &errorMsg, durationMs, true)
		}

		return sugarRes.NewAiErrorResponse("AI结果解密失败: " + err.Error()), nil
	}

	// 第六步：组合最终结果
	finalResult := validationMessage + decodedResult

	global.GVA_LOG.Info("智能匿名化分析完成，返回最终结果",
		zap.Int("finalLength", len(finalResult)))

	// 记录工具调用成功
	if logCtx != nil {
		durationMs := int(time.Since(toolCallStartTime).Milliseconds())
		toolResult := map[string]interface{}{
			"decoded_result":        decodedResult,
			"anonymized_data_count": len(anonymizedResult.AIReadyData),
			"validation_enabled":    enableDataValidation,
		}
		executionLogService.RecordToolCall(ctx, logCtx, toolCall.Function.Name, args, toolResult, nil, durationMs, true)
	}

	return sugarRes.NewAiSuccessResponseWithText(finalResult), nil
}

// DataValidationResult 数据验证结果
type DataValidationResult struct {
	IsDataAvailable   bool     `json:"is_data_available"`  // 数据是否可用
	ValidationMessage string   `json:"validation_message"` // 验证结果消息
	RecordCount       int      `json:"record_count"`       // 记录数量
	MissingDimensions []string `json:"missing_dimensions"` // 缺失的维度
}

// validateDataAvailability 验证数据可用性（不暴露原始数据）
func (s *SugarFormulaAiService) validateDataAvailability(ctx context.Context, modelName string, groupByDimensions []string, currentPeriodFilters, basePeriodFilters map[string]interface{}, userId string) (*DataValidationResult, error) {
	global.GVA_LOG.Info("开始验证数据可用性",
		zap.String("modelName", modelName),
		zap.Strings("groupByDimensions", groupByDimensions))

	result := &DataValidationResult{
		IsDataAvailable:   false,
		ValidationMessage: "",
		RecordCount:       0,
		MissingDimensions: make([]string, 0),
	}

	// 构建验证查询 - 使用实际的列进行最小化查询
	// 选择第一个分组维度作为返回列，这样可以统计记录数但不暴露敏感数据
	returnColumns := groupByDimensions[:1] // 只取第一个维度
	if len(returnColumns) == 0 {
		returnColumns = []string{"*"} // 如果没有分组维度，使用通配符
	}

	// 创建验证查询请求 - 合并筛选条件
	mergedFilters := make(map[string]interface{})
	for k, v := range currentPeriodFilters {
		mergedFilters[k] = v
	}

	validateReq := &sugarReq.SugarFormulaGetRequest{
		ModelName:     modelName,
		ReturnColumns: returnColumns,
		Filters:       mergedFilters,
	}

	// 执行验证查询
	formulaQueryService := SugarFormulaQueryService{}
	validateData, err := formulaQueryService.ExecuteGetFormula(ctx, validateReq, userId)
	if err != nil {
		return nil, fmt.Errorf("执行验证查询失败: %w", err)
	}
	if validateData.Error != "" {
		return nil, fmt.Errorf("验证查询错误: %s", validateData.Error)
	}

	// 统计实际返回的记录数
	result.RecordCount = len(validateData.Results)

	// 同时验证基期数据
	baseValidateReq := &sugarReq.SugarFormulaGetRequest{
		ModelName:     modelName,
		ReturnColumns: returnColumns,
		Filters:       basePeriodFilters,
	}

	baseValidateData, err := formulaQueryService.ExecuteGetFormula(ctx, baseValidateReq, userId)
	if err != nil {
		global.GVA_LOG.Warn("基期数据验证失败", zap.Error(err))
	} else if baseValidateData.Error != "" {
		global.GVA_LOG.Warn("基期数据查询错误", zap.String("error", baseValidateData.Error))
	}

	baseRecordCount := 0
	if baseValidateData != nil {
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

	global.GVA_LOG.Info("数据可用性验证完成",
		zap.Bool("isDataAvailable", result.IsDataAvailable),
		zap.Int("currentRecordCount", result.RecordCount),
		zap.Int("baseRecordCount", baseRecordCount),
		zap.String("message", result.ValidationMessage))

	return result, nil
}

// anonymizeNumericValue 对数值进行基础脱敏处理
func (s *SugarFormulaAiService) anonymizeNumericValue(value float64, valueType string) float64 {
	// 基础脱敏策略：
	// 1. 对于小数值（绝对值 < 1000），保留相对精度但添加小幅扰动
	// 2. 对于大数值，使用数量级保持和舍入策略
	// 3. 对于百分比类型，确保保持在合理范围内

	absValue := math.Abs(value)
	var anonymizedValue float64

	switch valueType {
	case "contribution":
		// 贡献度百分比：添加小幅随机扰动（±5%以内）
		maxPerturbation := 5.0
		perturbation := (rand.Float64() - 0.5) * 2 * maxPerturbation
		anonymizedValue = value + perturbation

		// 确保百分比在合理范围内
		if anonymizedValue > 100.0 {
			anonymizedValue = 100.0
		} else if anonymizedValue < -100.0 {
			anonymizedValue = -100.0
		}

	case "current", "base":
		// 本期值和基期值：根据数值大小应用不同脱敏策略
		if absValue < 1000 {
			// 小数值：添加5-15%的相对扰动
			perturbationRatio := 0.05 + rand.Float64()*0.10 // 5%-15%
			perturbation := absValue * perturbationRatio * (rand.Float64() - 0.5) * 2
			anonymizedValue = value + perturbation
		} else {
			// 大数值：保持数量级，添加一定扰动后舍入
			magnitude := math.Pow(10, math.Floor(math.Log10(absValue)))
			perturbationRatio := 0.10 + rand.Float64()*0.20 // 10%-30%
			perturbation := absValue * perturbationRatio * (rand.Float64() - 0.5) * 2
			anonymizedValue = value + perturbation

			// 根据数量级进行适当舍入
			if magnitude >= 1000 {
				roundTo := magnitude / 100 // 舍入到百位
				anonymizedValue = math.Round(anonymizedValue/roundTo) * roundTo
			}
		}

	case "change":
		// 变化值：保持符号一致性，但添加扰动
		if absValue < 100 {
			// 小变化值：添加10-25%扰动
			perturbationRatio := 0.10 + rand.Float64()*0.15
			perturbation := absValue * perturbationRatio * (rand.Float64() - 0.5) * 2
			anonymizedValue = value + perturbation
		} else {
			// 大变化值：添加15-35%扰动并舍入
			perturbationRatio := 0.15 + rand.Float64()*0.20
			perturbation := absValue * perturbationRatio * (rand.Float64() - 0.5) * 2
			anonymizedValue = value + perturbation

			// 舍入处理
			if absValue >= 1000 {
				anonymizedValue = math.Round(anonymizedValue/10) * 10
			}
		}

	default:
		// 默认策略：添加10%扰动
		perturbationRatio := 0.10
		perturbation := absValue * perturbationRatio * (rand.Float64() - 0.5) * 2
		anonymizedValue = value + perturbation
	}

	// 保留合理的精度（最多2位小数）
	anonymizedValue = math.Round(anonymizedValue*100) / 100

	// 安全处理除零错误
	var perturbationPercent float64
	if value != 0 {
		perturbationPercent = math.Abs((anonymizedValue-value)/value) * 100
	}

	global.GVA_LOG.Debug("数值脱敏处理",
		zap.String("valueType", valueType),
		zap.Float64("originalValue", value),
		zap.Float64("anonymizedValue", anonymizedValue),
		zap.Float64("perturbationPercent", perturbationPercent))

	return anonymizedValue
}

// TestAnonymizationEffect 测试匿名化效果的辅助方法
func (s *SugarFormulaAiService) TestAnonymizationEffect() {
	global.GVA_LOG.Info("开始测试匿名化效果")

	// 创建测试数据
	testContributions := []ContributionItem{
		{
			DimensionValues: map[string]interface{}{
				"区域名称": "济南市",
				"城市名称": "华东区",
			},
			ContributionPercent: 100.0,
			ChangeValue:         -46.89,
			CurrentValue:        19742.93,
			BaseValue:           19789.83,
			IsPositiveDriver:    true,
		},
		{
			DimensionValues: map[string]interface{}{
				"区域名称": "青岛市",
				"城市名称": "华东区",
			},
			ContributionPercent: 75.5,
			ChangeValue:         123.45,
			CurrentValue:        8567.12,
			BaseValue:           8443.67,
			IsPositiveDriver:    false,
		},
	}

	// 测试匿名化会话创建
	session, err := s.createAnonymizedSession(testContributions)
	if err != nil {
		global.GVA_LOG.Error("测试匿名化会话创建失败", zap.Error(err))
		return
	}

	// 验证匿名化结果
	global.GVA_LOG.Info("匿名化测试结果",
		zap.Int("原始数据条数", len(testContributions)),
		zap.Int("匿名化数据条数", len(session.AIReadyData)),
		zap.Int("映射关系数量", len(session.forwardMap)))

	// 检查数值是否被匿名化
	for i, aiItem := range session.AIReadyData {
		originalContrib := testContributions[i]

		// 获取匿名化后的数值
		anonContribPercent, _ := aiItem["contribution_percent"].(float64)
		anonChangeValue, _ := aiItem["change_value"].(float64)
		anonCurrentValue, _ := aiItem["current_value"].(float64)
		anonBaseValue, _ := aiItem["base_value"].(float64)

		global.GVA_LOG.Info("数值脱敏对比",
			zap.Int("itemIndex", i),
			zap.Float64("原始贡献度", originalContrib.ContributionPercent),
			zap.Float64("脱敏贡献度", anonContribPercent),
			zap.Float64("原始变化值", originalContrib.ChangeValue),
			zap.Float64("脱敏变化值", anonChangeValue),
			zap.Float64("原始本期值", originalContrib.CurrentValue),
			zap.Float64("脱敏本期值", anonCurrentValue),
			zap.Float64("原始基期值", originalContrib.BaseValue),
			zap.Float64("脱敏基期值", anonBaseValue))

		// 验证数值确实被修改了
		if anonContribPercent == originalContrib.ContributionPercent {
			global.GVA_LOG.Warn("贡献度未被脱敏", zap.Int("itemIndex", i))
		}
		if anonChangeValue == originalContrib.ChangeValue {
			global.GVA_LOG.Warn("变化值未被脱敏", zap.Int("itemIndex", i))
		}
		if anonCurrentValue == originalContrib.CurrentValue {
			global.GVA_LOG.Warn("本期值未被脱敏", zap.Int("itemIndex", i))
		}
		if anonBaseValue == originalContrib.BaseValue {
			global.GVA_LOG.Warn("基期值未被脱敏", zap.Int("itemIndex", i))
		}
	}

	// 测试序列化为文本
	aiDataText, err := s.serializeAnonymizedDataToText(session.AIReadyData)
	if err != nil {
		global.GVA_LOG.Error("序列化测试失败", zap.Error(err))
		return
	}

	global.GVA_LOG.Info("匿名化数据序列化测试通过",
		zap.Int("文本长度", len(aiDataText)),
		zap.String("预览", func() string {
			if len(aiDataText) > 200 {
				return aiDataText[:200] + "..."
			}
			return aiDataText
		}()))

	global.GVA_LOG.Info("匿名化效果测试完成")
}
