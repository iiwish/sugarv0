
// 自动生成模板SugarSemanticModels
package sugar
import (
	"time"
	"gorm.io/datatypes"
)

// Sugar指标语义表 结构体  SugarSemanticModels
type SugarSemanticModels struct {
  Id  *string `json:"id" form:"id" gorm:"primarykey;column:id;"`  //id字段
  Name  *string `json:"name" form:"name" gorm:"comment:模型的业务名称, 如“季度销售报告”;column:name;size:100;"`  //模型的业务名称, 如“季度销售报告”
  Description  *string `json:"description" form:"description" gorm:"column:description;"`  //description字段
  TeamId  *string `json:"teamId" form:"teamId" gorm:"column:team_id;"`  //teamId字段
  ConnectionId  *string `json:"connectionId" form:"connectionId" gorm:"comment:关联的数据库连接;column:connection_id;"`  //关联的数据库连接
  SourceTableName  *string `json:"sourceTableName" form:"sourceTableName" gorm:"comment:源数据库中的真实表名;column:source_table_name;size:255;"`  //源数据库中的真实表名
  ParameterConfig  datatypes.JSON `json:"parameterConfig" form:"parameterConfig" gorm:"comment:查询参数配置, 定义用户可用的筛选条件;column:parameter_config;" swaggertype:"object"`  //查询参数配置, 定义用户可用的筛选条件
  ReturnableColumnsConfig  datatypes.JSON `json:"returnableColumnsConfig" form:"returnableColumnsConfig" gorm:"comment:可返回字段配置, 定义用户可获取的数据列;column:returnable_columns_config;" swaggertype:"object"`  //可返回字段配置, 定义用户可获取的数据列
  PermissionKeyColumn  *string `json:"permissionKeyColumn" form:"permissionKeyColumn" gorm:"comment:用于行级权限判断的字段名, 如 city_code;column:permission_key_column;size:255;"`  //用于行级权限判断的字段名, 如 city_code
  CreatedBy  *string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:20;"`  //createdBy字段
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`  //createdAt字段
  UpdatedBy  *string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:20;"`  //updatedBy字段
  UpdatedAt  *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`  //updatedAt字段
  DeletedAt  *time.Time `json:"deletedAt" form:"deletedAt" gorm:"column:deleted_at;"`  //deletedAt字段
}


// TableName Sugar指标语义表 SugarSemanticModels自定义表名 sugar_semantic_models
func (SugarSemanticModels) TableName() string {
    return "sugar_semantic_models"
}





