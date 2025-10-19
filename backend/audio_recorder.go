package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// AudioRecorder 音频录制器
type AudioRecorder struct {
	isRecording bool
	recordCmd   *exec.Cmd
}

// NewAudioRecorder 创建录音器
func NewAudioRecorder() *AudioRecorder {
	return &AudioRecorder{
		isRecording: false,
	}
}

// StartRecording 开始录音
func (ar *AudioRecorder) StartRecording() (string, error) {
	if ar.isRecording {
		return "", fmt.Errorf("已经在录音中")
	}

	// 创建临时音频文件
	timestamp := time.Now().Format("20060102_150405")
	audioFile := filepath.Join(os.TempDir(), fmt.Sprintf("voice_%s.wav", timestamp))

	// 使用 arecord 录音 (Linux)
	// -d 5: 录制 5 秒
	// -f cd: CD 质量（44.1kHz, 16bit, 立体声）
	ar.recordCmd = exec.Command("arecord", "-d", "5", "-f", "cd", audioFile)

	err := ar.recordCmd.Start()
	if err != nil {
		return "", fmt.Errorf("启动录音失败: %v", err)
	}

	ar.isRecording = true
	return audioFile, nil
}

// StopRecording 停止录音
func (ar *AudioRecorder) StopRecording() error {
	if !ar.isRecording {
		return fmt.Errorf("当前没有在录音")
	}

	if ar.recordCmd != nil && ar.recordCmd.Process != nil {
		// 等待录音完成
		err := ar.recordCmd.Wait()
		ar.isRecording = false
		return err
	}

	ar.isRecording = false
	return nil
}

// IsRecording 检查是否正在录音
func (ar *AudioRecorder) IsRecording() bool {
	return ar.isRecording
}

// RecordAudio 录制音频（阻塞式，录制 5 秒）
func (ar *AudioRecorder) RecordAudio() (string, error) {
	audioFile, err := ar.StartRecording()
	if err != nil {
		return "", err
	}

	// 等待录音完成
	err = ar.StopRecording()
	if err != nil {
		return "", fmt.Errorf("录音过程出错: %v", err)
	}

	return audioFile, nil
}

// GetRecordingStatus 获取录音状态
func (ar *AudioRecorder) GetRecordingStatus() string {
	if ar.isRecording {
		return "recording"
	}
	return "idle"
}
