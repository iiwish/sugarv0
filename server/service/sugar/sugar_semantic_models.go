
package sugar

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarSemanticModelsService struct {}
// CreateSugarSemanticModels 创建Sugar指标语义表记录
// Author [yourname](https://github.com/yourname)
func (sugarSemanticModelsService *SugarSemanticModelsService) CreateSugarSemanticModels(ctx context.Context, sugarSemanticModels *sugar.SugarSemanticModels) (err error) {
	err = global.GVA_DB.Create(sugarSemanticModels).Error
	return err
}

// DeleteSugarSemanticModels 删除Sugar指标语义表记录
// Author [yourname](https://github.com/yourname)
func (sugarSemanticModelsService *SugarSemanticModelsService)DeleteSugarSemanticModels(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&sugar.SugarSemanticModels{},"id = ?",id).Error
	return err
}

// DeleteSugarSemanticModelsByIds 批量删除Sugar指标语义表记录
// Author [yourname](https://github.com/yourname)
func (sugarSemanticModelsService *SugarSemanticModelsService)DeleteSugarSemanticModelsByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]sugar.SugarSemanticModels{},"id in ?",ids).Error
	return err
}

// UpdateSugarSemanticModels 更新Sugar指标语义表记录
// Author [yourname](https://github.com/yourname)
func (sugarSemanticModelsService *SugarSemanticModelsService)UpdateSugarSemanticModels(ctx context.Context, sugarSemanticModels sugar.SugarSemanticModels) (err error) {
	err = global.GVA_DB.Model(&sugar.SugarSemanticModels{}).Where("id = ?",sugarSemanticModels.Id).Updates(&sugarSemanticModels).Error
	return err
}

// GetSugarSemanticModels 根据id获取Sugar指标语义表记录
// Author [yourname](https://github.com/yourname)
func (sugarSemanticModelsService *SugarSemanticModelsService)GetSugarSemanticModels(ctx context.Context, id string) (sugarSemanticModels sugar.SugarSemanticModels, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sugarSemanticModels).Error
	return
}
// GetSugarSemanticModelsInfoList 分页获取Sugar指标语义表记录
// Author [yourname](https://github.com/yourname)
func (sugarSemanticModelsService *SugarSemanticModelsService)GetSugarSemanticModelsInfoList(ctx context.Context, info sugarReq.SugarSemanticModelsSearch) (list []sugar.SugarSemanticModels, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&sugar.SugarSemanticModels{})
    var sugarSemanticModelss []sugar.SugarSemanticModels
    // 如果有条件搜索 下方会自动创建搜索语句
    
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&sugarSemanticModelss).Error
	return  sugarSemanticModelss, total, err
}
func (sugarSemanticModelsService *SugarSemanticModelsService)GetSugarSemanticModelsPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
