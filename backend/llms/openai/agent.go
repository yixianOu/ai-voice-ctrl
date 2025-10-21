// Package openai implements an OpenAI-based agent with function calling capabilities.
package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-voice-ctrl/backend/executor"
	"ai-voice-ctrl/backend/llms"

	"github.com/sashabaranov/go-openai"
)

// OpenAIAgent implements Agent interface using OpenAI Chat API with function calling
type OpenAIAgent struct {
	client   *openai.Client
	executor executor.Executor
	config   llms.AgentConfig
	messages []openai.ChatCompletionMessage
	tools    []openai.Tool
}

// NewOpenAIAgent creates a new OpenAI agent with executor injection
func NewOpenAIAgent(apiKey string, exec executor.Executor) llms.Agent {
	config := llms.DefaultAgentConfig()
	return NewOpenAIAgentWithConfig(apiKey, exec, config)
}

// NewOpenAIAgentWithConfig creates agent with custom config
func NewOpenAIAgentWithConfig(apiKey string, exec executor.Executor, config llms.AgentConfig) llms.Agent {
	client := openai.NewClient(apiKey)

	agent := &OpenAIAgent{
		client:   client,
		executor: exec,
		config:   config,
		messages: make([]openai.ChatCompletionMessage, 0),
		tools:    make([]openai.Tool, 0),
	}

	// Initialize system prompt
	if config.SystemPrompt != "" {
		agent.messages = append(agent.messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: config.SystemPrompt,
		})
	}

	// Convert executor tools to OpenAI format
	agent.prepareTools()

	return agent
}

// prepareTools converts executor tool schemas to OpenAI tool format
func (a *OpenAIAgent) prepareTools() {
	schemas := a.executor.ToolSchemas()
	a.tools = make([]openai.Tool, 0, len(schemas))

	for _, schema := range schemas {
		var params map[string]interface{}
		if err := json.Unmarshal(schema.Parameters, &params); err != nil {
			continue
		}

		a.tools = append(a.tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        schema.Name,
				Description: schema.Description,
				Parameters:  params,
			},
		})
	}
}

// Chat implements Agent.Chat
func (a *OpenAIAgent) Chat(ctx context.Context, message string) (string, error) {
	// Add user message
	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	// Execute conversation loop with tool calling
	maxIterations := a.config.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10
	}

	for i := 0; i < maxIterations; i++ {
		resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       a.config.Model,
			Messages:    a.messages,
			Tools:       a.tools,
			Temperature: a.config.Temperature,
		})
		if err != nil {
			return "", fmt.Errorf("openai api error: %w", err)
		}

		assistantMsg := resp.Choices[0].Message
		a.messages = append(a.messages, assistantMsg)

		// Check if conversation is finished
		if len(assistantMsg.ToolCalls) == 0 {
			return assistantMsg.Content, nil
		}

		// Execute tool calls
		for _, toolCall := range assistantMsg.ToolCalls {
			result, err := a.executeTool(ctx, toolCall)

			var resultJSON []byte
			if err != nil {
				// Tool execution failed, send error to LLM
				resultJSON, _ = json.Marshal(executor.ToolResult{
					Success: false,
					Message: err.Error(),
				})
			} else {
				resultJSON, _ = json.Marshal(result)
			}

			// Add tool result to messages
			a.messages = append(a.messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    string(resultJSON),
				ToolCallID: toolCall.ID,
			})
		}
	}

	return "", fmt.Errorf("max iterations (%d) reached without final response", maxIterations)
}

// executeTool executes a single tool call
func (a *OpenAIAgent) executeTool(ctx context.Context, toolCall openai.ToolCall) (executor.ToolResult, error) {
	call := executor.ToolCall{
		Name:      toolCall.Function.Name,
		Arguments: json.RawMessage(toolCall.Function.Arguments),
	}

	return a.executor.ExecuteTool(ctx, call)
}

// ChatWithHistory implements Agent.ChatWithHistory
func (a *OpenAIAgent) ChatWithHistory(ctx context.Context, messages []llms.Message) (llms.AgentResponse, error) {
	// Convert llms.Message to openai.ChatCompletionMessage
	a.messages = make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, msg := range messages {
		oaiMsg := openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}

		if len(msg.ToolCalls) > 0 {
			oaiMsg.ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				oaiMsg.ToolCalls[i] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolType(tc.Type),
					Function: openai.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}

		if msg.ToolCallID != "" {
			oaiMsg.ToolCallID = msg.ToolCallID
		}

		a.messages = append(a.messages, oaiMsg)
	}

	// Execute conversation
	text, err := a.Chat(ctx, "")
	if err != nil {
		return llms.AgentResponse{}, err
	}

	return llms.AgentResponse{
		Text: text,
	}, nil
}

// Reset implements Agent.Reset
func (a *OpenAIAgent) Reset() {
	a.messages = make([]openai.ChatCompletionMessage, 0)

	// Re-add system prompt if exists
	if a.config.SystemPrompt != "" {
		a.messages = append(a.messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.config.SystemPrompt,
		})
	}
}

// GetHistory implements Agent.GetHistory
func (a *OpenAIAgent) GetHistory() []llms.Message {
	history := make([]llms.Message, 0, len(a.messages))

	for _, msg := range a.messages {
		llmMsg := llms.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}

		if len(msg.ToolCalls) > 0 {
			llmMsg.ToolCalls = make([]llms.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				llmMsg.ToolCalls[i] = llms.ToolCall{
					ID:   tc.ID,
					Type: string(tc.Type),
				}
				llmMsg.ToolCalls[i].Function.Name = tc.Function.Name
				llmMsg.ToolCalls[i].Function.Arguments = tc.Function.Arguments
			}
		}

		if msg.ToolCallID != "" {
			llmMsg.ToolCallID = msg.ToolCallID
		}

		history = append(history, llmMsg)
	}

	return history
}

// SetSystemPrompt implements Agent.SetSystemPrompt
func (a *OpenAIAgent) SetSystemPrompt(prompt string) {
	a.config.SystemPrompt = prompt

	// Update messages
	if len(a.messages) > 0 && a.messages[0].Role == openai.ChatMessageRoleSystem {
		a.messages[0].Content = prompt
	} else {
		// Insert at beginning
		a.messages = append([]openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
		}, a.messages...)
	}
}
