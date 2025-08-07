package sugar

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarTeamsService struct{}

// CreateSugarTeams 创建团队信息表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamsService *SugarTeamsService) CreateSugarTeams(ctx context.Context, sugarTeams *sugar.SugarTeams) (err error) {
	err = global.GVA_DB.Create(sugarTeams).Error
	return err
}

// DeleteSugarTeams 删除团队信息表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamsService *SugarTeamsService) DeleteSugarTeams(ctx context.Context, id string, userId string) (err error) {
	err = global.GVA_DB.Where("owner_id = ?", userId).Delete(&sugar.SugarTeams{}, "id = ?", id).Error
	return err
}

// DeleteSugarTeamsByIds 批量删除团队信息表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamsService *SugarTeamsService) DeleteSugarTeamsByIds(ctx context.Context, ids []string, userId string) (err error) {
	err = global.GVA_DB.Where("owner_id = ?", userId).Delete(&[]sugar.SugarTeams{}, "id in ?", ids).Error
	return err
}

// UpdateSugarTeams 更新团队信息表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamsService *SugarTeamsService) UpdateSugarTeams(ctx context.Context, sugarTeams sugar.SugarTeams) (err error) {
	// 查询团队成员表，如果团队成员表中的角色不属于"owner"\"admin"，则不允许更新
	if sugarTeams.UpdatedBy == nil || *sugarTeams.UpdatedBy == "" {
		return errors.New("更新人不能为空")
	}
	var existingTeamMembers []sugar.SugarTeamMembers
	err = global.GVA_DB.Where("team_id = ? AND user_id = ?", sugarTeams.Id, *sugarTeams.UpdatedBy).Find(&existingTeamMembers).Error
	if err != nil {
		return err
	}
	if len(existingTeamMembers) == 0 {
		return errors.New("团队成员不存在或没有权限更新")
	}
	for _, member := range existingTeamMembers {
		if member.Role != "owner" && member.Role != "admin" {
			return errors.New("只有团队所有者或管理员可以更新团队信息")
		}
	}

	err = global.GVA_DB.Model(&sugar.SugarTeams{}).Where("id = ?", sugarTeams.Id).Updates(&sugarTeams).Error
	return err
}

// GetSugarTeams 根据id获取团队信息表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamsService *SugarTeamsService) GetSugarTeams(ctx context.Context, id string) (sugarTeams sugar.SugarTeams, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sugarTeams).Error
	return
}

// GetSugarTeamsInfoList 分页获取团队信息表记录
// Author [yourname](https://github.com/yourname)
func (sugarTeamsService *SugarTeamsService) GetSugarTeamsInfoList(ctx context.Context, info sugarReq.SugarTeamsSearch) (list []sugar.SugarTeams, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := global.GVA_DB.Model(&sugar.SugarTeams{})
	var sugarTeamss []sugar.SugarTeams
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.OwnerId != nil {
		db = db.Where("owner_id = ?", info.OwnerId)
	}
	if info.TeamName != nil {
		db = db.Where("team_name LIKE ?", "%"+*info.TeamName+"%")
	}

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Find(&sugarTeamss).Error
	return sugarTeamss, total, err
}
func (sugarTeamsService *SugarTeamsService) GetSugarTeamsPublic(ctx context.Context) {
	// 此方法为获取数据源定义的数据
	// 请自行实现
}
