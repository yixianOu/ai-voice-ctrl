# Executor 工具生命周期管理

## 概述

`Executor` 内置工具实例管理，支持 LLM 驱动的生命周期控制。工具实例由模型调用 `create_*` 函数动态创建。

backend/executor/executor.go: 类型定义，接口定义
backend/executor/aggregator.go: 单例模式，工具控制器
backend/executor/vscode.go: VSCode 具体工具生命周期实现
backend/executor/session_example_test.go: 使用示例代码和单元测试

## 核心概念

### 1. 工具生命周期

每个工具包含三类函数：

- **创建函数** (`create_*`)：由 LLM 调用以创建工具实例
- **操作函数** (`<tool>_<action>`)：执行具体操作，需要先创建实例
- **销毁函数** (`destroy_*`)：释放工具资源（可选）

### 2. 实例管理

`DefaultExecutor` 内部维护 `instances sync.Map`，存储当前会话中已创建的工具实例。

## 使用示例

### 基本流程

```go
// 1. 创建 executor
exec := executor.NewDefaultExecutor()

// 2. 注册工具（包含生命周期函数）
if err := executor.RegisterVSCodeLifecycle(exec); err != nil {
    log.Fatal(err)
}

// 3. 获取工具定义并传递给 LLM
schemas := exec.ToolSchemas()
// 将 schemas 发送给 LLM...

// 4. 处理 LLM 返回的 tool call
result, err := exec.ExecuteTool(ctx, toolCall)
```

### LLM 调用序列

```
用户："帮我用 VSCode 编辑 /code/main.go"

LLM 推理：
1. 需要先创建 VSCode 实例
2. 然后打开文件

调用序列：
→ create_vscode({"workspace": "/code"})
← {"success": true, "message": "VSCode instance created"}

→ vscode_open_file({"path": "main.go"})
← {"success": true, "message": "file opened"}
```

### 错误处理

如果 LLM 在未创建实例的情况下调用操作函数：

```go
// LLM 直接调用 vscode_open_file（未先调用 create_vscode）
result, err := exec.ExecuteTool(ctx, ToolCall{
    Name: "vscode_open_file",
    Arguments: json.RawMessage(`{"path": "test.go"}`),
})

// 返回：
// success: false
// message: "VSCode not created. Please call create_vscode first"
```

## 扩展新工具

每个工具在自己的文件中实现 `RegisterLifecycle` 方法：

```go
// vlc.go
type VLCTool struct {
    // ...
}

func (t *VLCTool) RegisterLifecycle(exec executor.Executor) error {
    definitions := map[string]executor.ToolDefinition{
        "create_vlc": {
            Name: "create_vlc",
            Executor: func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
                vlc := NewVLCTool()
                exec.StoreInstance("vlc", vlc)
                return executor.ToolResult{Success: true, Message: "VLC created"}, nil
            },
        },
        "vlc_play": {
            Name: "vlc_play",
            Executor: func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
                instance, ok := exec.LoadInstance("vlc")
                if !ok {
                    return executor.ToolResult{Success: false, Message: "VLC not created"}, errors.New("...")
                }
                // ...
            },
        },
        "destroy_vlc": {
            Name: "destroy_vlc",
            Executor: func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
                exec.DeleteInstance("vlc")
                return executor.ToolResult{Success: true, Message: "VLC destroyed"}, nil
            },
        },
    }
    return exec.RegisterTool("vlc", definitions)
}

// 便捷包装函数
func RegisterVLCLifecycle(exec executor.Executor) error {
    tool := &VLCTool{}
    return tool.RegisterLifecycle(exec)
}
```

## 系统提示词建议

```
你可以控制本地桌面应用。使用流程：

1. 使用 create_<tool> 创建工具实例（如 create_vscode）
2. 使用 <tool>_<action> 执行操作（如 vscode_open_file）
3. 使用 destroy_<tool> 关闭工具（可选，会话结束时自动清理）

注意：操作函数必须在创建实例后才能调用，否则会返回错误。

示例：
用户："帮我编辑 main.go"
助手：
→ create_vscode({"workspace": "/current/path"})
→ vscode_open_file({"path": "main.go"})
```

## 注意事项

1. **单实例限制**：当前设计每个工具类型只支持一个实例（如只能有一个 VSCode 实例）
2. **会话隔离**：不同 Executor 实例的工具状态互相独立
3. **并发安全**：内部使用 `sync.Map` 保证并发安全
4. **资源清理**：会话结束时应显式清理，或依赖 GC 回收
5. **工具定义位置**：每个工具的生命周期定义在其自己的文件中（如 `vscode.go`），保持模块独立性
