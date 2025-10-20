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

// ToolSchema exposes metadata required by LLM providers.
type ToolSchema struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// ToolCall represents an invocation request coming from an LLM.
type ToolCall struct {
	Name      string
	Arguments json.RawMessage
}

// ToolRegistry provides tool definitions for a specific application domain.
type ToolRegistry interface {
	Tools() map[string]ToolDefinition
}

// Executor aggregates multiple tool registries and dispatches tool calls.
type Executor interface {
	// RegisterRegistry adds a named registry for later lookup.
	RegisterRegistry(name string, registry ToolRegistry) error

	// ToolDefinitions returns the merged tool definitions keyed by tool name.
	ToolDefinitions() map[string]ToolDefinition

	// ToolSchemas returns lightweight schemas for LLM registration.
	ToolSchemas() []ToolSchema

	// ExecuteTool dispatches a tool call to the matching registry.
	ExecuteTool(ctx context.Context, call ToolCall) (ToolResult, error)
}
