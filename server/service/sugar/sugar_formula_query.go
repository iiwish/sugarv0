package sugar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	sugarRes "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/response"
	"go.uber.org/zap"
)

type SugarFormulaQueryService struct{}

// ExecuteCalcFormula 执行 SUGAR.CALC 公式
func (s *SugarFormulaQueryService) ExecuteCalcFormula(ctx context.Context, req *sugarReq.SugarFormulaCalcRequest, userId string) (*sugarRes.SugarFormulaCalcResponse, error) {
	// 验证计算方式
	if !req.ValidateCalcMethod() {
		return sugarRes.NewCalcErrorResponse("不支持的计算方式: " + req.CalcMethod), nil
	}

	// 获取语义模型
	model, err := s.getSemanticModel(ctx, req.ModelName, userId)
	if err != nil {
		return sugarRes.NewCalcErrorResponse(err.Error()), nil
	}

	// 构建SQL查询
	sql, args, err := s.buildCalcSQL(model, req)
	if err != nil {
		return sugarRes.NewCalcErrorResponse(err.Error()), nil
	}

	// 添加行级权限
	sql, args, err = s.addRowLevelPermissions(sql, args, model, userId)
	if err != nil {
		return sugarRes.NewCalcErrorResponse(err.Error()), nil
	}

	// 打印SQL查询用于调试
	global.GVA_LOG.Info("执行CALC查询",
		zap.String("sql", sql),
		zap.Any("args", args),
		zap.String("userId", userId))

	// 执行查询 - 使用指针类型处理 NULL 值
	var result *float64
	err = global.GVA_DB.Raw(sql, args...).Scan(&result).Error
	if err != nil {
		global.GVA_LOG.Error("CALC查询执行失败",
			zap.String("sql", sql),
			zap.Any("args", args),
			zap.Error(err))
		return sugarRes.NewCalcErrorResponse("查询执行失败: " + err.Error()), nil
	}

	// 处理 NULL 值情况
	var finalResult interface{}
	if result != nil {
		finalResult = *result
		global.GVA_LOG.Info("CALC查询成功",
			zap.Float64("result", *result))
	} else {
		finalResult = 0.0 // 当没有匹配数据时返回 0
		global.GVA_LOG.Info("CALC查询结果为NULL，返回默认值0.0")
	}

	return sugarRes.NewCalcSuccessResponse(finalResult), nil
}

// ExecuteGetFormula 执行 SUGAR.GET 公式
func (s *SugarFormulaQueryService) ExecuteGetFormula(ctx context.Context, req *sugarReq.SugarFormulaGetRequest, userId string) (*sugarRes.SugarFormulaGetResponse, error) {
	// 获取语义模型
	model, err := s.getSemanticModel(ctx, req.ModelName, userId)
	if err != nil {
		return sugarRes.NewGetErrorResponse(err.Error()), nil
	}

	// 构建SQL查询
	sql, args, err := s.buildGetSQL(model, req)
	if err != nil {
		return sugarRes.NewGetErrorResponse(err.Error()), nil
	}

	// 添加行级权限
	sql, args, err = s.addRowLevelPermissions(sql, args, model, userId)
	if err != nil {
		return sugarRes.NewGetErrorResponse(err.Error()), nil
	}

	// 打印SQL查询用于调试
	global.GVA_LOG.Info("执行GET查询",
		zap.String("sql", sql),
		zap.Any("args", args),
		zap.String("userId", userId))

	// 执行查询
	var results []map[string]interface{}
	err = global.GVA_DB.Raw(sql, args...).Scan(&results).Error
	if err != nil {
		global.GVA_LOG.Error("GET查询执行失败",
			zap.String("sql", sql),
			zap.Any("args", args),
			zap.Error(err))
		return sugarRes.NewGetErrorResponse("查询执行失败: " + err.Error()), nil
	}

	global.GVA_LOG.Info("GET查询成功",
		zap.Int("resultCount", len(results)))

	return sugarRes.NewGetSuccessResponse(results, req.ReturnColumns), nil
}

// getSemanticModel 获取语义模型
func (s *SugarFormulaQueryService) getSemanticModel(ctx context.Context, modelName, userId string) (*sugar.SugarSemanticModels, error) {
	var model sugar.SugarSemanticModels

	// 获取用户所属团队
	var teamIds []string
	err := global.GVA_DB.Table("sugar_team_members").Where("user_id = ?", userId).Pluck("team_id", &teamIds).Error
	if err != nil {
		return nil, errors.New("获取用户团队信息失败")
	}
	if len(teamIds) == 0 {
		return nil, errors.New("用户未加入任何团队")
	}

	// 查找语义模型
	err = global.GVA_DB.Where("name = ? AND team_id IN ?", modelName, teamIds).First(&model).Error
	if err != nil {
		return nil, errors.New("语义模型不存在或无权访问: " + modelName)
	}

	return &model, nil
}

