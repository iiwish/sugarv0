
package sugar

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarWorkspacesService struct {}
// CreateSugarWorkspaces 创建Sugar文件列表记录
// Author [yourname](https://github.com/yourname)
func (sugarWorkspacesService *SugarWorkspacesService) CreateSugarWorkspaces(ctx context.Context, sugarWorkspaces *sugar.SugarWorkspaces) (err error) {
	err = global.GVA_DB.Create(sugarWorkspaces).Error
	return err
}

// DeleteSugarWorkspaces 删除Sugar文件列表记录
// Author [yourname](https://github.com/yourname)
func (sugarWorkspacesService *SugarWorkspacesService)DeleteSugarWorkspaces(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&sugar.SugarWorkspaces{},"id = ?",id).Error
	return err
}

// DeleteSugarWorkspacesByIds 批量删除Sugar文件列表记录
// Author [yourname](https://github.com/yourname)
func (sugarWorkspacesService *SugarWorkspacesService)DeleteSugarWorkspacesByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]sugar.SugarWorkspaces{},"id in ?",ids).Error
	return err
}

// UpdateSugarWorkspaces 更新Sugar文件列表记录
// Author [yourname](https://github.com/yourname)
func (sugarWorkspacesService *SugarWorkspacesService)UpdateSugarWorkspaces(ctx context.Context, sugarWorkspaces sugar.SugarWorkspaces) (err error) {
	err = global.GVA_DB.Model(&sugar.SugarWorkspaces{}).Where("id = ?",sugarWorkspaces.Id).Updates(&sugarWorkspaces).Error
	return err
}

// GetSugarWorkspaces 根据id获取Sugar文件列表记录
// Author [yourname](https://github.com/yourname)
func (sugarWorkspacesService *SugarWorkspacesService)GetSugarWorkspaces(ctx context.Context, id string) (sugarWorkspaces sugar.SugarWorkspaces, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sugarWorkspaces).Error
	return
}
// GetSugarWorkspacesInfoList 分页获取Sugar文件列表记录
// Author [yourname](https://github.com/yourname)
func (sugarWorkspacesService *SugarWorkspacesService)GetSugarWorkspacesInfoList(ctx context.Context, info sugarReq.SugarWorkspacesSearch) (list []sugar.SugarWorkspaces, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&sugar.SugarWorkspaces{})
    var sugarWorkspacess []sugar.SugarWorkspaces
    // 如果有条件搜索 下方会自动创建搜索语句
    
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&sugarWorkspacess).Error
	return  sugarWorkspacess, total, err
}
func (sugarWorkspacesService *SugarWorkspacesService)GetSugarWorkspacesPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
