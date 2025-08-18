package anonymization_lite

import "time"

// LiteConfig 简化的匿名化配置
type LiteConfig struct {
	// 基础匿名化设置
	UseSemanticMapping bool `json:"use_semantic_mapping"` // 是否使用语义映射

	// 噪声控制
	NoiseLevel float64 `json:"noise_level"` // 统一的噪声级别 (0.0-1.0)

	// 随机种子
	RandomSeed int64 `json:"random_seed"`
}

// DefaultLiteConfig 返回默认的简化配置
func DefaultLiteConfig() *LiteConfig {
	return &LiteConfig{
		UseSemanticMapping: true,
		NoiseLevel:         0.1, // 10%的轻微噪声
		RandomSeed:         time.Now().UnixNano(),
	}
}

// ContributionItem 贡献度分析项（简化版）
type ContributionItem struct {
	DimensionValues     map[string]interface{} `json:"dimension_values"`     // 维度值组合
	ContributionPercent float64                `json:"contribution_percent"` // 贡献度百分比
	IsPositiveDriver    bool                   `json:"is_positive_driver"`   // 是否为正向驱动因子

	// 新增字段：提供更丰富的分析信息（比例化数据，避免泄露绝对值）
	ChangeRatePercent  float64 `json:"change_rate_percent"` // 变化率百分比（相对于基期的变化比例）
	TrendDirection     string  `json:"trend_direction"`     // 趋势方向："增长"、"下降"、"持平"
	ImpactLevel        string  `json:"impact_level"`        // 影响程度："高"、"中"、"低"
	RelativeImportance float64 `json:"relative_importance"` // 相对重要性（0-100，基于贡献度绝对值的排名百分位）
}

// DimensionSemanticInfo 维度语义信息
type DimensionSemanticInfo struct {
	AnonymizedName string `json:"anonymized_name"` // 匿名化后的维度名称
	OriginalName   string `json:"original_name"`   // 原始维度名称
	SemanticType   string `json:"semantic_type"`   // 语义类型：地区、产品、时间、部门等
	Description    string `json:"description"`     // 维度描述，用于AI理解
}

// LiteAnonymizationSession 简化的匿名化会话
type LiteAnonymizationSession struct {
	// 映射关系
	ForwardMap  map[string]string        `json:"forward_map"`   // 原始 -> 匿名
	ReverseMap  map[string]string        `json:"reverse_map"`   // 匿名 -> 原始
	AIReadyData []map[string]interface{} `json:"ai_ready_data"` // AI可读数据

	// 新增：维度语义信息
	DimensionSemantics map[string]*DimensionSemanticInfo `json:"dimension_semantics"` // 维度语义映射

	// 会话信息
	Config    *LiteConfig `json:"config"`
	CreatedAt time.Time   `json:"created_at"`
	UserID    string      `json:"user_id"`

	// 统计信息
	MappingCount      int `json:"mapping_count"`
	ContributionCount int `json:"contribution_count"`
}

// AIAnalysisRequest AI分析请求（简化版）
type AIAnalysisRequest struct {
	ModelName            string                 `json:"model_name"`
	TargetMetric         string                 `json:"target_metric"`
	CurrentPeriodFilters map[string]interface{} `json:"current_period_filters"`
	BasePeriodFilters    map[string]interface{} `json:"base_period_filters"`
	GroupByDimensions    []string               `json:"group_by_dimensions"`
	UserID               string                 `json:"user_id"`
}

// LiteAnonymizationError 简化的错误类型
type LiteAnonymizationError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *LiteAnonymizationError) Error() string {
	return e.Message
}

// NewLiteAnonymizationError 创建新的错误
func NewLiteAnonymizationError(message, code string) *LiteAnonymizationError {
	return &LiteAnonymizationError{
		Message: message,
		Code:    code,
	}
}
