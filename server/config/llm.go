package config

// LLM LLM配置结构
type LLM struct {
	// OpenAI兼容配置
	OpenAI OpenAIConfig `mapstructure:"openai" json:"openai" yaml:"openai"`
	// MCP配置
	MCP MCPConfig `mapstructure:"mcp" json:"mcp" yaml:"mcp"`
	// HTTP配置
	HTTP HTTPConfig `mapstructure:"http" json:"http" yaml:"http"`
}

// OpenAIConfig OpenAI兼容的LLM配置
type OpenAIConfig struct {
	BaseURL   string `mapstructure:"base-url" json:"base-url" yaml:"base-url"`       // OpenAI兼容的API地址
	Token     string `mapstructure:"token" json:"token" yaml:"token"`                // API密钥或访问令牌
	ModelName string `mapstructure:"model-name" json:"model-name" yaml:"model-name"` // 要使用的模型名称
}

// MCPConfig MCP模式的LLM配置
type MCPConfig struct {
	ModelName  string `mapstructure:"model-name" json:"model-name" yaml:"model-name"`    // MCP模式下使用的模型名称
	ServerName string `mapstructure:"server-name" json:"server-name" yaml:"server-name"` // MCP Server的名称
	Token      string `mapstructure:"token" json:"token" yaml:"token"`                   // API密钥或访问令牌
}

// HTTPConfig HTTP请求配置
type HTTPConfig struct {
	TimeoutSeconds int `mapstructure:"timeout-seconds" json:"timeout-seconds" yaml:"timeout-seconds"` // HTTP请求超时时间（秒）
}
