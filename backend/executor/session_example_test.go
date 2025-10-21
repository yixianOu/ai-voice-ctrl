package executor_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"ai-voice-ctrl/backend/executor"
)

// Example demonstrates LLM-driven tool lifecycle management.
func ExampleRegisterVSCodeLifecycle() {
	// 1. Create executor
	exec := executor.NewDefaultExecutor()

	// 2. Register VSCode with lifecycle functions
	if err := executor.RegisterVSCodeLifecycle(exec); err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return
	}

	// 3. Simulate LLM calling create_vscode
	createCall := executor.ToolCall{
		Name:      "create_vscode",
		Arguments: json.RawMessage(`{"workspace": "/home/orician/workspace/doc"}`),
	}
	result, err := exec.ExecuteTool(context.Background(), createCall)
	if err != nil {
		fmt.Printf("Create failed: %v\n", err)
		return
	}
	fmt.Printf("Create result: %s\n", result.Message)

	// 4. Simulate LLM calling vscode_open_file
	openCall := executor.ToolCall{
		Name:      "vscode_open_file",
		Arguments: json.RawMessage(`{"path": "main.go", "line": 10}`),
	}
	result, err = exec.ExecuteTool(context.Background(), openCall)
	if err != nil {
		fmt.Printf("Open failed: %v\n", err)
		return
	}
	fmt.Printf("Open result: %s\n", result.Message)

	// 5. Simulate LLM calling destroy_vscode
	destroyCall := executor.ToolCall{
		Name:      "destroy_vscode",
		Arguments: json.RawMessage(`{}`),
	}
	result, err = exec.ExecuteTool(context.Background(), destroyCall)
	if err != nil {
		fmt.Printf("Destroy failed: %v\n", err)
		return
	}
	fmt.Printf("Destroy result: %s\n", result.Message)
}

// TestSessionLifecycle tests the full lifecycle of tool instance management.
func TestSessionLifecycle(t *testing.T) {
	exec := executor.NewDefaultExecutor()

	if err := executor.RegisterVSCodeLifecycle(exec); err != nil {
		t.Fatalf("RegisterVSCodeLifecycle failed: %v", err)
	}

	ctx := context.Background()

	// Test 1: Call vscode_open_file without create_vscode should fail
	openCall := executor.ToolCall{
		Name:      "vscode_open_file",
		Arguments: json.RawMessage(`{"path": "test.go"}`),
	}
	result, err := exec.ExecuteTool(ctx, openCall)
	if err == nil {
		t.Error("Expected error when calling vscode_open_file without create_vscode")
	}
	if result.Success {
		t.Error("Result should indicate failure")
	}

	// Test 2: Create VSCode instance with current directory
	createCall := executor.ToolCall{
		Name:      "create_vscode",
		Arguments: json.RawMessage(`{"workspace": "."}`),
	}
	result, err = exec.ExecuteTool(ctx, createCall)
	if err != nil {
		t.Fatalf("create_vscode failed: %v", err)
	}
	if !result.Success {
		t.Error("create_vscode should succeed")
	}

	// Test 3: Now vscode_open_file should work (but may fail if code CLI not available)
	result, err = exec.ExecuteTool(ctx, openCall)
	// Skip verification if VSCode CLI is not available
	if err != nil && result.Message != "tool not found" {
		t.Logf("vscode_open_file result: %v (may fail without VSCode CLI)", err)
	}

	// Test 4: Destroy VSCode instance
	destroyCall := executor.ToolCall{
		Name:      "destroy_vscode",
		Arguments: json.RawMessage(`{}`),
	}
	result, err = exec.ExecuteTool(ctx, destroyCall)
	if err != nil {
		t.Fatalf("destroy_vscode failed: %v", err)
	}
	if !result.Success {
		t.Error("destroy_vscode should succeed")
	}

	// Test 5: After destroy, vscode_open_file should fail again
	result, err = exec.ExecuteTool(ctx, openCall)
	if err == nil {
		t.Error("Expected error when calling vscode_open_file after destroy")
	}
}
