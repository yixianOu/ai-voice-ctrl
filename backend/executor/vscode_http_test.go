package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVSCodeHTTPToolIntegration(t *testing.T) {
	// Create temporary workspace
	tmpDir, err := os.MkdirTemp("", "vscode-http-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create VSCode HTTP tool
	tool, err := NewVSCodeHTTPTool(tmpDir, 9527)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Wait for VSCode extension to be ready
	t.Log("Waiting for VSCode extension to start (ensure extension is running)...")
	time.Sleep(2 * time.Second)

	// Test: Open file
	t.Run("OpenFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("Hello VSCode"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"path": "test.txt",
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result, err := tool.openFile(ctx, payload)
		if err != nil {
			t.Fatalf("open file failed: %v", err)
		}

		if !result.Success {
			t.Errorf("expected success=true, got %v", result.Success)
		}

		t.Logf("Open file result: %+v", result)
	})

	// Test: Refresh file
	t.Run("RefreshFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "refresh_test.txt")
		if err := os.WriteFile(testFile, []byte("Original content"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Open file first
		openPayload, _ := json.Marshal(map[string]interface{}{
			"path": "refresh_test.txt",
		})

		ctx := context.Background()
		if _, err := tool.openFile(ctx, openPayload); err != nil {
			t.Fatalf("failed to open file: %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		// Modify file content
		if err := os.WriteFile(testFile, []byte("Updated content"), 0o644); err != nil {
			t.Fatalf("failed to update file: %v", err)
		}

		// Refresh file
		refreshPayload, _ := json.Marshal(map[string]interface{}{
			"path": "refresh_test.txt",
		})

		result, err := tool.refreshFile(ctx, refreshPayload)
		if err != nil {
			t.Fatalf("refresh file failed: %v", err)
		}

		if !result.Success {
			t.Errorf("expected success=true, got %v", result.Success)
		}

		t.Logf("Refresh file result: %+v", result)
	})

	// Test: Close file
	t.Run("CloseFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "close_test.txt")
		if err := os.WriteFile(testFile, []byte("Close me"), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Open file first
		openPayload, _ := json.Marshal(map[string]interface{}{
			"path": "close_test.txt",
		})

		ctx := context.Background()
		if _, err := tool.openFile(ctx, openPayload); err != nil {
			t.Fatalf("failed to open file: %v", err)
		}

		time.Sleep(500 * time.Millisecond)

		// Close file
		closePayload, _ := json.Marshal(map[string]interface{}{
			"path": "close_test.txt",
		})

		result, err := tool.closeFile(ctx, closePayload)
		if err != nil {
			t.Fatalf("close file failed: %v", err)
		}

		if !result.Success {
			t.Errorf("expected success=true, got %v", result.Success)
		}

		t.Logf("Close file result: %+v", result)
	})

	// Test: Open file with line number
	t.Run("OpenFileWithLine", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "multiline.txt")
		content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"
		if err := os.WriteFile(testFile, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		line := 3
		payload, _ := json.Marshal(map[string]interface{}{
			"path": "multiline.txt",
			"line": line,
		})

		ctx := context.Background()
		result, err := tool.openFile(ctx, payload)
		if err != nil {
			t.Fatalf("open file with line failed: %v", err)
		}

		if !result.Success {
			t.Errorf("expected success=true, got %v", result.Success)
		}

		t.Logf("Open file with line result: %+v", result)
	})

	// Test: Path resolution
	t.Run("PathResolution", func(t *testing.T) {
		tests := []struct {
			name      string
			input     string
			wantError bool
		}{
			{"relative path", "subdir/file.txt", false},
			{"absolute path", filepath.Join(tmpDir, "abs.txt"), false},
			{"empty path", "", true},
			{"current dir", ".", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resolved, err := tool.resolvePath(tt.input)
				if tt.wantError {
					if err == nil {
						t.Errorf("expected error, got nil")
					}
				} else {
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					if resolved == "" {
						t.Errorf("expected non-empty resolved path")
					}
					t.Logf("Resolved path: %s -> %s", tt.input, resolved)
				}
			})
		}
	})
}

func TestVSCodeHTTPToolRegistration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vscode-http-reg-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool, err := NewVSCodeHTTPTool(tmpDir, 9527)
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	exec := NewDefaultExecutor()

	// Test: Register lifecycle
	if err := tool.RegisterHTTPLifecycle(exec); err != nil {
		t.Fatalf("failed to register lifecycle: %v", err)
	}

	// Verify registered functions
	expectedFunctions := []string{
		"vscode_open_file",
		"vscode_close_file",
		"vscode_refresh_file",
		"vscode_close_window",
	}

	schemas := exec.ToolSchemas()
	registeredNames := make(map[string]bool)
	for _, schema := range schemas {
		registeredNames[schema.Name] = true
	}

	for _, funcName := range expectedFunctions {
		if !registeredNames[funcName] {
			t.Errorf("expected function %s not registered", funcName)
		}
	}

	t.Log("All expected functions registered successfully")
}

func TestVSCodeHTTPToolErrorHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vscode-http-err-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool, err := NewVSCodeHTTPTool(tmpDir, 9999) // Wrong port
	if err != nil {
		t.Fatalf("failed to create tool: %v", err)
	}

	// Test: Connection refused
	t.Run("ConnectionRefused", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]interface{}{
			"path": "test.txt",
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		_, err := tool.openFile(ctx, payload)
		if err == nil {
			t.Error("expected error due to wrong port, got nil")
		}
		t.Logf("Expected error: %v", err)
	})

	// Test: Invalid payload
	t.Run("InvalidPayload", func(t *testing.T) {
		invalidPayload := json.RawMessage(`{"invalid": "data"}`)

		ctx := context.Background()
		_, err := tool.openFile(ctx, invalidPayload)
		if err == nil {
			t.Error("expected error due to invalid payload, got nil")
		}
		t.Logf("Expected error: %v", err)
	})

	// Test: Empty path
	t.Run("EmptyPath", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]interface{}{
			"path": "",
		})

		ctx := context.Background()
		_, err := tool.openFile(ctx, payload)
		if err == nil {
			t.Error("expected error due to empty path, got nil")
		}
		t.Logf("Expected error: %v", err)
	})
}

func TestNewVSCodeHTTPTool(t *testing.T) {
	tests := []struct {
		name      string
		workspace string
		port      int
		wantError bool
	}{
		{"valid workspace", "/tmp/test", 9527, false},
		{"empty workspace", "", 9527, true},
		{"relative workspace", "./test", 9527, false},
		{"different port", "/tmp/test", 8080, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := NewVSCodeHTTPTool(tt.workspace, tt.port)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tool == nil {
					t.Error("expected non-nil tool")
				}
				if tool != nil && tool.bridgeURL != fmt.Sprintf("http://localhost:%d", tt.port) {
					t.Errorf("expected bridgeURL http://localhost:%d, got %s", tt.port, tool.bridgeURL)
				}
			}
		})
	}
}
