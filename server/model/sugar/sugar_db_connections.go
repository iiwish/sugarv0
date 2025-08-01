
// 自动生成模板SugarDbConnections
package sugar
import (
	"time"
)

// Sugar数据库配置表 结构体  SugarDbConnections
type SugarDbConnections struct {
  Id  *string `json:"id" form:"id" gorm:"primarykey;column:id;"`  //id字段
  Name  *string `json:"name" form:"name" gorm:"column:name;size:100;"`  //name字段
  TeamId  *string `json:"teamId" form:"teamId" gorm:"column:team_id;"`  //teamId字段
  DbType  *string `json:"dbType" form:"dbType" gorm:"column:db_type;size:20;"`  //dbType字段
  IsInternal  *bool `json:"isInternal" form:"isInternal" gorm:"comment:是否为Sugar内部数据库 (true代表本库);column:is_internal;"`  //是否为Sugar内部数据库 (true代表本库)
  Host  *string `json:"host" form:"host" gorm:"comment:内部数据库可为null;column:host;size:255;"`  //内部数据库可为null
  Port  *int `json:"port" form:"port" gorm:"comment:内部数据库可为null;column:port;size:10;"`  //内部数据库可为null
  Username  *string `json:"username" form:"username" gorm:"comment:内部数据库可为null;column:username;size:100;"`  //内部数据库可为null
  EncryptedPassword  *string `json:"encryptedPassword" form:"encryptedPassword" gorm:"comment:内部数据库可为null;column:encrypted_password;"`  //内部数据库可为null
  DatabaseName  *string `json:"databaseName" form:"databaseName" gorm:"column:database_name;size:100;"`  //databaseName字段
  SslConfig  string `json:"sslConfig" form:"sslConfig" gorm:"column:ssl_config;type:enum('');"`  //sslConfig字段
  CreatedBy  *string `json:"createdBy" form:"createdBy" gorm:"column:created_by;size:20;"`  //createdBy字段
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"column:created_at;"`  //createdAt字段
  UpdatedBy  *string `json:"updatedBy" form:"updatedBy" gorm:"column:updated_by;size:20;"`  //updatedBy字段
  UpdatedAt  *time.Time `json:"updatedAt" form:"updatedAt" gorm:"column:updated_at;"`  //updatedAt字段
  DeletedAt  *time.Time `json:"deletedAt" form:"deletedAt" gorm:"column:deleted_at;"`  //deletedAt字段
}


// TableName Sugar数据库配置表 SugarDbConnections自定义表名 sugar_db_connections
func (SugarDbConnections) TableName() string {
    return "sugar_db_connections"
}





