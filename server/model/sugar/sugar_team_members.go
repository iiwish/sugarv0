
// 自动生成模板SugarTeamMembers
package sugar
import (
	"time"
)

// sugarTeamMembers表 结构体  SugarTeamMembers
type SugarTeamMembers struct {
  Id  *int `json:"id" form:"id" gorm:"primarykey;column:id;size:19;"`  //id字段
  TeamId  *string `json:"teamId" form:"teamId" gorm:"column:team_id;" binding:"required"`  //teamId字段
  UserId  *string `json:"userId" form:"userId" gorm:"column:user_id;size:20;" binding:"required"`  //userId字段
  Role  string `json:"role" form:"role" gorm:"column:role;type:enum('');" binding:"required"`  //role字段
  CreatedBy  *string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:20;"`  //createdBy字段
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`  //createdAt字段
  UpdatedBy  *string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:20;"`  //updatedBy字段
  UpdatedAt  *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`  //updatedAt字段
}


// TableName sugarTeamMembers表 SugarTeamMembers自定义表名 sugar_team_members
func (SugarTeamMembers) TableName() string {
    return "sugar_team_members"
}





