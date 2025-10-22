// Package vscode tool
package vscode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	execCmd "os/exec"
	"path/filepath"
	"strings"
	"time"

	"ai-voice-ctrl/backend/executor"
)

const Vscode = "vscode"

// VSCodeTool 封装 VS Code 的 HTTP Bridge 操作。
type VSCodeTool struct {
	workspace  string
	codePath   string
	httpBridge *VSCodeHTTPTool
}

// NewVSCodeTool 创建一个 VS Code 工具实例。
func NewVSCodeTool(workspace string) (*VSCodeTool, error) {
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return nil, errors.New("workspace path is required")
	}

	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace path: %w", err)
	}

	codePath, err := execCmd.LookPath("code")
	if err != nil {
		return nil, fmt.Errorf("VS Code CLI not found: %w", err)
	}

	httpBridge, err := NewVSCodeHTTPTool(absWorkspace, 9527)
	if err != nil {
		return nil, fmt.Errorf("create HTTP bridge: %w", err)
	}

	return &VSCodeTool{
		workspace:  absWorkspace,
		codePath:   codePath,
		httpBridge: httpBridge,
	}, nil
}

func (t *VSCodeTool) openFile(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
	var args struct {
		Path string `json:"path"`
		Line *int   `json:"line"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return executor.ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode open file payload: %w", err)
	}

	resolvedPath, err := t.resolvePath(args.Path)
	if err != nil {
		return executor.ToolResult{Success: false, Message: err.Error()}, err
	}

	// Check if file exists, create if not
	fileCreated := false
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
			return executor.ToolResult{Success: false, Message: "failed to create parent directory"}, fmt.Errorf("create parent directories: %w", err)
		}

		// Create empty file with proper newline
		if err := os.WriteFile(resolvedPath, []byte("\n"), 0o644); err != nil {
			return executor.ToolResult{Success: false, Message: "failed to create file"}, fmt.Errorf("create file: %w", err)
		}
		fileCreated = true
	}

	// Open file in VSCode
	result, err := t.httpBridge.openFile(ctx, payload)
	if err != nil {
		return result, err
	}

	// Add file creation info to result
	if fileCreated {
		if result.Data == nil {
			result.Data = make(map[string]interface{})
		}
		result.Data["file_created"] = true
		result.Data["path"] = resolvedPath
		result.Message = fmt.Sprintf("file created and opened: %s", args.Path)
	}

	return result, nil
}

func (t *VSCodeTool) writeFile(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
	var args struct {
		Path         string `json:"path"`
		Content      string `json:"content"`
		OpenInEditor bool   `json:"open_in_editor"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return executor.ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode write file payload: %w", err)
	}

	resolvedPath, err := t.resolvePath(args.Path)
	if err != nil {
		return executor.ToolResult{Success: false, Message: err.Error()}, err
	}

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return executor.ToolResult{Success: false, Message: "failed to create parent directory"}, fmt.Errorf("create parent directories: %w", err)
	}

	// Write file content
	fileExisted := true
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		fileExisted = false
	}

	if err := os.WriteFile(resolvedPath, []byte(args.Content), 0o644); err != nil {
		return executor.ToolResult{Success: false, Message: "failed to write file"}, fmt.Errorf("write file: %w", err)
	}

	resultData := map[string]interface{}{
		"path":          resolvedPath,
		"file_created":  !fileExisted,
		"bytes_written": len(args.Content),
	}

	// Open in editor if requested
	if args.OpenInEditor {
		openPayload, _ := json.Marshal(map[string]interface{}{"path": args.Path})
		result, err := t.httpBridge.openFile(ctx, openPayload)
		if err != nil {
			return executor.ToolResult{Success: false, Message: err.Error()}, err
		}
		// Merge data
		for k, v := range resultData {
			if result.Data == nil {
				result.Data = make(map[string]interface{})
			}
			result.Data[k] = v
		}
		return result, nil
	}

	message := "file written"
	if !fileExisted {
		message = "file created and written"
	}

	return executor.ToolResult{
		Success: true,
		Message: message,
		Data:    resultData,
	}, nil
}

