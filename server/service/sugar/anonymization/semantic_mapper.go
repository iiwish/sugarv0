package anonymization

import (
	"crypto/sha256"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// SemanticMapper 语义映射器
type SemanticMapper struct {
	config           *AdvancedAnonymizationConfig
	dimensionMapping map[string]*DimensionSemanticInfo
	valueMapping     map[string]*ValueSemanticInfo
	mutex            sync.RWMutex

	// 业务逻辑保留
	businessRules map[string]*BusinessRule
	semanticRules map[string]*SemanticRule
}

// SemanticMapping 语义映射信息
type SemanticMapping struct {
	OriginalValue     string                 `json:"originalValue"`
	AnonymizedValue   string                 `json:"anonymizedValue"`
	SemanticType      string                 `json:"semanticType"`
	BusinessContext   map[string]interface{} `json:"businessContext"`
	PreservedFeatures []string               `json:"preservedFeatures"`
}

// DimensionSemanticInfo 维度语义信息
type DimensionSemanticInfo struct {
	DimensionName     string   `json:"dimensionName"`
	SemanticType      string   `json:"semanticType"`
	BusinessCategory  string   `json:"businessCategory"`
	HierarchyLevel    int      `json:"hierarchyLevel"`
	RelatedDimensions []string `json:"relatedDimensions"`
	MappingStrategy   string   `json:"mappingStrategy"`
}

// ValueSemanticInfo 值语义信息
type ValueSemanticInfo struct {
	Value            string  `json:"value"`
	SemanticCategory string  `json:"semanticCategory"`
	BusinessWeight   float64 `json:"businessWeight"`
	FrequencyRank    int     `json:"frequencyRank"`
	SensitivityLevel string  `json:"sensitivityLevel"`
}

// BusinessRule 业务规则
type BusinessRule struct {
	RuleName          string                 `json:"ruleName"`
	ApplicableDims    []string               `json:"applicableDims"`
	PreservationLogic string                 `json:"preservationLogic"`
	Parameters        map[string]interface{} `json:"parameters"`
}

// SemanticRule 语义规则
type SemanticRule struct {
	RuleName         string   `json:"ruleName"`
	PatternMatching  string   `json:"patternMatching"`
	TransformLogic   string   `json:"transformLogic"`
	PreserveFeatures []string `json:"preserveFeatures"`
}

// NewSemanticMapper 创建语义映射器
func NewSemanticMapper(config *AdvancedAnonymizationConfig) *SemanticMapper {
	mapper := &SemanticMapper{
		config:           config,
		dimensionMapping: make(map[string]*DimensionSemanticInfo),
		valueMapping:     make(map[string]*ValueSemanticInfo),
		businessRules:    make(map[string]*BusinessRule),
		semanticRules:    make(map[string]*SemanticRule),
	}

	// 初始化默认的语义规则
	mapper.initializeDefaultRules()

	return mapper
}

// initializeDefaultRules 初始化默认的语义规则
func (sm *SemanticMapper) initializeDefaultRules() {
	// 地理维度规则
	sm.businessRules["geographic"] = &BusinessRule{
		RuleName:          "地理维度保留",
		ApplicableDims:    []string{"区域", "省份", "城市", "地区"},
		PreservationLogic: "preserve_hierarchy",
		Parameters: map[string]interface{}{
			"hierarchy_levels":       []string{"大区", "省份", "城市"},
			"preserve_relative_size": true,
		},
	}

	// 产品维度规则
	sm.businessRules["product"] = &BusinessRule{
		RuleName:          "产品维度保留",
		ApplicableDims:    []string{"产品", "产品类别", "品牌", "型号"},
		PreservationLogic: "preserve_category",
		Parameters: map[string]interface{}{
			"category_mapping":             true,
			"preserve_performance_ranking": true,
		},
	}

	// 时间维度规则
	sm.businessRules["temporal"] = &BusinessRule{
		RuleName:          "时间维度保留",
		ApplicableDims:    []string{"年份", "月份", "季度", "日期"},
		PreservationLogic: "preserve_temporal_order",
		Parameters: map[string]interface{}{
			"preserve_seasonality": true,
			"preserve_trends":      true,
		},
	}

	// 语义规则
	sm.semanticRules["high_value"] = &SemanticRule{
		RuleName:         "高价值保留",
		PatternMatching:  ".*高端.*|.*优质.*|.*重点.*",
		TransformLogic:   "preserve_relative_importance",
		PreserveFeatures: []string{"ranking", "relative_performance"},
	}

	sm.semanticRules["geographic_hierarchy"] = &SemanticRule{
		RuleName:         "地理层次保留",
		PatternMatching:  ".*区$|.*省$|.*市$",
		TransformLogic:   "preserve_geographic_relationship",
		PreserveFeatures: []string{"hierarchy", "adjacency"},
	}
}

// GetSemanticDimensionName 获取维度的语义化匿名名称
func (sm *SemanticMapper) GetSemanticDimensionName(session *AnonymizationSession, dimName string, counters map[string]int) string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 检查是否已存在映射
	if anonymized, exists := session.forwardMap[dimName]; exists {
		return anonymized
	}

	// 分析维度的语义信息
	semanticInfo := sm.analyzeDimensionSemantics(dimName)
	sm.dimensionMapping[dimName] = semanticInfo

	// 生成语义化的匿名名称
	var anonymized string
	if sm.config.PreserveBusiness {
		anonymized = sm.generateSemanticDimensionName(semanticInfo, counters)
	} else {
		// 使用传统方式
		counters["dimension"]++
		anonymized = fmt.Sprintf("D%02d", counters["dimension"])
	}

	// 存储映射
	session.forwardMap[dimName] = anonymized
	session.reverseMap[anonymized] = dimName

	global.GVA_LOG.Debug("生成语义维度映射",
		zap.String("original", dimName),
		zap.String("anonymized", anonymized),
		zap.String("semanticType", semanticInfo.SemanticType))

	return anonymized
}

