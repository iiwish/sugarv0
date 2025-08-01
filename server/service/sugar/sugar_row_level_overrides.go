
package sugar

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarRowLevelOverridesService struct {}
// CreateSugarRowLevelOverrides 创建Sugar行级权限豁免表记录
// Author [yourname](https://github.com/yourname)
func (sugarRowLevelOverridesService *SugarRowLevelOverridesService) CreateSugarRowLevelOverrides(ctx context.Context, sugarRowLevelOverrides *sugar.SugarRowLevelOverrides) (err error) {
	err = global.GVA_DB.Create(sugarRowLevelOverrides).Error
	return err
}

// DeleteSugarRowLevelOverrides 删除Sugar行级权限豁免表记录
// Author [yourname](https://github.com/yourname)
func (sugarRowLevelOverridesService *SugarRowLevelOverridesService)DeleteSugarRowLevelOverrides(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&sugar.SugarRowLevelOverrides{},"id = ?",id).Error
	return err
}

// DeleteSugarRowLevelOverridesByIds 批量删除Sugar行级权限豁免表记录
// Author [yourname](https://github.com/yourname)
func (sugarRowLevelOverridesService *SugarRowLevelOverridesService)DeleteSugarRowLevelOverridesByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]sugar.SugarRowLevelOverrides{},"id in ?",ids).Error
	return err
}

// UpdateSugarRowLevelOverrides 更新Sugar行级权限豁免表记录
// Author [yourname](https://github.com/yourname)
func (sugarRowLevelOverridesService *SugarRowLevelOverridesService)UpdateSugarRowLevelOverrides(ctx context.Context, sugarRowLevelOverrides sugar.SugarRowLevelOverrides) (err error) {
	err = global.GVA_DB.Model(&sugar.SugarRowLevelOverrides{}).Where("id = ?",sugarRowLevelOverrides.Id).Updates(&sugarRowLevelOverrides).Error
	return err
}

// GetSugarRowLevelOverrides 根据id获取Sugar行级权限豁免表记录
// Author [yourname](https://github.com/yourname)
func (sugarRowLevelOverridesService *SugarRowLevelOverridesService)GetSugarRowLevelOverrides(ctx context.Context, id string) (sugarRowLevelOverrides sugar.SugarRowLevelOverrides, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sugarRowLevelOverrides).Error
	return
}
// GetSugarRowLevelOverridesInfoList 分页获取Sugar行级权限豁免表记录
// Author [yourname](https://github.com/yourname)
func (sugarRowLevelOverridesService *SugarRowLevelOverridesService)GetSugarRowLevelOverridesInfoList(ctx context.Context, info sugarReq.SugarRowLevelOverridesSearch) (list []sugar.SugarRowLevelOverrides, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&sugar.SugarRowLevelOverrides{})
    var sugarRowLevelOverridess []sugar.SugarRowLevelOverrides
    // 如果有条件搜索 下方会自动创建搜索语句
    
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&sugarRowLevelOverridess).Error
	return  sugarRowLevelOverridess, total, err
}
func (sugarRowLevelOverridesService *SugarRowLevelOverridesService)GetSugarRowLevelOverridesPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
