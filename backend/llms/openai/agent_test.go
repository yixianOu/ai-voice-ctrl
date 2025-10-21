package openai_test

import (
	"context"
	"os"
	"testing"

	"ai-voice-ctrl/backend/executor"
	"ai-voice-ctrl/backend/executor/vscode"
	"ai-voice-ctrl/backend/llms"
	openai_llms "ai-voice-ctrl/backend/llms/openai"
)

func TestOpenAIAgentBasic(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Initialize executor
	exec := executor.NewDefaultExecutor()
	if err := vscode.RegisterVSCodeLifecycle(exec); err != nil {
		t.Fatalf("Failed to register VSCode: %v", err)
	}

	// Create agent
	config := llms.DefaultAgentConfig()
	config.SystemPrompt = "You are a helpful assistant that can manage VSCode editor."

	agent := openai_llms.NewOpenAIAgentWithConfig(apiKey, exec, config)

	ctx := context.Background()

	// Test single command
	response, err := agent.Chat(ctx, `Create a VSCode workspace in "/tmp/agent_test"`)
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	t.Logf("Response: %s", response)
}

func TestOpenAIAgentMultiTurn(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	exec := executor.NewDefaultExecutor()
	if err := vscode.RegisterVSCodeLifecycle(exec); err != nil {
		t.Fatalf("Failed to register VSCode: %v", err)
	}

	agent := openai_llms.NewOpenAIAgent(apiKey, exec)
	ctx := context.Background()

	// Turn 1: Create workspace
	resp1, err := agent.Chat(ctx, `Create workspace "/tmp/multi_turn_test"`)
	if err != nil {
		t.Fatalf("Turn 1 failed: %v", err)
	}
	t.Logf("Turn 1: %s", resp1)

	// Turn 2: Write file (context preserved)
	resp2, err := agent.Chat(ctx, `Write a file called "hello.md" with content "# Hello World"`)
	if err != nil {
		t.Fatalf("Turn 2 failed: %v", err)
	}
	t.Logf("Turn 2: %s", resp2)

	// Turn 3: Append content
	resp3, err := agent.Chat(ctx, `Append "## New Section" to the file`)
	if err != nil {
		t.Fatalf("Turn 3 failed: %v", err)
	}
	t.Logf("Turn 3: %s", resp3)

	// Check history
	history := agent.GetHistory()
	t.Logf("Total messages in history: %d", len(history))
}

func TestOpenAIAgentReset(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	exec := executor.NewDefaultExecutor()
	if err := vscode.RegisterVSCodeLifecycle(exec); err != nil {
		t.Fatalf("Failed to register VSCode: %v", err)
	}

	agent := openai_llms.NewOpenAIAgent(apiKey, exec)
	ctx := context.Background()

	// Add some messages
	agent.Chat(ctx, "Create workspace /tmp/test")

	beforeReset := agent.GetHistory()
	t.Logf("Messages before reset: %d", len(beforeReset))

	// Reset
	agent.Reset()

	afterReset := agent.GetHistory()
	t.Logf("Messages after reset: %d", len(afterReset))

	if len(afterReset) >= len(beforeReset) {
		t.Error("Reset did not clear messages")
	}
}
