
// 自动生成模板SugarWorkspaces
package sugar
import (
	"time"
	"gorm.io/datatypes"
)

// Sugar文件列表 结构体  SugarWorkspaces
type SugarWorkspaces struct {
  Id  *string `json:"id" form:"id" gorm:"primarykey;column:id;"`  //id字段
  Name  *string `json:"name" form:"name" gorm:"column:name;size:255;"`  //name字段
  Type  string `json:"type" form:"type" gorm:"column:type;type:enum('');"`  //type字段
  ParentId  *string `json:"parentId" form:"parentId" gorm:"column:parent_id;"`  //parentId字段
  TeamId  *string `json:"teamId" form:"teamId" gorm:"comment:资源统一归属于团队;column:team_id;"`  //资源统一归属于团队
  Content  datatypes.JSON `json:"content" form:"content" gorm:"column:content;" swaggertype:"object"`  //content字段
  CreatedBy  *string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:20;"`  //createdBy字段
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`  //createdAt字段
  UpdatedBy  *string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:20;"`  //updatedBy字段
  UpdatedAt  *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`  //updatedAt字段
  DeletedAt  *time.Time `json:"deletedAt" form:"deletedAt" gorm:"column:deleted_at;"`  //deletedAt字段
}


// TableName Sugar文件列表 SugarWorkspaces自定义表名 sugar_workspaces
func (SugarWorkspaces) TableName() string {
    return "sugar_workspaces"
}