// buildCalcSQL 构建计算SQL
func (s *SugarFormulaQueryService) buildCalcSQL(model *sugar.SugarSemanticModels, req *sugarReq.SugarFormulaCalcRequest) (string, []interface{}, error) {
	// 解析可返回列配置
	var returnableColumns map[string]map[string]interface{}
	if err := json.Unmarshal(model.ReturnableColumnsConfig, &returnableColumns); err != nil {
		return "", nil, errors.New("解析可返回列配置失败")
	}

	// 检查计算列是否存在
	columnConfig, exists := returnableColumns[req.CalcColumn]
	if !exists {
		return "", nil, errors.New("计算列不存在: " + req.CalcColumn)
	}

	// 获取实际列名
	actualColumn, ok := columnConfig["column"].(string)
	if !ok {
		return "", nil, errors.New("计算列配置错误")
	}

	// 构建SELECT子句
	selectClause := fmt.Sprintf("SELECT %s(t.%s)", req.CalcMethod, actualColumn)

	// 构建FROM子句
	fromClause := fmt.Sprintf("FROM %s AS t", *model.SourceTableName)

	// 构建WHERE子句
	whereClause, args, err := s.buildWhereClause(model, req.Filters)
	if err != nil {
		return "", nil, err
	}

	sql := selectClause + " " + fromClause
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}

	return sql, args, nil
}

// buildGetSQL 构建查询SQL
func (s *SugarFormulaQueryService) buildGetSQL(model *sugar.SugarSemanticModels, req *sugarReq.SugarFormulaGetRequest) (string, []interface{}, error) {
	// 解析可返回列配置
	var returnableColumns map[string]map[string]interface{}
	if err := json.Unmarshal(model.ReturnableColumnsConfig, &returnableColumns); err != nil {
		return "", nil, errors.New("解析可返回列配置失败")
	}

	// 构建SELECT子句
	var selectColumns []string
	for _, columnName := range req.ReturnColumns {
		columnConfig, exists := returnableColumns[columnName]
		if !exists {
			return "", nil, errors.New("返回列不存在: " + columnName)
		}

		actualColumn, ok := columnConfig["column"].(string)
		if !ok {
			return "", nil, errors.New("返回列配置错误: " + columnName)
		}

		selectColumns = append(selectColumns, fmt.Sprintf("t.%s AS `%s`", actualColumn, columnName))
	}

	selectClause := "SELECT " + strings.Join(selectColumns, ", ")

	// 构建FROM子句
	fromClause := fmt.Sprintf("FROM %s AS t", *model.SourceTableName)

	// 构建WHERE子句
	whereClause, args, err := s.buildWhereClause(model, req.Filters)
	if err != nil {
		return "", nil, err
	}

	sql := selectClause + " " + fromClause
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}

	return sql, args, nil
}

// buildWhereClause 构建WHERE子句
func (s *SugarFormulaQueryService) buildWhereClause(model *sugar.SugarSemanticModels, filters map[string]interface{}) (string, []interface{}, error) {
	if len(filters) == 0 {
		return "", []interface{}{}, nil
	}

	// 解析参数配置
	var parameterConfig map[string]map[string]interface{}
	if err := json.Unmarshal(model.ParameterConfig, &parameterConfig); err != nil {
		return "", nil, errors.New("解析参数配置失败")
	}

	var conditions []string
	var args []interface{}

	for filterKey, filterValue := range filters {
		paramConfig, exists := parameterConfig[filterKey]
		if !exists {
			return "", nil, errors.New("筛选条件不存在: " + filterKey)
		}

		actualColumn, ok := paramConfig["column"].(string)
		if !ok {
			return "", nil, errors.New("筛选条件配置错误: " + filterKey)
		}

		operator, ok := paramConfig["operator"].(string)
		if !ok {
			operator = "="
		}

		switch operator {
		case "=":
			conditions = append(conditions, fmt.Sprintf("t.%s = ?", actualColumn))
			args = append(args, filterValue)
		case "LIKE":
			conditions = append(conditions, fmt.Sprintf("t.%s LIKE ?", actualColumn))
			args = append(args, "%"+fmt.Sprintf("%v", filterValue)+"%")
		case ">", "<", ">=", "<=", "!=":
			conditions = append(conditions, fmt.Sprintf("t.%s %s ?", actualColumn, operator))
			args = append(args, filterValue)
		default:
			return "", nil, errors.New("不支持的操作符: " + operator)
		}
	}

	return strings.Join(conditions, " AND "), args, nil
}

// addRowLevelPermissions 添加行级权限
func (s *SugarFormulaQueryService) addRowLevelPermissions(sql string, args []interface{}, model *sugar.SugarSemanticModels, userId string) (string, []interface{}, error) {
	// 如果没有配置权限列，直接返回
	if model.PermissionKeyColumn == nil || *model.PermissionKeyColumn == "" {
		return sql, args, nil
	}

	// 检查用户是否在豁免表中
	var count int64
	err := global.GVA_DB.Table("sugar_row_level_overrides").Where("user_id = ?", userId).Count(&count).Error
	if err == nil && count > 0 {
		// 用户在豁免表中，跳过权限检查
		return sql, args, nil
	}

	// 获取用户的城市权限
	var cityCodes []string
	err = global.GVA_DB.Table("sugar_city_permissions").Where("user_id = ?", userId).Pluck("city_code", &cityCodes).Error
	if err != nil {
		return "", nil, errors.New("获取用户权限失败")
	}

	if len(cityCodes) == 0 {
		return "", nil, errors.New("用户无数据访问权限")
	}

	// 添加权限条件
	permissionColumn := *model.PermissionKeyColumn

	// 检查SQL是否已有WHERE子句
	if strings.Contains(strings.ToUpper(sql), "WHERE") {
		sql += " AND "
	} else {
		sql += " WHERE "
	}

	// 构建IN条件
	placeholders := make([]string, len(cityCodes))
	for i := range cityCodes {
		placeholders[i] = "?"
		args = append(args, cityCodes[i])
	}

	sql += fmt.Sprintf("t.%s IN (%s)", permissionColumn, strings.Join(placeholders, ","))

	return sql, args, nil
}
