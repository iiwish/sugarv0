
// 自动生成模板SugarExecutionLogs
package sugar
import (
	"time"
	"gorm.io/datatypes"
)

// sugar操作日志表 结构体  SugarExecutionLogs
type SugarExecutionLogs struct {
  Id  *int `json:"id" form:"id" gorm:"primarykey;column:id;size:19;"`  //id字段
  LogType  string `json:"logType" form:"logType" gorm:"column:log_type;type:enum('');"`  //logType字段
  WorkspaceId  *string `json:"workspaceId" form:"workspaceId" gorm:"column:workspace_id;"`  //workspaceId字段
  UserId  *string `json:"userId" form:"userId" gorm:"column:user_id;size:20;"`  //userId字段
  ConnectionId  *string `json:"connectionId" form:"connectionId" gorm:"column:connection_id;"`  //connectionId字段
  AgentId  *string `json:"agentId" form:"agentId" gorm:"column:agent_id;"`  //agentId字段
  InputPayload  datatypes.JSON `json:"inputPayload" form:"inputPayload" gorm:"column:input_payload;" swaggertype:"object"`  //inputPayload字段
  Status  string `json:"status" form:"status" gorm:"column:status;type:enum('');"`  //status字段
  ResultSummary  *string `json:"resultSummary" form:"resultSummary" gorm:"column:result_summary;"`  //resultSummary字段
  DurationMs  *int `json:"durationMs" form:"durationMs" gorm:"column:duration_ms;size:10;"`  //durationMs字段
  ExecutedAt  *time.Time `json:"executedAt" form:"executedAt" gorm:"column:executed_at;"`  //executedAt字段
}


// TableName sugar操作日志表 SugarExecutionLogs自定义表名 sugar_execution_logs
func (SugarExecutionLogs) TableName() string {
    return "sugar_execution_logs"
}





