package backend

import (
	"fmt"
	"strings"
)

// VoiceHandler 处理语音命令
type VoiceHandler struct {
	audioRecorder *AudioRecorder
}

// NewVoiceHandler 创建语音处理器
func NewVoiceHandler() *VoiceHandler {
	return &VoiceHandler{
		audioRecorder: NewAudioRecorder(),
	}
}

// ProcessCommand 处理语音命令
func (vh *VoiceHandler) ProcessCommand(command string) string {
	command = strings.TrimSpace(strings.ToLower(command))

	return fmt.Sprintf("抱歉，我还不知道如何执行：%s", command)
}

// StartRecording 开始录音
func (vh *VoiceHandler) StartRecording() (string, error) {
	return vh.audioRecorder.StartRecording()
}

// StopRecording 停止录音
func (vh *VoiceHandler) StopRecording() error {
	return vh.audioRecorder.StopRecording()
}

// GetRecordingStatus 获取录音状态
func (vh *VoiceHandler) GetRecordingStatus() string {
	return vh.audioRecorder.GetRecordingStatus()
}
