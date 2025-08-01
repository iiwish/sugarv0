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
		sugarRouter.InitSugarExecutionLogsRouter(privateGroup, publicGroup) // 占位方法，保证文件可以正确加载，避免go空变量检测报错，请勿删除。
		sugarRouter.InitSugarWorkspacesRouter(privateGroup, publicGroup)
	}
}
