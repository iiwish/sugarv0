package system

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

type SysLLMService struct{}

// LLMConfig 定义了调用LLM所需的配置
type LLMConfig struct {
	BaseURL             string `json:"baseUrl"`              // OpenAI兼容的API地址，例如 https://api.openai.com/v1
	Token               string `json:"token"`                // API密钥或访问令牌
	ModelName           string `json:"modelName"`            // 要使用的模型名称
	EnableAnonymization bool   `json:"enable_anonymization"` //
}

// ChatMessage 聊天消息结构
type ChatMessage struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"` // 消息内容
}

// ToolDefinition 工具定义结构
type ToolDefinition struct {
	Name        string                 `json:"name"`        // 工具名称
	Description string                 `json:"description"` // 工具描述
	Parameters  map[string]interface{} `json:"parameters"`  // 工具参数schema
}

// OpenAI兼容的请求结构
type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Tools       []OpenAITool  `json:"tools,omitempty"`
	ToolChoice  interface{}   `json:"tool_choice,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// OpenAI兼容的工具结构
type OpenAITool struct {
	Type     string             `json:"type"`
	Function OpenAIToolFunction `json:"function"`
}

type OpenAIToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OpenAI兼容的响应结构
type OpenAIResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []OpenAIChoice         `json:"choices"`
	Usage   map[string]interface{} `json:"usage,omitempty"`
}

type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type OpenAIMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
}

type OpenAIToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function OpenAIToolCallFunction `json:"function"`
}

type OpenAIToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatWithTools 发起一个支持工具调用的会话（直接使用OpenAI兼容接口）
func (s *SysLLMService) ChatWithTools(ctx context.Context, config LLMConfig, messages []ChatMessage, tools []ToolDefinition) (string, error) {
	return s.chatWithOpenAI(ctx, config, messages, tools)
}

// ChatSimple 发起一个简单的聊天会话（不带工具）
func (s *SysLLMService) ChatSimple(ctx context.Context, config LLMConfig, messages []ChatMessage) (string, error) {
	return s.ChatWithTools(ctx, config, messages, nil)
}

// chatWithOpenAI 使用OpenAI兼容接口进行聊天
func (s *SysLLMService) chatWithOpenAI(ctx context.Context, config LLMConfig, messages []ChatMessage, tools []ToolDefinition) (string, error) {
	// 构造请求
	request := OpenAIRequest{
		Model:    config.ModelName,
		Messages: messages,
		Tools:    s.convertToolsToOpenAI(tools),
	}

	// 如果有工具，设置tool_choice为auto
	if len(tools) > 0 {
		request.ToolChoice = "auto"
	}

	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		global.GVA_LOG.Error("序列化LLM请求失败", zap.Error(err))
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构造HTTP请求
	url := strings.TrimSuffix(config.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		global.GVA_LOG.Error("创建HTTP请求失败", zap.Error(err))
		return "", fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	if config.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+config.Token)
	}

	// 发送请求
	timeout := time.Duration(global.GVA_CONFIG.LLM.HTTP.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 180 * time.Second // 默认超时时间
	}
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		global.GVA_LOG.Error("发送HTTP请求失败", zap.Error(err))
		return "", fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		global.GVA_LOG.Error("读取LLM响应失败", zap.Error(err))
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		global.GVA_LOG.Error("LLM API请求失败",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("response", string(responseBody)))
		return "", fmt.Errorf("API请求失败，状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(responseBody, &openAIResp); err != nil {
		// 如果解析失败，可能是纯文本响应
		global.GVA_LOG.Warn("解析LLM JSON响应失败，可能为纯文本", zap.Error(err))
		return string(responseBody), nil
	}

	// 提取回复内容
	if len(openAIResp.Choices) == 0 {
		return "", errors.New("API未返回任何选择")
	}

	choice := openAIResp.Choices[0]

	// 如果有工具调用，将完整的工具调用信息序列化为JSON字符串返回
	if len(choice.Message.ToolCalls) > 0 {
		toolCallsJSON, err := json.Marshal(choice.Message.ToolCalls)
		if err != nil {
			global.GVA_LOG.Error("序列化工具调用失败", zap.Error(err))
			return "", fmt.Errorf("序列化工具调用失败: %w", err)
		}
		// 返回一个可识别的JSON结构，表示这是一个工具调用
		return fmt.Sprintf(`{"type": "tool_call", "content": %s}`, string(toolCallsJSON)), nil
	}
	// 如果是普通文本消息，直接返回内容
	return choice.Message.Content, nil
}

// ParseLLMConfigFromJSON 从JSON字符串解析LLM配置
func (s *SysLLMService) ParseLLMConfigFromJSON(configJSON string) (*LLMConfig, error) {
	var config LLMConfig
	err := json.Unmarshal([]byte(configJSON), &config)
	if err != nil {
		return nil, fmt.Errorf("解析LLM配置失败: %w", err)
	}

	// 验证必要字段
	if config.BaseURL == "" {
		return nil, errors.New("BaseURL不能为空")
	}
	if config.ModelName == "" {
		return nil, errors.New("ModelName不能为空")
	}

	return &config, nil
}

// GetDefaultLLMConfig 获取默认的LLM配置（从全局配置中读取）
func (s *SysLLMService) GetDefaultLLMConfig() *LLMConfig {
	config := global.GVA_CONFIG.LLM.OpenAI
	return &LLMConfig{
		BaseURL:   config.BaseURL,
		Token:     config.Token,
		ModelName: config.ModelName,
	}
}

// convertToolsToOpenAI 将ToolDefinition转换为OpenAI格式
func (s *SysLLMService) convertToolsToOpenAI(tools []ToolDefinition) []OpenAITool {
	if tools == nil {
		return nil
	}

	openAITools := make([]OpenAITool, len(tools))
	for i, tool := range tools {
		openAITools[i] = OpenAITool{
			Type: "function",
			Function: OpenAIToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		}
	}
	return openAITools
}
