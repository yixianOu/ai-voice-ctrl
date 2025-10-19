package backend

import (
	"fmt"
	"os/exec"
	"strings"
)

// MusicController 音乐控制器
type MusicController struct{}

// NewMusicController 创建音乐控制器
func NewMusicController() *MusicController {
	return &MusicController{}
}

// ControlMusic 控制音乐播放
func (mc *MusicController) ControlMusic(command string) string {
	// 播放
	if strings.Contains(command, "播放") {
		return mc.play()
	}

	// 暂停
	if strings.Contains(command, "暂停") {
		return mc.pause()
	}

	// 下一首
	if strings.Contains(command, "下一首") || strings.Contains(command, "下一曲") {
		return mc.next()
	}

	// 上一首
	if strings.Contains(command, "上一首") || strings.Contains(command, "上一曲") {
		return mc.previous()
	}

	return "未识别的音乐控制命令"
}

// play 播放音乐
func (mc *MusicController) play() string {
	cmd := exec.Command("playerctl", "play")
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("播放失败: %v", err)
	}
	return "✅ 已开始播放"
}

// pause 暂停音乐
func (mc *MusicController) pause() string {
	cmd := exec.Command("playerctl", "pause")
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("暂停失败: %v", err)
	}
	return "✅ 已暂停播放"
}

// next 下一首
func (mc *MusicController) next() string {
	cmd := exec.Command("playerctl", "next")
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("切换失败: %v", err)
	}
	return "✅ 已切换到下一首"
}

// previous 上一首
func (mc *MusicController) previous() string {
	cmd := exec.Command("playerctl", "previous")
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("切换失败: %v", err)
	}
	return "✅ 已切换到上一首"
}

// GetStatus 获取播放状态
func (mc *MusicController) GetStatus() (string, error) {
	cmd := exec.Command("playerctl", "status")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCurrentSong 获取当前歌曲信息
func (mc *MusicController) GetCurrentSong() (string, error) {
	cmd := exec.Command("playerctl", "metadata", "--format", "{{ artist }} - {{ title }}")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
