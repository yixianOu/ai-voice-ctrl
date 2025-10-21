package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ai-voice-ctrl/backend/executor" // 确保路径正确
	"ai-voice-ctrl/backend/executor/browser"
)

func main() {
	// 1. 创建执行器实例
	exec := executor.NewDefaultExecutor()

	// 2. 注册新的浏览器工具
	if err := browser.RegisterBrowserTool(exec); err != nil {
		log.Fatalf("注册浏览器工具失败: %v", err)
	}

	// --- 模拟 LLM 对话流程 ---
	ctx := context.Background()

	// 3. 模拟 LLM 调用 create_browser_session (使用默认浏览器)
	fmt.Println("--- 步骤 1: 创建浏览器会话 (默认) ---")
	createCall := executor.ToolCall{
		Name:      "create_browser_session",
		Arguments: json.RawMessage(`{}`), // 空参数表示使用默认浏览器
	}
	result, err := exec.ExecuteTool(ctx, createCall)
	if err != nil || !result.Success {
		log.Fatalf("创建会话失败: %v, %s", err, result.Message)
	}
	fmt.Println("结果:", result.Message)

	// 4. 模拟 LLM 调用 browser_open_url
	fmt.Println("\n--- 步骤 2: 打开 Google ---")
	openArgs, _ := json.Marshal(map[string]string{"url": "https://www.google.com"})
	openCall := executor.ToolCall{
		Name:      "browser_open_url",
		Arguments: json.RawMessage(openArgs),
	}
	result, err = exec.ExecuteTool(ctx, openCall)
	if err != nil || !result.Success {
		log.Fatalf("打开网页失败: %v, %s", err, result.Message)
	}
	fmt.Println("结果:", result.Message)

	// 5. 模拟 LLM 调用 destroy_browser_session
	fmt.Println("\n--- 步骤 3: 销毁会话 ---")
	destroyCall := executor.ToolCall{Name: "destroy_browser_session"}
	result, err = exec.ExecuteTool(ctx, destroyCall)
	if err != nil || !result.Success {
		log.Fatalf("销毁会话失败: %v, %s", err, result.Message)
	}
	fmt.Println("结果:", result.Message)
}
