package sugar

import (
	"context"
	"encoding/json"
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

type SugarFormulaAiService struct{}

var llmService = system.SysLLMService{}

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

	// 4. 构建用户消息
	userMessage := s.buildUserMessage(req.Description, agent.Semantic, req.DataRange)
	global.GVA_LOG.Debug("构建用户消息", zap.String("userMessage", userMessage))

	// 5. 准备工具定义（硬编码）
	tools := []system.ToolDefinition{
		{
			Name:        "semantic_data_fetcher",
			Description: "根据指定的语义模型、维度、度量和筛选条件，从数据源获取数据。这是AIFETCH公式的内部实现工具。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"modelName": map[string]interface{}{
						"type":        "string",
						"description": "要查询的语义模型名称。",
					},
					"returnColumns": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "需要返回的列名数组，可以是维度或度量。",
					},
					"filters": map[string]interface{}{
						"type":        "object",
						"description": "筛选条件，格式为 {\"列名\": \"筛选值\"}。",
					},
					"userId": map[string]interface{}{
						"type":        "string",
						"description": "发起请求的用户ID，工具内部需要此参数进行鉴权。",
					},
				},
				"required": []string{"modelName", "returnColumns", "userId"},
			},
		},
	}

	// 6. 构建消息列表
	messages := []system.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	// 7. 调用LLM，传入工具定义
	global.GVA_LOG.Info("开始调用LLM",
		zap.String("model", llmConfig.ModelName),
		zap.Int("toolsCount", len(tools)),
		zap.Int("messagesCount", len(messages)))

	llmResponse, err := llmService.ChatWithTools(ctx, *llmConfig, messages, tools)
	if err != nil {
		global.GVA_LOG.Error("AIFETCH LLM调用失败", zap.Error(err), zap.String("agent", req.AgentName))
		return sugarRes.NewAiErrorResponse("AI分析失败: " + err.Error()), nil
	}

	global.GVA_LOG.Info("LLM调用成功", zap.String("responseLength", fmt.Sprintf("%d", len(llmResponse))))
	global.GVA_LOG.Debug("LLM原始响应", zap.String("llmResponse", llmResponse))

	// 8. 解析响应并处理可能的工具调用
	return s.processAiFetchResponse(ctx, llmResponse, userId, req, agent, llmConfig)
}

