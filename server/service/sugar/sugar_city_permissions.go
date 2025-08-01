
package sugar

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarCityPermissionsService struct {}
// CreateSugarCityPermissions 创建sugarCityPermissions表记录
// Author [yourname](https://github.com/yourname)
func (sugarCityPermissionsService *SugarCityPermissionsService) CreateSugarCityPermissions(ctx context.Context, sugarCityPermissions *sugar.SugarCityPermissions) (err error) {
	err = global.GVA_DB.Create(sugarCityPermissions).Error
	return err
}

// DeleteSugarCityPermissions 删除sugarCityPermissions表记录
// Author [yourname](https://github.com/yourname)
func (sugarCityPermissionsService *SugarCityPermissionsService)DeleteSugarCityPermissions(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&sugar.SugarCityPermissions{},"id = ?",id).Error
	return err
}

// DeleteSugarCityPermissionsByIds 批量删除sugarCityPermissions表记录
// Author [yourname](https://github.com/yourname)
func (sugarCityPermissionsService *SugarCityPermissionsService)DeleteSugarCityPermissionsByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]sugar.SugarCityPermissions{},"id in ?",ids).Error
	return err
}

// UpdateSugarCityPermissions 更新sugarCityPermissions表记录
// Author [yourname](https://github.com/yourname)
func (sugarCityPermissionsService *SugarCityPermissionsService)UpdateSugarCityPermissions(ctx context.Context, sugarCityPermissions sugar.SugarCityPermissions) (err error) {
	err = global.GVA_DB.Model(&sugar.SugarCityPermissions{}).Where("id = ?",sugarCityPermissions.Id).Updates(&sugarCityPermissions).Error
	return err
}

// GetSugarCityPermissions 根据id获取sugarCityPermissions表记录
// Author [yourname](https://github.com/yourname)
func (sugarCityPermissionsService *SugarCityPermissionsService)GetSugarCityPermissions(ctx context.Context, id string) (sugarCityPermissions sugar.SugarCityPermissions, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sugarCityPermissions).Error
	return
}
// GetSugarCityPermissionsInfoList 分页获取sugarCityPermissions表记录
// Author [yourname](https://github.com/yourname)
func (sugarCityPermissionsService *SugarCityPermissionsService)GetSugarCityPermissionsInfoList(ctx context.Context, info sugarReq.SugarCityPermissionsSearch) (list []sugar.SugarCityPermissions, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&sugar.SugarCityPermissions{})
    var sugarCityPermissionss []sugar.SugarCityPermissions
    // 如果有条件搜索 下方会自动创建搜索语句
    
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&sugarCityPermissionss).Error
	return  sugarCityPermissionss, total, err
}
func (sugarCityPermissionsService *SugarCityPermissionsService)GetSugarCityPermissionsPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
