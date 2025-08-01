
// 自动生成模板SugarAgents
package sugar
import (
	"time"
)

// sugar智能体表 结构体  SugarAgents
type SugarAgents struct {
  Id  *string `json:"id" form:"id" gorm:"primarykey;column:id;"`  //id字段
  Name  *string `json:"name" form:"name" gorm:"column:name;size:100;"`  //name字段
  Description  *string `json:"description" form:"description" gorm:"column:description;"`  //description字段
  AgentType  string `json:"agentType" form:"agentType" gorm:"comment:系统预置, 团队自定义;column:agent_type;type:enum('');"`  //系统预置, 团队自定义
  TeamId  *string `json:"teamId" form:"teamId" gorm:"column:team_id;"`  //teamId字段
  EndpointConfig  string `json:"endpointConfig" form:"endpointConfig" gorm:"comment:定义 Agent 的调用方式, 如 API URL, headers 等;column:endpoint_config;type:enum('');"`  //定义 Agent 的调用方式, 如 API URL, headers 等
  CreatedBy  *string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:20;"`  //createdBy字段
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`  //createdAt字段
  UpdatedBy  *string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:20;"`  //updatedBy字段
  UpdatedAt  *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`  //updatedAt字段
  DeletedAt  *time.Time `json:"deletedAt" form:"deletedAt" gorm:"column:deleted_at;"`  //deletedAt字段
}


// TableName sugar智能体表 SugarAgents自定义表名 sugar_agents
func (SugarAgents) TableName() string {
    return "sugar_agents"
}





