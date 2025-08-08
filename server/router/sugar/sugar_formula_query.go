package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarFormulaQueryRouter struct{}

// InitSugarFormulaQueryRouter 初始化 Sugar公式查询 路由信息
func (s *SugarFormulaQueryRouter) InitSugarFormulaQueryRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	sugarFormulaQueryRouter := Router.Group("sugarFormulaQuery").Use(middleware.OperationRecord())
	sugarFormulaQueryRouterWithoutRecord := Router.Group("sugarFormulaQuery")
	{
		sugarFormulaQueryRouter.POST("executeCalc", sugarFormulaQueryApi.ExecuteSugarCalc) // 执行 SUGAR.CALC 公式
		sugarFormulaQueryRouter.POST("executeGet", sugarFormulaQueryApi.ExecuteSugarGet)   // 执行 SUGAR.GET 公式
	}
	{
		// 如果需要不记录操作日志的接口，可以在这里添加
		_ = sugarFormulaQueryRouterWithoutRecord
	}
}
