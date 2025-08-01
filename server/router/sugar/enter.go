package sugar

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1"

type RouterGroup struct {
	SugarTeamsRouter
	SugarTeamMembersRouter
	SugarDbConnectionsRouter
	SugarSemanticModelsRouter
	SugarAgentsRouter
	SugarCityPermissionsRouter
	SugarRowLevelOverridesRouter
	SugarExecutionLogsRouter
	SugarWorkspacesRouter
}

var (
	sugarTeamsApi             = api.ApiGroupApp.SugarApiGroup.SugarTeamsApi
	sugarTeamMembersApi       = api.ApiGroupApp.SugarApiGroup.SugarTeamMembersApi
	sugarDbConnectionsApi     = api.ApiGroupApp.SugarApiGroup.SugarDbConnectionsApi
	sugarSemanticModelsApi    = api.ApiGroupApp.SugarApiGroup.SugarSemanticModelsApi
	sugarAgentsApi            = api.ApiGroupApp.SugarApiGroup.SugarAgentsApi
	sugarCityPermissionsApi   = api.ApiGroupApp.SugarApiGroup.SugarCityPermissionsApi
	sugarRowLevelOverridesApi = api.ApiGroupApp.SugarApiGroup.SugarRowLevelOverridesApi
	sugarExecutionLogsApi     = api.ApiGroupApp.SugarApiGroup.SugarExecutionLogsApi
	sugarWorkspacesApi        = api.ApiGroupApp.SugarApiGroup.SugarWorkspacesApi
)
