package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// AudioRecorder 音频录制器
type AudioRecorder struct {
	mu          sync.Mutex
	isRecording bool
	recordCmd   *exec.Cmd
	currentFile string
	lastFile    string
	waitCh      chan error
}

// NewAudioRecorder 创建录音器
func NewAudioRecorder() *AudioRecorder { return &AudioRecorder{} }

// StartRecording 开始录音
func (ar *AudioRecorder) StartRecording() (string, error) {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if ar.isRecording {
		return "", fmt.Errorf("已经在录音中")
	}

	timestamp := time.Now().Format("20060102_150405")
	filePath := filepath.Join(os.TempDir(), fmt.Sprintf("voice_%s.wav", timestamp))

	cmd := exec.Command("arecord", "-q", "-f", "cd", "-t", "wav", filePath)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("启动录音失败: %v", err)
	}

	ar.isRecording = true
	ar.recordCmd = cmd
	ar.currentFile = filePath
	ar.lastFile = filePath
	ar.waitCh = make(chan error, 1)

	go func(c *exec.Cmd, ch chan<- error) {
		err := c.Wait()
		ar.mu.Lock()
		ar.isRecording = false
		ar.recordCmd = nil
		ar.currentFile = ""
		ar.mu.Unlock()
		ch <- err
	}(cmd, ar.waitCh)

	return filePath, nil
}

// StopRecording 停止录音
func (ar *AudioRecorder) StopRecording() (string, error) {
	ar.mu.Lock()
	cmd := ar.recordCmd
	filePath := ar.lastFile
	waitCh := ar.waitCh
	if !ar.isRecording || cmd == nil || cmd.Process == nil {
		ar.mu.Unlock()
		return "", fmt.Errorf("当前没有在录音")
	}
	ar.mu.Unlock()

	// 尝试通过 SIGINT 优雅停止
	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		if killErr := cmd.Process.Kill(); killErr != nil {
			return "", fmt.Errorf("停止录音失败: %v", err)
		}
	}

	var waitErr error
	if waitCh != nil {
		waitErr = <-waitCh
		ar.mu.Lock()
		ar.waitCh = nil
		ar.mu.Unlock()
	} else {
		waitErr = cmd.Wait()
	}

	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signal() == syscall.SIGINT {
				waitErr = nil
			}
		}
	}

	if waitErr != nil {
		return "", fmt.Errorf("录音过程出错: %v", waitErr)
	}

	return filePath, nil
}

// IsRecording 检查是否正在录音
func (ar *AudioRecorder) IsRecording() bool {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	return ar.isRecording
}

// RecordAudio 录制音频（阻塞式，录制 5 秒）
func (ar *AudioRecorder) RecordAudio() (string, error) {
	if _, err := ar.StartRecording(); err != nil {
		return "", err
	}

	time.Sleep(5 * time.Second)

	return ar.StopRecording()
}

// GetRecordingStatus 获取录音状态
func (ar *AudioRecorder) GetRecordingStatus() string {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	if ar.isRecording {
		return "recording"
	}
	return "idle"
}

// CurrentFile 获取当前录音文件路径
func (ar *AudioRecorder) CurrentFile() string {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	return ar.currentFile
}

// LastFile 返回最近一次录音生成的文件路径
func (ar *AudioRecorder) LastFile() string {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	return ar.lastFile
}
