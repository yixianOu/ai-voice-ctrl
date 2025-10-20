package executor

import (
	"context"
	"fmt"
)

// DefaultExecutor implements the Executor interface and aggregates tool registries.
type DefaultExecutor struct {
	registries map[string]ToolRegistry
}

// NewDefaultExecutor creates an empty DefaultExecutor instance.
func NewDefaultExecutor() *DefaultExecutor {
	return &DefaultExecutor{
		registries: make(map[string]ToolRegistry),
	}
}

// RegisterRegistry adds a ToolRegistry under the given name.
func (e *DefaultExecutor) RegisterRegistry(name string, registry ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("registry %s is nil", name)
	}

	if _, exists := e.registries[name]; exists {
		return fmt.Errorf("registry %s already registered", name)
	}

	e.registries[name] = registry
	return nil
}

// ToolDefinitions returns the merged tool definitions from all registries.
func (e *DefaultExecutor) ToolDefinitions() map[string]ToolDefinition {
	result := make(map[string]ToolDefinition)

	for _, registry := range e.registries {
		for name, def := range registry.Tools() {
			result[name] = def
		}
	}

	return result
}

// ToolSchemas returns lightweight schema information for LLM registration.
func (e *DefaultExecutor) ToolSchemas() []ToolSchema {
	definitions := e.ToolDefinitions()
	schemas := make([]ToolSchema, 0, len(definitions))
	for _, def := range definitions {
		schemas = append(schemas, ToolSchema{
			Name:        def.Name,
			Description: def.Description,
			Parameters:  def.Parameters,
		})
	}
	return schemas
}

// ExecuteTool dispatches a tool call to the matching ToolDefinition.
func (e *DefaultExecutor) ExecuteTool(ctx context.Context, call ToolCall) (ToolResult, error) {
	definitions := e.ToolDefinitions()
	definition, ok := definitions[call.Name]
	if !ok {
		return ToolResult{Success: false, Message: "tool not found"}, fmt.Errorf("tool %s not registered", call.Name)
	}

	return definition.Executor(ctx, call.Arguments)
}

var _ Executor = (*DefaultExecutor)(nil)
