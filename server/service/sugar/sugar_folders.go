package sugar

import (
	"context"
	"errors"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	sugarRes "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SugarFoldersService struct{}

// GetWorkspaceTree 获取工作空间文件夹树形结构
func (s *SugarFoldersService) GetWorkspaceTree(ctx context.Context, req *sugarReq.SugarFoldersGetWorkspaceTreeRequest, userId string) (*sugarRes.SugarFoldersGetWorkspaceTreeResponse, error) {
	var workspaces []sugar.SugarWorkspaces

	// 构建查询条件
	query := global.GVA_DB.Where("deleted_at IS NULL")

	if req.TeamId != nil && *req.TeamId != "" {
		// 指定团队ID
		query = query.Where("team_id = ?", *req.TeamId)
	} else {
		// 获取用户所属的所有团队
		var teamIds []string
		err := global.GVA_DB.Table("sugar_team_members").Where("user_id = ?", userId).Pluck("team_id", &teamIds).Error
		if err != nil {
			global.GVA_LOG.Error("获取用户团队信息失败", zap.Error(err))
			return nil, errors.New("获取用户团队信息失败")
		}
		if len(teamIds) == 0 {
			return sugarRes.NewWorkspaceTreeSuccessResponse([]*sugarRes.SugarFoldersWorkspaceTreeNode{}), nil
		}
		query = query.Where("team_id IN ?", teamIds)
	}

	// 查询所有工作空间项目
	err := query.Order("created_at ASC").Find(&workspaces).Error
	if err != nil {
		global.GVA_LOG.Error("查询工作空间失败", zap.Error(err))
		return nil, errors.New("查询工作空间失败")
	}

	// 构建树形结构
	tree := s.buildTree(workspaces)

	return sugarRes.NewWorkspaceTreeSuccessResponse(tree), nil
}

// CreateFolder 创建文件夹
func (s *SugarFoldersService) CreateFolder(ctx context.Context, req *sugarReq.SugarFoldersCreateFolderRequest, userId string) (*sugarRes.SugarFoldersCreateFolderResponse, error) {
	// 验证类型
	if !req.ValidateType() {
		return nil, errors.New("无效的类型，只支持 folder 或 file")
	}

	// 验证用户是否有权限在该团队创建文件夹
	var count int64
	err := global.GVA_DB.Table("sugar_team_members").Where("user_id = ? AND team_id = ?", userId, req.TeamId).Count(&count).Error
	if err != nil || count == 0 {
		return nil, errors.New("无权限在该团队创建文件夹")
	}

	// 如果指定了父文件夹，验证父文件夹是否存在且为文件夹类型
	if req.ParentId != nil && *req.ParentId != "" {
		var parent sugar.SugarWorkspaces
		err := global.GVA_DB.Where("id = ? AND team_id = ? AND type = ? AND deleted_at IS NULL", *req.ParentId, req.TeamId, "folder").First(&parent).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("父文件夹不存在")
			}
			return nil, errors.New("查询父文件夹失败")
		}
	}

	// 检查同级目录下是否已存在同名项目
	var existCount int64
	query := global.GVA_DB.Model(&sugar.SugarWorkspaces{}).Where("name = ? AND team_id = ? AND deleted_at IS NULL", req.Name, req.TeamId)
	if req.ParentId != nil {
		query = query.Where("parent_id = ?", *req.ParentId)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	err = query.Count(&existCount).Error
	if err != nil {
		return nil, errors.New("检查名称重复失败")
	}
	if existCount > 0 {
		return nil, errors.New("同级目录下已存在同名项目")
	}

	// 创建新的工作空间项目
	now := time.Now()
	id := uuid.New().String()
	workspace := sugar.SugarWorkspaces{
		Id:        &id,
		Name:      &req.Name,
		Type:      req.Type,
		ParentId:  req.ParentId,
		TeamId:    &req.TeamId,
		CreatedBy: &userId,
		CreatedAt: &now,
		UpdatedBy: &userId,
		UpdatedAt: &now,
	}

	err = global.GVA_DB.Create(&workspace).Error
	if err != nil {
		global.GVA_LOG.Error("创建工作空间项目失败", zap.Error(err))
		return nil, errors.New("创建失败")
	}

	return sugarRes.NewCreateFolderSuccessResponse(workspace), nil
}

