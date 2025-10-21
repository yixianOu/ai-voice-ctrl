package executor_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"ai-voice-ctrl/backend/executor"
	"ai-voice-ctrl/backend/executor/vscode"

	"github.com/sashabaranov/go-openai"
)

// TestOpenAIVSCodeIntegration tests VSCode tool integration with OpenAI function calling
func TestOpenAIVSCodeIntegration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping OpenAI integration test")
	}

	// 1. Initialize executor and register VSCode lifecycle
	exec := executor.NewDefaultExecutor()
	if err := vscode.RegisterVSCodeLifecycle(exec); err != nil {
		t.Fatalf("Failed to register VSCode lifecycle: %v", err)
	}

	// 2. Prepare OpenAI client and tools
	client := openai.NewClient(apiKey)
	schemas := exec.ToolSchemas()

	// Convert to OpenAI tool format
	tools := make([]openai.Tool, 0, len(schemas))
	for _, schema := range schemas {
		var params map[string]interface{}
		if err := json.Unmarshal(schema.Parameters, &params); err != nil {
			t.Fatalf("Failed to parse parameters for %s: %v", schema.Name, err)
		}

		tools = append(tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        schema.Name,
				Description: schema.Description,
				Parameters:  params,
			},
		})
	}

	t.Logf("Registered %d tools for OpenAI", len(tools))

	// 3. Create conversation with OpenAI
	ctx := context.Background()
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleSystem,
			Content: `You are a helpful assistant that can manage VSCode editor. 
Follow these steps:
1. Create a VSCode instance with workspace "/tmp/openai_test"
2. Write a file called "hello.md" with some markdown content
3. Append additional content to the file
4. Destroy the VSCode instance

Execute each step using the available tools.`,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Please help me create a VSCode workspace and manage some files as described.",
		},
	}

	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		t.Logf("--- Iteration %d ---", i+1)

		resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    openai.GPT4TurboPreview,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			t.Fatalf("OpenAI API error: %v", err)
		}

		assistantMsg := resp.Choices[0].Message
		messages = append(messages, assistantMsg)

		// Check if conversation is finished
		if len(assistantMsg.ToolCalls) == 0 {
			t.Logf("Assistant response: %s", assistantMsg.Content)
			break
		}

		// Execute tool calls
		for _, toolCall := range assistantMsg.ToolCalls {
			t.Logf("Executing tool: %s with args: %s", toolCall.Function.Name, toolCall.Function.Arguments)

			call := executor.ToolCall{
				Name:      toolCall.Function.Name,
				Arguments: json.RawMessage(toolCall.Function.Arguments),
			}

			result, err := exec.ExecuteTool(ctx, call)
			if err != nil {
				t.Logf("Tool execution error: %v", err)
			}

			resultJSON, _ := json.Marshal(result)
			t.Logf("Tool result: %s", string(resultJSON))

			// Add tool result to messages
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    string(resultJSON),
				ToolCallID: toolCall.ID,
			})
		}
	}

	t.Log("OpenAI VSCode integration test completed")
}

// TestOpenAIVSCodeManualSteps demonstrates manual step-by-step execution
func TestOpenAIVSCodeManualSteps(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping OpenAI integration test")
	}

	exec := executor.NewDefaultExecutor()
	if err := vscode.RegisterVSCodeLifecycle(exec); err != nil {
		t.Fatalf("Failed to register VSCode lifecycle: %v", err)
	}

	client := openai.NewClient(apiKey)
	schemas := exec.ToolSchemas()

	tools := make([]openai.Tool, 0, len(schemas))
	for _, schema := range schemas {
		var params map[string]interface{}
		if err := json.Unmarshal(schema.Parameters, &params); err != nil {
			t.Fatalf("Failed to parse parameters: %v", err)
		}

		tools = append(tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        schema.Name,
				Description: schema.Description,
				Parameters:  params,
			},
		})
	}

	ctx := context.Background()

	// Test single tool call: Create VSCode
	t.Log("Step 1: Ask OpenAI to create VSCode instance")
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4TurboPreview,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: `Create a VSCode workspace in directory "/tmp/test"`,
			},
		},
		Tools: tools,
	})
	if err != nil {
		t.Fatalf("OpenAI API error: %v", err)
	}

	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		toolCall := resp.Choices[0].Message.ToolCalls[0]
		t.Logf("OpenAI suggested tool: %s", toolCall.Function.Name)
		t.Logf("Arguments: %s", toolCall.Function.Arguments)

		// Execute the tool
		call := executor.ToolCall{
			Name:      toolCall.Function.Name,
			Arguments: json.RawMessage(toolCall.Function.Arguments),
		}

		result, err := exec.ExecuteTool(ctx, call)
		if err != nil {
			t.Fatalf("Tool execution failed: %v", err)
		}

		resultJSON, _ := json.Marshal(result)
		t.Logf("Execution result: %s", string(resultJSON))

		if !result.Success {
			t.Errorf("Expected successful tool execution")
		}
	} else {
		t.Error("Expected OpenAI to suggest a tool call")
	}

	t.Log("Manual steps test completed")
}

// TestOpenAIToolSchemaFormat verifies tool schemas are properly formatted for OpenAI
func TestOpenAIToolSchemaFormat(t *testing.T) {
	exec := executor.NewDefaultExecutor()
	if err := vscode.RegisterVSCodeLifecycle(exec); err != nil {
		t.Fatalf("Failed to register VSCode lifecycle: %v", err)
	}

	schemas := exec.ToolSchemas()
	t.Logf("Total tools registered: %d", len(schemas))

	for _, schema := range schemas {
		t.Logf("Tool: %s", schema.Name)
		t.Logf("  Description: %s", schema.Description)

		var params map[string]interface{}
		if err := json.Unmarshal(schema.Parameters, &params); err != nil {
			t.Errorf("Invalid JSON schema for %s: %v", schema.Name, err)
			continue
		}

		// Verify required fields
		if schema.Name == "" {
			t.Errorf("Tool name is empty")
		}
		if schema.Description == "" {
			t.Errorf("Tool %s has no description", schema.Name)
		}
		if params["type"] != "object" {
			t.Errorf("Tool %s parameters type should be 'object', got: %v", schema.Name, params["type"])
		}

		// Pretty print parameters
		prettyParams, _ := json.MarshalIndent(params, "  ", "  ")
		t.Logf("  Parameters:\n  %s", string(prettyParams))
	}
}

// Helper function to pretty print messages for debugging
func printMessages(t *testing.T, messages []openai.ChatCompletionMessage) {
	for i, msg := range messages {
		t.Logf("Message %d [%s]:", i+1, msg.Role)
		if msg.Content != "" {
			t.Logf("  Content: %s", msg.Content)
		}
		if len(msg.ToolCalls) > 0 {
			for j, tc := range msg.ToolCalls {
				t.Logf("  ToolCall %d: %s(%s)", j+1, tc.Function.Name, tc.Function.Arguments)
			}
		}
		if msg.ToolCallID != "" {
			t.Logf("  ToolCallID: %s", msg.ToolCallID)
		}
	}
}
