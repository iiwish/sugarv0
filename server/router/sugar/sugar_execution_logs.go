package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarExecutionLogsRouter struct {}

// InitSugarExecutionLogsRouter 初始化 sugar操作日志表 路由信息
func (s *SugarExecutionLogsRouter) InitSugarExecutionLogsRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	sugarExecutionLogsRouter := Router.Group("sugarExecutionLogs").Use(middleware.OperationRecord())
	sugarExecutionLogsRouterWithoutRecord := Router.Group("sugarExecutionLogs")
	sugarExecutionLogsRouterWithoutAuth := PublicRouter.Group("sugarExecutionLogs")
	{
		sugarExecutionLogsRouter.POST("createSugarExecutionLogs", sugarExecutionLogsApi.CreateSugarExecutionLogs)   // 新建sugar操作日志表
		sugarExecutionLogsRouter.DELETE("deleteSugarExecutionLogs", sugarExecutionLogsApi.DeleteSugarExecutionLogs) // 删除sugar操作日志表
		sugarExecutionLogsRouter.DELETE("deleteSugarExecutionLogsByIds", sugarExecutionLogsApi.DeleteSugarExecutionLogsByIds) // 批量删除sugar操作日志表
		sugarExecutionLogsRouter.PUT("updateSugarExecutionLogs", sugarExecutionLogsApi.UpdateSugarExecutionLogs)    // 更新sugar操作日志表
	}
	{
		sugarExecutionLogsRouterWithoutRecord.GET("findSugarExecutionLogs", sugarExecutionLogsApi.FindSugarExecutionLogs)        // 根据ID获取sugar操作日志表
		sugarExecutionLogsRouterWithoutRecord.GET("getSugarExecutionLogsList", sugarExecutionLogsApi.GetSugarExecutionLogsList)  // 获取sugar操作日志表列表
	}
	{
	    sugarExecutionLogsRouterWithoutAuth.GET("getSugarExecutionLogsPublic", sugarExecutionLogsApi.GetSugarExecutionLogsPublic)  // sugar操作日志表开放接口
	}
}
