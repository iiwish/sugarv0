package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarFoldersRouter struct{}

// InitSugarFoldersRouter 初始化 Sugar文件夹管理 路由信息
func (s *SugarFoldersRouter) InitSugarFoldersRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	sugarFoldersRouter := Router.Group("sugarFolders").Use(middleware.OperationRecord())
	sugarFoldersRouterWithoutRecord := Router.Group("sugarFolders")
	{
		sugarFoldersRouter.GET("getWorkspaceTree", sugarFoldersApi.GetWorkspaceTree) // 获取工作空间文件夹树形结构
		sugarFoldersRouter.POST("createFolder", sugarFoldersApi.CreateFolder)        // 创建文件夹
		sugarFoldersRouter.PUT("rename", sugarFoldersApi.RenameItem)                 // 重命名文件夹或文件
		sugarFoldersRouter.PUT("move", sugarFoldersApi.MoveItem)                     // 移动文件夹或文件
		sugarFoldersRouter.DELETE("deleteItem", sugarFoldersApi.DeleteItem)          // 删除文件夹或文件
		sugarFoldersRouter.GET("getFolderContent", sugarFoldersApi.GetFolderContent) // 获取文件夹内容
	}
	{
		// 如果需要不记录操作日志的接口，可以在这里添加
		_ = sugarFoldersRouterWithoutRecord
	}
}
