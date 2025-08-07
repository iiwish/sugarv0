package sugar

import (
	"context"
	"errors"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
)

type SugarWorkspacesService struct{}

// CreateSugarWorkspaces 创建Sugar文件列表记录
func (s *SugarWorkspacesService) CreateSugarWorkspaces(ctx context.Context, workspace *sugar.SugarWorkspaces) (err error) {
	err = global.GVA_DB.Create(workspace).Error
	return err
}

// DeleteSugarWorkspaces 删除Sugar文件列表记录
func (s *SugarWorkspacesService) DeleteSugarWorkspaces(ctx context.Context, id string, userId string) (err error) {
	var workspace sugar.SugarWorkspaces
	if err = global.GVA_DB.Where("id = ?", id).First(&workspace).Error; err != nil {
		return errors.New("文件或文件夹不存在")
	}
	// 简化权限：只有创建者能删除
	if *workspace.CreatedBy != userId {
		return errors.New("无权删除")
	}
	err = global.GVA_DB.Delete(&sugar.SugarWorkspaces{}, "id = ?", id).Error
	return err
}

// DeleteSugarWorkspacesByIds 批量删除Sugar文件列表记录
func (s *SugarWorkspacesService) DeleteSugarWorkspacesByIds(ctx context.Context, ids []string, userId string) (err error) {
	// 简化权限：只批量删除用户自己创建的文件
	err = global.GVA_DB.Where("id IN ? AND created_by = ?", ids, userId).Delete(&[]sugar.SugarWorkspaces{}).Error
	return err
}

// UpdateSugarWorkspaces 更新Sugar文件列表记录
func (s *SugarWorkspacesService) UpdateSugarWorkspaces(ctx context.Context, workspace sugar.SugarWorkspaces, userId string) (err error) {
	var oldWorkspace sugar.SugarWorkspaces
	if err = global.GVA_DB.Where("id = ?", workspace.Id).First(&oldWorkspace).Error; err != nil {
		return errors.New("文件或文件夹不存在")
	}
	// 简化权限：只有创建者能更新
	if *oldWorkspace.CreatedBy != userId {
		return errors.New("无权更新")
	}
	err = global.GVA_DB.Model(&sugar.SugarWorkspaces{}).Where("id = ?", workspace.Id).Updates(&workspace).Error
	return err
}

// GetSugarWorkspaces 根据id获取Sugar文件列表记录
func (s *SugarWorkspacesService) GetSugarWorkspaces(ctx context.Context, id string, userId string) (workspace sugar.SugarWorkspaces, err error) {
	if err = global.GVA_DB.Where("id = ?", id).First(&workspace).Error; err != nil {
		return workspace, errors.New("文件或文件夹不存在")
	}
	// TODO: 此处应增加对 sugar_workspace_permissions 表的校验，判断用户是否有权限查看
	return workspace, nil
}

// GetSugarWorkspacesInfoListByUser 分页获取Sugar文件列表记录
func (s *SugarWorkspacesService) GetSugarWorkspacesInfoListByUser(ctx context.Context, info sugarReq.SugarWorkspacesSearch, userId string) (list []sugar.SugarWorkspaces, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)

	var teamIds []string
	err = global.GVA_DB.Table("sugar_team_members").Where("user_id = ?", userId).Pluck("team_id", &teamIds).Error
	if err != nil {
		return nil, 0, err
	}
	if len(teamIds) == 0 {
		return []sugar.SugarWorkspaces{}, 0, nil
	}

	db := global.GVA_DB.Model(&sugar.SugarWorkspaces{}).Where("team_id IN ?", teamIds)

	// 根据 parent_id 筛选
	if info.ParentId != nil {
		db = db.Where("parent_id = ?", *info.ParentId)
	} else {
		db = db.Where("parent_id IS NULL")
	}

	var sugarWorkspacess []sugar.SugarWorkspaces
	err = db.Count(&total).Error
	if err != nil {
		return
	}

	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}

	err = db.Order("type asc, updated_at desc").Find(&sugarWorkspacess).Error
	return sugarWorkspacess, total, err
}
