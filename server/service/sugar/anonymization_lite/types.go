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
	// 移除了 CurrentValue, BaseValue, ChangeValue - 只保留贡献度
}

// LiteAnonymizationSession 简化的匿名化会话
type LiteAnonymizationSession struct {
	// 映射关系
	ForwardMap  map[string]string        `json:"forward_map"`   // 原始 -> 匿名
	ReverseMap  map[string]string        `json:"reverse_map"`   // 匿名 -> 原始
	AIReadyData []map[string]interface{} `json:"ai_ready_data"` // AI可读数据

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
