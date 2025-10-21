package handler

import (
	"ai-voice-ctrl/backend/llms"
	openaillms "ai-voice-ctrl/backend/llms/openai"
	tools "ai-voice-ctrl/backend/tools/audio"
	"context"
	"fmt"
)

// AudioHandler coordinates audio capture and speech recognition
// It acts as a dispatcher/orchestrator between frontend requests,
// audio capture tools, and ASR services
type AudioHandler struct {
	// Audio capture
	recorder *tools.AudioRecorder

	// Speech recognition service
	asrService llms.ASRService

	// Configuration
	vadConfig tools.VADConfig

	// State management
	lastWavData    []byte
	lastTranscript string
}

// AudioHandlerConfig contains configuration for AudioHandler
type AudioHandlerConfig struct {
	// VAD configuration
	VADConfig tools.VADConfig

	// ASR service (if nil, will use OpenAI by default)
	ASRService llms.ASRService

	// OpenAI API key (used if ASRService is nil)
	OpenAIAPIKey string

	// Enable VAD by default
	EnableVAD bool
}

// DefaultAudioHandlerConfig returns default configuration
func DefaultAudioHandlerConfig(apiKey string) AudioHandlerConfig {
	return AudioHandlerConfig{
		VADConfig:    tools.DefaultVADConfig(),
		EnableVAD:    true,
		OpenAIAPIKey: apiKey,
	}
}

// TODO：integrate these two methods into one with options
// NewAudioHandler creates a new audio handler with default config
func NewAudioHandler(apiKey string) *AudioHandler {
	return NewAudioHandlerWithConfig(DefaultAudioHandlerConfig(apiKey))
}

// NewAudioHandlerWithConfig creates a new audio handler with custom config
func NewAudioHandlerWithConfig(config AudioHandlerConfig) *AudioHandler {
	// Initialize ASR service
	var asrService llms.ASRService
	if config.ASRService != nil {
		asrService = config.ASRService
	} else if config.OpenAIAPIKey != "" {
		// Default to OpenAI Whisper
		asrService = openaillms.NewOpenAIASR(config.OpenAIAPIKey)
	}

	// If VADConfig is not set (zero value), use default
	vadConfig := config.VADConfig
	if vadConfig.SilenceThreshold == 0 && vadConfig.SilenceDuration == 0 {
		// VADConfig is zero value, use default
		vadConfig = tools.DefaultVADConfig()
	}

	// Initialize audio recorder
	var recorder *tools.AudioRecorder
	if config.EnableVAD {
		recorder = tools.NewAudioRecorderWithVAD(vadConfig)
	} else {
		recorder = tools.NewAudioRecorder()
	}

	return &AudioHandler{
		recorder:   recorder,
		asrService: asrService,
		vadConfig:  vadConfig,
	}
}

// ==================== Audio Capture Methods ====================
// These methods delegate to the audio recorder

// StartRecording starts audio recording (manual mode)
func (ah *AudioHandler) StartRecording() error {
	return ah.recorder.StartRecording()
}

// StopRecording stops audio recording and returns WAV data (manual mode)
func (ah *AudioHandler) StopRecording() ([]byte, error) {
	wavData, err := ah.recorder.StopRecording()
	if err != nil {
		return nil, err
	}
	ah.lastWavData = wavData
	return wavData, nil
}

// RecordAudio records audio for specified duration (blocking)
// Use this when you need a fixed-duration recording
// func (ah *AudioHandler) RecordAudio(seconds int) ([]byte, error) {
// 	duration := time.Duration(seconds) * time.Second
// 	wavData, err := ah.recorder.RecordAudio(duration)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to record audio: %w", err)
// 	}
// 	ah.lastWavData = wavData
// 	return wavData, nil
// }

// RecordAudioWithVAD records audio with VAD (blocking until silence detected)
// Use this for automatic speech detection
func (ah *AudioHandler) recordAudioWithVAD() ([]byte, error) {
	wavData, err := ah.recorder.RecordAudioWithVAD()
	if err != nil {
		return nil, fmt.Errorf("failed to record audio with VAD: %w", err)
	}
	ah.lastWavData = wavData
	return wavData, nil
}

// ==================== Speech Recognition Methods ====================
// These methods coordinate between audio capture and ASR service

// TranscribeAudioData transcribes pre-recorded audio data
// Use this when you already have audio data (e.g., from file or network)

// TODO: integrate these two methods into one with options
func (ah *AudioHandler) transcribeAudioData(ctx context.Context, audioData []byte) (string, error) {
	if ah.asrService == nil {
		return "", fmt.Errorf("ASR service not configured")
	}

	text, err := ah.asrService.Transcribe(ctx, audioData)
	if err != nil {
		return "", fmt.Errorf("ASR transcription failed: %w", err)
	}

	ah.lastTranscript = text
	return text, nil
}

// TranscribeAudioDataWithOptions transcribes audio data with custom options
func (ah *AudioHandler) transcribeAudioDataWithOptions(ctx context.Context, audioData []byte, opts llms.TranscribeOptions) (llms.TranscribeResponse, error) {
	if ah.asrService == nil {
		return llms.TranscribeResponse{}, fmt.Errorf("ASR service not configured")
	}

	response, err := ah.asrService.TranscribeWithOptions(ctx, audioData, opts)
	if err != nil {
		return llms.TranscribeResponse{}, fmt.Errorf("ASR transcription failed: %w", err)
	}

	ah.lastTranscript = response.Text
	return response, nil
}

