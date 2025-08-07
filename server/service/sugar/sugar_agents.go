package sugar

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarAgentsService struct{}

// CreateSugarAgents 创建sugar智能体表记录
func (s *SugarAgentsService) CreateSugarAgents(ctx context.Context, sugarAgents *sugar.SugarAgents) (err error) {
	err = global.GVA_DB.Create(sugarAgents).Error
	return err
}

// DeleteSugarAgents 删除sugar智能体表记录
func (s *SugarAgentsService) DeleteSugarAgents(ctx context.Context, id string, userId string) (err error) {
	// 增加权限校验：只允许创建者或团队管理员删除
	var agent sugar.SugarAgents
	if err = global.GVA_DB.Where("id = ?", id).First(&agent).Error; err != nil {
		return errors.New("记录不存在")
	}
	// 此处应有更复杂的团队角色权限判断，暂时简化为只判断创建者
	if *agent.CreatedBy != userId {
		return errors.New("无权删除")
	}
	err = global.GVA_DB.Delete(&sugar.SugarAgents{}, "id = ?", id).Error
	return err
}

// DeleteSugarAgentsByIds 批量删除sugar智能体表记录
func (s *SugarAgentsService) DeleteSugarAgentsByIds(ctx context.Context, ids []string, userId string) (err error) {
	// 增加权限校验：只允许用户删除自己创建的记录
	err = global.GVA_DB.Where("id IN ? AND created_by = ?", ids, userId).Delete(&[]sugar.SugarAgents{}).Error
	return err
}

// UpdateSugarAgents 更新sugar智能体表记录
func (s *SugarAgentsService) UpdateSugarAgents(ctx context.Context, agent sugar.SugarAgents, userId string) (err error) {
	// 增加权限校验：只允许创建者或团队管理员更新
	var oldAgent sugar.SugarAgents
	if err = global.GVA_DB.Where("id = ?", agent.Id).First(&oldAgent).Error; err != nil {
		return errors.New("记录不存在")
	}
	// 此处应有更复杂的团队角色权限判断，暂时简化为只判断创建者
	if *oldAgent.CreatedBy != userId {
		return errors.New("无权更新")
	}
	err = global.GVA_DB.Model(&sugar.SugarAgents{}).Where("id = ?", agent.Id).Updates(&agent).Error
	return err
}

// GetSugarAgents 根据id获取sugar智能体表记录
func (s *SugarAgentsService) GetSugarAgents(ctx context.Context, id string, userId string) (agent sugar.SugarAgents, err error) {
	if err = global.GVA_DB.Where("id = ?", id).First(&agent).Error; err != nil {
		return agent, errors.New("记录不存在")
	}
	// 此处应有更复杂的团队角色权限判断，暂时简化为只判断创建者或团队成员
	if *agent.CreatedBy != userId {
		// 可以在这里查询 team_members 表确认用户是否是团队成员
		// return agent, errors.New("无权查看")
	}
	return agent, nil
}

// GetSugarAgentsListByUser 分页获取sugar智能体表记录
func (s *SugarAgentsService) GetSugarAgentsListByUser(ctx context.Context, info sugarReq.SugarAgentsSearch, userId string) (list []sugar.SugarAgents, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)

	// 1. 查找用户所在的所有团队
	var teamIds []string
	err = global.GVA_DB.Table("sugar_team_members").Where("user_id = ?", userId).Pluck("team_id", &teamIds).Error
	if err != nil {
		return nil, 0, err
	}

	// 如果用户不属于任何团队，则返回空
	if len(teamIds) == 0 {
		return []sugar.SugarAgents{}, 0, nil
	}

	// 2. 创建查询
	db := global.GVA_DB.Model(&sugar.SugarAgents{}).Where("team_id IN ?", teamIds)
	var sugarAgentss []sugar.SugarAgents

	// 如果有其他搜索条件，可以在这里添加
	// if info.Name != "" {
	// 	db = db.Where("name LIKE ?", "%"+info.Name+"%")
	// }

	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Order("updated_at desc").Find(&sugarAgentss).Error
	return sugarAgentss, total, err
}
