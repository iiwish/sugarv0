
package sugar

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarAgentsService struct {}
// CreateSugarAgents 创建sugar智能体表记录
// Author [yourname](https://github.com/yourname)
func (sugarAgentsService *SugarAgentsService) CreateSugarAgents(ctx context.Context, sugarAgents *sugar.SugarAgents) (err error) {
	err = global.GVA_DB.Create(sugarAgents).Error
	return err
}

// DeleteSugarAgents 删除sugar智能体表记录
// Author [yourname](https://github.com/yourname)
func (sugarAgentsService *SugarAgentsService)DeleteSugarAgents(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&sugar.SugarAgents{},"id = ?",id).Error
	return err
}

// DeleteSugarAgentsByIds 批量删除sugar智能体表记录
// Author [yourname](https://github.com/yourname)
func (sugarAgentsService *SugarAgentsService)DeleteSugarAgentsByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]sugar.SugarAgents{},"id in ?",ids).Error
	return err
}

// UpdateSugarAgents 更新sugar智能体表记录
// Author [yourname](https://github.com/yourname)
func (sugarAgentsService *SugarAgentsService)UpdateSugarAgents(ctx context.Context, sugarAgents sugar.SugarAgents) (err error) {
	err = global.GVA_DB.Model(&sugar.SugarAgents{}).Where("id = ?",sugarAgents.Id).Updates(&sugarAgents).Error
	return err
}

// GetSugarAgents 根据id获取sugar智能体表记录
// Author [yourname](https://github.com/yourname)
func (sugarAgentsService *SugarAgentsService)GetSugarAgents(ctx context.Context, id string) (sugarAgents sugar.SugarAgents, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sugarAgents).Error
	return
}
// GetSugarAgentsInfoList 分页获取sugar智能体表记录
// Author [yourname](https://github.com/yourname)
func (sugarAgentsService *SugarAgentsService)GetSugarAgentsInfoList(ctx context.Context, info sugarReq.SugarAgentsSearch) (list []sugar.SugarAgents, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&sugar.SugarAgents{})
    var sugarAgentss []sugar.SugarAgents
    // 如果有条件搜索 下方会自动创建搜索语句
    
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&sugarAgentss).Error
	return  sugarAgentss, total, err
}
func (sugarAgentsService *SugarAgentsService)GetSugarAgentsPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
