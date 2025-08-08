package response

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
)

// SugarFoldersWorkspaceTreeNode 工作空间树节点
type SugarFoldersWorkspaceTreeNode struct {
	Id       string                           `json:"id"`       // 节点ID
	Name     string                           `json:"name"`     // 节点名称
	Type     string                           `json:"type"`     // 节点类型：folder 或 file
	ParentId *string                          `json:"parentId"` // 父节点ID
	TeamId   string                           `json:"teamId"`   // 团队ID
	Children []*SugarFoldersWorkspaceTreeNode `json:"children"` // 子节点
}

// SugarFoldersGetWorkspaceTreeResponse 获取工作空间文件夹树形结构响应
type SugarFoldersGetWorkspaceTreeResponse struct {
	Tree []*SugarFoldersWorkspaceTreeNode `json:"tree"` // 树形结构数据
}

// SugarFoldersCreateFolderResponse 创建文件夹响应
type SugarFoldersCreateFolderResponse struct {
	SugarWorkspace sugar.SugarWorkspaces `json:"sugarWorkspace"` // 创建的工作空间项目
}

// SugarFoldersRenameResponse 重命名响应
type SugarFoldersRenameResponse struct {
	SugarWorkspace sugar.SugarWorkspaces `json:"sugarWorkspace"` // 重命名后的工作空间项目
}

// SugarFoldersMoveResponse 移动响应
type SugarFoldersMoveResponse struct {
	SugarWorkspace sugar.SugarWorkspaces `json:"sugarWorkspace"` // 移动后的工作空间项目
}

// SugarFoldersDeleteResponse 删除响应
type SugarFoldersDeleteResponse struct {
	Message string `json:"message"` // 删除结果消息
}

// SugarFoldersGetFolderContentResponse 获取文件夹内容响应
type SugarFoldersGetFolderContentResponse struct {
	List     []sugar.SugarWorkspaces `json:"list"`     // 文件夹内容列表
	Total    int64                   `json:"total"`    // 总数
	Page     int                     `json:"page"`     // 当前页码
	PageSize int                     `json:"pageSize"` // 每页数量
}

// NewWorkspaceTreeSuccessResponse 创建成功的工作空间树响应
func NewWorkspaceTreeSuccessResponse(tree []*SugarFoldersWorkspaceTreeNode) *SugarFoldersGetWorkspaceTreeResponse {
	return &SugarFoldersGetWorkspaceTreeResponse{
		Tree: tree,
	}
}

// NewCreateFolderSuccessResponse 创建成功的文件夹响应
func NewCreateFolderSuccessResponse(workspace sugar.SugarWorkspaces) *SugarFoldersCreateFolderResponse {
	return &SugarFoldersCreateFolderResponse{
		SugarWorkspace: workspace,
	}
}

// NewRenameSuccessResponse 创建成功的重命名响应
func NewRenameSuccessResponse(workspace sugar.SugarWorkspaces) *SugarFoldersRenameResponse {
	return &SugarFoldersRenameResponse{
		SugarWorkspace: workspace,
	}
}

// NewMoveSuccessResponse 创建成功的移动响应
func NewMoveSuccessResponse(workspace sugar.SugarWorkspaces) *SugarFoldersMoveResponse {
	return &SugarFoldersMoveResponse{
		SugarWorkspace: workspace,
	}
}

// NewDeleteSuccessResponse 创建成功的删除响应
func NewDeleteSuccessResponse(message string) *SugarFoldersDeleteResponse {
	return &SugarFoldersDeleteResponse{
		Message: message,
	}
}

// NewGetFolderContentSuccessResponse 创建成功的获取文件夹内容响应
func NewGetFolderContentSuccessResponse(list []sugar.SugarWorkspaces, total int64, page int, pageSize int) *SugarFoldersGetFolderContentResponse {
	return &SugarFoldersGetFolderContentResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
}