// RenameItem 重命名文件夹或文件
func (s *SugarFoldersService) RenameItem(ctx context.Context, req *sugarReq.SugarFoldersRenameRequest, userId string) (*sugarRes.SugarFoldersRenameResponse, error) {
	// 查找要重命名的项目
	var workspace sugar.SugarWorkspaces
	err := global.GVA_DB.Where("id = ? AND deleted_at IS NULL", req.Id).First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, errors.New("查询项目失败")
	}

	// 验证用户是否有权限
	var count int64
	err = global.GVA_DB.Table("sugar_team_members").Where("user_id = ? AND team_id = ?", userId, *workspace.TeamId).Count(&count).Error
	if err != nil || count == 0 {
		return nil, errors.New("无权限操作该项目")
	}

	// 检查同级目录下是否已存在同名项目
	var existCount int64
	query := global.GVA_DB.Model(&sugar.SugarWorkspaces{}).Where("name = ? AND team_id = ? AND id != ? AND deleted_at IS NULL", req.Name, *workspace.TeamId, req.Id)
	if workspace.ParentId != nil {
		query = query.Where("parent_id = ?", *workspace.ParentId)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	err = query.Count(&existCount).Error
	if err != nil {
		return nil, errors.New("检查名称重复失败")
	}
	if existCount > 0 {
		return nil, errors.New("同级目录下已存在同名项目")
	}

	// 更新名称
	now := time.Now()
	err = global.GVA_DB.Model(&workspace).Updates(map[string]interface{}{
		"name":       req.Name,
		"updated_by": userId,
		"updated_at": now,
	}).Error
	if err != nil {
		global.GVA_LOG.Error("重命名失败", zap.Error(err))
		return nil, errors.New("重命名失败")
	}

	// 重新查询更新后的数据
	err = global.GVA_DB.Where("id = ?", req.Id).First(&workspace).Error
	if err != nil {
		return nil, errors.New("查询更新后的数据失败")
	}

	return sugarRes.NewRenameSuccessResponse(workspace), nil
}

// MoveItem 移动文件夹或文件
func (s *SugarFoldersService) MoveItem(ctx context.Context, req *sugarReq.SugarFoldersMoveRequest, userId string) (*sugarRes.SugarFoldersMoveResponse, error) {
	// 查找要移动的项目
	var workspace sugar.SugarWorkspaces
	err := global.GVA_DB.Where("id = ? AND deleted_at IS NULL", req.Id).First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, errors.New("查询项目失败")
	}

	// 验证用户是否有权限操作原项目
	var count int64
	err = global.GVA_DB.Table("sugar_team_members").Where("user_id = ? AND team_id = ?", userId, *workspace.TeamId).Count(&count).Error
	if err != nil || count == 0 {
		return nil, errors.New("无权限操作该项目")
	}

	// 验证用户是否有权限操作目标团队
	err = global.GVA_DB.Table("sugar_team_members").Where("user_id = ? AND team_id = ?", userId, req.TeamId).Count(&count).Error
	if err != nil || count == 0 {
		return nil, errors.New("无权限移动到目标团队")
	}

	// 如果指定了目标父文件夹，验证父文件夹是否存在且为文件夹类型
	if req.ParentId != nil && *req.ParentId != "" {
		// 不能移动到自己或自己的子目录
		if *req.ParentId == req.Id {
			return nil, errors.New("不能移动到自己")
		}

		// 检查是否移动到自己的子目录
		if workspace.Type == "folder" {
			isChild, err := s.isChildFolder(req.Id, *req.ParentId)
			if err != nil {
				return nil, err
			}
			if isChild {
				return nil, errors.New("不能移动到自己的子目录")
			}
		}

		var parent sugar.SugarWorkspaces
		err := global.GVA_DB.Where("id = ? AND team_id = ? AND type = ? AND deleted_at IS NULL", *req.ParentId, req.TeamId, "folder").First(&parent).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("目标父文件夹不存在")
			}
			return nil, errors.New("查询目标父文件夹失败")
		}
	}

	// 检查目标位置是否已存在同名项目
	var existCount int64
	query := global.GVA_DB.Model(&sugar.SugarWorkspaces{}).Where("name = ? AND team_id = ? AND id != ? AND deleted_at IS NULL", *workspace.Name, req.TeamId, req.Id)
	if req.ParentId != nil {
		query = query.Where("parent_id = ?", *req.ParentId)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	err = query.Count(&existCount).Error
	if err != nil {
		return nil, errors.New("检查名称重复失败")
	}
	if existCount > 0 {
		return nil, errors.New("目标位置已存在同名项目")
	}

	// 更新位置
	now := time.Now()
	err = global.GVA_DB.Model(&workspace).Updates(map[string]interface{}{
		"parent_id":  req.ParentId,
		"team_id":    req.TeamId,
		"updated_by": userId,
		"updated_at": now,
	}).Error
	if err != nil {
		global.GVA_LOG.Error("移动失败", zap.Error(err))
		return nil, errors.New("移动失败")
	}

	// 重新查询更新后的数据
	err = global.GVA_DB.Where("id = ?", req.Id).First(&workspace).Error
	if err != nil {
		return nil, errors.New("查询更新后的数据失败")
	}

	return sugarRes.NewMoveSuccessResponse(workspace), nil
}

