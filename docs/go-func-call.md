根据你的需求"使用 Go 快速实现"，我建议：

---

## **结论：优先用 Function Calling，而非 MCP**

| 维度 | **Function Calling** ✅ | **MCP** |
|------|------------------------|---------|
| **实现复杂度** | 极低（只需返回 JSON Schema + 执行函数） | 中等（需实现完整 JSON-RPC 2.0 协议） |
| **开发时间** | 30 分钟 - 2 小时 | 2-5 天 |
| **Go 生态支持** | 原生支持（任何 HTTP 库即可） | 需自建或找第三方库 |
| **调试难度** | 简单（直接 HTTP 调试） | 复杂（需处理 stdio/SSE 协议层） |
| **适用场景** | 直接与 LLM API 集成（OpenAI/Anthropic/Azure） | 需要标准化工具接口、多客户端共享 |

---

### 1. **极简实现（Go 示例）**
```go
// main.gopackage main

import (
        "context""encoding/json""fmt""log""os""os/exec""runtime""github.com/sashabaranov/go-openai"
)

// main 函数是整个程序的入口func main() {
        // --- 1. 设置 ---
        apiKey := os.Getenv("OPENAI_API_KEY")
        if apiKey == "" {
                log.Fatal("请设置 OPENAI_API_KEY 环境变量")
        }
        client := openai.NewClient(apiKey)
        ctx := context.Background()

        // 用户的自然语言指令
        userInput := "can you open vscode for me?"// --- 2. 定义我们的工具 (Function Call 的 Go 语言表示) ---// 这部分代码精确地对应了我们第一步设计的 JSON 结构
        openAppTool := openai.Tool{
                Type: openai.ToolTypeFunction,
                Function: &openai.FunctionDefinition{
                        Name:        "open_application",
                        Description: "Opens a specified application on the user's computer.",
                        Parameters:  json.RawMessage(`...`), // 在下方填充
                },
        }
        // 为了代码清晰，将 JSON 参数定义为字符串
        openAppTool.Function.Parameters = json.RawMessage(`{
                "type": "object",
                "properties": {
                        "app_name": {
                                "type": "string",
                                "description": "The name of the application to open, e.g., 'Visual Studio Code', 'Slack'."
                        }
                },
                "required": ["app_name"]
        }`)

        // --- 3. 调用 LLM，让它进行决策 ---
        fmt.Println(">> 正在向 OpenAI 发送请求，让它决定使用哪个工具...")
        resp, err := client.CreateChatCompletion(
                ctx,
                openai.ChatCompletionRequest{
                        Model: openai.GPT4o, // 推荐使用支持工具调用的新模型
                        Messages: []openai.ChatCompletionMessage{
                                {
                                        Role:    openai.ChatMessageRoleUser,
                                        Content: userInput,
                                },
                        },
                        Tools: []openai.Tool{openAppTool}, // 把我们的工具清单发给 LLM
                },
        )

        if err != nil {
                log.Fatalf("ChatCompletion 错误: %v", err)
        }

        // --- 4. 解析 LLM 的响应并执行操作 ---
        message := resp.Choices[0].Message
        // 检查 LLM 是否决定要调用我们的工具if len(message.ToolCalls) > 0 {
                toolCall := message.ToolCalls[0]
                functionName := toolCall.Function.Name
                fmt.Printf(">> OpenAI 决定调用工具: %s\n", functionName)

                // 使用 switch 来处理不同的工具调用，这使得扩展新功能变得容易switch functionName {
                case "open_application":
                        // 解析 LLM 提供的参数var args struct {
                                AppName string `json:"app_name"`
                        }
                        err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
                        if err != nil {
                                log.Fatalf("解析工具参数失败: %v", err)
                        }
                        
                        // 调用我们自己编写的、安全的执行函数
                        err = executeOpenApplication(args.AppName)
                        if err != nil {
                                fmt.Printf("!! 执行失败: %v\n", err)
                        } else {
                                fmt.Printf("✅ 成功执行: 已尝试打开 '%s'\n", args.AppName)
                        }
                default:
                        fmt.Printf("!! 未知的工具: %s\n", functionName)
                }
        } else {
                // 如果 LLM 没调用工具，而是直接回复了文本
                fmt.Println(">> OpenAI 直接回复:")
                fmt.Println(message.Content)
        }
}

// executeOpenApplication 是真正的“执行层”代码 (MCP 的一部分)// 它负责与操作系统交互，是安全且预先编写好的。func executeOpenApplication(appName string) error {
        fmt.Printf(">> 正在尝试在系统 (%s) 上打开: %s\n", runtime.GOOS, appName)
        var cmd *exec.Cmd

        // 为了跨平台兼容性，我们检测当前的操作系统switch runtime.GOOS {
        case "darwin": // macOS// 在 macOS, 'open -a' 是打开应用的标准方式
                cmd = exec.Command("open", "-a", appName)
        case "windows":
                // 在 Windows, 'start' 命令可以用来启动应用// 注意：这里的 "" 是为了处理应用名称中可能包含空格的情况
                cmd = exec.Command("cmd", "/C", "start", "", appName)
        case "linux":
                // 在 Linux, 我们假设应用的可执行文件在系统的 PATH 中
                cmd = exec.Command(appName)
        default:
                return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
        }

        // 执行命令并返回任何可能发生的错误return cmd.Run()
}
```
