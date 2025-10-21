package vscode

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"ai-voice-ctrl/backend/executor"
)

// TestSessionLifecycle tests the full lifecycle of tool instance management.
func TestSessionLifecycle(t *testing.T) {
	exec := executor.NewDefaultExecutor()

	if err := RegisterVSCodeLifecycle(exec); err != nil {
		t.Fatalf("RegisterVSCodeLifecycle failed: %v", err)
	}

	ctx := context.Background()

	// Test 1: Call vscode_open_file without create_vscode should fail
	openCall := executor.ToolCall{
		Name:      "vscode_open_file",
		Arguments: json.RawMessage(`{"path": "artical.md"}`),
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
		Arguments: json.RawMessage(`{"workspace": "/tmp/test"}`),
	}
	result, err = exec.ExecuteTool(ctx, createCall)
	if err != nil {
		t.Fatalf("create_vscode failed: %v", err)
	}
	if !result.Success {
		t.Error("create_vscode should succeed")
	}

	go func() {
		// Phase 1: 前5秒执行多文件写入测试
		t.Log("Phase 1: Multiple files write test (5 seconds)")
		phase1Deadline := time.Now().Add(5 * time.Second)
		fileIndex := 0

		for time.Now().Before(phase1Deadline) {
			fileIndex++
			fileName := fmt.Sprintf("test_file_%d.md", fileIndex)

			// 写入不同文件
			writeCall := executor.ToolCall{
				Name: "vscode_write_file",
				Arguments: json.RawMessage(fmt.Sprintf(`{
			"path": "%s",
			"content": "# Test File %d\n\nCreated at: %s\n\n## Content\n\nThis is test file number %d.\n",
			"open_in_editor": true
		}`, fileName, fileIndex, time.Now().Format(time.RFC3339), fileIndex)),
			}
			result, err = exec.ExecuteTool(ctx, writeCall)
			if err != nil {
				t.Logf("Phase 1 - Write file %d failed: %v", fileIndex, err)
			} else {
				t.Logf("Phase 1 - Completed file %d: %s", fileIndex, fileName)
			}

			time.Sleep(300 * time.Millisecond)
		}

		// Phase 2: 后5秒执行单文件多次写入和修改
		t.Log("Phase 2: Single file multiple modifications (5 seconds)")
		phase2Deadline := time.Now().Add(5 * time.Second)
		singleFileName := "target_file.md"
		modificationIndex := 0

		// 初始创建文件
		writeCall := executor.ToolCall{
			Name: "vscode_write_file",
			Arguments: json.RawMessage(fmt.Sprintf(`{
			"path": "%s",
			"content": "# Target File\n\nInitial content created at: %s\n\n",
			"open_in_editor": true
		}`, singleFileName, time.Now().Format(time.RFC3339))),
		}
		result, err = exec.ExecuteTool(ctx, writeCall)
		if err != nil {
			t.Logf("Phase 2 - Initial write failed: %v", err)
		}

		for time.Now().Before(phase2Deadline) {
			modificationIndex++

			// 交替执行写入和追加
			if modificationIndex%2 == 0 {
				// 重写文件
				writeCall := executor.ToolCall{
					Name: "vscode_write_file",
					Arguments: json.RawMessage(fmt.Sprintf(`{
			"path": "%s",
			"content": "# Target File\n\nRewritten at: %s\n\n## Modification %d\n\nThis is rewrite number %d.\n",
			"open_in_editor": true
		}`, singleFileName, time.Now().Format(time.RFC3339), modificationIndex, modificationIndex)),
				}
				result, err = exec.ExecuteTool(ctx, writeCall)
				if err != nil {
					t.Logf("Phase 2 - Rewrite %d failed: %v", modificationIndex, err)
				} else {
					t.Logf("Phase 2 - Rewrite %d completed", modificationIndex)
				}
			} else {
				// 追加内容
				appendCall := executor.ToolCall{
					Name: "vscode_append_file",
					Arguments: json.RawMessage(fmt.Sprintf(`{
			"path": "%s",
			"content": "\n### Append %d\n\nAppended at: %s\n",
			"open_in_editor": false
		}`, singleFileName, modificationIndex, time.Now().Format(time.RFC3339))),
				}
				result, err = exec.ExecuteTool(ctx, appendCall)
				if err != nil {
					t.Logf("Phase 2 - Append %d failed: %v", modificationIndex, err)
				} else {
					t.Logf("Phase 2 - Append %d completed", modificationIndex)
				}
			}

			time.Sleep(300 * time.Millisecond)
		}

		t.Logf("Test completed: Phase 1 created %d files, Phase 2 made %d modifications", fileIndex, modificationIndex)
	}()

	<-time.After(10 * time.Second)

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
