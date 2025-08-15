package anonymization

import (
	"math"
	"math/rand"
	"sync"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// NoiseGenerator 噪声生成器
type NoiseGenerator struct {
	config *AdvancedAnonymizationConfig
	rand   *rand.Rand
	mutex  sync.Mutex

	// 噪声统计
	noiseStats map[string]*NoiseStats
}

// NoiseStats 噪声统计信息
type NoiseStats struct {
	TotalNoise   float64 `json:"totalNoise"`
	NoiseCount   int     `json:"noiseCount"`
	AverageNoise float64 `json:"averageNoise"`
	MaxNoise     float64 `json:"maxNoise"`
	MinNoise     float64 `json:"minNoise"`
	Scale        float64 `json:"scale"`
}

// NewNoiseGenerator 创建噪声生成器
func NewNoiseGenerator(config *AdvancedAnonymizationConfig) *NoiseGenerator {
	return &NoiseGenerator{
		config:     config,
		rand:       rand.New(rand.NewSource(config.RandomSeed)),
		noiseStats: make(map[string]*NoiseStats),
	}
}

// GenerateLaplaceNoise 生成拉普拉斯噪声（用于差分隐私）
func (ng *NoiseGenerator) GenerateLaplaceNoise(scale float64) float64 {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	// 使用反函数方法生成拉普拉斯分布
	u := ng.rand.Float64() - 0.5
	noise := -scale * math.Copysign(math.Log(1-2*math.Abs(u)), u)

	global.GVA_LOG.Debug("生成拉普拉斯噪声",
		zap.Float64("scale", scale),
		zap.Float64("noise", noise))

	return noise
}

// GenerateGaussianNoise 生成高斯噪声
func (ng *NoiseGenerator) GenerateGaussianNoise(mean, stddev float64) float64 {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	// Box-Muller变换生成高斯分布
	if ng.rand == nil {
		ng.rand = rand.New(rand.NewSource(ng.config.RandomSeed))
	}

	u1 := ng.rand.Float64()
	u2 := ng.rand.Float64()

	z0 := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
	noise := mean + stddev*z0

	global.GVA_LOG.Debug("生成高斯噪声",
		zap.Float64("mean", mean),
		zap.Float64("stddev", stddev),
		zap.Float64("noise", noise))

	return noise
}

// GenerateExponentialNoise 生成指数噪声
func (ng *NoiseGenerator) GenerateExponentialNoise(lambda float64) float64 {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	u := ng.rand.Float64()
	noise := -math.Log(1-u) / lambda

	return noise
}

// GenerateAdaptiveNoise 生成自适应噪声
func (ng *NoiseGenerator) GenerateAdaptiveNoise(baseNoise, sensitivityScore float64) float64 {
	// 根据敏感性分数调整噪声强度
	adaptiveFactor := 1.0 + sensitivityScore*ng.config.NoiseVariance
	adaptedNoise := baseNoise * adaptiveFactor

	global.GVA_LOG.Debug("生成自适应噪声",
		zap.Float64("baseNoise", baseNoise),
		zap.Float64("sensitivityScore", sensitivityScore),
		zap.Float64("adaptedNoise", adaptedNoise))

	return adaptedNoise
}

// RecordNoiseParameters 记录噪声参数
func (ng *NoiseGenerator) RecordNoiseParameters(metricType string, noise, scale float64) {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	if ng.noiseStats[metricType] == nil {
		ng.noiseStats[metricType] = &NoiseStats{
			MaxNoise: noise,
			MinNoise: noise,
			Scale:    scale,
		}
	}

	stats := ng.noiseStats[metricType]
	stats.TotalNoise += math.Abs(noise)
	stats.NoiseCount++
	stats.AverageNoise = stats.TotalNoise / float64(stats.NoiseCount)

	if math.Abs(noise) > math.Abs(stats.MaxNoise) {
		stats.MaxNoise = noise
	}
	if math.Abs(noise) < math.Abs(stats.MinNoise) {
		stats.MinNoise = noise
	}
}

// GetNoiseStatistics 获取噪声统计信息
func (ng *NoiseGenerator) GetNoiseStatistics() map[string]*NoiseStats {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	// 返回副本以确保线程安全
	result := make(map[string]*NoiseStats)
	for key, stats := range ng.noiseStats {
		result[key] = &NoiseStats{
			TotalNoise:   stats.TotalNoise,
			NoiseCount:   stats.NoiseCount,
			AverageNoise: stats.AverageNoise,
			MaxNoise:     stats.MaxNoise,
			MinNoise:     stats.MinNoise,
			Scale:        stats.Scale,
		}
	}

	return result
}

// CalculateOptimalEpsilon 计算最优隐私预算
func (ng *NoiseGenerator) CalculateOptimalEpsilon(dataSize int, targetUtility float64) float64 {
	// 基于数据大小和目标效用计算最优ε值
	// 这是一个简化的启发式方法

	baseEpsilon := 0.5
	sizeAdjustment := math.Log(float64(dataSize)) / 10.0
	utilityAdjustment := targetUtility * 0.5

	optimalEpsilon := baseEpsilon + sizeAdjustment + utilityAdjustment

	// 限制在合理范围内
	if optimalEpsilon < 0.01 {
		optimalEpsilon = 0.01
	}
	if optimalEpsilon > 2.0 {
		optimalEpsilon = 2.0
	}

	global.GVA_LOG.Info("计算最优隐私预算",
		zap.Int("dataSize", dataSize),
		zap.Float64("targetUtility", targetUtility),
		zap.Float64("optimalEpsilon", optimalEpsilon))

	return optimalEpsilon
}

// AddCalibratedNoise 添加校准噪声（确保满足差分隐私）
func (ng *NoiseGenerator) AddCalibratedNoise(value, sensitivity, epsilon float64) float64 {
	scale := sensitivity / epsilon
	noise := ng.GenerateLaplaceNoise(scale)
	noisyValue := value + noise

	// 记录参数以便审计
	ng.RecordNoiseParameters("calibrated", noise, scale)

	return noisyValue
}

// ComputePrivacyLoss 计算隐私损失
func (ng *NoiseGenerator) ComputePrivacyLoss(originalValue, noisyValue, sensitivity float64) float64 {
	if sensitivity <= 0 {
		return 0
	}

	difference := math.Abs(noisyValue - originalValue)
	privacyLoss := difference / sensitivity

	return privacyLoss
}

// Reset 重置噪声生成器状态
func (ng *NoiseGenerator) Reset() {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	ng.noiseStats = make(map[string]*NoiseStats)
	ng.rand = rand.New(rand.NewSource(ng.config.RandomSeed))
}
