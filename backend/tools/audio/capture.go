package tools

import (
	"fmt"
	"math"
	"sync"
	"time"
	"unsafe"

	"github.com/gen2brain/malgo"
)

// VADConfig Voice Activity Detection configuration
type VADConfig struct {
	SilenceThreshold  float64       // Audio volume threshold (0.0 - 1.0), typically 0.01-0.03
	SilenceDuration   time.Duration // Duration to consider silence (500ms - 1000ms recommended)
	MinSpeechDuration time.Duration // Minimum speech duration (300ms recommended)
	MaxRecordDuration time.Duration // Maximum recording duration (30s recommended)
	InitialDelay      time.Duration // Initial delay before VAD starts (grace period for user to start speaking)
}

// DefaultVADConfig returns default VAD configuration
func DefaultVADConfig() VADConfig {
	return VADConfig{
		SilenceThreshold:  0.01,                    // 1% volume threshold
		SilenceDuration:   700 * time.Millisecond,  // 700ms silence triggers stop
		MinSpeechDuration: 300 * time.Millisecond,  // At least 300ms of speech
		MaxRecordDuration: 30 * time.Second,        // Max 30 seconds
		InitialDelay:      1000 * time.Millisecond, // 1000ms grace period
	}
}

// AudioRecorder uses malgo for cross-platform audio recording with VAD support
type AudioRecorder struct {
	ctx           *malgo.AllocatedContext
	device        *malgo.Device
	isRecording   bool
	pcmData       []int16
	sampleRate    uint32
	channels      uint32
	bitsPerSample uint16

	// VAD related fields
	vadConfig          VADConfig
	vadEnabled         bool
	lastSoundTime      time.Time
	recordingStartTime time.Time
	silenceDetected    chan bool
	mu                 sync.Mutex
}

// NewAudioRecorder creates an audio recorder with VAD disabled
func NewAudioRecorder() *AudioRecorder {
	return &AudioRecorder{
		sampleRate:    SAMPLE_RATE,
		channels:      CHANNELS,
		bitsPerSample: 16,
		vadEnabled:    false,
	}
}

// NewAudioRecorderWithVAD creates an audio recorder with VAD enabled
func NewAudioRecorderWithVAD(config VADConfig) *AudioRecorder {
	return &AudioRecorder{
		sampleRate:      SAMPLE_RATE,
		channels:        CHANNELS,
		bitsPerSample:   16,
		vadEnabled:      true,
		vadConfig:       config,
		silenceDetected: make(chan bool, 1),
	}
}

// calculateRMS calculates Root Mean Square (volume level) of audio samples
func calculateRMS(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}

	var sum float64
	for _, sample := range samples {
		normalized := float64(sample) / 32768.0 // Normalize to -1.0 to 1.0
		sum += normalized * normalized
	}

	rms := math.Sqrt(sum / float64(len(samples)))
	return rms
}

