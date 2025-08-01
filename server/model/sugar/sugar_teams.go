
// 自动生成模板SugarTeams
package sugar
import (
	"time"
)

// 团队信息表 结构体  SugarTeams
type SugarTeams struct {
  Id  *string `json:"id" form:"id" gorm:"primarykey;column:id;"`  //id字段
  TeamName  *string `json:"teamName" form:"teamName" gorm:"column:team_name;size:100;"`  //teamName字段
  OwnerId  *string `json:"ownerId" form:"ownerId" gorm:"comment:团队创建者/个人空间的所有者;column:owner_id;size:20;"`  //团队创建者/个人空间的所有者
  IsPersonal  *bool `json:"isPersonal" form:"isPersonal" gorm:"comment:是否为个人空间团队 (true代表个人空间);column:is_personal;"`  //是否为个人空间团队 (true代表个人空间)
  CreatedBy  *string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:20;"`  //createdBy字段
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`  //createdAt字段
  UpdatedBy  *string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:20;"`  //updatedBy字段
  UpdatedAt  *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`  //updatedAt字段
}


// TableName 团队信息表 SugarTeams自定义表名 sugar_teams
func (SugarTeams) TableName() string {
    return "sugar_teams"
}





