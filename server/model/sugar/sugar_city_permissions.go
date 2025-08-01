
// 自动生成模板SugarCityPermissions
package sugar
import (
	"time"
)

// sugarCityPermissions表 结构体  SugarCityPermissions
type SugarCityPermissions struct {
  Id  *int `json:"id" form:"id" gorm:"primarykey;column:id;size:19;"`  //id字段
  UserId  *string `json:"userId" form:"userId" gorm:"column:user_id;size:20;"`  //userId字段
  CityCode  *string `json:"cityCode" form:"cityCode" gorm:"comment:城市编码;column:city_code;size:50;"`  //城市编码
  CreatedBy  *string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:20;"`  //createdBy字段
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`  //createdAt字段
  UpdatedBy  *string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:20;"`  //updatedBy字段
  UpdatedAt  *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`  //updatedAt字段
  DeletedAt  *time.Time `json:"deletedAt" form:"deletedAt" gorm:"column:deleted_at;"`  //deletedAt字段
}


// TableName sugarCityPermissions表 SugarCityPermissions自定义表名 sugar_city_permissions
func (SugarCityPermissions) TableName() string {
    return "sugar_city_permissions"
}





