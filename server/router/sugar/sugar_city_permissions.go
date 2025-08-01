package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarCityPermissionsRouter struct {}

// InitSugarCityPermissionsRouter 初始化 sugarCityPermissions表 路由信息
func (s *SugarCityPermissionsRouter) InitSugarCityPermissionsRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	sugarCityPermissionsRouter := Router.Group("sugarCityPermissions").Use(middleware.OperationRecord())
	sugarCityPermissionsRouterWithoutRecord := Router.Group("sugarCityPermissions")
	sugarCityPermissionsRouterWithoutAuth := PublicRouter.Group("sugarCityPermissions")
	{
		sugarCityPermissionsRouter.POST("createSugarCityPermissions", sugarCityPermissionsApi.CreateSugarCityPermissions)   // 新建sugarCityPermissions表
		sugarCityPermissionsRouter.DELETE("deleteSugarCityPermissions", sugarCityPermissionsApi.DeleteSugarCityPermissions) // 删除sugarCityPermissions表
		sugarCityPermissionsRouter.DELETE("deleteSugarCityPermissionsByIds", sugarCityPermissionsApi.DeleteSugarCityPermissionsByIds) // 批量删除sugarCityPermissions表
		sugarCityPermissionsRouter.PUT("updateSugarCityPermissions", sugarCityPermissionsApi.UpdateSugarCityPermissions)    // 更新sugarCityPermissions表
	}
	{
		sugarCityPermissionsRouterWithoutRecord.GET("findSugarCityPermissions", sugarCityPermissionsApi.FindSugarCityPermissions)        // 根据ID获取sugarCityPermissions表
		sugarCityPermissionsRouterWithoutRecord.GET("getSugarCityPermissionsList", sugarCityPermissionsApi.GetSugarCityPermissionsList)  // 获取sugarCityPermissions表列表
	}
	{
	    sugarCityPermissionsRouterWithoutAuth.GET("getSugarCityPermissionsPublic", sugarCityPermissionsApi.GetSugarCityPermissionsPublic)  // sugarCityPermissions表开放接口
	}
}
