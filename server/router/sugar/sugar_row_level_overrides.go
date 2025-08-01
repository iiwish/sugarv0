package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarRowLevelOverridesRouter struct {}

// InitSugarRowLevelOverridesRouter 初始化 Sugar行级权限豁免表 路由信息
func (s *SugarRowLevelOverridesRouter) InitSugarRowLevelOverridesRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	sugarRowLevelOverridesRouter := Router.Group("sugarRowLevelOverrides").Use(middleware.OperationRecord())
	sugarRowLevelOverridesRouterWithoutRecord := Router.Group("sugarRowLevelOverrides")
	sugarRowLevelOverridesRouterWithoutAuth := PublicRouter.Group("sugarRowLevelOverrides")
	{
		sugarRowLevelOverridesRouter.POST("createSugarRowLevelOverrides", sugarRowLevelOverridesApi.CreateSugarRowLevelOverrides)   // 新建Sugar行级权限豁免表
		sugarRowLevelOverridesRouter.DELETE("deleteSugarRowLevelOverrides", sugarRowLevelOverridesApi.DeleteSugarRowLevelOverrides) // 删除Sugar行级权限豁免表
		sugarRowLevelOverridesRouter.DELETE("deleteSugarRowLevelOverridesByIds", sugarRowLevelOverridesApi.DeleteSugarRowLevelOverridesByIds) // 批量删除Sugar行级权限豁免表
		sugarRowLevelOverridesRouter.PUT("updateSugarRowLevelOverrides", sugarRowLevelOverridesApi.UpdateSugarRowLevelOverrides)    // 更新Sugar行级权限豁免表
	}
	{
		sugarRowLevelOverridesRouterWithoutRecord.GET("findSugarRowLevelOverrides", sugarRowLevelOverridesApi.FindSugarRowLevelOverrides)        // 根据ID获取Sugar行级权限豁免表
		sugarRowLevelOverridesRouterWithoutRecord.GET("getSugarRowLevelOverridesList", sugarRowLevelOverridesApi.GetSugarRowLevelOverridesList)  // 获取Sugar行级权限豁免表列表
	}
	{
	    sugarRowLevelOverridesRouterWithoutAuth.GET("getSugarRowLevelOverridesPublic", sugarRowLevelOverridesApi.GetSugarRowLevelOverridesPublic)  // Sugar行级权限豁免表开放接口
	}
}
