package sugar

import (
	"context"
	"errors"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
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

// CreateWorkbookFile 创建新的工作簿文件
func (s *SugarWorkspacesService) CreateWorkbookFile(ctx context.Context, name string, parentId *string, teamId string, userId string, defaultContent datatypes.JSON) (*sugar.SugarWorkspaces, error) {
	// 验证用户是否有权限在该团队创建文件
	var count int64
	err := global.GVA_DB.Table("sugar_team_members").Where("user_id = ? AND team_id = ?", userId, teamId).Count(&count).Error
	if err != nil || count == 0 {
		return nil, errors.New("无权限在该团队创建文件")
	}

	// 如果指定了父文件夹，验证父文件夹是否存在且为文件夹类型
	if parentId != nil && *parentId != "" {
		var parent sugar.SugarWorkspaces
		err := global.GVA_DB.Where("id = ? AND team_id = ? AND type = ? AND deleted_at IS NULL", *parentId, teamId, "folder").First(&parent).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("父文件夹不存在")
			}
			return nil, errors.New("查询父文件夹失败")
		}
	}

	// 检查同级目录下是否已存在同名文件
	var existCount int64
	query := global.GVA_DB.Model(&sugar.SugarWorkspaces{}).Where("name = ? AND team_id = ? AND type = ? AND deleted_at IS NULL", name, teamId, "file")
	if parentId != nil {
		query = query.Where("parent_id = ?", *parentId)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	err = query.Count(&existCount).Error
	if err != nil {
		return nil, errors.New("检查名称重复失败")
	}
	if existCount > 0 {
		return nil, errors.New("同级目录下已存在同名文件")
	}

	// 创建新的工作簿文件
	now := time.Now()
	id := uuid.New().String()
	workspace := sugar.SugarWorkspaces{
		Id:        &id,
		Name:      &name,
		Type:      "file",
		ParentId:  parentId,
		TeamId:    &teamId,
		Content:   defaultContent,
		CreatedBy: &userId,
		CreatedAt: &now,
		UpdatedBy: &userId,
		UpdatedAt: &now,
	}

	err = global.GVA_DB.Create(&workspace).Error
	if err != nil {
		global.GVA_LOG.Error("创建工作簿文件失败", zap.Error(err))
		return nil, errors.New("创建文件失败")
	}

	global.GVA_LOG.Info("工作簿文件创建成功", zap.String("id", id), zap.String("name", name))
	return &workspace, nil
}

// SaveWorkbookContent 保存工作簿内容
func (s *SugarWorkspacesService) SaveWorkbookContent(ctx context.Context, id string, content datatypes.JSON, userId string) error {
	// 查找要保存的文件
	var workspace sugar.SugarWorkspaces
	err := global.GVA_DB.Where("id = ? AND type = ? AND deleted_at IS NULL", id, "file").First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文件不存在")
		}
		return errors.New("查询文件失败")
	}

	// 验证用户是否有权限操作该文件
	var count int64
	err = global.GVA_DB.Table("sugar_team_members").Where("user_id = ? AND team_id = ?", userId, *workspace.TeamId).Count(&count).Error
	if err != nil || count == 0 {
		return errors.New("无权限操作该文件")
	}

	// 更新文件内容
	now := time.Now()
	err = global.GVA_DB.Model(&workspace).Updates(map[string]interface{}{
		"content":    content,
		"updated_by": userId,
		"updated_at": now,
	}).Error
	if err != nil {
		global.GVA_LOG.Error("保存工作簿内容失败", zap.Error(err))
		return errors.New("保存文件失败")
	}

	global.GVA_LOG.Info("工作簿内容保存成功", zap.String("id", id))
	return nil
}

// GetWorkbookContent 获取工作簿内容
func (s *SugarWorkspacesService) GetWorkbookContent(ctx context.Context, id string, userId string) (datatypes.JSON, error) {
	// 查找文件
	var workspace sugar.SugarWorkspaces
	err := global.GVA_DB.Where("id = ? AND type = ? AND deleted_at IS NULL", id, "file").First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文件不存在")
		}
		return nil, errors.New("查询文件失败")
	}

	// 验证用户是否有权限访问该文件
	var count int64
	err = global.GVA_DB.Table("sugar_team_members").Where("user_id = ? AND team_id = ?", userId, *workspace.TeamId).Count(&count).Error
	if err != nil || count == 0 {
		return nil, errors.New("无权限访问该文件")
	}

	return workspace.Content, nil
}
