package llms

import "context"

// Agent interface for LLM with function calling
type Agent interface {
	// Chat processes user message and returns assistant response
	Chat(ctx context.Context, message string) (string, error)

	// ChatWithHistory processes conversation with full history
	ChatWithHistory(ctx context.Context, messages []Message) (AgentResponse, error)

	// Reset clears conversation history
	Reset()

	// GetHistory returns conversation history
	GetHistory() []Message

	// SetSystemPrompt sets system prompt
	SetSystemPrompt(prompt string)
}

// Message represents a conversation message
type Message struct {
	Role       string     `json:"role"` // "system", "user", "assistant", "tool"
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // for assistant messages
	ToolCallID string     `json:"tool_call_id,omitempty"` // for tool messages
}

// ToolCall represents a function call from LLM
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON string
	} `json:"function"`
}

// AgentResponse contains the agent's response with metadata
type AgentResponse struct {
	// Text response from assistant
	Text string

	// ToolCalls made during conversation
	ToolCalls []ToolCall

	// FinishReason indicates why generation stopped
	FinishReason string

	// Usage statistics
	Usage *Usage
}

// Usage contains token usage statistics
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// AgentConfig contains configuration for agent
type AgentConfig struct {
	// Model name
	Model string

	// System prompt
	SystemPrompt string

	// Temperature (0.0-2.0)
	Temperature float32

	// Maximum iterations for tool calling loops
	MaxIterations int
}

// DefaultAgentConfig returns default configuration
func DefaultAgentConfig() AgentConfig {
	return AgentConfig{
		Model:         "gpt-4-turbo-preview",
		Temperature:   0.7,
		MaxIterations: 10,
	}
}