// processAiFetchResponse 处理AIFETCH的响应，可能包含工具调用
func (s *SugarFormulaAiService) processAiFetchResponse(ctx context.Context, llmResponse string, userId string, req *sugarReq.SugarFormulaAiFetchRequest, agent *sugar.SugarAgents, llmConfig *system.LLMConfig) (*sugarRes.SugarFormulaAiResponse, error) {
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

		if toolCall.Function.Name == "semantic_data_fetcher" {
			// 解析工具调用参数
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				global.GVA_LOG.Error("解析工具调用参数失败", zap.Error(err), zap.String("arguments", toolCall.Function.Arguments))
				return sugarRes.NewAiErrorResponse("解析工具调用参数失败: " + err.Error()), nil
			}

			// 提取参数
			modelName, _ := args["modelName"].(string)
			returnColumnsInterface, _ := args["returnColumns"].([]interface{})
			filters, _ := args["filters"].(map[string]interface{})

			var returnColumns []string
			for _, item := range returnColumnsInterface {
				if str, ok := item.(string); ok {
					returnColumns = append(returnColumns, str)
				}
			}

			global.GVA_LOG.Info("提取工具调用参数",
				zap.String("modelName", modelName),
				zap.Strings("returnColumns", returnColumns),
				zap.Any("filters", filters))

			// 构建并执行SUGAR.GET请求
			getRequest := &sugarReq.SugarFormulaGetRequest{
				ModelName:     modelName,
				ReturnColumns: returnColumns,
				Filters:       filters,
			}

			global.GVA_LOG.Info("开始执行内部数据查询", zap.String("modelName", modelName))
			formulaQueryService := SugarFormulaQueryService{}
			getResult, err := formulaQueryService.ExecuteGetFormula(ctx, getRequest, userId)
			if err != nil {
				global.GVA_LOG.Error("内部数据查询失败", zap.Error(err))
				return sugarRes.NewAiErrorResponse("内部数据查询失败: " + err.Error()), nil
			}
			if getResult.Error != "" {
				global.GVA_LOG.Error("内部数据查询返回错误", zap.String("error", getResult.Error))
				return sugarRes.NewAiErrorResponse("内部数据查询失败: " + getResult.Error), nil
			}

			global.GVA_LOG.Info("内部数据查询成功",
				zap.Int("resultCount", len(getResult.Results)),
				zap.Strings("columns", getResult.Columns))

			// 将查询结果转换为文本格式，供AI分析
			data := s.convertMapsToSlice(getResult.Results, getResult.Columns)
			dataText, err := s.serializeDataToText(data)
			if err != nil {
				return sugarRes.NewAiErrorResponse("数据序列化失败: " + err.Error()), nil
			}

			// 进行二次AI分析
			analysisResult, err := s.performDataAnalysis(ctx, dataText, req.Description, agent, llmConfig)
			if err != nil {
				return sugarRes.NewAiErrorResponse("AI数据分析失败: " + err.Error()), nil
			}

			return sugarRes.NewAiSuccessResponseWithText(analysisResult), nil
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

	// 1. 序列化数据为可读格式
	dataText, err := s.serializeDataToText(req.DataSource)
	if err != nil {
		global.GVA_LOG.Error("数据序列化失败", zap.Error(err))
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

	// 3. 构建用户消息
	userMessage := fmt.Sprintf("请分析以下数据：\n\n%s\n\n分析要求：%s", dataText, req.Description)
	global.GVA_LOG.Debug("构建用户消息", zap.String("userMessage", userMessage))

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
		return sugarRes.NewAiErrorResponse("AI分析失败: " + err.Error()), nil
	}

	global.GVA_LOG.Info("AIEXPLAIN分析完成", zap.String("responseLength", fmt.Sprintf("%d", len(response))))
	global.GVA_LOG.Debug("AIEXPLAIN响应内容", zap.String("response", response))
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

// buildSystemPrompt 构建系统提示词
func (s *SugarFormulaAiService) buildSystemPrompt(agent *sugar.SugarAgents, userId string) string {
	basePrompt := ""
	if agent.Prompt != nil && *agent.Prompt != "" {
		basePrompt = *agent.Prompt
	} else {
		basePrompt = "你是一个专业的数据分析师，请根据用户的需求进行数据分析。"
	}

	// 在系统提示词中补充用户ID信息，供工具调用时使用
	return basePrompt + fmt.Sprintf("\n\n重要提示：当前用户ID为 %s，在调用 semantic_data_fetcher 工具时，请务必将此用户ID作为 userId 参数传递。", userId)
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

// performDataAnalysis 对获取的数据进行AI分析
func (s *SugarFormulaAiService) performDataAnalysis(ctx context.Context, dataText string, userDescription string, agent *sugar.SugarAgents, llmConfig *system.LLMConfig) (string, error) {
	global.GVA_LOG.Info("开始执行数据分析",
		zap.String("userDescription", userDescription),
		zap.String("dataLength", fmt.Sprintf("%d", len(dataText))))

	// 构建分析提示词
	systemPrompt := s.buildSystemPrompt(agent, "")
	global.GVA_LOG.Debug("构建分析系统提示词", zap.String("systemPrompt", systemPrompt))

	// 构建用户消息，包含数据和分析要求
	userMessage := fmt.Sprintf("基于以下数据进行分析：\n\n%s\n\n用户的分析需求：%s\n\n请提供详细的分析结论和洞察。", dataText, userDescription)
	global.GVA_LOG.Debug("构建分析用户消息", zap.String("userMessage", userMessage))

	// 构建消息列表
	messages := []system.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	// 调用LLM进行分析
	global.GVA_LOG.Info("开始调用LLM进行数据分析", zap.String("model", llmConfig.ModelName))
	response, err := llmService.ChatSimple(ctx, *llmConfig, messages)
	if err != nil {
		global.GVA_LOG.Error("AI数据分析调用失败", zap.Error(err))
		return "", fmt.Errorf("AI分析失败: %w", err)
	}

	global.GVA_LOG.Info("数据分析完成", zap.String("responseLength", fmt.Sprintf("%d", len(response))))
	global.GVA_LOG.Debug("数据分析响应", zap.String("response", response))
	return response, nil
}
