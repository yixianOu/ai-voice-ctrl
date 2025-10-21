
## **核心问题：工具实例由llm创建!**

### **模式 B：LLM 按需创建（Lazy Initialization）**
```go
// LLM 决定何时创建工具
// 用户："帮我打开 VSCode 编辑 main.go"
// LLM 调用：create_vscode → edit_file
func (s *Session) HandleFunctionCall(call FunctionCall) {
    switch call.Name {
    case "create_vscode":
        s.Tools["vscode"] = createVSCode(call.Args)
    case "vscode_edit_file":
        s.Tools["vscode"].EditFile(...)
    }
}
```

---

## **Function Call 工具定义示例**

### **方式 1：显式创建 + 操作分离** ✅ 推荐

```json
{
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "create_vscode",
        "description": "创建 VSCode 实例（打开工作区）",
        "parameters": {
          "type": "object",
          "properties": {
            "workspace": {"type": "string", "description": "工作区路径"}
          },
          "required": ["workspace"]
        }
      }
    },
    {
      "type": "function",
      "function": {
        "name": "vscode_open_file",
        "description": "在 VSCode 中打开文件（需要先调用 create_vscode）",
        "parameters": {
          "type": "object",
          "properties": {
            "file": {"type": "string", "description": "文件路径"}
          }
        }
      }
    },
    {
      "type": "function",
      "function": {
        "name": "destroy_vscode",
        "description": "关闭 VSCode 实例"
      }
    }
  ]
}
```

**LLM 调用流程：**
```
用户："帮我用 VSCode 编辑 /code/main.go"

LLM 推理：
1. 需要先创建 VSCode 实例
2. 然后打开文件

调用序列：
→ create_vscode({"workspace": "/code"})
← "VSCode 已启动"
→ vscode_open_file({"file": "/code/main.go"})
← "文件已打开"
```

---

## **针对你的 Go 客户端 + Function Call 场景的最佳实践**

### **✅ 推荐方案：显式创建 + 状态追踪**

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
)

type DesktopController struct {
    Registry *ToolRegistry
    Session  *Session
}

func (dc *DesktopController) ChatLoop(ctx context.Context) {
    for {
        userInput := getUserInput()
        
        // 1. 调用 LLM（传递可用工具列表）
        response := dc.callLLM(userInput, dc.getToolDefinitions())
        
        // 2. 处理 tool_calls
        for _, toolCall := range response.ToolCalls {
            result, err := dc.handleToolCall(toolCall)
            
            // 3. 将结果返回给 LLM
            dc.sendToolResult(toolCall.ID, result, err)
        }
        
        // 4. 输出最终回复
        fmt.Println(response.Content)
    }
}

func (dc *DesktopController) handleToolCall(call ToolCall) (string, error) {
    switch call.Function.Name {
    case "create_vscode":
        workspace := call.Function.Arguments["workspace"].(string)
        instance, err := dc.Registry.GetOrCreate("vscode", map[string]interface{}{
            "workspace": workspace,
        })
        if err != nil {
            return "", err
        }
        dc.Session.Tools["vscode"] = instance
        return fmt.Sprintf("VSCode 已打开工作区: %s", workspace), nil
        
    case "vscode_open_file":
        instance := dc.Session.Tools["vscode"]
        if instance == nil {
            return "", fmt.Errorf("VSCode 未启动，请先调用 create_vscode")
        }
        
        file := call.Function.Arguments["file"].(string)
        return instance.Adapter.Execute("open_file", map[string]interface{}{
            "file": file,
        })
        
    case "create_vlc":
        instance, err := dc.Registry.GetOrCreate("vlc", nil)
        if err != nil {
            return "", err
        }
        dc.Session.Tools["vlc"] = instance
        return "VLC 已启动", nil
        
    case "vlc_play":
        instance := dc.Session.Tools["vlc"]
        if instance == nil {
            return "", fmt.Errorf("VLC 未启动，请先调用 create_vlc")
        }
        
        file := call.Function.Arguments["file"].(string)
        return instance.Adapter.Execute("play", map[string]interface{}{
            "file": file,
        })
    }
    
    return "", fmt.Errorf("unknown tool: %s", call.Function.Name)
}
```

---

## **关键设计要点**

### **1. 工具定义中明确依赖关系**

```json
{
  "name": "vscode_open_file",
  "description": "在 VSCode 中打开文件。⚠️ 前置条件：必须先调用 create_vscode 创建实例",
  "parameters": {...}
}
```

### **2. 错误提示要明确**

```go
if instance == nil {
    return "", fmt.Errorf(
        "VSCode 未启动。请先调用 create_vscode 函数创建实例，" +
        "例如：create_vscode({\"workspace\": \"/path/to/project\"})",
    )
}
```

### **3. 在系统提示词中说明生命周期**

```
你可以控制桌面应用。使用流程：
1. 使用 create_<tool> 创建工具实例（如 create_vscode）
2. 使用 <tool>_<action> 执行操作（如 vscode_open_file）
3. 使用 destroy_<tool> 关闭工具（可选，系统会自动清理）

示例对话：
用户："帮我编辑 main.go"
助手：
→ create_vscode({"workspace": "/code"})
→ vscode_open_file({"file": "/code/main.go"})
```

## **最终建议**

### **对于你的 Go 客户端项目：**

✅ **采用：LLM 按需创建 + 显式生命周期管理**

**原因：**
1. **资源效率**：不会在启动时打开所有应用
2. **灵活性**：支持多实例（同时打开 2 个 VSCode）
3. **LLM 友好**：符合自然语言理解（"先打开，再操作"）
4. **易于扩展**：新增工具只需注册，无需修改启动逻辑

**实现步骤：**
1. ✅ 为每个工具定义 `create_*` 和 `*_action` 函数
2. ✅ 在 `handleToolCall` 中检查实例是否存在
3. ✅ 错误提示中引导 LLM 先创建实例
4. ⚠️ 可选：对轻量级工具（clipboard、system_info）用预加载
