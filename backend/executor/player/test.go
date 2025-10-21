package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	spotifyauth "github.com/zmb3/spotify/v2/auth" // 使用 spotifyauth
	"golang.org/x/oauth2"
)

// 替换为你自己的值
const redirectURI = "http://127.0.0.1:20721/callback"
const tokenFilePath = "token.json" // 保存 token 的文件名

// 从 Spotify 开发者后台获取
var (
	clientID     = "577ac022b7a2428a9d2ab5f4058d47a8" // 替换
	clientSecret = "78e782b341b14dd3a9da54a5516109e0" // 替换
)

// 全局变量，用于在 HTTP handler 和 main 之间传递
var (
	ch   = make(chan *oauth2.Token) // 用于接收 token
	auth *spotifyauth.Authenticator
)

func getToken() {
	// 0. 检查你的 ID 和 Secret 是否填写
	if clientID == "YOUR_CLIENT_ID" || clientSecret == "YOUR_CLIENT_SECRET" {
		fmt.Println("!!! 错误：请在代码中填入你的 Client ID 和 Client Secret。")
		return
	}

	// 1. 创建 Spotify Authenticator
	// Scopes 决定了你的应用拥有哪些权限
	// 这是你 AI 助手所需要的最关键权限
	auth = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserModifyPlaybackState, // 播放/暂停/切歌 (!!!)
			spotifyauth.ScopeUserReadPlaybackState,   // 读取当前播放状态 (!!!)
			spotifyauth.ScopeUserReadPrivate,         // 读取用户信息
			spotifyauth.ScopePlaylistReadPrivate,     // 读取私有播放列表
		),
	)

	// 2. 启动一个本地 Web 服务器来接收回调
	http.HandleFunc("/callback", callbackHandler)
	go func() {
		log.Println("回调服务器启动于 http://127.0.0.1:20721")
		if err := http.ListenAndServe("127.0.0.1:20721", nil); err != nil {
			log.Fatalf("无法启动回调服务器: %v", err)
		}
	}()

	// 3. 生成并打印授权 URL
	// 我们需要一个 "state" 字符串来防止 CSRF 攻击，这里简单用 "state"
	authURL := auth.AuthURL("state")
	fmt.Println("--- Spotify 授权 ---")
	fmt.Println("请在你的浏览器中打开以下 URL 进行授权:")
	fmt.Println(authURL)
	fmt.Println("--------------------")

	// 4. 等待回调 handler 通过 channel 发送 token
	token := <-ch

	// 5. 将获取到的 token 保存到文件
	saveToken(token)

	fmt.Println("--- 授权成功! ---")
	fmt.Println("Token 已保存到 token.json，你可以关闭此程序了。")
	fmt.Println("现在可以运行你的 AI 助手主程序了。")
}

// callbackHandler 是我们的 HTTP 回调处理器
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	// 检查 state 是否匹配
	if st := r.URL.Query().Get("state"); st != "state" {
		http.Error(w, "State 不匹配", http.StatusBadRequest)
		return
	}

	// 从 URL query 中获取授权码 (code)
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "未找到 Code", http.StatusBadRequest)
		return
	}

	// 2. 使用 code 换取 token
	token, err := auth.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("无法换取 Token: %v", err), http.StatusInternalServerError)
		return
	}

	// 3. 授权成功，向用户显示消息
	fmt.Fprintln(w, "授权成功！你现在可以关闭这个浏览器窗口，返回你的 Go 程序。")

	// 4. 通过 channel 将 token 发送回 main 函数
	ch <- token
}

// saveToken 将 token 保存为 JSON 文件
func saveToken(token *oauth2.Token) {
	file, err := os.Create(tokenFilePath)
	if err != nil {
		log.Fatalf("无法创建 token 文件: %v", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(token); err != nil {
		log.Fatalf("无法将 token 写入文件: %v", err)
	}
}
