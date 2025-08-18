package sugar

type ServiceGroup struct {
	SugarTeamsService
	SugarTeamMembersService
	SugarDbConnectionsService
	SugarSemanticModelsService
	SugarAgentsService
	SugarCityPermissionsService
	SugarRowLevelOverridesService
	SugarExecutionLogsService
	SugarWorkspacesService
	SugarFormulaQueryService
	SugarFormulaAiService
	SugarFoldersService
}

// GetSugarFormulaAiService 获取AI服务单例实例
func (s *ServiceGroup) GetSugarFormulaAiService() *SugarFormulaAiService {
	return GetSugarFormulaAiService()
}
