package backend

import (
	"fmt"
	"strings"
)

// VoiceHandler 处理语音命令
type VoiceHandler struct {
	appController    *AppController
	volumeController *VolumeController
	musicController  *MusicController
	fileController   *FileController
}

// NewVoiceHandler 创建语音处理器
func NewVoiceHandler() *VoiceHandler {
	return &VoiceHandler{
		appController:    NewAppController(),
		volumeController: NewVolumeController(),
		musicController:  NewMusicController(),
		fileController:   NewFileController(),
	}
}

// ProcessCommand 处理语音命令
func (vh *VoiceHandler) ProcessCommand(command string) string {
	command = strings.TrimSpace(strings.ToLower(command))

	// 打开应用
	if strings.Contains(command, "打开") {
		return vh.appController.OpenApp(command)
	}

	// 音量控制
	if strings.Contains(command, "音量") {
		return vh.volumeController.ControlVolume(command)
	}

	// 音乐控制
	if strings.Contains(command, "音乐") || strings.Contains(command, "播放") ||
		strings.Contains(command, "暂停") || strings.Contains(command, "上一首") ||
		strings.Contains(command, "下一首") {
		return vh.musicController.ControlMusic(command)
	}

	// 文件操作
	if strings.Contains(command, "创建文件") || strings.Contains(command, "写入") {
		return vh.fileController.HandleFile(command)
	}

	return fmt.Sprintf("抱歉，我还不知道如何执行：%s", command)
}
