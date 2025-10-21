package handler_test

import (
	"context"
	"os"
	"testing"

	"ai-voice-ctrl/backend/executor"
	"ai-voice-ctrl/backend/executor/vscode"
	"ai-voice-ctrl/backend/handler"
	openai_llms "ai-voice-ctrl/backend/llms/openai"
)

func TestCommandHandlerTextCommand(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Setup
	exec := executor.NewDefaultExecutor()
	if err := vscode.RegisterVSCodeLifecycle(exec); err != nil {
		t.Fatalf("Failed to register VSCode: %v", err)
	}

	agent := openai_llms.NewOpenAIAgent(apiKey, exec)
	audioHandler := handler.NewAudioHandler(apiKey)
	cmdHandler := handler.NewCommandHandler(audioHandler, agent)

	ctx := context.Background()

	// Test text command
	response, err := cmdHandler.ProcessTextCommand(ctx, `Create a VSCode workspace in "/tmp/cmd_test"`)
	if err != nil {
		t.Fatalf("ProcessTextCommand failed: %v", err)
	}

	t.Logf("Response: %s", response)

	// Test multi-turn
	resp2, err := cmdHandler.ProcessTextCommand(ctx, `Write a file "test.txt" with "Hello"`)
	if err != nil {
		t.Fatalf("Second command failed: %v", err)
	}
	t.Logf("Response 2: %s", resp2)

	// Check history
	history := cmdHandler.GetConversationHistory()
	t.Logf("Conversation history has %d messages", len(history))
}

func TestCommandHandlerReset(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	exec := executor.NewDefaultExecutor()
	if err := vscode.RegisterVSCodeLifecycle(exec); err != nil {
		t.Fatalf("Failed to register VSCode: %v", err)
	}

	agent := openai_llms.NewOpenAIAgent(apiKey, exec)
	audioHandler := handler.NewAudioHandler(apiKey)
	cmdHandler := handler.NewCommandHandler(audioHandler, agent)

	ctx := context.Background()

	// Add conversation
	cmdHandler.ProcessTextCommand(ctx, "Create workspace /tmp/test")

	beforeReset := len(cmdHandler.GetConversationHistory())
	t.Logf("Messages before reset: %d", beforeReset)

	// Reset
	cmdHandler.ResetConversation()

	afterReset := len(cmdHandler.GetConversationHistory())
	t.Logf("Messages after reset: %d", afterReset)

	if afterReset >= beforeReset {
		t.Error("Reset did not reduce message count")
	}
}
