package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"ai-voice-ctrl/backend/executor"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	SpotifyInstanceName = "spotify_client"
	tokenFilePath       = "token.json"
	redirectURI         = "http://127.0.0.1:20721/callback" // 除了端口外均不可改动
	localServerAddr     = "127.0.0.1:20721"                 // 与上方保持一致
)

// spotifyClientInstance 包装了 spotify 客户端以供会话存储
type spotifyClientInstance struct {
	Client *spotify.Client
}

// searchSongArgs 定义了搜索工具的参数结构
type searchSongArgs struct {
	Query string `json:"query"`
}

// RegisterSpotifyTool 向执行器注册所有与 Spotify 相关的工具
func RegisterSpotifyTool(exec executor.Executor) error {
	// 定义 "create_spotify_client" 工具
	createClientDef := executor.ToolDefinition{
		Name:        "create_spotify_client",
		Description: "创建并验证一个 Spotify 客户端。必须在调用任何其他 Spotify 工具之前调用此工具。",
		Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`), // 无参数
		Executor: func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
			// 检查实例是否已存在
			if _, ok := exec.LoadInstance(SpotifyInstanceName); ok {
				return executor.ToolResult{Success: true, Message: "Spotify 客户端已存在，无需重复创建。"}, nil
			}

			// 获取认证后的客户端
			client, err := getAuthenticatedClient(ctx)
			if err != nil {
				return executor.ToolResult{Success: false, Message: fmt.Sprintf("创建 Spotify 客户端失败: %v", err)}, err
			}

			// 存储实例
			instance := &spotifyClientInstance{Client: client}
			exec.StoreInstance(SpotifyInstanceName, instance)

			return executor.ToolResult{Success: true, Message: "Spotify 客户端创建并验证成功。"}, nil
		},
	}

	// 定义 "spotify_search_song" 工具
	searchSongDef := executor.ToolDefinition{
		Name:        "spotify_search_song",
		Description: "在 Spotify 上搜索歌曲并返回其公开链接。",
		Parameters: json.RawMessage(`{
            "type": "object",
            "properties": {
                "query": {
                    "type": "string",
                    "description": "搜索关键词，例如：'晴天 周杰伦'"
                }
            },
            "required": ["query"]
        }`),
		Executor: func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
			// 加载实例
			instance, ok := exec.LoadInstance(SpotifyInstanceName)
			if !ok {
				return executor.ToolResult{Success: false, Message: "错误：Spotify 客户端不存在。请先调用 create_spotify_client。"}, nil
			}
			clientInstance, ok := instance.(*spotifyClientInstance)
			if !ok {
				return executor.ToolResult{Success: false, Message: "内部错误：存储的 Spotify 实例类型不正确。"}, nil
			}

			// 解析参数
			var args searchSongArgs
			if err := json.Unmarshal(payload, &args); err != nil {
				return executor.ToolResult{Success: false, Message: fmt.Sprintf("参数解析失败: %v", err)}, err
			}

			// 执行核心逻辑
			results, err := clientInstance.Client.Search(ctx, args.Query, spotify.SearchTypeTrack)
			if err != nil {
				return executor.ToolResult{Success: false, Message: fmt.Sprintf("Spotify API 搜索失败: %v", err)}, err
			}

			if len(results.Tracks.Tracks) == 0 {
				return executor.ToolResult{Success: true, Message: fmt.Sprintf("没有为查询 '%s' 找到歌曲。", args.Query)}, nil
			}

			track := results.Tracks.Tracks[0]
			trackURL := track.ExternalURLs["spotify"]

			// 返回成功结果
			return executor.ToolResult{
				Success: true,
				Message: fmt.Sprintf("已为查询 '%s' 找到歌曲。", args.Query),
				Data: map[string]interface{}{
					"name":   track.Name,
					"artist": track.Artists[0].Name,
					"url":    trackURL,
				},
			}, nil
		},
	}

	// 定义 "destroy_spotify_client" 工具
	destroyClientDef := executor.ToolDefinition{
		Name:        "destroy_spotify_client",
		Description: "销毁并清理当前的 Spotify 客户端会话。",
		Parameters:  json.RawMessage(`{"type": "object", "properties": {}}`),
		Executor: func(ctx context.Context, payload json.RawMessage) (executor.ToolResult, error) {
			exec.DeleteInstance(SpotifyInstanceName)
			return executor.ToolResult{Success: true, Message: "Spotify 客户端会话已销毁。"}, nil
		},
	}

	// 将所有定义组合到一个 map 中
	spotifyDefinitions := map[string]executor.ToolDefinition{
		createClientDef.Name:  createClientDef,
		searchSongDef.Name:    searchSongDef,
		destroyClientDef.Name: destroyClientDef,
	}

	// 注册整个 "spotify" 工具组
	return exec.RegisterTool("spotify", spotifyDefinitions)
}

// --- Helper Functions ---

// getAuthenticatedClient 处理 OAuth 流程并返回一个可用的客户端
func getAuthenticatedClient(ctx context.Context) (*spotify.Client, error) {
	// 从环境变量或安全存储中获取
	clientID := os.Getenv("SPOTIFY_ID")
	clientSecret := os.Getenv("SPOTIFY_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("环境变量 SPOTIFY_ID 或 SPOTIFY_SECRET 未设置")
	}

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserReadPlaybackState),
	)

	// 尝试从文件加载 token
	token, err := loadTokenFromFile()
	if err != nil {
		// 如果 token 不存在或无效，启动授权流程
		fmt.Println("无法加载 token，启动新的授权流程...")
		token, err = performAuthFlow(ctx, auth)
		if err != nil {
			return nil, err
		}
	}

	httpClient := auth.Client(ctx, token)
	client := spotify.New(httpClient)
	return client, nil
}

// performAuthFlow 启动一个本地服务器来完成 OAuth 授权
func performAuthFlow(ctx context.Context, auth *spotifyauth.Authenticator) (*oauth2.Token, error) {
	var wg sync.WaitGroup
	wg.Add(1)

	var token *oauth2.Token
	var authErr error

	server := &http.Server{Addr: localServerAddr}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		t, err := auth.Token(r.Context(), "state", r)
		if err != nil {
			http.Error(w, "无法获取 token", http.StatusBadRequest)
			authErr = err
			return
		}
		token = t
		fmt.Fprintln(w, "授权成功！您可以关闭此窗口。")
		go func() {
			// 短暂延迟后关闭服务器
			server.Shutdown(context.Background())
		}()
	})

	authURL := auth.AuthURL("state")
	fmt.Println("--- Spotify 授权 ---")
	fmt.Println("请在浏览器中打开以下 URL 完成授权:")
	fmt.Println(authURL)

	// 启动服务器并等待回调
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP 服务器错误: %v", err)
		}
	}()

	wg.Wait() // 等待回调完成

	if authErr != nil {
		return nil, fmt.Errorf("授权流程失败: %w", authErr)
	}
	if token == nil {
		return nil, fmt.Errorf("未能获取 token")
	}

	saveTokenToFile(token)
	return token, nil
}

func loadTokenFromFile() (*oauth2.Token, error) {
	file, err := os.Open(tokenFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var t oauth2.Token
	if err := json.NewDecoder(file).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func saveTokenToFile(token *oauth2.Token) {
	file, err := os.Create(tokenFilePath)
	if err != nil {
		log.Printf("无法创建 token 文件: %v", err)
		return
	}
	defer file.Close()
	json.NewEncoder(file).Encode(token)
}