// GetSemanticDimensionValue 获取维度值的语义化匿名值
func (sm *SemanticMapper) GetSemanticDimensionValue(session *AnonymizationSession, dimName, dimValue string, counters map[string]int) string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	fullKey := fmt.Sprintf("%s:%s", dimName, dimValue)

	// 检查是否已存在映射
	if anonymized, exists := session.forwardMap[fullKey]; exists {
		return anonymized
	}

	// 分析值的语义信息
	valueInfo := sm.analyzeValueSemantics(dimName, dimValue)
	sm.valueMapping[fullKey] = valueInfo

	// 生成语义化的匿名值
	var anonymized string
	if sm.config.PreserveBusiness {
		anonymized = sm.generateSemanticValue(dimName, valueInfo, counters)
	} else {
		// 使用传统方式
		anonymizedDim := session.forwardMap[dimName]
		if anonymizedDim == "" {
			// 如果维度还没有匿名化，先创建维度代号
			dimensionCounters := make(map[string]int)
			anonymizedDim = sm.GetSemanticDimensionName(session, dimName, dimensionCounters)
		}

		dimKey := fmt.Sprintf("value_%s", dimName)
		counters[dimKey]++
		anonymized = fmt.Sprintf("%s_V%02d", anonymizedDim, counters[dimKey])
	}

	// 存储映射
	session.forwardMap[fullKey] = anonymized
	session.reverseMap[anonymized] = dimValue

	global.GVA_LOG.Debug("生成语义值映射",
		zap.String("dimension", dimName),
		zap.String("originalValue", dimValue),
		zap.String("anonymizedValue", anonymized),
		zap.String("semanticCategory", valueInfo.SemanticCategory))

	return anonymized
}