// DeleteItem 删除文件夹或文件
func (s *SugarFoldersService) DeleteItem(ctx context.Context, req *sugarReq.SugarFoldersDeleteRequest, userId string) (*sugarRes.SugarFoldersDeleteResponse, error) {
	// 查找要删除的项目
	var workspace sugar.SugarWorkspaces
	err := global.GVA_DB.Where("id = ? AND deleted_at IS NULL", req.Id).First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("项目不存在")
		}
		return nil, errors.New("查询项目失败")
	}

	// 验证用户是否有权限
	var count int64
	err = global.GVA_DB.Table("sugar_team_members").Where("user_id = ? AND team_id = ?", userId, *workspace.TeamId).Count(&count).Error
	if err != nil || count == 0 {
		return nil, errors.New("无权限操作该项目")
	}

	// 如果是文件夹，检查是否有子项目
	if workspace.Type == "folder" {
		var childCount int64
		err = global.GVA_DB.Model(&sugar.SugarWorkspaces{}).Where("parent_id = ? AND deleted_at IS NULL", req.Id).Count(&childCount).Error
		if err != nil {
			return nil, errors.New("检查子项目失败")
		}
		if childCount > 0 {
			return nil, errors.New("文件夹不为空，无法删除")
		}
	}

	// 软删除
	now := time.Now()
	err = global.GVA_DB.Model(&workspace).Updates(map[string]interface{}{
		"deleted_at": now,
		"updated_by": userId,
		"updated_at": now,
	}).Error
	if err != nil {
		global.GVA_LOG.Error("删除失败", zap.Error(err))
		return nil, errors.New("删除失败")
	}

	return sugarRes.NewDeleteSuccessResponse("删除成功"), nil
}

// GetFolderContent 获取文件夹内容
func (s *SugarFoldersService) GetFolderContent(ctx context.Context, req *sugarReq.SugarFoldersGetFolderContentRequest, userId string) (*sugarRes.SugarFoldersGetFolderContentResponse, error) {
	// 查找文件夹
	var folder sugar.SugarWorkspaces
	err := global.GVA_DB.Where("id = ? AND type = ? AND deleted_at IS NULL", req.FolderId, "folder").First(&folder).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文件夹不存在")
		}
		return nil, errors.New("查询文件夹失败")
	}

	// 验证用户是否有权限
	var count int64
	err = global.GVA_DB.Table("sugar_team_members").Where("user_id = ? AND team_id = ?", userId, *folder.TeamId).Count(&count).Error
	if err != nil || count == 0 {
		return nil, errors.New("无权限访问该文件夹")
	}

	// 查询文件夹内容
	var workspaces []sugar.SugarWorkspaces
	var total int64

	query := global.GVA_DB.Model(&sugar.SugarWorkspaces{}).Where("parent_id = ? AND deleted_at IS NULL", req.FolderId)

	// 计算总数
	err = query.Count(&total).Error
	if err != nil {
		return nil, errors.New("查询总数失败")
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	err = query.Order("type ASC, created_at ASC").Offset(offset).Limit(req.PageSize).Find(&workspaces).Error
	if err != nil {
		global.GVA_LOG.Error("查询文件夹内容失败", zap.Error(err))
		return nil, errors.New("查询文件夹内容失败")
	}

	return sugarRes.NewGetFolderContentSuccessResponse(workspaces, total, req.Page, req.PageSize), nil
}

// buildTree 构建树形结构
func (s *SugarFoldersService) buildTree(workspaces []sugar.SugarWorkspaces) []*sugarRes.SugarFoldersWorkspaceTreeNode {
	// 创建节点映射
	nodeMap := make(map[string]*sugarRes.SugarFoldersWorkspaceTreeNode)
	var rootNodes []*sugarRes.SugarFoldersWorkspaceTreeNode

	// 创建所有节点
	for _, workspace := range workspaces {
		node := &sugarRes.SugarFoldersWorkspaceTreeNode{
			Id:       *workspace.Id,
			Name:     *workspace.Name,
			Type:     workspace.Type,
			ParentId: workspace.ParentId,
			TeamId:   *workspace.TeamId,
			Children: []*sugarRes.SugarFoldersWorkspaceTreeNode{},
		}
		nodeMap[*workspace.Id] = node
	}

	// 构建父子关系
	for _, workspace := range workspaces {
		node := nodeMap[*workspace.Id]
		if workspace.ParentId == nil {
			// 根节点
			rootNodes = append(rootNodes, node)
		} else {
			// 子节点
			if parent, exists := nodeMap[*workspace.ParentId]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return rootNodes
}

// isChildFolder 检查 childId 是否是 parentId 的子文件夹
func (s *SugarFoldersService) isChildFolder(parentId, childId string) (bool, error) {
	var workspace sugar.SugarWorkspaces
	err := global.GVA_DB.Where("id = ? AND deleted_at IS NULL", childId).First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.New("查询子文件夹失败")
	}

	// 递归检查父级
	if workspace.ParentId == nil {
		return false, nil
	}

	if *workspace.ParentId == parentId {
		return true, nil
	}

	return s.isChildFolder(parentId, *workspace.ParentId)
}
