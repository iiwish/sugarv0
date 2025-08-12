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
	// 1. 获取Agent信息
	agent, err := s.getAgentByName(ctx, req.AgentName, userId)
	if err != nil {
		return sugarRes.NewAiErrorResponse(err.Error()), nil
	}

	// 2. 获取LLM配置
	var llmConfig *system.LLMConfig
	if agent.EndpointConfig != "" {
		llmConfig, err = llmService.ParseLLMConfigFromJSON(agent.EndpointConfig)
		if err != nil {
			global.GVA_LOG.Warn("解析Agent LLM配置失败，使用默认LLM配置", zap.Error(err))
			llmConfig = llmService.GetDefaultLLMConfig()
		}
	} else {
		llmConfig = llmService.GetDefaultLLMConfig()
	}

	// 3. 构建系统提示词
	systemPrompt := s.buildSystemPrompt(agent)

	// 4. 构建用户消息
	userMessage := s.buildUserMessage(req.Description, agent.Semantic)

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
	llmResponse, err := llmService.ChatWithTools(ctx, *llmConfig, messages, tools)
	if err != nil {
		global.GVA_LOG.Error("AIFETCH LLM调用失败", zap.Error(err), zap.String("agent", req.AgentName))
		return sugarRes.NewAiErrorResponse("AI分析失败: " + err.Error()), nil
	}

	// 8. 解析响应并处理可能的工具调用
	return s.processAiFetchResponse(ctx, llmResponse, userId)
}

// processAiFetchResponse 处理AIFETCH的响应，可能包含工具调用
func (s *SugarFormulaAiService) processAiFetchResponse(ctx context.Context, llmResponse string, userId string) (*sugarRes.SugarFormulaAiResponse, error) {
	var toolCallResp ToolCallResponse
	err := json.Unmarshal([]byte(llmResponse), &toolCallResp)

	// 如果解析失败或不是工具调用，则认为是普通文本响应
	if err != nil || toolCallResp.Type != "tool_call" {
		return sugarRes.NewAiSuccessResponseWithText(llmResponse), nil
	}

	// 处理工具调用
	if len(toolCallResp.Content) > 0 {
		toolCall := toolCallResp.Content[0]
		if toolCall.Function.Name == "semantic_data_fetcher" {
			// 解析工具调用参数
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
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

			// 构建并执行SUGAR.GET请求
			getRequest := &sugarReq.SugarFormulaGetRequest{
				ModelName:     modelName,
				ReturnColumns: returnColumns,
				Filters:       filters,
			}

			formulaQueryService := SugarFormulaQueryService{}
			getResult, err := formulaQueryService.ExecuteGetFormula(ctx, getRequest, userId)
			if err != nil {
				return sugarRes.NewAiErrorResponse("内部数据查询失败: " + err.Error()), nil
			}
			if getResult.Error != "" {
				return sugarRes.NewAiErrorResponse("内部数据查询失败: " + getResult.Error), nil
			}

			// 将查询结果包装成AI Response
			data := s.convertMapsToSlice(getResult.Results, getResult.Columns)
			return sugarRes.NewAiSuccessResponseWithData(data), nil
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
	// 1. 序列化数据为可读格式
	dataText, err := s.serializeDataToText(req.DataSource)
	if err != nil {
		return sugarRes.NewAiErrorResponse("数据序列化失败: " + err.Error()), nil
	}

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
		systemPrompt = s.buildSystemPrompt(agent)
	}

	// 3. 构建用户消息
	userMessage := fmt.Sprintf("请分析以下数据：\n\n%s\n\n分析要求：%s", dataText, req.Description)

	// 4. 构建消息列表
	messages := []system.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	// 5. 直接调用OpenAI兼容接口（不带工具）
	response, err := llmService.ChatSimple(ctx, *llmConfig, messages)
	if err != nil {
		global.GVA_LOG.Error("AIEXPLAIN OpenAI调用失败", zap.Error(err))
		return sugarRes.NewAiErrorResponse("AI分析失败: " + err.Error()), nil
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

// buildSystemPrompt 构建系统提示词
func (s *SugarFormulaAiService) buildSystemPrompt(agent *sugar.SugarAgents) string {
	if agent.Prompt != nil && *agent.Prompt != "" {
		return *agent.Prompt
	}
	return "你是一个专业的数据分析师，请根据用户的需求进行数据分析。"
}

// buildUserMessage 构建用户消息
func (s *SugarFormulaAiService) buildUserMessage(description string, semantic *string) string {
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
		if column, ok := paramConfig["column"].(string); ok {
			builder.WriteString(fmt.Sprintf(" [字段: %s]", column))
		}
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
		if column, ok := columnConfig["column"].(string); ok {
			builder.WriteString(fmt.Sprintf(" [字段: %s]", column))
		}
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
