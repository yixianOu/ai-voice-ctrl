package backend

import (
	"fmt"
	"os/exec"
	"strings"
)

// AppController 应用程序控制器
type AppController struct {
	appMap map[string]string
}

// NewAppController 创建应用控制器
func NewAppController() *AppController {
	return &AppController{
		appMap: map[string]string{
			"firefox":  "firefox",
			"火狐":       "firefox",
			"chrome":   "google-chrome",
			"谷歌":       "google-chrome",
			"浏览器":      "firefox",
			"vscode":   "code",
			"vs code":  "code",
			"终端":       "gnome-terminal",
			"terminal": "gnome-terminal",
			"文件管理器":    "nautilus",
			"files":    "nautilus",
		},
	}
}

// OpenApp 打开应用程序
func (ac *AppController) OpenApp(command string) string {
	for key, app := range ac.appMap {
		if strings.Contains(command, key) {
			cmd := exec.Command("xdg-open", app)
			err := cmd.Start()
			if err != nil {
				return fmt.Sprintf("打开 %s 失败: %v", app, err)
			}
			return fmt.Sprintf("✅ 已打开 %s", app)
		}
	}

	return "未找到要打开的应用程序"
}

// AddApp 添加新的应用映射
func (ac *AppController) AddApp(keyword, appName string) {
	ac.appMap[keyword] = appName
}