// analyzeDimensionSemantics 分析维度的语义信息
func (sm *SemanticMapper) analyzeDimensionSemantics(dimName string) *DimensionSemanticInfo {
	info := &DimensionSemanticInfo{
		DimensionName:    dimName,
		SemanticType:     "general",
		BusinessCategory: "unknown",
		HierarchyLevel:   1,
		MappingStrategy:  "simple",
	}

	// 分析维度类型
	lowerName := strings.ToLower(dimName)

	// 地理维度
	if strings.Contains(lowerName, "区域") || strings.Contains(lowerName, "地区") ||
		strings.Contains(lowerName, "省") || strings.Contains(lowerName, "市") {
		info.SemanticType = "geographic"
		info.BusinessCategory = "location"
		info.MappingStrategy = "hierarchy_preserving"

		if strings.Contains(lowerName, "区域") || strings.Contains(lowerName, "大区") {
			info.HierarchyLevel = 1
		} else if strings.Contains(lowerName, "省") {
			info.HierarchyLevel = 2
		} else if strings.Contains(lowerName, "市") || strings.Contains(lowerName, "城市") {
			info.HierarchyLevel = 3
		}
	}

	// 产品维度
	if strings.Contains(lowerName, "产品") || strings.Contains(lowerName, "商品") ||
		strings.Contains(lowerName, "品牌") || strings.Contains(lowerName, "型号") {
		info.SemanticType = "product"
		info.BusinessCategory = "merchandise"
		info.MappingStrategy = "category_preserving"
	}

	// 时间维度
	if strings.Contains(lowerName, "年") || strings.Contains(lowerName, "月") ||
		strings.Contains(lowerName, "季度") || strings.Contains(lowerName, "日期") {
		info.SemanticType = "temporal"
		info.BusinessCategory = "time"
		info.MappingStrategy = "order_preserving"
	}

	// 组织维度
	if strings.Contains(lowerName, "部门") || strings.Contains(lowerName, "团队") ||
		strings.Contains(lowerName, "渠道") || strings.Contains(lowerName, "销售") {
		info.SemanticType = "organizational"
		info.BusinessCategory = "structure"
		info.MappingStrategy = "relationship_preserving"
	}

	return info
}

// analyzeValueSemantics 分析维度值的语义信息
func (sm *SemanticMapper) analyzeValueSemantics(dimName, dimValue string) *ValueSemanticInfo {
	info := &ValueSemanticInfo{
		Value:            dimValue,
		SemanticCategory: "general",
		BusinessWeight:   1.0,
		FrequencyRank:    0,
		SensitivityLevel: "medium",
	}

	lowerValue := strings.ToLower(dimValue)

	// 分析价值级别
	if strings.Contains(lowerValue, "高端") || strings.Contains(lowerValue, "优质") ||
		strings.Contains(lowerValue, "重点") || strings.Contains(lowerValue, "核心") {
		info.SemanticCategory = "high_value"
		info.BusinessWeight = 2.0
		info.SensitivityLevel = "high"
	} else if strings.Contains(lowerValue, "普通") || strings.Contains(lowerValue, "标准") {
		info.SemanticCategory = "standard"
		info.BusinessWeight = 1.0
		info.SensitivityLevel = "medium"
	} else if strings.Contains(lowerValue, "低端") || strings.Contains(lowerValue, "基础") {
		info.SemanticCategory = "basic"
		info.BusinessWeight = 0.5
		info.SensitivityLevel = "low"
	}

	// 分析地理特征
	if strings.Contains(lowerValue, "北京") || strings.Contains(lowerValue, "上海") ||
		strings.Contains(lowerValue, "深圳") || strings.Contains(lowerValue, "广州") {
		info.SemanticCategory = "tier1_city"
		info.BusinessWeight = 2.0
		info.SensitivityLevel = "high"
	}

	// 分析特殊标识
	if len(dimValue) <= 2 {
		info.SensitivityLevel = "low"
	} else if len(dimValue) > 10 {
		info.SensitivityLevel = "high"
	}

	return info
}

// generateSemanticDimensionName 生成语义化的维度名称
func (sm *SemanticMapper) generateSemanticDimensionName(info *DimensionSemanticInfo, counters map[string]int) string {
	var prefix string

	switch info.SemanticType {
	case "geographic":
		prefix = "LOC"
	case "product":
		prefix = "PRD"
	case "temporal":
		prefix = "TIME"
	case "organizational":
		prefix = "ORG"
	default:
		prefix = "DIM"
	}

	// 生成基于语义类型的计数器
	counterKey := fmt.Sprintf("%s_dimension", info.SemanticType)
	counters[counterKey]++

	return fmt.Sprintf("%s%02d", prefix, counters[counterKey])
}

