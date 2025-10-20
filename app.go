package main

import (
	"context"
	"fmt"

	"ai-voice-ctrl/backend"
)

// App struct
type App struct {
	ctx          context.Context
	voiceHandler *backend.VoiceHandler
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		voiceHandler: backend.NewVoiceHandler(),
	}
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
	return a.voiceHandler.ProcessCommand(command)
}

// StartRecording 开始录音
func (a *App) StartRecording() (string, error) { return a.voiceHandler.StartRecording() }

// StopRecording 停止录音
func (a *App) StopRecording() (string, error) { return a.voiceHandler.StopRecording() }

// GetRecordingStatus 获取录音状态
func (a *App) GetRecordingStatus() string {
	return a.voiceHandler.GetRecordingStatus()
}
