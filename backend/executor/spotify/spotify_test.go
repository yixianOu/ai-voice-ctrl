package spotify_test

import (
	"context"
	"encoding/json"
	"testing"

	"ai-voice-ctrl/backend/executor"
	spotifyExecutor "ai-voice-ctrl/backend/executor/spotify"
)

// TestSpotifyToolLifecycle 测试 Spotify 工具的完整生命周期：
// create -> search -> destroy
func TestSpotifyToolLifecycle(t *testing.T) {
	// --- 前置条件检查 ---
	// 这个测试需要真实调用 Spotify API，因此需要设置环境变量
	// if os.Getenv("SPOTIFY_ID") == "" || os.Getenv("SPOTIFY_SECRET") == "" {
	// 	t.Skip("跳过 Spotify 生命周期测试：环境变量 SPOTIFY_ID 和 SPOTIFY_SECRET 未设置。")
	// }
	// 清理旧的 token 文件，确保测试环境纯净
	// os.Remove("token.json")
	// defer os.Remove("spotify_token.json") // 测试结束后再次清理

	// --- 1. 初始化执行器并注册工具 ---
	exec := executor.NewDefaultExecutor()
	if err := spotifyExecutor.RegisterSpotifyTool(exec); err != nil {
		t.Fatalf("RegisterSpotifyTool 失败: %v", err)
	}

	ctx := context.Background()

	// --- 2. 模拟 LLM 行为：在创建客户端之前就尝试搜索 ---
	t.Run("Search before create", func(t *testing.T) {
		searchArgs, _ := json.Marshal(map[string]string{"query": "周杰伦"})
		searchCall := executor.ToolCall{
			Name:      "spotify_search_song",
			Arguments: json.RawMessage(searchArgs),
		}

		result, err := exec.ExecuteTool(ctx, searchCall)
		if err != nil {
			t.Errorf("ExecuteTool 不应返回顶层错误，而是通过 ToolResult 传递失败信息, got: %v", err)
		}
		if result.Success {
			t.Errorf("预期调用会失败，但结果为成功。 Message: %s", result.Message)
		}
		expectedMsg := "错误：Spotify 客户端不存在。请先调用 create_spotify_client。"
		if result.Message != expectedMsg {
			t.Errorf("预期的错误信息是 '%s', 但得到 '%s'", expectedMsg, result.Message)
		}
	})

	// --- 3. 模拟 LLM 行为：创建 Spotify 客户端 ---
	// 注意：首次运行时，这会需要您在浏览器中手动授权一次。
	t.Run("Create client", func(t *testing.T) {
		createCall := executor.ToolCall{Name: "create_spotify_client"}
		result, err := exec.ExecuteTool(ctx, createCall)
		if err != nil || !result.Success {
			t.Fatalf("创建客户端失败: err=%v, message=%s", err, result.Message)
		}
		if result.Message != "Spotify 客户端创建并验证成功。" {
			t.Errorf("预期的成功信息不匹配, got: %s", result.Message)
		}
	})

	// --- 4. 模拟 LLM 行为：重复创建客户端 ---
	t.Run("Create client again", func(t *testing.T) {
		createCall := executor.ToolCall{Name: "create_spotify_client"}
		result, err := exec.ExecuteTool(ctx, createCall)
		if err != nil || !result.Success {
			t.Fatalf("重复创建客户端失败: err=%v, message=%s", err, result.Message)
		}
		if result.Message != "Spotify 客户端已存在，无需重复创建。" {
			t.Errorf("预期的信息不匹配, got: %s", result.Message)
		}
	})

	// --- 5. 模拟 LLM 行为：成功创建后进行搜索 ---
	t.Run("Search after create", func(t *testing.T) {
		searchArgs, _ := json.Marshal(map[string]string{"query": "牵丝戏"})
		searchCall := executor.ToolCall{
			Name:      "spotify_search_song",
			Arguments: json.RawMessage(searchArgs),
		}
		result, err := exec.ExecuteTool(ctx, searchCall)
		if err != nil || !result.Success {
			t.Fatalf("搜索歌曲失败: err=%v, message=%s", err, result.Message)
		}
		if result.Data == nil {
			t.Fatal("预期结果中应包含 Data 字段，但得到 nil")
		}
		if _, ok := result.Data["url"]; !ok {
			t.Error("预期 Data 字段中应包含 'url' 键")
		}
		t.Logf("搜索成功，返回数据: %v", result.Data)
	})

	// --- 6. 模拟 LLM 行为：销毁客户端 ---
	t.Run("Destroy client", func(t *testing.T) {
		destroyCall := executor.ToolCall{Name: "destroy_spotify_client"}
		result, err := exec.ExecuteTool(ctx, destroyCall)
		if err != nil || !result.Success {
			t.Fatalf("销毁客户端失败: err=%v, message=%s", err, result.Message)
		}
		if result.Message != "Spotify 客户端会话已销毁。" {
			t.Errorf("预期的信息不匹配, got: %s", result.Message)
		}
	})

	// --- 7. 模拟 LLM 行为：销毁后再次搜索 ---
	t.Run("Search after destroy", func(t *testing.T) {
		searchArgs, _ := json.Marshal(map[string]string{"query": "周杰伦"})
		searchCall := executor.ToolCall{
			Name:      "spotify_search_song",
			Arguments: json.RawMessage(searchArgs),
		}
		result, _ := exec.ExecuteTool(ctx, searchCall)
		if result.Success {
			t.Errorf("预期在销毁后调用会失败，但结果为成功。 Message: %s", result.Message)
		}
		expectedMsg := "错误：Spotify 客户端不存在。请先调用 create_spotify_client。"
		if result.Message != expectedMsg {
			t.Errorf("预期的错误信息是 '%s', 但得到 '%s'", expectedMsg, result.Message)
		}
	})
}
