
package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	
)

type SugarTeamMembersSearch struct{
      TeamId  *string `json:"teamId" form:"teamId"` 
      UserId  *string `json:"userId" form:"userId"` 
    request.PageInfo
}
