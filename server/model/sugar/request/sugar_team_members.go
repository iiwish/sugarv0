package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

type SugarTeamMembersCreateRequest struct {
	TeamId    string  `json:"teamId" form:"teamId" binding:"required"`
	UserId    string  `json:"userId" form:"userId" binding:"required"`
	Role      string  `json:"role" form:"role" binding:"required"` //role字段
	CreatedBy *string `json:"createdBy" form:"createdBy"`          // 创建者ID
}

type SugarTeamMembersUpdateRequest struct {
	Id     int     `json:"id" form:"id" binding:"required"`
	TeamId *string `json:"teamId" form:"teamId" binding:"required"`
	UserId *string `json:"userId" form:"userId" binding:"required"`
	Role   *string `json:"role" form:"role"`
	Status *string `json:"status" form:"status"`
}

type SugarTeamMembersSearch struct {
	TeamId *string `json:"teamId" form:"teamId"`
	UserId *string `json:"userId" form:"userId"`
	Role   *string `json:"role" form:"role"`
	Status *string `json:"status" form:"status"`
	request.PageInfo
}
