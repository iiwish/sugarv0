package sugar

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type SugarTeamMembersRouter struct {}

// InitSugarTeamMembersRouter 初始化 sugarTeamMembers表 路由信息
func (s *SugarTeamMembersRouter) InitSugarTeamMembersRouter(Router *gin.RouterGroup,PublicRouter *gin.RouterGroup) {
	sugarTeamMembersRouter := Router.Group("sugarTeamMembers").Use(middleware.OperationRecord())
	sugarTeamMembersRouterWithoutRecord := Router.Group("sugarTeamMembers")
	sugarTeamMembersRouterWithoutAuth := PublicRouter.Group("sugarTeamMembers")
	{
		sugarTeamMembersRouter.POST("createSugarTeamMembers", sugarTeamMembersApi.CreateSugarTeamMembers)   // 新建sugarTeamMembers表
		sugarTeamMembersRouter.DELETE("deleteSugarTeamMembers", sugarTeamMembersApi.DeleteSugarTeamMembers) // 删除sugarTeamMembers表
		sugarTeamMembersRouter.DELETE("deleteSugarTeamMembersByIds", sugarTeamMembersApi.DeleteSugarTeamMembersByIds) // 批量删除sugarTeamMembers表
		sugarTeamMembersRouter.PUT("updateSugarTeamMembers", sugarTeamMembersApi.UpdateSugarTeamMembers)    // 更新sugarTeamMembers表
	}
	{
		sugarTeamMembersRouterWithoutRecord.GET("findSugarTeamMembers", sugarTeamMembersApi.FindSugarTeamMembers)        // 根据ID获取sugarTeamMembers表
		sugarTeamMembersRouterWithoutRecord.GET("getSugarTeamMembersList", sugarTeamMembersApi.GetSugarTeamMembersList)  // 获取sugarTeamMembers表列表
	}
	{
	    sugarTeamMembersRouterWithoutAuth.GET("getSugarTeamMembersPublic", sugarTeamMembersApi.GetSugarTeamMembersPublic)  // sugarTeamMembers表开放接口
	}
}
