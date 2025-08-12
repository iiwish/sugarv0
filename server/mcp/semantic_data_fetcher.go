package mcpTool

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
)

func init() {
	RegisterTool(&SemanticDataFetcher{})
}

type SemanticDataFetcher struct{}

// Handle 语义数据获取工具的处理器
// 这个Handle实际上不应该被直接调用。它的主要目的是为LLM提供一个工具定义的schema。
// 真正的执行逻辑应该在上层服务（如sugar_formula_ai_service）中根据LLM的响应来处理。
func (t *SemanticDataFetcher) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 这个工具的实现在sugar_formula_ai_service.go中，这里返回一个提示性错误。
	return nil, errors.New("semantic_data_fetcher不应直接通过MCP服务器调用，其逻辑由AIFETCH公式在内部处理")
}

// New 返回工具注册信息
func (t *SemanticDataFetcher) New() mcp.Tool {
	return mcp.NewTool("semantic_data_fetcher",
		mcp.WithDescription("根据指定的语义模型、维度、度量和筛选条件，从数据源获取数据。这是AIFETCH公式的内部实现工具。"),
		mcp.WithString("modelName",
			mcp.Required(),
			mcp.Description("要查询的语义模型名称。"),
		),
		mcp.WithArray("returnColumns",
			mcp.Required(),
			mcp.Description("需要返回的列名数组，可以是维度或度量。"),
			mcp.Items(mcp.WithString("", mcp.Description("列名"))),
		),
		mcp.WithObject("filters",
			mcp.Description("筛选条件，格式为 {\"列名\": \"筛选值\"}。"),
		),
		mcp.WithString("userId",
			mcp.Required(),
			mcp.Description("发起请求的用户ID，工具内部需要此参数进行鉴权。"),
		),
	)
}
