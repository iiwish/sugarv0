// 自动生成模板SugarExecutionLogs
package sugar

import (
	"time"

	"gorm.io/datatypes"
)

type SugarExecutionLogs struct {
	Id                   int64          `json:"id" form:"id" gorm:"primaryKey;column:id;autoIncrement;type:bigint;comment:id字段"`
	LogType              string         `json:"logType" form:"logType" gorm:"column:log_type;type:enum('db_query','ai_agent');not null;comment:日志类型: 数据库查询或AI Agent调用"`
	WorkspaceId          *string        `json:"workspaceId" form:"workspaceId" gorm:"column:workspace_id;type:char(36);comment:关联的工作区ID"`
	UserId               *string        `json:"userId" form:"userId" gorm:"column:user_id;type:varchar(20);comment:发起操作的用户ID"`
	ConnectionId         *string        `json:"connectionId" form:"connectionId" gorm:"column:connection_id;type:char(36);comment:如果适用，关联的数据库连接ID"`
	AgentId              *string        `json:"agentId" form:"agentId" gorm:"column:agent_id;type:char(36);comment:如果适用，关联的AI Agent ID"`
	InputPayload         datatypes.JSON `json:"inputPayload" form:"inputPayload" gorm:"column:input_payload;type:json;comment:原始的、未经处理的输入负载" swaggertype:"object"`
	Status               string         `json:"status" form:"status" gorm:"column:status;type:enum('pending','success','failed','timeout');not null;comment:任务执行状态"`
	FinalResult          *string        `json:"finalResult" form:"finalResult" gorm:"column:final_result;type:text;comment:最终返回给用户的、已解码的可读结果"`
	DurationMs           *int           `json:"durationMs" form:"durationMs" gorm:"column:duration_ms;type:int;comment:任务执行耗时（毫秒）"`
	AnonymizationEnabled bool           `json:"anonymizationEnabled" form:"anonymizationEnabled" gorm:"column:anonymization_enabled;type:boolean;not null;default:false;comment:标记本次调用是否启动了数据匿名化流程"`
	AnonymizedInput      datatypes.JSON `json:"anonymizedInput" form:"anonymizedInput" gorm:"column:anonymized_input;type:json;comment:匿名化后，实际发送给AI模型的输入（包括数据和提示词）" swaggertype:"object"`
	AnonymizedOutput     *string        `json:"anonymizedOutput" form:"anonymizedOutput" gorm:"column:anonymized_output;type:text;comment:从AI模型收到的、解码前的原始输出"`
	ErrorMessage         *string        `json:"errorMessage" form:"errorMessage" gorm:"column:error_message;type:text;comment:如果执行失败，记录错误信息"`
	ExecutedAt           time.Time      `json:"executedAt" form:"executedAt" gorm:"column:executed_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:执行时间"`

	// AI交互详细记录字段
	SystemPrompt   *string        `json:"systemPrompt" form:"systemPrompt" gorm:"column:system_prompt;type:text;comment:AI系统提示词"`
	UserMessage    *string        `json:"userMessage" form:"userMessage" gorm:"column:user_message;type:text;comment:用户输入消息"`
	LlmConfig      datatypes.JSON `json:"llmConfig" form:"llmConfig" gorm:"column:llm_config;type:json;comment:使用的LLM配置信息" swaggertype:"object"`
	AiInteractions datatypes.JSON `json:"aiInteractions" form:"aiInteractions" gorm:"column:ai_interactions;type:json;comment:完整的AI交互记录，包含所有消息和响应" swaggertype:"object"`
	ToolCalls      datatypes.JSON `json:"toolCalls" form:"toolCalls" gorm:"column:tool_calls;type:json;comment:AI工具调用详情记录" swaggertype:"object"`
	RawLlmResponse *string        `json:"rawLlmResponse" form:"rawLlmResponse" gorm:"column:raw_llm_response;type:text;comment:LLM的原始响应内容"`
	TokenUsage     datatypes.JSON `json:"tokenUsage" form:"tokenUsage" gorm:"column:token_usage;type:json;comment:Token使用统计信息" swaggertype:"object"`
}

// TableName sugar操作日志表 SugarExecutionLogs自定义表名 sugar_execution_logs
func (SugarExecutionLogs) TableName() string {
	return "sugar_execution_logs"
}