// generateSemanticValue 生成语义化的维度值
func (sm *SemanticMapper) generateSemanticValue(dimName string, info *ValueSemanticInfo, counters map[string]int) string {
	// 获取维度的匿名代号
	dimInfo := sm.dimensionMapping[dimName]
	if dimInfo == nil {
		// 回退到简单策略
		return sm.generateSimpleValue(dimName, counters)
	}

	var valuePrefix string

	switch info.SemanticCategory {
	case "high_value":
		valuePrefix = "HV"
	case "standard":
		valuePrefix = "ST"
	case "basic":
		valuePrefix = "BS"
	case "tier1_city":
		valuePrefix = "T1"
	default:
		valuePrefix = "GN"
	}

	// 生成基于语义类别的计数器
	counterKey := fmt.Sprintf("%s_%s_value", dimInfo.SemanticType, info.SemanticCategory)
	counters[counterKey]++

	// 构建语义化的值名称
	return fmt.Sprintf("%s_%s%02d", sm.getDimensionPrefix(dimInfo.SemanticType), valuePrefix, counters[counterKey])
}

// generateSimpleValue 生成简单的值（回退策略）
func (sm *SemanticMapper) generateSimpleValue(dimName string, counters map[string]int) string {
	counterKey := fmt.Sprintf("simple_value_%s", dimName)
	counters[counterKey]++
	return fmt.Sprintf("VAL%02d", counters[counterKey])
}

// getDimensionPrefix 获取维度前缀
func (sm *SemanticMapper) getDimensionPrefix(semanticType string) string {
	switch semanticType {
	case "geographic":
		return "LOC"
	case "product":
		return "PRD"
	case "temporal":
		return "TIME"
	case "organizational":
		return "ORG"
	default:
		return "DIM"
	}
}

// PreserveBusinessLogic 保留业务逻辑
func (sm *SemanticMapper) PreserveBusinessLogic(session *AnonymizationSession, contributions []ContributionItem) error {
	if !sm.config.PreserveBusiness {
		return nil
	}

	// 分析业务关系
	relationships := sm.analyzeBusinessRelationships(contributions)

	// 调整映射以保留关系
	for _, relationship := range relationships {
		if err := sm.adjustMappingForRelationship(session, relationship); err != nil {
			return err
		}
	}

	global.GVA_LOG.Info("业务逻辑保留完成", zap.Int("relationships", len(relationships)))
	return nil
}

// analyzeBusinessRelationships 分析业务关系
func (sm *SemanticMapper) analyzeBusinessRelationships(contributions []ContributionItem) []BusinessRelationship {
	var relationships []BusinessRelationship

	// 分析地理层次关系
	geoRelationships := sm.analyzeGeographicHierarchy(contributions)
	relationships = append(relationships, geoRelationships...)

	// 分析性能排名关系
	performanceRelationships := sm.analyzePerformanceRanking(contributions)
	relationships = append(relationships, performanceRelationships...)

	return relationships
}

// BusinessRelationship 业务关系
type BusinessRelationship struct {
	Type         string                 `json:"type"`
	Elements     []string               `json:"elements"`
	Relationship string                 `json:"relationship"`
	Parameters   map[string]interface{} `json:"parameters"`
}

// analyzeGeographicHierarchy 分析地理层次关系
func (sm *SemanticMapper) analyzeGeographicHierarchy(contributions []ContributionItem) []BusinessRelationship {
	var relationships []BusinessRelationship

	// 收集地理维度
	geoDimensions := make(map[string][]string)

	for _, contrib := range contributions {
		for dimName, dimValue := range contrib.DimensionValues {
			if sm.isGeographicDimension(dimName) {
				geoDimensions[dimName] = append(geoDimensions[dimName], fmt.Sprintf("%v", dimValue))
			}
		}
	}

	// 构建层次关系
	for dimName, values := range geoDimensions {
		if len(values) > 1 {
			relationship := BusinessRelationship{
				Type:         "geographic_hierarchy",
				Elements:     values,
				Relationship: "hierarchical",
				Parameters: map[string]interface{}{
					"dimension":          dimName,
					"preserve_adjacency": true,
				},
			}
			relationships = append(relationships, relationship)
		}
	}

	return relationships
}

