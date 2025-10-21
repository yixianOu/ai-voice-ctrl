package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VSCodeTool 封装 VS Code 的 CLI 操作。
type VSCodeTool struct {
	workspace string
	codePath  string
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

	codePath, err := exec.LookPath("code")
	if err != nil {
		return nil, fmt.Errorf("VS Code CLI not found: %w", err)
	}

	return &VSCodeTool{
		workspace: absWorkspace,
		codePath:  codePath,
	}, nil
}

func (t *VSCodeTool) openFile(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	var args struct {
		Path string `json:"path"`
		Line *int   `json:"line"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode open file payload: %w", err)
	}

	resolvedPath, err := t.resolvePath(args.Path)
	if err != nil {
		return ToolResult{Success: false, Message: err.Error()}, err
	}

	ctx = ensureContext(ctx)

	var commandArgs []string
	if args.Line != nil {
		if *args.Line < 1 {
			return ToolResult{Success: false, Message: "line must be at least 1"}, errors.New("line out of range")
		}
		commandArgs = []string{"-g", fmt.Sprintf("%s:%d", resolvedPath, *args.Line)}
	} else {
		commandArgs = []string{resolvedPath}
	}

	stdout, err := t.runVSCodeCommand(ctx, commandArgs...)
	if err != nil {
		return ToolResult{Success: false, Message: stderrOrFallback(stdout, err)}, fmt.Errorf("open file with VS Code: %w", err)
	}

	return ToolResult{Success: true, Message: "file opened", Data: map[string]interface{}{"path": resolvedPath}}, nil
}

func (t *VSCodeTool) writeFile(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	var args struct {
		Path         string `json:"path"`
		Content      string `json:"content"`
		OpenInEditor bool   `json:"open_in_editor"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode write file payload: %w", err)
	}

	resolvedPath, err := t.resolvePath(args.Path)
	if err != nil {
		return ToolResult{Success: false, Message: err.Error()}, err
	}

	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return ToolResult{Success: false, Message: "failed to create parent directory"}, fmt.Errorf("create parent directories: %w", err)
	}

	if err := os.WriteFile(resolvedPath, []byte(args.Content), 0o644); err != nil {
		return ToolResult{Success: false, Message: "failed to write file"}, fmt.Errorf("write file: %w", err)
	}

	if args.OpenInEditor {
		ctx = ensureContext(ctx)
		if output, runErr := t.runVSCodeCommand(ctx, resolvedPath); runErr != nil {
			return ToolResult{Success: false, Message: stderrOrFallback(output, runErr)}, fmt.Errorf("open written file: %w", runErr)
		}
	}

	return ToolResult{Success: true, Message: "file written", Data: map[string]interface{}{"path": resolvedPath}}, nil
}

func (t *VSCodeTool) appendFile(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	var args struct {
		Path         string `json:"path"`
		Content      string `json:"content"`
		OpenInEditor bool   `json:"open_in_editor"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode append file payload: %w", err)
	}

	resolvedPath, err := t.resolvePath(args.Path)
	if err != nil {
		return ToolResult{Success: false, Message: err.Error()}, err
	}

	if err := os.MkdirAll(filepath.Dir(resolvedPath), 0o755); err != nil {
		return ToolResult{Success: false, Message: "failed to create parent directory"}, fmt.Errorf("create parent directories: %w", err)
	}

	file, err := os.OpenFile(resolvedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return ToolResult{Success: false, Message: "failed to open file"}, fmt.Errorf("open file for append: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(args.Content); err != nil {
		return ToolResult{Success: false, Message: "failed to append text"}, fmt.Errorf("append content: %w", err)
	}

	if args.OpenInEditor {
		ctx = ensureContext(ctx)
		if output, runErr := t.runVSCodeCommand(ctx, resolvedPath); runErr != nil {
			return ToolResult{Success: false, Message: stderrOrFallback(output, runErr)}, fmt.Errorf("open appended file: %w", runErr)
		}
	}

	return ToolResult{Success: true, Message: "content appended", Data: map[string]interface{}{"path": resolvedPath}}, nil
}

func (t *VSCodeTool) executeCommand(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	var args struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode command payload: %w", err)
	}

	if strings.TrimSpace(args.Command) == "" {
		return ToolResult{Success: false, Message: "command is required"}, errors.New("missing command")
	}

	ctx = ensureContext(ctx)
	commandArgs := append([]string{"--command", args.Command}, args.Args...)
	stdout, err := t.runVSCodeCommand(ctx, commandArgs...)
	if err != nil {
		return ToolResult{Success: false, Message: stderrOrFallback(stdout, err)}, fmt.Errorf("execute VS Code command: %w", err)
	}

	response := ToolResult{Success: true, Message: "command executed"}
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
	cmd := exec.CommandContext(ctx, t.codePath, args...)
	cmd.Dir = t.workspace
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
func (t *VSCodeTool) RegisterLifecycle(exec Executor) error {
	definitions := map[string]ToolDefinition{
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

func (t *VSCodeTool) createVSCodeExecutor(exec Executor) ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
		var args struct {
			Workspace string `json:"workspace"`
		}
		if err := json.Unmarshal(payload, &args); err != nil {
			return ToolResult{Success: false, Message: "invalid payload"}, err
		}

		tool, err := NewVSCodeTool(args.Workspace)
		if err != nil {
			return ToolResult{Success: false, Message: err.Error()}, err
		}

		exec.StoreInstance("vscode", tool)
		return ToolResult{
			Success: true,
			Message: "VSCode instance created",
			Data:    map[string]interface{}{"workspace": args.Workspace},
		}, nil
	}
}

func (t *VSCodeTool) destroyVSCodeExecutor(exec Executor) ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
		exec.DeleteInstance("vscode")
		return ToolResult{Success: true, Message: "VSCode instance destroyed"}, nil
	}
}

func (t *VSCodeTool) getVSCodeInstance(exec Executor) (*VSCodeTool, error) {
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

func (t *VSCodeTool) vsCodeOpenFileExecutor(exec Executor) ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
		tool, err := t.getVSCodeInstance(exec)
		if err != nil {
			return ToolResult{Success: false, Message: err.Error()}, err
		}
		return tool.openFile(ctx, payload)
	}
}

func (t *VSCodeTool) vsCodeWriteFileExecutor(exec Executor) ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
		tool, err := t.getVSCodeInstance(exec)
		if err != nil {
			return ToolResult{Success: false, Message: err.Error()}, err
		}
		return tool.writeFile(ctx, payload)
	}
}

func (t *VSCodeTool) vsCodeAppendFileExecutor(exec Executor) ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
		tool, err := t.getVSCodeInstance(exec)
		if err != nil {
			return ToolResult{Success: false, Message: err.Error()}, err
		}
		return tool.appendFile(ctx, payload)
	}
}

func (t *VSCodeTool) vsCodeExecuteCommandExecutor(exec Executor) ToolExecutor {
	return func(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
		tool, err := t.getVSCodeInstance(exec)
		if err != nil {
			return ToolResult{Success: false, Message: err.Error()}, err
		}
		return tool.executeCommand(ctx, payload)
	}
}

// RegisterVSCodeLifecycle registers VSCode tool with LLM-driven lifecycle management.
// This is a convenience wrapper that creates a temporary VSCodeTool instance for registration.
func RegisterVSCodeLifecycle(exec Executor) error {
	// Use a dummy workspace for registration - actual workspace will be set by create_vscode
	tool := &VSCodeTool{}
	return tool.RegisterLifecycle(exec)
}
