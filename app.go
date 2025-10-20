package main

import (
	"context"
	"fmt"
	"os"

	"ai-voice-ctrl/backend/handler"
	"ai-voice-ctrl/backend/llms"
	openaillms "ai-voice-ctrl/backend/llms/openai"
)

// App struct
type App struct {
	ctx          context.Context
	audioHandler *handler.AudioHandler
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")

	return &App{
		audioHandler: handler.NewAudioHandler(apiKey),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	// Clean up resources
	if a.audioHandler != nil {
		a.audioHandler.Close()
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// ==================== Audio Recording Methods ====================

// StartRecording starts audio recording (manual mode)
func (a *App) StartRecording() error {
	return a.audioHandler.StartRecording()
}

// StopRecording stops audio recording and returns WAV data
func (a *App) StopRecording() ([]byte, error) {
	return a.audioHandler.StopRecording()
}

// RecordAudio records audio for specified seconds
func (a *App) RecordAudio(seconds int) ([]byte, error) {
	return a.audioHandler.RecordAudio(seconds)
}

// RecordAudioWithVAD records audio with VAD (auto-stops when user stops speaking)
func (a *App) RecordAudioWithVAD() ([]byte, error) {
	return a.audioHandler.RecordAudioWithVAD()
}

// ==================== Speech Recognition Methods ====================

// RecordAndTranscribe records audio with VAD and transcribes it
// This is the recommended method for simple voice input
func (a *App) RecordAndTranscribe() (string, error) {
	return a.audioHandler.RecordAndTranscribe(a.ctx)
}

// RecordAndTranscribeWithLanguage records and transcribes with specified language
func (a *App) RecordAndTranscribeWithLanguage(language string) (string, error) {
	opts := llms.TranscribeOptions{
		Language: language,
	}
	response, err := a.audioHandler.RecordAndTranscribeWithOptions(a.ctx, opts)
	if err != nil {
		return "", err
	}
	return response.Text, nil
}

// TranscribeAudioData transcribes pre-recorded audio data
func (a *App) TranscribeAudioData(audioData []byte) (string, error) {
	return a.audioHandler.TranscribeAudioData(a.ctx, audioData)
}

// ==================== Complete Workflow Methods ====================

// ExecuteVoiceCommandWorkflow executes complete voice command workflow
// Returns: transcript, commandResult, error
func (a *App) ExecuteVoiceCommandWorkflow() (string, string, error) {
	workflow, err := a.audioHandler.ExecuteVoiceCommandWorkflow(a.ctx)
	if err != nil {
		return "", "", err
	}
	return workflow.Transcript, workflow.CommandResult, nil
}

// ExecuteVoiceCommandWorkflowWithLanguage executes workflow with language specification
func (a *App) ExecuteVoiceCommandWorkflowWithLanguage(language string) (string, string, error) {
	opts := llms.TranscribeOptions{
		Language: language,
	}
	workflow, err := a.audioHandler.ExecuteVoiceCommandWorkflowWithOptions(a.ctx, opts)
	if err != nil {
		return "", "", err
	}
	return workflow.Transcript, workflow.CommandResult, nil
}

// RecordTranscribeAndProcess records, transcribes, and processes the command
// Legacy method - kept for backward compatibility
// Deprecated: Use ExecuteVoiceCommandWorkflow instead
func (a *App) RecordTranscribeAndProcess() (transcription string, result string, err error) {
	workflow, err := a.audioHandler.ExecuteVoiceCommandWorkflow(a.ctx)
	if err != nil {
		return "", "", err
	}
	return workflow.Transcript, workflow.CommandResult, nil
}

// ==================== Command Processing Methods ====================

// ProcessVoiceCommand processes voice recognition commands (placeholder)
// TODO: Will be implemented with LLM integration
func (a *App) ProcessVoiceCommand(command string) string {
	// This is currently a placeholder
	// Future: Integrate with LLM and function calling
	return fmt.Sprintf("Command received: '%s' (Processing not yet implemented. Will be integrated with LLM and function calling)", command)
}

// ==================== Status and State Methods ====================

// GetRecordingStatus gets current recording status
func (a *App) GetRecordingStatus() string {
	return a.audioHandler.GetRecordingStatus()
}

// IsRecording checks if currently recording
func (a *App) IsRecording() bool {
	return a.audioHandler.IsRecording()
}

// GetLastTranscript returns the last transcribed text
func (a *App) GetLastTranscript() string {
	return a.audioHandler.GetLastTranscript()
}

// GetASRServiceName returns the name of current ASR service
func (a *App) GetASRServiceName() string {
	return a.audioHandler.GetASRServiceName()
}

// ==================== Configuration Methods ====================

// SetAPIKey updates the OpenAI API key
func (a *App) SetAPIKey(apiKey string) {
	// Create new OpenAI ASR service with new API key
	newASR := openaillms.NewOpenAIASR(apiKey)
	a.audioHandler.SetASRService(newASR)
}
