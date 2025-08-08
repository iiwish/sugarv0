package sugar

import "github.com/flipped-aurora/gin-vue-admin/server/service"

type ApiGroup struct {
	SugarTeamsApi
	SugarTeamMembersApi
	SugarDbConnectionsApi
	SugarSemanticModelsApi
	SugarAgentsApi
	SugarCityPermissionsApi
	SugarRowLevelOverridesApi
	SugarExecutionLogsApi
	SugarWorkspacesApi
	SugarFormulaQueryApi
}

var (
	sugarTeamsService             = service.ServiceGroupApp.SugarServiceGroup.SugarTeamsService
	sugarTeamMembersService       = service.ServiceGroupApp.SugarServiceGroup.SugarTeamMembersService
	sugarDbConnectionsService     = service.ServiceGroupApp.SugarServiceGroup.SugarDbConnectionsService
	sugarSemanticModelsService    = service.ServiceGroupApp.SugarServiceGroup.SugarSemanticModelsService
	sugarAgentsService            = service.ServiceGroupApp.SugarServiceGroup.SugarAgentsService
	sugarCityPermissionsService   = service.ServiceGroupApp.SugarServiceGroup.SugarCityPermissionsService
	sugarRowLevelOverridesService = service.ServiceGroupApp.SugarServiceGroup.SugarRowLevelOverridesService
	sugarExecutionLogsService     = service.ServiceGroupApp.SugarServiceGroup.SugarExecutionLogsService
	sugarWorkspacesService        = service.ServiceGroupApp.SugarServiceGroup.SugarWorkspacesService
	sugarFormulaQueryService      = service.ServiceGroupApp.SugarServiceGroup.SugarFormulaQueryService
)