// RecordAndTranscribe records audio with VAD and transcribes it
// This is the main method for voice input with automatic speech detection
func (ah *AudioHandler) RecordAndTranscribe(ctx context.Context) (string, error) {
	// Step 1: Record audio with VAD (auto-stops when user stops speaking)
	wavData, err := ah.recordAudioWithVAD()
	if err != nil {
		return "", err
	}

	// Step 2: Transcribe using ASR service
	text, err := ah.transcribeAudioData(ctx, wavData)
	if err != nil {
		return "", err
	}

	return text, nil
}

// RecordAndTranscribeWithOptions records and transcribes with custom options
func (ah *AudioHandler) RecordAndTranscribeWithOptions(ctx context.Context, opts llms.TranscribeOptions) (llms.TranscribeResponse, error) {
	// Step 1: Record audio with VAD
	wavData, err := ah.recordAudioWithVAD()
	if err != nil {
		return llms.TranscribeResponse{}, err
	}

	// Step 2: Transcribe with options
	response, err := ah.transcribeAudioDataWithOptions(ctx, wavData, opts)
	if err != nil {
		return llms.TranscribeResponse{}, err
	}

	return response, nil
}

// ==================== High-Level Workflow Methods ====================
// These methods represent complete workflows for different use cases

// VoiceCommandWorkflow represents a complete voice command workflow result
// type VoiceCommandWorkflow struct {
// 	// Audio data captured
// 	AudioData []byte

// 	// Transcribed text from ASR
// 	Transcript string

// 	// Detailed transcription response (if available)
// 	TranscriptDetails *llms.TranscribeResponse

// 	// Command processing result (for future implementation)
// 	CommandResult string

// 	// Timestamps
// 	RecordingStartTime time.Time
// 	RecordingEndTime   time.Time
// 	TranscriptionTime  time.Time
// }

// ExecuteVoiceCommandWorkflow executes a complete voice command workflow
// 1. Record audio with VAD
// 2. Transcribe audio to text
// 3. (Future) Process command through LLM with function calling
// Returns detailed workflow result
// func (ah *AudioHandler) ExecuteVoiceCommandWorkflow(ctx context.Context) (*VoiceCommandWorkflow, error) {
// 	result := &VoiceCommandWorkflow{
// 		RecordingStartTime: time.Now(),
// 	}

// 	// Step 1: Record audio
// 	wavData, err := ah.RecordAudioWithVAD()
// 	if err != nil {
// 		return nil, fmt.Errorf("recording failed: %w", err)
// 	}
// 	result.AudioData = wavData
// 	result.RecordingEndTime = time.Now()

// 	// Step 2: Transcribe
// 	text, err := ah.TranscribeAudioData(ctx, wavData)
// 	if err != nil {
// 		return nil, fmt.Errorf("transcription failed: %w", err)
// 	}
// 	result.Transcript = text
// 	result.TranscriptionTime = time.Now()

// 	// Step 3: (Future) Process command with LLM
// 	// This will be implemented later with prompts and function calling
// 	result.CommandResult = ah.processCommandPlaceholder(text)

// 	return result, nil
// }

// ExecuteVoiceCommandWorkflowWithOptions executes workflow with custom options
// func (ah *AudioHandler) ExecuteVoiceCommandWorkflowWithOptions(ctx context.Context, opts llms.TranscribeOptions) (*VoiceCommandWorkflow, error) {
// 	result := &VoiceCommandWorkflow{
// 		RecordingStartTime: time.Now(),
// 	}

// 	// Step 1: Record audio
// 	wavData, err := ah.RecordAudioWithVAD()
// 	if err != nil {
// 		return nil, fmt.Errorf("recording failed: %w", err)
// 	}
// 	result.AudioData = wavData
// 	result.RecordingEndTime = time.Now()

// 	// Step 2: Transcribe with options
// 	transcriptResp, err := ah.TranscribeAudioDataWithOptions(ctx, wavData, opts)
// 	if err != nil {
// 		return nil, fmt.Errorf("transcription failed: %w", err)
// 	}
// 	result.Transcript = transcriptResp.Text
// 	result.TranscriptDetails = &transcriptResp
// 	result.TranscriptionTime = time.Now()

// 	// Step 3: (Future) Process command
// 	result.CommandResult = ah.processCommandPlaceholder(transcriptResp.Text)

// 	return result, nil
// }

// ==================== State Management Methods ====================

// IsRecording checks if currently recording
func (ah *AudioHandler) IsRecording() bool {
	return ah.recorder.IsRecording()
}

// GetRecordingStatus gets current recording status
// func (ah *AudioHandler) GetRecordingStatus() string {
// 	return ah.recorder.GetRecordingStatus()
// }

// GetLastWavData returns the last recorded audio data
// func (ah *AudioHandler) GetLastWavData() []byte {
// 	return ah.lastWavData
// }

// GetLastTranscript returns the last transcribed text
// func (ah *AudioHandler) GetLastTranscript() string {
// 	return ah.lastTranscript
// }

// ==================== Configuration Methods ====================

// SetASRService updates the ASR service
// Use this to switch between different ASR providers at runtime
func (ah *AudioHandler) SetASRService(service llms.ASRService) {
	ah.asrService = service
}

// GetASRServiceName returns the name of current ASR service
func (ah *AudioHandler) GetASRServiceName() string {
	if ah.asrService == nil {
		return "None"
	}
	return ah.asrService.GetServiceName()
}

// UpdateVADConfig updates the VAD configuration
// Note: This only affects future recordings
func (ah *AudioHandler) UpdateVADConfig(config tools.VADConfig) error {
	ah.vadConfig = config
	// Reinitialize recorder with new config
	if ah.recorder != nil {
		ah.recorder.Close()
	}
	ah.recorder = tools.NewAudioRecorderWithVAD(config)
	return nil
}

// ==================== Resource Management ====================

// Close releases all resources
func (ah *AudioHandler) Close() error {
	if ah.recorder != nil {
		ah.recorder.Close()
	}
	return nil
}
