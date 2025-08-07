package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarAgentsRouter struct{}

// InitSugarAgentsRouter 初始化 sugar智能体表 路由信息
func (s *SugarAgentsRouter) InitSugarAgentsRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	sugarAgentsRouter := Router.Group("sugarAgents").Use(middleware.OperationRecord())
	sugarAgentsRouterWithoutRecord := Router.Group("sugarAgents")
	{
		sugarAgentsRouter.POST("createSugarAgents", sugarAgentsApi.CreateSugarAgents)             // 新建sugar智能体表
		sugarAgentsRouter.DELETE("deleteSugarAgents", sugarAgentsApi.DeleteSugarAgents)           // 删除sugar智能体表
		sugarAgentsRouter.DELETE("deleteSugarAgentsByIds", sugarAgentsApi.DeleteSugarAgentsByIds) // 批量删除sugar智能体表
		sugarAgentsRouter.PUT("updateSugarAgents", sugarAgentsApi.UpdateSugarAgents)              // 更新sugar智能体表
	}
	{
		sugarAgentsRouterWithoutRecord.GET("findSugarAgents", sugarAgentsApi.FindSugarAgents)       // 根据ID获取sugar智能体表
		sugarAgentsRouterWithoutRecord.GET("getSugarAgentsList", sugarAgentsApi.GetSugarAgentsList) // 获取sugar智能体表列表
	}
}
