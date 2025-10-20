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

// VSCodeToolRegistry exposes VS Code operations via the ToolRegistry pattern.
type VSCodeToolRegistry struct {
	workspace string
	codePath  string
}

// NewVSCodeToolRegistry verifies the VS Code CLI and normalizes the workspace path.
func NewVSCodeToolRegistry(workspace string) (*VSCodeToolRegistry, error) {
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

	return &VSCodeToolRegistry{
		workspace: absWorkspace,
		codePath:  codePath,
	}, nil
}

// Tools returns the VS Code function-calling definitions.
func (r *VSCodeToolRegistry) Tools() map[string]ToolDefinition {
	return map[string]ToolDefinition{
		"vscode_open_file": {
			Name:        "vscode_open_file",
			Description: "Open a file in VS Code, optionally jumping to a specific line",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "File path relative to the workspace root"
					},
					"line": {
						"type": "integer",
						"minimum": 1,
						"description": "1-based line number to focus"
					}
				},
				"required": ["path"]
			}`),
			Executor: r.openFile,
		},
		"vscode_write_file": {
			Name:        "vscode_write_file",
			Description: "Create or overwrite a file with provided content, optionally open it",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "File path relative to the workspace root"
					},
					"content": {
						"type": "string",
						"description": "Full file content"
					},
					"open_in_editor": {
						"type": "boolean",
						"description": "Open the file in VS Code after writing"
					}
				},
				"required": ["path", "content"]
			}`),
			Executor: r.writeFile,
		},
		"vscode_append_file": {
			Name:        "vscode_append_file",
			Description: "Append text to an existing file, creating it if necessary",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {
						"type": "string",
						"description": "File path relative to the workspace root"
					},
					"content": {
						"type": "string",
						"description": "Text to append"
					},
					"open_in_editor": {
						"type": "boolean",
						"description": "Open the file in VS Code after appending"
					}
				},
				"required": ["path", "content"]
			}`),
			Executor: r.appendFile,
		},
		"vscode_execute_command": {
			Name:        "vscode_execute_command",
			Description: "Run a VS Code command through the CLI",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"command": {
						"type": "string",
						"description": "Command identifier, e.g. workbench.action.files.save"
					},
					"args": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Optional command arguments"
					}
				},
				"required": ["command"]
			}`),
			Executor: r.executeCommand,
		},
	}
}

var _ ToolRegistry = (*VSCodeToolRegistry)(nil)

func (r *VSCodeToolRegistry) openFile(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	var args struct {
		Path string `json:"path"`
		Line *int   `json:"line"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode open file payload: %w", err)
	}

	resolvedPath, err := r.resolvePath(args.Path)
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

	stdout, err := r.runVSCodeCommand(ctx, commandArgs...)
	if err != nil {
		return ToolResult{Success: false, Message: stderrOrFallback(stdout, err)}, fmt.Errorf("open file with VS Code: %w", err)
	}

	return ToolResult{Success: true, Message: "file opened", Data: map[string]interface{}{"path": resolvedPath}}, nil
}

func (r *VSCodeToolRegistry) writeFile(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	var args struct {
		Path         string `json:"path"`
		Content      string `json:"content"`
		OpenInEditor bool   `json:"open_in_editor"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode write file payload: %w", err)
	}

	resolvedPath, err := r.resolvePath(args.Path)
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
		if output, runErr := r.runVSCodeCommand(ctx, resolvedPath); runErr != nil {
			return ToolResult{Success: false, Message: stderrOrFallback(output, runErr)}, fmt.Errorf("open written file: %w", runErr)
		}
	}

	return ToolResult{Success: true, Message: "file written", Data: map[string]interface{}{"path": resolvedPath}}, nil
}

func (r *VSCodeToolRegistry) appendFile(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	var args struct {
		Path         string `json:"path"`
		Content      string `json:"content"`
		OpenInEditor bool   `json:"open_in_editor"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return ToolResult{Success: false, Message: "invalid payload"}, fmt.Errorf("decode append file payload: %w", err)
	}

	resolvedPath, err := r.resolvePath(args.Path)
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
		if output, runErr := r.runVSCodeCommand(ctx, resolvedPath); runErr != nil {
			return ToolResult{Success: false, Message: stderrOrFallback(output, runErr)}, fmt.Errorf("open appended file: %w", runErr)
		}
	}

	return ToolResult{Success: true, Message: "content appended", Data: map[string]interface{}{"path": resolvedPath}}, nil
}

func (r *VSCodeToolRegistry) executeCommand(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
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
	stdout, err := r.runVSCodeCommand(ctx, commandArgs...)
	if err != nil {
		return ToolResult{Success: false, Message: stderrOrFallback(stdout, err)}, fmt.Errorf("execute VS Code command: %w", err)
	}

	response := ToolResult{Success: true, Message: "command executed"}
	if stdout != "" {
		response.Data = map[string]interface{}{"output": stdout}
	}
	return response, nil
}

func (r *VSCodeToolRegistry) resolvePath(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", errors.New("path is required")
	}

	var joined string
	if filepath.IsAbs(input) {
		joined = input
	} else {
		joined = filepath.Join(r.workspace, input)
	}

	absTarget, err := filepath.Abs(joined)
	if err != nil {
		return "", fmt.Errorf("resolve target path: %w", err)
	}

	workspacePrefix := r.workspace
	if workspacePrefix != "" && workspacePrefix[len(workspacePrefix)-1] != os.PathSeparator {
		workspacePrefix += string(os.PathSeparator)
	}

	if workspacePrefix != "" && absTarget != r.workspace && !strings.HasPrefix(absTarget, workspacePrefix) {
		return "", fmt.Errorf("path %s is outside the workspace", input)
	}

	return absTarget, nil
}

func (r *VSCodeToolRegistry) runVSCodeCommand(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, r.codePath, args...)
	cmd.Dir = r.workspace
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
