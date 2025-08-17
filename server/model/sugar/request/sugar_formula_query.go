package request

// SugarFormulaCalcRequest SUGAR.CALC 公式请求结构
type SugarFormulaCalcRequest struct {
	ModelName  string                 `json:"modelName" binding:"required"`  // 语义模型名称
	CalcColumn string                 `json:"calcColumn" binding:"required"` // 计算列名
	CalcMethod string                 `json:"calcMethod" binding:"required"` // 计算方式: SUM, AVG, COUNT, MAX, MIN
	Filters    map[string]interface{} `json:"filters"`                       // 筛选条件键值对
}

// SugarFormulaGetRequest SUGAR.GET 公式请求结构
type SugarFormulaGetRequest struct {
	ModelName     string                 `json:"modelName" binding:"required"`     // 语义模型名称
	ReturnColumns []string               `json:"returnColumns" binding:"required"` // 返回列名列表
	Filters       map[string]interface{} `json:"filters"`                          // 筛选条件键值对
	GroupBy       []string               `json:"groupBy"`                          // 分组字段列表，用于聚合查询
}

// ValidateCalcMethod 验证计算方式是否有效
func (r *SugarFormulaCalcRequest) ValidateCalcMethod() bool {
	validMethods := map[string]bool{
		"SUM":   true,
		"AVG":   true,
		"COUNT": true,
		"MAX":   true,
		"MIN":   true,
	}
	return validMethods[r.CalcMethod]
}
