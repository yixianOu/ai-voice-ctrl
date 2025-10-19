package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// ProcessVoiceCommand 处理语音识别的命令
func (a *App) ProcessVoiceCommand(command string) string {
	command = strings.TrimSpace(strings.ToLower(command))

	// 打开应用
	if strings.Contains(command, "打开") {
		return a.handleOpenApp(command)
	}

	// 音量控制
	if strings.Contains(command, "音量") {
		return a.handleVolume(command)
	}

	// 音乐控制
	if strings.Contains(command, "音乐") || strings.Contains(command, "播放") ||
		strings.Contains(command, "暂停") || strings.Contains(command, "上一首") ||
		strings.Contains(command, "下一首") {
		return a.handleMusic(command)
	}

	// 文件操作
	if strings.Contains(command, "创建文件") || strings.Contains(command, "写入") {
		return a.handleFile(command)
	}

	return fmt.Sprintf("抱歉，我还不知道如何执行：%s", command)
}

// 打开应用程序
func (a *App) handleOpenApp(command string) string {
	appMap := map[string]string{
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
	}

	for key, app := range appMap {
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

// 音量控制
func (a *App) handleVolume(command string) string {
	// 提取音量数字
	var volume string
	if strings.Contains(command, "设为") || strings.Contains(command, "调到") {
		// 简单提取数字
		parts := strings.Fields(command)
		for _, part := range parts {
			part = strings.TrimSuffix(part, "%")
			if len(part) > 0 && part[0] >= '0' && part[0] <= '9' {
				volume = part
				break
			}
		}

		if volume != "" {
			cmd := exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", volume+"%")
			err := cmd.Run()
			if err != nil {
				return fmt.Sprintf("设置音量失败: %v", err)
			}
			return fmt.Sprintf("✅ 音量已设为 %s%%", volume)
		}
	}

	if strings.Contains(command, "增大") || strings.Contains(command, "加大") {
		cmd := exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", "+10%")
		cmd.Run()
		return "✅ 音量已增大"
	}

	if strings.Contains(command, "减小") || strings.Contains(command, "降低") {
		cmd := exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", "-10%")
		cmd.Run()
		return "✅ 音量已减小"
	}

	return "未识别的音量命令"
}

// 音乐控制
func (a *App) handleMusic(command string) string {
	if strings.Contains(command, "播放") {
		cmd := exec.Command("playerctl", "play")
		err := cmd.Run()
		if err != nil {
			return fmt.Sprintf("播放失败: %v", err)
		}
		return "✅ 已开始播放"
	}

	if strings.Contains(command, "暂停") {
		cmd := exec.Command("playerctl", "pause")
		err := cmd.Run()
		if err != nil {
			return fmt.Sprintf("暂停失败: %v", err)
		}
		return "✅ 已暂停播放"
	}

	if strings.Contains(command, "下一首") || strings.Contains(command, "下一曲") {
		cmd := exec.Command("playerctl", "next")
		err := cmd.Run()
		if err != nil {
			return fmt.Sprintf("切换失败: %v", err)
		}
		return "✅ 已切换到下一首"
	}

	if strings.Contains(command, "上一首") || strings.Contains(command, "上一曲") {
		cmd := exec.Command("playerctl", "previous")
		err := cmd.Run()
		if err != nil {
			return fmt.Sprintf("切换失败: %v", err)
		}
		return "✅ 已切换到上一首"
	}

	return "未识别的音乐控制命令"
}

// 文件操作
func (a *App) handleFile(command string) string {
	if strings.Contains(command, "创建文件") {
		// 提取文件名
		parts := strings.Split(command, "创建文件")
		if len(parts) > 1 {
			filename := strings.TrimSpace(parts[1])
			if filename != "" {
				cmd := exec.Command("touch", filename)
				err := cmd.Run()
				if err != nil {
					return fmt.Sprintf("创建文件失败: %v", err)
				}
				return fmt.Sprintf("✅ 已创建文件: %s", filename)
			}
		}
		return "请指定文件名"
	}

	return "未识别的文件操作命令"
}
