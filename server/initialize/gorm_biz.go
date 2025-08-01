package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
)

func bizModel() error {
	db := global.GVA_DB
	err := db.AutoMigrate(sugar.SugarTeams{}, sugar.SugarTeamMembers{}, sugar.SugarDbConnections{}, sugar.SugarSemanticModels{}, sugar.SugarAgents{}, sugar.SugarCityPermissions{}, sugar.SugarRowLevelOverrides{}, sugar.SugarExecutionLogs{}, sugar.SugarWorkspaces{})
	if err != nil {
		return err
	}
	return nil
}
