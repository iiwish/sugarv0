
// 自动生成模板SugarRowLevelOverrides
package sugar
import (
	"time"
)

// Sugar行级权限豁免表 结构体  SugarRowLevelOverrides
type SugarRowLevelOverrides struct {
  Id  *int `json:"id" form:"id" gorm:"primarykey;column:id;size:19;"`  //id字段
  UserId  *string `json:"userId" form:"userId" gorm:"column:user_id;size:20;"`  //userId字段
  Description  *string `json:"description" form:"description" gorm:"comment:配置原因说明;column:description;size:255;"`  //配置原因说明
  CreatedBy  *string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:20;"`  //createdBy字段
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`  //createdAt字段
  UpdatedBy  *string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:20;"`  //updatedBy字段
  UpdatedAt  *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`  //updatedAt字段
  DeletedAt  *time.Time `json:"deletedAt" form:"deletedAt" gorm:"column:deleted_at;"`  //deletedAt字段
}


// TableName Sugar行级权限豁免表 SugarRowLevelOverrides自定义表名 sugar_row_level_overrides
func (SugarRowLevelOverrides) TableName() string {
    return "sugar_row_level_overrides"
}





