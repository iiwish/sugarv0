package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarTeamsRouter struct {}

// InitSugarTeamsRouter 初始化 团队信息表 路由信息
func (s *SugarTeamsRouter) InitSugarTeamsRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	sugarTeamsRouter := Router.Group("sugarTeams").Use(middleware.OperationRecord())
	sugarTeamsRouterWithoutRecord := Router.Group("sugarTeams")
	sugarTeamsRouterWithoutAuth := PublicRouter.Group("sugarTeams")
	{
		sugarTeamsRouter.POST("createSugarTeams", sugarTeamsApi.CreateSugarTeams)   // 新建团队信息表
		sugarTeamsRouter.DELETE("deleteSugarTeams", sugarTeamsApi.DeleteSugarTeams) // 删除团队信息表
		sugarTeamsRouter.DELETE("deleteSugarTeamsByIds", sugarTeamsApi.DeleteSugarTeamsByIds) // 批量删除团队信息表
		sugarTeamsRouter.PUT("updateSugarTeams", sugarTeamsApi.UpdateSugarTeams)    // 更新团队信息表
	}
	{
		sugarTeamsRouterWithoutRecord.GET("findSugarTeams", sugarTeamsApi.FindSugarTeams)        // 根据ID获取团队信息表
		sugarTeamsRouterWithoutRecord.GET("getSugarTeamsList", sugarTeamsApi.GetSugarTeamsList)  // 获取团队信息表列表
	}
	{
	    sugarTeamsRouterWithoutAuth.GET("getSugarTeamsPublic", sugarTeamsApi.GetSugarTeamsPublic)  // 团队信息表开放接口
	}
}