// StartRecording starts audio recording
func (ar *AudioRecorder) StartRecording() error {
	if ar.isRecording {
		return fmt.Errorf("already recording")
	}

	if err := ar.ensureContext(); err != nil {
		return fmt.Errorf("failed to initialize audio context: %w", err)
	}

	// Reset PCM data buffer
	ar.pcmData = make([]int16, 0)

	// Initialize VAD state
	if ar.vadEnabled {
		ar.lastSoundTime = time.Now()
		ar.recordingStartTime = time.Now()
		ar.silenceDetected = make(chan bool, 1)

		// Debug: Print VAD configuration
		fmt.Printf("调试: VAD 配置\n")
		fmt.Printf("  准备时间: %.1f 秒\n", ar.vadConfig.InitialDelay.Seconds())
		fmt.Printf("  静音阈值: %.3f\n", ar.vadConfig.SilenceThreshold)
		fmt.Printf("  静音时长: %.1f 秒\n", ar.vadConfig.SilenceDuration.Seconds())
		fmt.Printf("  最小说话: %.1f 秒\n", ar.vadConfig.MinSpeechDuration.Seconds())
	}

	config := malgo.DefaultDeviceConfig(malgo.Capture)
	config.SampleRate = ar.sampleRate
	config.Capture.Channels = ar.channels
	config.Capture.Format = malgo.FormatS16

	callbacks := malgo.DeviceCallbacks{
		Data: func(outputSamples, inputSamples []byte, framecount uint32) {
			if len(inputSamples) == 0 {
				return
			}
			// Convert byte slice to int16 slice
			sampleCount := framecount * ar.channels
			samples := unsafe.Slice((*int16)(unsafe.Pointer(&inputSamples[0])), sampleCount)

			ar.mu.Lock()
			ar.pcmData = append(ar.pcmData, samples...)
			ar.mu.Unlock()

			// VAD: Check if user is speaking
			if ar.vadEnabled {
				ar.mu.Lock()
				recordingDuration := time.Since(ar.recordingStartTime)
				ar.mu.Unlock()

				rms := calculateRMS(samples)

				if rms > ar.vadConfig.SilenceThreshold {
					// Sound detected
					ar.mu.Lock()
					ar.lastSoundTime = time.Now()
					ar.mu.Unlock()
				} else {
					// Still in grace period - reset lastSoundTime to prevent premature stop
					if recordingDuration < ar.vadConfig.InitialDelay {
						ar.mu.Lock()
						ar.lastSoundTime = time.Now()
						ar.mu.Unlock()
					} else {
						// After grace period, check if silence duration exceeded
						ar.mu.Lock()
						silenceDuration := time.Since(ar.lastSoundTime)
						ar.mu.Unlock()

						// Stop if: 1) silence detected after minimum speech, OR 2) max duration reached
						if (silenceDuration >= ar.vadConfig.SilenceDuration &&
							recordingDuration >= ar.vadConfig.MinSpeechDuration) ||
							recordingDuration >= ar.vadConfig.MaxRecordDuration {
							select {
							case ar.silenceDetected <- true:
							default:
							}
						}
					}
				}
			}
		},
	}

	device, err := malgo.InitDevice(ar.ctx.Context, config, callbacks)
	if err != nil {
		return fmt.Errorf("failed to initialize recording device: %w", err)
	}

	if err := device.Start(); err != nil {
		device.Uninit()
		return fmt.Errorf("failed to start recording device: %w", err)
	}

	ar.device = device
	ar.isRecording = true

	return nil
}

// StopRecording stops audio recording and returns WAV data
func (ar *AudioRecorder) StopRecording() ([]byte, error) {
	if !ar.isRecording || ar.device == nil {
		return nil, fmt.Errorf("not currently recording")
	}

	device := ar.device
	ar.isRecording = false
	ar.device = nil

	if err := device.Stop(); err != nil {
		device.Uninit()
		return nil, fmt.Errorf("failed to stop recording: %w", err)
	}

	device.Uninit()

	// Copy PCM data
	ar.mu.Lock()
	pcmData := append([]int16(nil), ar.pcmData...)
	pcmDataLen := len(ar.pcmData)
	ar.pcmData = nil
	ar.mu.Unlock()

	if len(pcmData) == 0 {
		return nil, fmt.Errorf("no audio data captured")
	}

	// Calculate actual recording duration
	duration := float64(pcmDataLen) / float64(ar.sampleRate) / float64(ar.channels)
	fmt.Printf("调试: 捕获了 %d 个样本 (%.2f 秒)\n", pcmDataLen, duration)

	// Encode to WAV format
	wavData, err := EncodeToWAV(pcmData, ar.sampleRate, ar.channels)
	if err != nil {
		return nil, fmt.Errorf("failed to encode WAV: %w", err)
	}

	return wavData, nil
}

// RecordAudio records audio (blocking, records for specified duration)
func (ar *AudioRecorder) RecordAudio(duration time.Duration) ([]byte, error) {
	if err := ar.StartRecording(); err != nil {
		return nil, err
	}

	time.Sleep(duration)

	return ar.StopRecording()
}

// RecordAudioWithVAD records audio with Voice Activity Detection
// Automatically stops when user stops speaking
func (ar *AudioRecorder) RecordAudioWithVAD() ([]byte, error) {
	if !ar.vadEnabled {
		return nil, fmt.Errorf("VAD is not enabled, use NewAudioRecorderWithVAD() to create recorder")
	}

	if err := ar.StartRecording(); err != nil {
		return nil, err
	}

	// Wait for silence detection or timeout
	<-ar.silenceDetected

	return ar.StopRecording()
}

// IsRecording checks if currently recording
func (ar *AudioRecorder) IsRecording() bool {
	return ar.isRecording
}

// GetRecordingStatus gets recording status
func (ar *AudioRecorder) GetRecordingStatus() string {
	if ar.isRecording {
		return "recording"
	}
	return "idle"
}

// Close releases underlying resources
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
