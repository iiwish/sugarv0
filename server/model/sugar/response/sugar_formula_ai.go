package response

// SugarFormulaAiResponse AI公式通用响应结构
type SugarFormulaAiResponse struct {
	Result [][]interface{} `json:"result,omitempty"` // AI返回的二维数组结果（用于AIFETCH）
	Text   string          `json:"text,omitempty"`   // AI返回的文本分析结果（用于AIEXPLAIN）
	Error  string          `json:"error,omitempty"`  // 执行过程中的错误信息
}

// NewAiSuccessResponseWithData 创建成功的AI响应（带数据结果）
func NewAiSuccessResponseWithData(result [][]interface{}) *SugarFormulaAiResponse {
	return &SugarFormulaAiResponse{
		Result: result,
		Text:   "",
		Error:  "",
	}
}

// NewAiSuccessResponseWithText 创建成功的AI响应（带文本结果）
func NewAiSuccessResponseWithText(text string) *SugarFormulaAiResponse {
	return &SugarFormulaAiResponse{
		Result: nil,
		Text:   text,
		Error:  "",
	}
}

// NewAiSuccessResponseWithBoth 创建成功的AI响应（带数据和文本结果）
func NewAiSuccessResponseWithBoth(result [][]interface{}, text string) *SugarFormulaAiResponse {
	return &SugarFormulaAiResponse{
		Result: result,
		Text:   text,
		Error:  "",
	}
}

// NewAiErrorResponse 创建错误的AI响应
func NewAiErrorResponse(error string) *SugarFormulaAiResponse {
	return &SugarFormulaAiResponse{
		Result: nil,
		Text:   "",
		Error:  error,
	}
}
