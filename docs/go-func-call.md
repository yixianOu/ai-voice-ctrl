# 基于 go-openai 的 Function Calling 示例

下面的示例使用社区常用的 `github.com/sashabaranov/go-openai` 库，实现一个完整的工具调用流程：

1. 定义工具 `search_database`；
2. 将工具清单随对话请求发送给 OpenAI 模型；
3. 解析模型返回的 `tool_calls` 并执行本地逻辑；
4. 将工具执行结果再次发送给模型，获取最终回复。

```go
package main

import (
        "context"
        "encoding/json"
        "fmt"
        "log"

        openai "github.com/sashabaranov/go-openai"
)

type toolArgs struct {
        Query string `json:"query"`
}

func main() {
        client := openai.NewClientFromEnv() // 需提前设置 OPENAI_API_KEY
        ctx := context.Background()

        tool := openai.Tool{
                Type: openai.ToolTypeFunction,
                Function: &openai.FunctionDefinition{
                        Name:        "search_database",
                        Description: "搜索数据库中的用户信息",
                        Parameters: json.RawMessage(`{
                                "type": "object",
                                "properties": {
                                        "query": {
                                                "type": "string",
                                                "description": "搜索关键词"
                                        }
                                },
                                "required": ["query"]
                        }`),
                },
        }

        resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
                Model: openai.GPT4oMini,
                Messages: []openai.ChatCompletionMessage{{
                        Role:    openai.ChatMessageRoleUser,
                        Content: "帮我搜索张三",
                }},
                Tools: []openai.Tool{tool},
        })
        if err != nil {
                log.Fatalf("调用 OpenAI 失败: %v", err)
        }

        choice := resp.Choices[0].Message
        if len(choice.ToolCalls) == 0 {
                log.Println("模型未调用工具，直接回复:", choice.Content)
                return
        }

        call := choice.ToolCalls[0]
        if call.Function.Name != "search_database" {
                log.Fatalf("未知工具: %s", call.Function.Name)
        }

        var args toolArgs
        if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
                log.Fatalf("解析参数失败: %v", err)
        }

        toolResult := searchDatabase(args.Query)

        followResp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
                Model: openai.GPT4oMini,
                Messages: []openai.ChatCompletionMessage{
                        choice,
                        {
                                Role:       openai.ChatMessageRoleTool,
                                ToolCallID: call.ID,
                                Content:    toolResult,
                        },
                },
        })
        if err != nil {
                log.Fatalf("二次请求失败: %v", err)
        }

        fmt.Println("最终回复:", followResp.Choices[0].Message.Content)
}

func searchDatabase(query string) string {
        // TODO: 替换为真实逻辑
        return fmt.Sprintf("找到 3 条关于 '%s' 的结果", query)
}
```

> 生产环境中建议增加：超时控制、重试策略、日志以及对多次工具调用的循环处理。