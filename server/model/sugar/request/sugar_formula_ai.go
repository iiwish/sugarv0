package request

// SugarFormulaAiFetchRequest AIFETCH 公式请求结构
type SugarFormulaAiFetchRequest struct {
	AgentName   string `json:"agentName" binding:"required"` // Agent 名称
	Description string `json:"description"`                  // 用户输入的自然语言分析需求
	DataRange   string `json:"dataRange,omitempty"`          // 可选的数据范围，如果提供则优先使用
}

// SugarFormulaAiExplainRequest AIEXPLAIN 公式请求结构
type SugarFormulaAiExplainRangeRequest struct {
	DataSource  [][]interface{} `json:"dataSource" binding:"required"`  // 前端传入的二维数据
	Description string          `json:"description" binding:"required"` // 用户的分析需求
}