func (t *VSCodeTool) appendFile(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
	var args struct {
		Path         string `json:"path"`
		Content      string `json:"content"`
		OpenInEditor bool   `json:"open_in_editor"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return executor.ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode append file payload: %w", err)
	}

	resolvedPath, err := t.resolvePath(args.Path)
	if err != nil {
		return executor.ToolResult{Success: false, Message: err.Error()}, err
	}

	// Create parent directories if needed
	if err = os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return executor.ToolResult{Success: false, Message: "failed to create parent directory"}, fmt.Errorf("create parent directories: %w", err)
	}

	// Check if file exists
	fileExisted := true
	if _, err := os.Stat(resolvedPath); os.IsNotExist(err) {
		fileExisted = false
	}

	// Open file for appending (creates if not exists)
	file, err := os.OpenFile(resolvedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return executor.ToolResult{Success: false, Message: "failed to open file"}, fmt.Errorf("open file for append: %w", err)
	}
	defer file.Close()

	// Append content
	bytesWritten, err := file.WriteString(args.Content)
	if err != nil {
		return executor.ToolResult{Success: false, Message: "failed to append text"}, fmt.Errorf("append content: %w", err)
	}

	resultData := map[string]interface{}{
		"path":          resolvedPath,
		"file_created":  !fileExisted,
		"bytes_written": bytesWritten,
	}

	// Open in editor if requested
	if args.OpenInEditor {
		openPayload, _ := json.Marshal(map[string]interface{}{"path": args.Path})
		result, err := t.httpBridge.openFile(ctx, openPayload)
		if err != nil {
			return executor.ToolResult{Success: false, Message: err.Error()}, err
		}
		// Merge data
		for k, v := range resultData {
			if result.Data == nil {
				result.Data = make(map[string]interface{})
			}
			result.Data[k] = v
		}
		return result, nil
	}

	message := "content appended"
	if !fileExisted {
		message = "file created and content appended"
	}

	return executor.ToolResult{
		Success: true,
		Message: message,
		Data:    resultData,
	}, nil
}

func (t *VSCodeTool) executeCommand(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
	var args struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return executor.ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode command payload: %w", err)
	}

	if strings.TrimSpace(args.Command) == "" {
		return executor.ToolResult{Success: false, Message: "command is required"}, errors.New("missing command")
	}

	ctx = ensureContext(ctx)
	commandArgs := append([]string{t.workspace, "--command", args.Command}, args.Args...)
	stdout, err := t.runVSCodeCommand(ctx, commandArgs...)
	if err != nil {
		return executor.ToolResult{Success: false, Message: stderrOrFallback(stdout, err)}, fmt.Errorf("execute VS Code command: %w", err)
	}

	response := executor.ToolResult{Success: true, Message: "command executed"}
	if stdout != "" {
		response.Data = map[string]interface{}{"output": stdout}
	}
	return response, nil
}

func (t *VSCodeTool) resolvePath(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.New("path is required")
	}

	var joined string
	if filepath.IsAbs(input) {
		joined = input
	} else {
		joined = filepath.Join(t.workspace, input)
	}

	absTarget, err := filepath.Abs(joined)
	if err != nil {
		return "", fmt.Errorf("resolve target path: %w", err)
	}

	workspacePrefix := t.workspace
	if workspacePrefix != "" && workspacePrefix[len(workspacePrefix)-1] != os.PathSeparator {
		workspacePrefix += string(os.PathSeparator)
	}

	if workspacePrefix != "" && absTarget != t.workspace && !strings.HasPrefix(absTarget, workspacePrefix) {
		return "", fmt.Errorf("path %s is outside the workspace", input)
	}

	return absTarget, nil
}

func (t *VSCodeTool) runVSCodeCommand(ctx context.Context, args ...string) (string, error) {
	cmd := execCmd.CommandContext(ctx, t.codePath, args...)
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func ensureContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}

func stderrOrFallback(output string, err error) string {
	message := strings.TrimSpace(output)
	if message != "" {
		return message
	}
	if err != nil {
		return err.Error()
	}
	return "operation failed"
}

// RegisterLifecycle registers VSCode tool with create/destroy/operation functions.
func (t *VSCodeTool) RegisterLifecycle(exec executor.Executor) error {
	definitions := map[string]executor.ToolDefinition{
		"create_vscode": {
			Name:        "create_vscode",
			Description: "Create VSCode instance with workspace",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"workspace": {
						"type": "string",
						"description": "Workspace directory path"
					}
				},
				"required": ["workspace"]
			}`),
			Executor: t.createVSCodeExecutor(exec),
		},
		"vscode_open_file": {
			Name:        "vscode_open_file",
			Description: "Open a file in VSCode (requires create_vscode first)",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "File path relative to workspace"
					},
					"line": {
						"type": "integer",
						"minimum": 1,
						"description": "Line number to jump to"
					}
				},
				"required": ["path"]
			}`),
			Executor: t.vsCodeOpenFileExecutor(exec),
		},
		"vscode_write_file": {
			Name:        "vscode_write_file",
			Description: "Write content to a file (requires create_vscode first)",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "File path relative to workspace"
					},
					"content": {
						"type": "string",
						"description": "File content"
					},
					"open_in_editor": {
						"type": "boolean",
						"description": "Open file after writing"
					}
				},
				"required": ["path", "content"]
			}`),
			Executor: t.vsCodeWriteFileExecutor(exec),
		},
		"vscode_append_file": {
			Name:        "vscode_append_file",
			Description: "Append content to a file (requires create_vscode first)",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "File path relative to workspace"
					},
					"content": {
						"type": "string",
						"description": "Content to append"
					},
					"open_in_editor": {
						"type": "boolean",
						"description": "Open file after appending"
					}
				},
				"required": ["path", "content"]
			}`),
			Executor: t.vsCodeAppendFileExecutor(exec),
		},
		"vscode_execute_command": {
			Name:        "vscode_execute_command",
			Description: "Execute a VSCode command (requires create_vscode first)",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"command": {
						"type": "string",
						"description": "Command identifier"
					},
					"args": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Command arguments"
					}
				},
				"required": ["command"]
			}`),
			Executor: t.vsCodeExecuteCommandExecutor(exec),
		},
		"destroy_vscode": {
			Name:        "destroy_vscode",
			Description: "Destroy VSCode instance and release resources",
			Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
			Executor:    t.destroyVSCodeExecutor(exec),
		},
	}

	return exec.RegisterTool("vscode", definitions)
}

