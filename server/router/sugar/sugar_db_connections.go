package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarDbConnectionsRouter struct {}

// InitSugarDbConnectionsRouter 初始化 Sugar数据库配置表 路由信息
func (s *SugarDbConnectionsRouter) InitSugarDbConnectionsRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	sugarDbConnectionsRouter := Router.Group("sugarDbConnections").Use(middleware.OperationRecord())
	sugarDbConnectionsRouterWithoutRecord := Router.Group("sugarDbConnections")
	sugarDbConnectionsRouterWithoutAuth := PublicRouter.Group("sugarDbConnections")
	{
		sugarDbConnectionsRouter.POST("createSugarDbConnections", sugarDbConnectionsApi.CreateSugarDbConnections)   // 新建Sugar数据库配置表
		sugarDbConnectionsRouter.DELETE("deleteSugarDbConnections", sugarDbConnectionsApi.DeleteSugarDbConnections) // 删除Sugar数据库配置表
		sugarDbConnectionsRouter.DELETE("deleteSugarDbConnectionsByIds", sugarDbConnectionsApi.DeleteSugarDbConnectionsByIds) // 批量删除Sugar数据库配置表
		sugarDbConnectionsRouter.PUT("updateSugarDbConnections", sugarDbConnectionsApi.UpdateSugarDbConnections)    // 更新Sugar数据库配置表
	}
	{
		sugarDbConnectionsRouterWithoutRecord.GET("findSugarDbConnections", sugarDbConnectionsApi.FindSugarDbConnections)        // 根据ID获取Sugar数据库配置表
		sugarDbConnectionsRouterWithoutRecord.GET("getSugarDbConnectionsList", sugarDbConnectionsApi.GetSugarDbConnectionsList)  // 获取Sugar数据库配置表列表
	}
	{
	    sugarDbConnectionsRouterWithoutAuth.GET("getSugarDbConnectionsPublic", sugarDbConnectionsApi.GetSugarDbConnectionsPublic)  // Sugar数据库配置表开放接口
	}
}
