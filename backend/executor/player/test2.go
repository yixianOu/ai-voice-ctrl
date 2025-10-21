package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// loadToken 从文件加载 token
func loadToken() (*oauth2.Token, error) {
	file, err := os.Open(tokenFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var token oauth2.Token
	if err := json.NewDecoder(file).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

// 这是你 AI 助手的核心逻辑
func main() {
	// 0. get token
	getToken()
	// 1. 加载我们之前获取的 token
	token, err := loadToken()
	if err != nil {
		log.Fatalf("无法加载 token.json: %v\n请先运行 authorize.go 完成授权。", err)
	}

	// 2. 创建 Authenticator
	// (注意：这里的 RedirectURL 和 Scopes 必须和 authorize.go 中的完全一致)
	auth := spotifyauth.New(spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		// 即使我们只是加载 token，也需要这些信息来让它自动刷新
	)

	// 3. 创建一个 HTTP 客户端，它会自动处理 token 刷新
	ctx := context.Background()
	httpClient := auth.Client(ctx, token)

	// 4. 创建 Spotify 客户端 (!!!)
	// 这是你真正用来调用 API 的对象
	client := spotify.New(httpClient)

	fmt.Println("--- Spotify AI 助手已准备就绪 ---")

	// -----------------------------------------------------------------
	// 演示：这就是你的 AI 助手的 Function Call 最终要调用的地方
	// -----------------------------------------------------------------

	// 模拟 LLM 返回了 `searchMusic(query="牵丝戏")`
	fmt.Println("\n[演示] 正在执行 searchMusic(query='牵丝戏')...")
	searchResult, err := client.Search(ctx, "牵丝戏", spotify.SearchTypeTrack)
	if err != nil {
		log.Fatalf("搜索失败: %v", err)
	}

	if len(searchResult.Tracks.Tracks) == 0 {
		fmt.Println("没有找到歌曲。")
		return
	}

	// 获取搜索到的第一首歌
	track := searchResult.Tracks.Tracks[0]

	// 从歌曲信息中获取公开的 URL
	trackURL, exists := track.ExternalURLs["spotify"]
	if !exists {
		log.Fatal("无法找到该歌曲的 Spotify 链接。")
	}

	fmt.Printf("✅ 找到歌曲: %s - %s\n", track.Name, track.Artists[0].Name)
	fmt.Printf("🔗 链接: %s\n", trackURL)
}
