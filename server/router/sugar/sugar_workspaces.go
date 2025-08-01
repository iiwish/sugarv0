package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarWorkspacesRouter struct {}

// InitSugarWorkspacesRouter 初始化 Sugar文件列表 路由信息
func (s *SugarWorkspacesRouter) InitSugarWorkspacesRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	sugarWorkspacesRouter := Router.Group("sugarWorkspaces").Use(middleware.OperationRecord())
	sugarWorkspacesRouterWithoutRecord := Router.Group("sugarWorkspaces")
	sugarWorkspacesRouterWithoutAuth := PublicRouter.Group("sugarWorkspaces")
	{
		sugarWorkspacesRouter.POST("createSugarWorkspaces", sugarWorkspacesApi.CreateSugarWorkspaces)   // 新建Sugar文件列表
		sugarWorkspacesRouter.DELETE("deleteSugarWorkspaces", sugarWorkspacesApi.DeleteSugarWorkspaces) // 删除Sugar文件列表
		sugarWorkspacesRouter.DELETE("deleteSugarWorkspacesByIds", sugarWorkspacesApi.DeleteSugarWorkspacesByIds) // 批量删除Sugar文件列表
		sugarWorkspacesRouter.PUT("updateSugarWorkspaces", sugarWorkspacesApi.UpdateSugarWorkspaces)    // 更新Sugar文件列表
	}
	{
		sugarWorkspacesRouterWithoutRecord.GET("findSugarWorkspaces", sugarWorkspacesApi.FindSugarWorkspaces)        // 根据ID获取Sugar文件列表
		sugarWorkspacesRouterWithoutRecord.GET("getSugarWorkspacesList", sugarWorkspacesApi.GetSugarWorkspacesList)  // 获取Sugar文件列表列表
	}
	{
	    sugarWorkspacesRouterWithoutAuth.GET("getSugarWorkspacesPublic", sugarWorkspacesApi.GetSugarWorkspacesPublic)  // Sugar文件列表开放接口
	}
}
