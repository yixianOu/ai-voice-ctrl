// Package executor 提供用于注册和执行工具的通用接口。
package executor

import (
	"context"
	"encoding/json"
)

// ToolResult represents the normalized response returned to the LLM.
type ToolResult struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// ToolExecutor is the function signature used by the ToolRegistry.
type ToolExecutor func(ctx context.Context, payload json.RawMessage) (ToolResult, error)

// ToolDefinition describes a callable tool, including its JSON schema.
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  json.RawMessage
	Executor    ToolExecutor
}

// ToolSchema exposes metadata required by LLM providers。
type ToolSchema struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// ToolCall represents an invocation request coming from an LLM。
type ToolCall struct {
	Name      string
	Arguments json.RawMessage
}

// Executor 负责注册/管理工具实例并调度调用。
type Executor interface {
	RegisterTool(name string, definitions map[string]ToolDefinition) error
	UnregisterTool(name string) error
	ToolSchemas() []ToolSchema
	ExecuteTool(ctx context.Context, call ToolCall) (ToolResult, error)
}
