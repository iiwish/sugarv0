package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

type SugarTeamsSearch struct {
	OwnerId  *string `json:"ownerId" form:"ownerId"`   //团队创建者/个人空间的所有者
	TeamName *string `json:"teamName" form:"teamName"` // 团队名称
	request.PageInfo
}

// SugarTeamsCreateRequest 创建团队请求结构体
type SugarTeamsCreateRequest struct {
	TeamName   string `json:"teamName" binding:"required"` // 团队名称（必填）
	IsPersonal bool   `json:"isPersonal"`                  // 是否个人空间
}

// SugarTeamsUpdateRequest 更新团队请求结构体
type SugarTeamsUpdateRequest struct {
	Id       string  `json:"id" binding:"required"` // 团队ID（必填）
	TeamName *string `json:"teamName"`              // 团队名称（可选）
}