func (t *VSCodeTool) createVSCodeExecutor(exec executor.Executor) executor.ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
		var args struct {
			Workspace string `json:"workspace"`
		}
		if err := json.Unmarshal(payload, &args); err != nil {
			return executor.ToolResult{Success: false, Message: "invalid payload"}, err
		}

		tool, err := NewVSCodeTool(args.Workspace)
		if err != nil {
			return executor.ToolResult{Success: false, Message: err.Error()}, err
		}

		ctx = ensureContext(ctx)
		vsCmd := execCmd.CommandContext(ctx, tool.codePath, "--new-window", tool.workspace)
		vsCmd.Env = os.Environ()
		if err := vsCmd.Start(); err != nil {
			return executor.ToolResult{Success: false, Message: "failed to open workspace"}, fmt.Errorf("open workspace: %w", err)
		}

		time.Sleep(10 * time.Second)

		exec.StoreInstance(Vscode, tool)
		return executor.ToolResult{
			Success: true,
			Message: "VSCode instance created and workspace opened",
			Data:    map[string]interface{}{"workspace": args.Workspace, "pid": vsCmd.Process.Pid},
		}, nil
	}
}

func (t *VSCodeTool) destroyVSCodeExecutor(exec executor.Executor) executor.ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
		tool, err := t.getVSCodeInstance(exec)
		if err != nil {
			return executor.ToolResult{Success: false, Message: err.Error()}, err
		}

		result, _ := tool.httpBridge.closeWindow(ctx, json.RawMessage(`{}`))
		exec.DeleteInstance("vscode")

		return executor.ToolResult{
			Success: true,
			Message: "VSCode instance destroyed and window closed",
			Data:    result.Data,
		}, nil
	}
}

func (t *VSCodeTool) getVSCodeInstance(exec executor.Executor) (*VSCodeTool, error) {
	instance, ok := exec.LoadInstance("vscode")
	if !ok {
		return nil, fmt.Errorf("VSCode not created. Please call create_vscode first")
	}
	tool, ok := instance.(*VSCodeTool)
	if !ok {
		return nil, fmt.Errorf("invalid VSCode instance type")
	}
	return tool, nil
}

func (t *VSCodeTool) vsCodeOpenFileExecutor(exec executor.Executor) executor.ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
		tool, err := t.getVSCodeInstance(exec)
		if err != nil {
			return executor.ToolResult{Success: false, Message: err.Error()}, err
		}
		return tool.openFile(ctx, payload)
	}
}

func (t *VSCodeTool) vsCodeWriteFileExecutor(exec executor.Executor) executor.ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
		tool, err := t.getVSCodeInstance(exec)
		if err != nil {
			return executor.ToolResult{Success: false, Message: err.Error()}, err
		}
		return tool.writeFile(ctx, payload)
	}
}

func (t *VSCodeTool) vsCodeAppendFileExecutor(exec executor.Executor) executor.ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
		tool, err := t.getVSCodeInstance(exec)
		if err != nil {
			return executor.ToolResult{Success: false, Message: err.Error()}, err
		}
		return tool.appendFile(ctx, payload)
	}
}

func (t *VSCodeTool) vsCodeExecuteCommandExecutor(exec executor.Executor) executor.ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
		tool, err := t.getVSCodeInstance(exec)
		if err != nil {
			return executor.ToolResult{Success: false, Message: err.Error()}, err
		}
		return tool.executeCommand(ctx, payload)
	}
}

// RegisterVSCodeLifecycle registers VSCode tool with LLM-driven lifecycle management.
// This is a convenience wrapper that creates a temporary VSCodeTool instance for registration.
func RegisterVSCodeLifecycle(exec executor.Executor) error {
	// Use a dummy workspace for registration - actual workspace will be set by create_vscode
	tool := &VSCodeTool{}
	return tool.RegisterLifecycle(exec)
}
