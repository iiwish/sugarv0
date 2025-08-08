package response

// SugarFormulaCalcResponse SUGAR.CALC 公式响应结构
type SugarFormulaCalcResponse struct {
	Result interface{} `json:"result"` // 计算结果，可能是数字或错误信息
	Error  string      `json:"error"`  // 错误信息
}

// SugarFormulaGetResponse SUGAR.GET 公式响应结构
type SugarFormulaGetResponse struct {
	Results []map[string]interface{} `json:"results"` // 查询结果列表
	Count   int                      `json:"count"`   // 结果数量
	Error   string                   `json:"error"`   // 错误信息
}

// NewCalcSuccessResponse 创建成功的计算响应
func NewCalcSuccessResponse(result interface{}) *SugarFormulaCalcResponse {
	return &SugarFormulaCalcResponse{
		Result: result,
		Error:  "",
	}
}

// NewCalcErrorResponse 创建错误的计算响应
func NewCalcErrorResponse(error string) *SugarFormulaCalcResponse {
	return &SugarFormulaCalcResponse{
		Result: nil,
		Error:  error,
	}
}

// NewGetSuccessResponse 创建成功的查询响应
func NewGetSuccessResponse(results []map[string]interface{}) *SugarFormulaGetResponse {
	return &SugarFormulaGetResponse{
		Results: results,
		Count:   len(results),
		Error:   "",
	}
}

// NewGetErrorResponse 创建错误的查询响应
func NewGetErrorResponse(error string) *SugarFormulaGetResponse {
	return &SugarFormulaGetResponse{
		Results: []map[string]interface{}{},
		Count:   0,
		Error:   error,
	}
}
