package executor

import (
	"context"
	"fmt"
	"sync"
)

// DefaultExecutor 实现 Executor，负责管理工具定义与实例生命周期。
type DefaultExecutor struct {
	tools     map[string]map[string]ToolDefinition
	instances sync.Map
}

// NewDefaultExecutor creates an empty DefaultExecutor instance.
func NewDefaultExecutor() *DefaultExecutor {
	return &DefaultExecutor{
		tools: make(map[string]map[string]ToolDefinition),
	}
}

// RegisterTool registers a tool class and its definitions.
func (e *DefaultExecutor) RegisterTool(name string, definitions map[string]ToolDefinition) error {
	if name == "" {
		return fmt.Errorf("tool name is required")
	}

	if len(definitions) == 0 {
		return fmt.Errorf("tool %s definitions is empty", name)
	}

	if _, exists := e.tools[name]; exists {
		return fmt.Errorf("tool %s already registered", name)
	}

	copy := make(map[string]ToolDefinition, len(definitions))
	for fn, def := range definitions {
		copy[fn] = def
	}

	e.tools[name] = copy
	return nil
}

// UnregisterTool removes a tool class.
func (e *DefaultExecutor) UnregisterTool(name string) error {
	if _, ok := e.tools[name]; !ok {
		return fmt.Errorf("tool %s not registered", name)
	}
	delete(e.tools, name)
	return nil
}

// ToolSchemas returns lightweight schema information for LLM registration.
func (e *DefaultExecutor) ToolSchemas() []ToolSchema {
	schemas := make([]ToolSchema, 0)
	for _, defs := range e.tools {
		for _, def := range defs {
			schemas = append(schemas, ToolSchema{
				Name:        def.Name,
				Description: def.Description,
				Parameters:  def.Parameters,
			})
		}
	}
	return schemas
}

// ExecuteTool dispatches a tool call to the matching ToolDefinition.
func (e *DefaultExecutor) ExecuteTool(ctx context.Context, call ToolCall) (ToolResult, error) {
	for _, defs := range e.tools {
		if def, ok := defs[call.Name]; ok {
			return def.Executor(ctx, call.Arguments)
		}
	}
	return ToolResult{Success: false, Message: "tool not found"}, fmt.Errorf("tool %s not registered", call.Name)
}

// StoreInstance stores a tool instance for the session.
func (e *DefaultExecutor) StoreInstance(name string, instance interface{}) {
	e.instances.Store(name, instance)
}

// LoadInstance retrieves a tool instance from the session.
func (e *DefaultExecutor) LoadInstance(name string) (interface{}, bool) {
	return e.instances.Load(name)
}

// DeleteInstance removes a tool instance from the session.
func (e *DefaultExecutor) DeleteInstance(name string) {
	e.instances.Delete(name)
}

var _ Executor = (*DefaultExecutor)(nil)
