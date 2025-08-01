package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarSemanticModelsRouter struct {}

// InitSugarSemanticModelsRouter 初始化 Sugar指标语义表 路由信息
func (s *SugarSemanticModelsRouter) InitSugarSemanticModelsRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	sugarSemanticModelsRouter := Router.Group("sugarSemanticModels").Use(middleware.OperationRecord())
	sugarSemanticModelsRouterWithoutRecord := Router.Group("sugarSemanticModels")
	sugarSemanticModelsRouterWithoutAuth := PublicRouter.Group("sugarSemanticModels")
	{
		sugarSemanticModelsRouter.POST("createSugarSemanticModels", sugarSemanticModelsApi.CreateSugarSemanticModels)   // 新建Sugar指标语义表
		sugarSemanticModelsRouter.DELETE("deleteSugarSemanticModels", sugarSemanticModelsApi.DeleteSugarSemanticModels) // 删除Sugar指标语义表
		sugarSemanticModelsRouter.DELETE("deleteSugarSemanticModelsByIds", sugarSemanticModelsApi.DeleteSugarSemanticModelsByIds) // 批量删除Sugar指标语义表
		sugarSemanticModelsRouter.PUT("updateSugarSemanticModels", sugarSemanticModelsApi.UpdateSugarSemanticModels)    // 更新Sugar指标语义表
	}
	{
		sugarSemanticModelsRouterWithoutRecord.GET("findSugarSemanticModels", sugarSemanticModelsApi.FindSugarSemanticModels)        // 根据ID获取Sugar指标语义表
		sugarSemanticModelsRouterWithoutRecord.GET("getSugarSemanticModelsList", sugarSemanticModelsApi.GetSugarSemanticModelsList)  // 获取Sugar指标语义表列表
	}
	{
	    sugarSemanticModelsRouterWithoutAuth.GET("getSugarSemanticModelsPublic", sugarSemanticModelsApi.GetSugarSemanticModelsPublic)  // Sugar指标语义表开放接口
	}
}
