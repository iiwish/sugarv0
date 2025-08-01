
package sugar

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarTeamMembersService struct {}
// CreateSugarTeamMembers 创建sugarTeamMembers表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamMembersService *SugarTeamMembersService) CreateSugarTeamMembers(ctx context.Context, sugarTeamMembers *sugar.SugarTeamMembers) (err error) {
	err = global.GVA_DB.Create(sugarTeamMembers).Error
	return err
}

// DeleteSugarTeamMembers 删除sugarTeamMembers表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamMembersService *SugarTeamMembersService)DeleteSugarTeamMembers(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&sugar.SugarTeamMembers{},"id = ?",id).Error
	return err
}

// DeleteSugarTeamMembersByIds 批量删除sugarTeamMembers表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamMembersService *SugarTeamMembersService)DeleteSugarTeamMembersByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]sugar.SugarTeamMembers{},"id in ?",ids).Error
	return err
}

// UpdateSugarTeamMembers 更新sugarTeamMembers表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamMembersService *SugarTeamMembersService)UpdateSugarTeamMembers(ctx context.Context, sugarTeamMembers sugar.SugarTeamMembers) (err error) {
	err = global.GVA_DB.Model(&sugar.SugarTeamMembers{}).Where("id = ?",sugarTeamMembers.Id).Updates(&sugarTeamMembers).Error
	return err
}

// GetSugarTeamMembers 根据id获取sugarTeamMembers表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamMembersService *SugarTeamMembersService)GetSugarTeamMembers(ctx context.Context, id string) (sugarTeamMembers sugar.SugarTeamMembers, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sugarTeamMembers).Error
	return
}
// GetSugarTeamMembersInfoList 分页获取sugarTeamMembers表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamMembersService *SugarTeamMembersService)GetSugarTeamMembersInfoList(ctx context.Context, info sugarReq.SugarTeamMembersSearch) (list []sugar.SugarTeamMembers, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&sugar.SugarTeamMembers{})
    var sugarTeamMemberss []sugar.SugarTeamMembers
    // 如果有条件搜索 下方会自动创建搜索语句
    
    if info.TeamId != nil && *info.TeamId != "" {
        db = db.Where("team_id = ?", *info.TeamId)
    }
    if info.UserId != nil && *info.UserId != "" {
        db = db.Where("user_id = ?", *info.UserId)
    }
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&sugarTeamMemberss).Error
	return  sugarTeamMemberss, total, err
}
func (sugarTeamMembersService *SugarTeamMembersService)GetSugarTeamMembersPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
