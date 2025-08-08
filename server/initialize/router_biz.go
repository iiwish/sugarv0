package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/router"
	"github.com/gin-gonic/gin"
)

func holder(routers ...*gin.RouterGroup) {
	_ = routers
	_ = router.RouterGroupApp
}
func initBizRouter(routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]
	holder(publicGroup, privateGroup)
	{
		sugarRouter := router.RouterGroupApp.Sugar
		sugarRouter.InitSugarTeamsRouter(privateGroup, publicGroup)
		sugarRouter.InitSugarTeamMembersRouter(privateGroup, publicGroup)
		sugarRouter.InitSugarDbConnectionsRouter(privateGroup, publicGroup)
		sugarRouter.InitSugarSemanticModelsRouter(privateGroup, publicGroup)
		sugarRouter.InitSugarAgentsRouter(privateGroup, publicGroup)
		sugarRouter.InitSugarCityPermissionsRouter(privateGroup, publicGroup)
		sugarRouter.InitSugarRowLevelOverridesRouter(privateGroup, publicGroup)
		sugarRouter.InitSugarExecutionLogsRouter(privateGroup, publicGroup)
		sugarRouter.InitSugarWorkspacesRouter(privateGroup, publicGroup)
		sugarRouter.InitSugarFormulaQueryRouter(privateGroup, publicGroup) // Sugar公式查询路由
		sugarRouter.InitSugarFoldersRouter(privateGroup, publicGroup)      // Sugar文件夹管理路由
	}
}
