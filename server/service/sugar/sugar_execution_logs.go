
package sugar

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarExecutionLogsService struct {}
// CreateSugarExecutionLogs 创建sugar操作日志表记录
// Author [yourname](https://github.com/yourname)
func (sugarExecutionLogsService *SugarExecutionLogsService) CreateSugarExecutionLogs(ctx context.Context, sugarExecutionLogs *sugar.SugarExecutionLogs) (err error) {
	err = global.GVA_DB.Create(sugarExecutionLogs).Error
	return err
}

// DeleteSugarExecutionLogs 删除sugar操作日志表记录
// Author [yourname](https://github.com/yourname)
func (sugarExecutionLogsService *SugarExecutionLogsService)DeleteSugarExecutionLogs(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&sugar.SugarExecutionLogs{},"id = ?",id).Error
	return err
}

// DeleteSugarExecutionLogsByIds 批量删除sugar操作日志表记录
// Author [yourname](https://github.com/yourname)
func (sugarExecutionLogsService *SugarExecutionLogsService)DeleteSugarExecutionLogsByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]sugar.SugarExecutionLogs{},"id in ?",ids).Error
	return err
}

// UpdateSugarExecutionLogs 更新sugar操作日志表记录
// Author [yourname](https://github.com/yourname)
func (sugarExecutionLogsService *SugarExecutionLogsService)UpdateSugarExecutionLogs(ctx context.Context, sugarExecutionLogs sugar.SugarExecutionLogs) (err error) {
	err = global.GVA_DB.Model(&sugar.SugarExecutionLogs{}).Where("id = ?",sugarExecutionLogs.Id).Updates(&sugarExecutionLogs).Error
	return err
}

// GetSugarExecutionLogs 根据id获取sugar操作日志表记录
// Author [yourname](https://github.com/yourname)
func (sugarExecutionLogsService *SugarExecutionLogsService)GetSugarExecutionLogs(ctx context.Context, id string) (sugarExecutionLogs sugar.SugarExecutionLogs, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sugarExecutionLogs).Error
	return
}
// GetSugarExecutionLogsInfoList 分页获取sugar操作日志表记录
// Author [yourname](https://github.com/yourname)
func (sugarExecutionLogsService *SugarExecutionLogsService)GetSugarExecutionLogsInfoList(ctx context.Context, info sugarReq.SugarExecutionLogsSearch) (list []sugar.SugarExecutionLogs, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&sugar.SugarExecutionLogs{})
    var sugarExecutionLogss []sugar.SugarExecutionLogs
    // 如果有条件搜索 下方会自动创建搜索语句
    
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&sugarExecutionLogss).Error
	return  sugarExecutionLogss, total, err
}
func (sugarExecutionLogsService *SugarExecutionLogsService)GetSugarExecutionLogsPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
