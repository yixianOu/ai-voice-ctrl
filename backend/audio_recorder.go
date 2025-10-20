package backend

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gen2brain/malgo"
)

// AudioRecorder 使用 malgo 进行跨平台录音
type AudioRecorder struct {
	ctx           *malgo.AllocatedContext
	device        *malgo.Device
	isRecording   bool
	dataBuffer    bytes.Buffer
	currentFile   string
	lastFile      string
	sampleRate    uint32
	channels      uint32
	bitsPerSample uint16
}

// NewAudioRecorder 创建录音器
func NewAudioRecorder() *AudioRecorder {
	return &AudioRecorder{
		sampleRate:    44100,
		channels:      1,
		bitsPerSample: 16,
	}
}

// StartRecording 开始录音
func (ar *AudioRecorder) StartRecording() (string, error) {
	if ar.isRecording {
		return "", fmt.Errorf("已经在录音中")
	}

	if err := ar.ensureContext(); err != nil {
		return "", fmt.Errorf("初始化音频上下文失败: %w", err)
	}

	ar.dataBuffer.Reset()

	timestamp := time.Now().Format("20060102_150405")
	filePath := filepath.Join(os.TempDir(), fmt.Sprintf("voice_%s.wav", timestamp))

	config := malgo.DefaultDeviceConfig(malgo.Capture)
	config.SampleRate = ar.sampleRate
	config.Capture.Channels = ar.channels
	config.Capture.Format = malgo.FormatS16

	callbacks := malgo.DeviceCallbacks{
		Data: func(outputSamples, inputSamples []byte, framecount uint32) {
			if len(inputSamples) == 0 {
				return
			}
			ar.dataBuffer.Write(inputSamples)
		},
	}

	device, err := malgo.InitDevice(ar.ctx.Context, config, callbacks)
	if err != nil {
		return "", fmt.Errorf("初始化录音设备失败: %w", err)
	}

	if err := device.Start(); err != nil {
		device.Uninit()
		return "", fmt.Errorf("启动录音设备失败: %w", err)
	}

	ar.device = device
	ar.isRecording = true
	ar.currentFile = filePath
	ar.lastFile = filePath

	return filePath, nil
}

// StopRecording 停止录音
func (ar *AudioRecorder) StopRecording() (string, error) {
	if !ar.isRecording || ar.device == nil {
		return "", fmt.Errorf("当前没有在录音")
	}
	device := ar.device
	filePath := ar.currentFile
	ar.isRecording = false
	ar.device = nil
	ar.currentFile = ""

	if err := device.Stop(); err != nil {
		device.Uninit()
		return "", fmt.Errorf("停止录音失败: %w", err)
	}

	device.Uninit()

	data := append([]byte(nil), ar.dataBuffer.Bytes()...)
	ar.dataBuffer.Reset()

	if len(data) == 0 {
		return "", fmt.Errorf("未捕获到音频数据")
	}

	if err := savePCMAsWAV(filePath, data, ar.sampleRate, ar.channels, ar.bitsPerSample); err != nil {
		return "", err
	}

	ar.lastFile = filePath

	return filePath, nil
}

// RecordAudio 录制音频（阻塞式，录制 5 秒）
func (ar *AudioRecorder) RecordAudio() (string, error) {
	if _, err := ar.StartRecording(); err != nil {
		return "", err
	}

	time.Sleep(5 * time.Second)

	return ar.StopRecording()
}

// IsRecording 检查是否正在录音
func (ar *AudioRecorder) IsRecording() bool {
	return ar.isRecording
}

// GetRecordingStatus 获取录音状态
func (ar *AudioRecorder) GetRecordingStatus() string {
	if ar.isRecording {
		return "recording"
	}
	return "idle"
}

// CurrentFile 获取当前录音文件路径
func (ar *AudioRecorder) CurrentFile() string {
	return ar.currentFile
}

// LastFile 返回最近一次录音生成的文件路径
func (ar *AudioRecorder) LastFile() string {
	return ar.lastFile
}

// Close 释放底层资源
func (ar *AudioRecorder) Close() {
	if ar.device != nil {
		ar.device.Stop()
		ar.device.Uninit()
		ar.device = nil
	}
	if ar.ctx != nil {
		ar.ctx.Uninit()
		ar.ctx = nil
	}
}

func (ar *AudioRecorder) ensureContext() error {
	if ar.ctx != nil {
		return nil
	}
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return err
	}
	ar.ctx = ctx
	return nil
}

func savePCMAsWAV(path string, data []byte, sampleRate, channels uint32, bitsPerSample uint16) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建音频文件失败: %w", err)
	}
	defer f.Close()

	dataSize := uint32(len(data))
	byteRate := sampleRate * channels * uint32(bitsPerSample) / 8
	blockAlign := uint16(channels * uint32(bitsPerSample) / 8)

	if _, err := f.Write([]byte("RIFF")); err != nil {
		return fmt.Errorf("写入 RIFF 头失败: %w", err)
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(36+dataSize)); err != nil {
		return fmt.Errorf("写入 RIFF 大小失败: %w", err)
	}
	if _, err := f.Write([]byte("WAVE")); err != nil {
		return fmt.Errorf("写入 WAVE 头失败: %w", err)
	}
	if _, err := f.Write([]byte("fmt ")); err != nil {
		return fmt.Errorf("写入 fmt 标记失败: %w", err)
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(16)); err != nil {
		return fmt.Errorf("写入 fmt 长度失败: %w", err)
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(1)); err != nil {
		return fmt.Errorf("写入音频格式失败: %w", err)
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(channels)); err != nil {
		return fmt.Errorf("写入声道数失败: %w", err)
	}
	if err := binary.Write(f, binary.LittleEndian, sampleRate); err != nil {
		return fmt.Errorf("写入采样率失败: %w", err)
	}
	if err := binary.Write(f, binary.LittleEndian, byteRate); err != nil {
		return fmt.Errorf("写入字节率失败: %w", err)
	}
	if err := binary.Write(f, binary.LittleEndian, blockAlign); err != nil {
		return fmt.Errorf("写入块对齐失败: %w", err)
	}
	if err := binary.Write(f, binary.LittleEndian, bitsPerSample); err != nil {
		return fmt.Errorf("写入位深失败: %w", err)
	}
	if _, err := f.Write([]byte("data")); err != nil {
		return fmt.Errorf("写入 data 标记失败: %w", err)
	}
	if err := binary.Write(f, binary.LittleEndian, dataSize); err != nil {
		return fmt.Errorf("写入数据长度失败: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("写入音频数据失败: %w", err)
	}

	return nil
}
