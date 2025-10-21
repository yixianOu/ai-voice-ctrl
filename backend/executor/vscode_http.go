package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// VSCodeHTTPTool communicates with VSCode via extension HTTP bridge.
type VSCodeHTTPTool struct {
	workspace  string
	bridgeURL  string
	httpClient *http.Client
}

// NewVSCodeHTTPTool creates a VSCode HTTP tool instance.
func NewVSCodeHTTPTool(workspace string, bridgePort int) (*VSCodeHTTPTool, error) {
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return nil, errors.New("workspace path is required")
	}

	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace path: %w", err)
	}

	return &VSCodeHTTPTool{
		workspace: absWorkspace,
		bridgeURL: fmt.Sprintf("http://localhost:%d", bridgePort),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (t *VSCodeHTTPTool) sendCommand(ctx context.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	payload := map[string]interface{}{
		"action": action,
		"params": params,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal command: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.bridgeURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if success, ok := result["success"].(bool); !ok || !success {
		msg := "command failed"
		if m, ok := result["message"].(string); ok {
			msg = m
		}
		return nil, errors.New(msg)
	}

	return result, nil
}

func (t *VSCodeHTTPTool) openFile(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	var args struct {
		Path string `json:"path"`
		Line *int   `json:"line"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return ToolResult{Success: false, Message: "invalid payload"}, err
	}

	resolvedPath, err := t.resolvePath(args.Path)
	if err != nil {
		return ToolResult{Success: false, Message: err.Error()}, err
	}

	params := map[string]interface{}{"path": resolvedPath}
	if args.Line != nil {
		params["line"] = *args.Line
	}

	result, err := t.sendCommand(ctx, "open_file", params)
	if err != nil {
		return ToolResult{Success: false, Message: err.Error()}, err
	}

	return ToolResult{Success: true, Message: "file opened", Data: result}, nil
}

func (t *VSCodeHTTPTool) closeFile(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return ToolResult{Success: false, Message: "invalid payload"}, err
	}

	resolvedPath, err := t.resolvePath(args.Path)
	if err != nil {
		return ToolResult{Success: false, Message: err.Error()}, err
	}

	result, err := t.sendCommand(ctx, "close_file", map[string]interface{}{"path": resolvedPath})
	if err != nil {
		return ToolResult{Success: false, Message: err.Error()}, err
	}

	return ToolResult{Success: true, Message: "file closed", Data: result}, nil
}

func (t *VSCodeHTTPTool) refreshFile(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return ToolResult{Success: false, Message: "invalid payload"}, err
	}

	resolvedPath, err := t.resolvePath(args.Path)
	if err != nil {
		return ToolResult{Success: false, Message: err.Error()}, err
	}

	result, err := t.sendCommand(ctx, "refresh_file", map[string]interface{}{"path": resolvedPath})
	if err != nil {
		return ToolResult{Success: false, Message: err.Error()}, err
	}

	return ToolResult{Success: true, Message: "file refreshed", Data: result}, nil
}

func (t *VSCodeHTTPTool) closeWindow(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
	result, err := t.sendCommand(ctx, "close_window", map[string]interface{}{})
	if err != nil {
		return ToolResult{Success: false, Message: err.Error()}, err
	}

	return ToolResult{Success: true, Message: "window closed", Data: result}, nil
}

func (t *VSCodeHTTPTool) resolvePath(input string) (string, error) {
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

	return absTarget, nil
}

// RegisterHTTPLifecycle registers VSCode HTTP bridge tool.
func (t *VSCodeHTTPTool) RegisterHTTPLifecycle(exec Executor) error {
	definitions := map[string]ToolDefinition{
		"vscode_open_file": {
			Name:        "vscode_open_file",
			Description: "Open file in VSCode via HTTP bridge",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {"type": "string", "description": "File path"},
					"line": {"type": "integer", "minimum": 1, "description": "Line number"}
				},
				"required": ["path"]
			}`),
			Executor: func(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
				return t.openFile(ctx, payload)
			},
		},
		"vscode_close_file": {
			Name:        "vscode_close_file",
			Description: "Close file in VSCode via HTTP bridge",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {"type": "string", "description": "File path"}
				},
				"required": ["path"]
			}`),
			Executor: func(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
				return t.closeFile(ctx, payload)
			},
		},
		"vscode_refresh_file": {
			Name:        "vscode_refresh_file",
			Description: "Refresh file content in VSCode via HTTP bridge",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"path": {"type": "string", "description": "File path"}
				},
				"required": ["path"]
			}`),
			Executor: func(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
				return t.refreshFile(ctx, payload)
			},
		},
		"vscode_close_window": {
			Name:        "vscode_close_window",
			Description: "Close VSCode window via HTTP bridge",
			Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
			Executor: func(ctx context.Context, payload json.RawMessage) (ToolResult, error) {
				return t.closeWindow(ctx, payload)
			},
		},
	}

	return exec.RegisterTool("vscode_http", definitions)
}