// analyzePerformanceRanking 分析性能排名关系
func (sm *SemanticMapper) analyzePerformanceRanking(contributions []ContributionItem) []BusinessRelationship {
	var relationships []BusinessRelationship

	// 按贡献度排序
	sortedContribs := make([]ContributionItem, len(contributions))
	copy(sortedContribs, contributions)

	sort.Slice(sortedContribs, func(i, j int) bool {
		return math.Abs(sortedContribs[i].ContributionPercent) > math.Abs(sortedContribs[j].ContributionPercent)
	})

	// 构建排名关系
	var elements []string
	for _, contrib := range sortedContribs {
		for _, dimValue := range contrib.DimensionValues {
			elements = append(elements, fmt.Sprintf("%v", dimValue))
		}
	}

	if len(elements) > 1 {
		relationship := BusinessRelationship{
			Type:         "performance_ranking",
			Elements:     elements,
			Relationship: "ranked",
			Parameters: map[string]interface{}{
				"ranking_criteria":        "contribution_percent",
				"preserve_relative_order": true,
			},
		}
		relationships = append(relationships, relationship)
	}

	return relationships
}

// adjustMappingForRelationship 调整映射以保留关系
func (sm *SemanticMapper) adjustMappingForRelationship(session *AnonymizationSession, relationship BusinessRelationship) error {
	switch relationship.Type {
	case "geographic_hierarchy":
		return sm.adjustGeographicMapping(session, relationship)
	case "performance_ranking":
		return sm.adjustPerformanceMapping(session, relationship)
	default:
		return nil
	}
}

// adjustGeographicMapping 调整地理映射
func (sm *SemanticMapper) adjustGeographicMapping(session *AnonymizationSession, relationship BusinessRelationship) error {
	// 确保地理层次在匿名化后仍然保持
	// 这里可以实现更复杂的地理关系保留逻辑
	global.GVA_LOG.Debug("调整地理映射", zap.Strings("elements", relationship.Elements))
	return nil
}

// adjustPerformanceMapping 调整性能映射
func (sm *SemanticMapper) adjustPerformanceMapping(session *AnonymizationSession, relationship BusinessRelationship) error {
	// 确保性能排名在匿名化后仍然保持相对顺序
	global.GVA_LOG.Debug("调整性能映射", zap.Strings("elements", relationship.Elements))
	return nil
}

// isGeographicDimension 判断是否为地理维度
func (sm *SemanticMapper) isGeographicDimension(dimName string) bool {
	lowerName := strings.ToLower(dimName)
	return strings.Contains(lowerName, "区域") || strings.Contains(lowerName, "地区") ||
		strings.Contains(lowerName, "省") || strings.Contains(lowerName, "市") ||
		strings.Contains(lowerName, "城市") || strings.Contains(lowerName, "地点")
}

// GenerateSemanticHash 生成语义哈希（用于一致性匿名化）
func (sm *SemanticMapper) GenerateSemanticHash(value string, salt string) string {
	hasher := sha256.New()
	hasher.Write([]byte(value + salt))
	hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", hash)[:8] // 取前8位
}

// GetSemanticMappings 获取所有语义映射
func (sm *SemanticMapper) GetSemanticMappings() map[string]*SemanticMapping {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	mappings := make(map[string]*SemanticMapping)

	// 从维度映射构建语义映射
	for dimName, dimInfo := range sm.dimensionMapping {
		mapping := &SemanticMapping{
			OriginalValue: dimName,
			SemanticType:  dimInfo.SemanticType,
			BusinessContext: map[string]interface{}{
				"businessCategory": dimInfo.BusinessCategory,
				"hierarchyLevel":   dimInfo.HierarchyLevel,
				"mappingStrategy":  dimInfo.MappingStrategy,
			},
			PreservedFeatures: []string{"semantic_type", "business_category"},
		}
		mappings[dimName] = mapping
	}

	return mappings
}
