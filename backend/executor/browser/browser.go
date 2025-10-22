package browser

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-voice-ctrl/backend/executor"

	"github.com/toqueteos/webbrowser"
)

const (
	BrowserInstanceKey = "browser_session"
)

// browserSessionInstance 包装了浏览器会话的状态。
// 在这个简单的例子中，它只保存了要使用的浏览器名称。
type browserSessionInstance struct {
	BrowserName string
}

// createBrowserSessionArgs 定义了创建会话工具的参数结构
type createBrowserSessionArgs struct {
	// Browser 字段可以为空，表示使用系统默认浏览器
	Browser string `json:"browser,omitempty"`
}

// openURLArgs 定义了打开网页工具的参数结构
type openURLArgs struct {
	URL string `json:"url"`
}

// RegisterBrowserTool 向执行器注册所有与浏览器相关的工具
func RegisterBrowserTool(exec executor.Executor) error {
	// 定义 "create_browser_session" 工具
	createSessionDef := executor.ToolDefinition{
		Name:        "create_browser_session",
		Description: "初始化一个浏览器会话。在打开任何网页之前，应首先调用此工具来指定要使用的浏览器。",
		Parameters: json.RawMessage(`{
            "type": "object",
            "properties": {
                "browser": {
                    "type": "string",
                    "description": "要使用的浏览器名称，例如 'chrome', 'firefox'。如果留空或省略，将使用系统默认浏览器。"
                }
            }
        }`),
		Executor: func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
			if _, ok := exec.LoadInstance(BrowserInstanceKey); ok {
				return executor.ToolResult{Success: true, Message: "浏览器会话已存在，无需重复创建。"}, nil
			}

			var args createBrowserSessionArgs
			if err := json.Unmarshal(payload, &args); err != nil {
				// 即使解析失败也继续，因为 browser 参数是可选的
				args.Browser = "" // 使用默认值
			}

			browserName := args.Browser
			if browserName == "" {
				browserName = "default"
			}

			instance := &browserSessionInstance{BrowserName: browserName}
			exec.StoreInstance(BrowserInstanceKey, instance)

			return executor.ToolResult{Success: true, Message: fmt.Sprintf("浏览器会话已创建，将使用 '%s' 浏览器。", browserName)}, nil
		},
	}

	// 定义 "browser_open_url" 工具
	openURLDef := executor.ToolDefinition{
		Name:        "browser_open_url",
		Description: "在当前会话指定的浏览器中打开一个网页。",
		Parameters: json.RawMessage(`{
            "type": "object",
            "properties": {
                "url": {
                    "type": "string",
                    "description": "要打开的完整网址，必须以 http:// 或 https:// 开头。"
                }
            },
            "required": ["url"]
        }`),
		Executor: func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
			instance, ok := exec.LoadInstance(BrowserInstanceKey)
			if !ok {
				return executor.ToolResult{Success: false, Message: "错误：浏览器会话不存在。请先调用 create_browser_session。"}, nil
			}
			session, ok := instance.(*browserSessionInstance)
			if !ok {
				return executor.ToolResult{Success: false, Message: "内部错误：存储的浏览器会话实例类型不正确。"}, nil
			}

			var args openURLArgs
			if err := json.Unmarshal(payload, &args); err != nil {
				return executor.ToolResult{Success: false, Message: fmt.Sprintf("参数解析失败: %v", err)}, err
			}
			if args.URL == "" {
				return executor.ToolResult{Success: false, Message: "错误：'url' 参数不能为空。"}, nil
			}

			if err := webbrowser.Open(args.URL); err != nil {
				return executor.ToolResult{Success: false, Message: fmt.Sprintf("打开网页失败: %v", err)}, err
			}

			return executor.ToolResult{Success: true, Message: fmt.Sprintf("已在 '%s' 浏览器中成功打开网页: %s", session.BrowserName, args.URL)}, nil
		},
	}

	// 定义 "destroy_browser_session" 工具
	destroySessionDef := executor.ToolDefinition{
		Name:        "destroy_browser_session",
		Description: "销毁并清理当前的浏览器会话。",
		Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
		Executor: func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
			exec.DeleteInstance(BrowserInstanceKey)
			return executor.ToolResult{Success: true, Message: "浏览器会话已销毁。"}, nil
		},
	}

	// 注册所有工具
	browserDefinitions := map[string]executor.ToolDefinition{
		createSessionDef.Name:  createSessionDef,
		openURLDef.Name:        openURLDef,
		destroySessionDef.Name: destroySessionDef,
	}

	return exec.RegisterTool("browser", browserDefinitions)
}
