package sugar

import "github.com/flipped-aurora/gin-vue-admin/server/service/system"

// AnonymizationSession 匿名化会话，为单次请求保存状态
type AnonymizationSession struct {
	// forwardMap 用于编码： "华东区域" -> "D01_V01"
	forwardMap map[string]string
	// reverseMap 用于解码： "D01_V01" -> "华东区域"
	reverseMap map[string]string
	// AIReadyData 是准备好发送给AI的、完全匿名化的数据
	AIReadyData []map[string]interface{}
}

// ToolCallResponse 用于解析LLM返回的工具调用指令
type ToolCallResponse struct {
	Type    string                  `json:"type"`
	Content []system.OpenAIToolCall `json:"content"`
}

// ContributionItem 表示单个维度组合的贡献度分析结果
type ContributionItem struct {
	DimensionValues     map[string]interface{} // 维度值组合，如 {"区域": "华东", "产品": "A产品"}
	CurrentValue        float64                // 本期值
	BaseValue           float64                // 基期值
	ChangeValue         float64                // 变化值 (本期值 - 基期值)
	ContributionPercent float64                // 贡献度百分比
	IsPositiveDriver    bool                   // 是否为正向驱动因子
}

// DataValidationResult 数据验证结果
type DataValidationResult struct {
	IsDataAvailable   bool     `json:"is_data_available"`  // 数据是否可用
	ValidationMessage string   `json:"validation_message"` // 验证结果消息
	RecordCount       int      `json:"record_count"`       // 记录数量
	MissingDimensions []string `json:"missing_dimensions"` // 缺失的维度
}

// DataScopeInfo 数据范围信息结构
type DataScopeInfo struct {
	TotalRecords       int                      `json:"total_records"`       // 总记录数
	DimensionCoverage  map[string][]string      `json:"dimension_coverage"`  // 各维度的可用值列表
	SampleData         []map[string]interface{} `json:"sample_data"`         // 样本数据
	DataQualityInfo    map[string]interface{}   `json:"data_quality_info"`   // 数据质量信息
	RecommendedFilters map[string]interface{}   `json:"recommended_filters"` // 推荐的筛选条件
}
