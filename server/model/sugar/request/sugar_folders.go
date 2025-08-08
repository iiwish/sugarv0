package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

// SugarFoldersGetWorkspaceTreeRequest 获取工作空间文件夹树形结构请求
type SugarFoldersGetWorkspaceTreeRequest struct {
	TeamId *string `json:"teamId" form:"teamId"` // 团队ID，不传则获取当前用户所有团队的工作空间
}

// SugarFoldersCreateFolderRequest 创建文件夹请求
type SugarFoldersCreateFolderRequest struct {
	Name     string  `json:"name" binding:"required"`   // 文件夹名称
	ParentId *string `json:"parentId"`                  // 父文件夹ID，为空则创建在根目录
	TeamId   string  `json:"teamId" binding:"required"` // 团队ID
	Type     string  `json:"type" binding:"required"`   // 类型：folder 或 file
}

// SugarFoldersRenameRequest 重命名文件夹或文件请求
type SugarFoldersRenameRequest struct {
	Id   string `json:"id" binding:"required"`   // 要重命名的项目ID
	Name string `json:"name" binding:"required"` // 新名称
}

// SugarFoldersMoveRequest 移动文件夹或文件请求
type SugarFoldersMoveRequest struct {
	Id       string  `json:"id" binding:"required"`     // 要移动的项目ID
	ParentId *string `json:"parentId"`                  // 新的父文件夹ID，为空则移动到根目录
	TeamId   string  `json:"teamId" binding:"required"` // 目标团队ID
}

// SugarFoldersDeleteRequest 删除文件夹或文件请求
type SugarFoldersDeleteRequest struct {
	Id string `json:"id" form:"id" binding:"required"` // 要删除的项目ID
}

// SugarFoldersGetFolderContentRequest 获取文件夹内容请求
type SugarFoldersGetFolderContentRequest struct {
	FolderId         string `json:"folderId" form:"folderId" binding:"required"` // 文件夹ID
	request.PageInfo        // 分页信息
}

// ValidateType 验证类型是否有效
func (r *SugarFoldersCreateFolderRequest) ValidateType() bool {
	validTypes := map[string]bool{
		"folder": true,
		"file":   true,
	}
	return validTypes[r.Type]
}
