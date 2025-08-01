
package sugar

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarDbConnectionsService struct {}
// CreateSugarDbConnections 创建Sugar数据库配置表记录
// Author [yourname](https://github.com/yourname)
func (sugarDbConnectionsService *SugarDbConnectionsService) CreateSugarDbConnections(ctx context.Context, sugarDbConnections *sugar.SugarDbConnections) (err error) {
	err = global.GVA_DB.Create(sugarDbConnections).Error
	return err
}

// DeleteSugarDbConnections 删除Sugar数据库配置表记录
// Author [yourname](https://github.com/yourname)
func (sugarDbConnectionsService *SugarDbConnectionsService)DeleteSugarDbConnections(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&sugar.SugarDbConnections{},"id = ?",id).Error
	return err
}

// DeleteSugarDbConnectionsByIds 批量删除Sugar数据库配置表记录
// Author [yourname](https://github.com/yourname)
func (sugarDbConnectionsService *SugarDbConnectionsService)DeleteSugarDbConnectionsByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]sugar.SugarDbConnections{},"id in ?",ids).Error
	return err
}

// UpdateSugarDbConnections 更新Sugar数据库配置表记录
// Author [yourname](https://github.com/yourname)
func (sugarDbConnectionsService *SugarDbConnectionsService)UpdateSugarDbConnections(ctx context.Context, sugarDbConnections sugar.SugarDbConnections) (err error) {
	err = global.GVA_DB.Model(&sugar.SugarDbConnections{}).Where("id = ?",sugarDbConnections.Id).Updates(&sugarDbConnections).Error
	return err
}

// GetSugarDbConnections 根据id获取Sugar数据库配置表记录
// Author [yourname](https://github.com/yourname)
func (sugarDbConnectionsService *SugarDbConnectionsService)GetSugarDbConnections(ctx context.Context, id string) (sugarDbConnections sugar.SugarDbConnections, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sugarDbConnections).Error
	return
}
// GetSugarDbConnectionsInfoList 分页获取Sugar数据库配置表记录
// Author [yourname](https://github.com/yourname)
func (sugarDbConnectionsService *SugarDbConnectionsService)GetSugarDbConnectionsInfoList(ctx context.Context, info sugarReq.SugarDbConnectionsSearch) (list []sugar.SugarDbConnections, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&sugar.SugarDbConnections{})
    var sugarDbConnectionss []sugar.SugarDbConnections
    // 如果有条件搜索 下方会自动创建搜索语句
    
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&sugarDbConnectionss).Error
	return  sugarDbConnectionss, total, err
}
func (sugarDbConnectionsService *SugarDbConnectionsService)GetSugarDbConnectionsPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
