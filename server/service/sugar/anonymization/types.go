package anonymization

// AIAnalysisRequest 定义了一个用于贡献度归因分析的请求。
// 它描述了需要比较的两个数据集、分析的目标指标以及分组维度。
type AIAnalysisRequest struct {
	ModelName            string                 `json:"modelName"`            // 语义模型名称
	TargetMetric         string                 `json:"targetMetric"`         // 需要分析的核心指标列名，例如 "指标金额"
	CurrentPeriodFilters map[string]interface{} `json:"currentPeriodFilters"` // 用于获取"本期"数据的筛选条件
	BasePeriodFilters    map[string]interface{} `json:"basePeriodFilters"`    // 用于获取"基期"（如上期或预算）数据的筛选条件
	GroupByDimensions    []string               `json:"groupByDimensions"`    // 进行分组和归因分析的维度列名列表
}

// AnonymizationSession 为单次请求保存所有状态，特别是编码/解码所需的映射表。
// 这确保了整个服务的无状态和线程安全。
type AnonymizationSession struct {
	// forwardMap 用于编码： "华东区域" -> "D01_V01"
	forwardMap map[string]string

	// reverseMap 用于解码： "D01_V01" -> "华东区域"
	reverseMap map[string]string

	// AIReadyData 是准备好发送给AI的、完全匿名化的数据。
	// 结构是 []map[string]interface{}，例如：
	// [{"item_code": "D01_V01", "contribution_percent": 55.5, "is_positive_driver": true}]
	AIReadyData []map[string]interface{}
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

// Validate 验证AIAnalysisRequest的有效性
func (req *AIAnalysisRequest) Validate() error {
	if req.ModelName == "" {
		return NewValidationError("modelName不能为空")
	}
	if req.TargetMetric == "" {
		return NewValidationError("targetMetric不能为空")
	}
	if len(req.GroupByDimensions) == 0 {
		return NewValidationError("groupByDimensions不能为空")
	}
	if req.CurrentPeriodFilters == nil {
		return NewValidationError("currentPeriodFilters不能为空")
	}
	if req.BasePeriodFilters == nil {
		return NewValidationError("basePeriodFilters不能为空")
	}
	return nil
}
