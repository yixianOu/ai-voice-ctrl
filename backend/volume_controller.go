package backend

import (
	"fmt"
	"os/exec"
	"strings"
)

// VolumeController 音量控制器
type VolumeController struct{}

// NewVolumeController 创建音量控制器
func NewVolumeController() *VolumeController {
	return &VolumeController{}
}

// ControlVolume 控制音量
func (vc *VolumeController) ControlVolume(command string) string {
	// 设置音量到指定值
	if strings.Contains(command, "设为") || strings.Contains(command, "调到") {
		volume := vc.extractVolume(command)
		if volume != "" {
			return vc.setVolume(volume)
		}
	}

	// 增大音量
	if strings.Contains(command, "增大") || strings.Contains(command, "加大") {
		return vc.adjustVolume("+10%")
	}

	// 减小音量
	if strings.Contains(command, "减小") || strings.Contains(command, "降低") {
		return vc.adjustVolume("-10%")
	}

	return "未识别的音量命令"
}

// setVolume 设置音量到指定值
func (vc *VolumeController) setVolume(volume string) string {
	cmd := exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", volume+"%")
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("设置音量失败: %v", err)
	}
	return fmt.Sprintf("✅ 音量已设为 %s%%", volume)
}

// adjustVolume 调整音量
func (vc *VolumeController) adjustVolume(delta string) string {
	cmd := exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", delta)
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("调整音量失败: %v", err)
	}

	if strings.HasPrefix(delta, "+") {
		return "✅ 音量已增大"
	}
	return "✅ 音量已减小"
}

// extractVolume 从命令中提取音量数字
func (vc *VolumeController) extractVolume(command string) string {
	parts := strings.Fields(command)
	for _, part := range parts {
		part = strings.TrimSuffix(part, "%")
		if len(part) > 0 && part[0] >= '0' && part[0] <= '9' {
			return part
		}
	}
	return ""
}
