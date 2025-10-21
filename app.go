package main

import (
	"context"
	"fmt"
	"os"

	"ai-voice-ctrl/backend/executor"
	"ai-voice-ctrl/backend/executor/browser"
	"ai-voice-ctrl/backend/executor/spotify"
	"ai-voice-ctrl/backend/executor/vscode"
	"ai-voice-ctrl/backend/handler"
	"ai-voice-ctrl/backend/llms"
	openaillms "ai-voice-ctrl/backend/llms/openai"
)

// App struct
type App struct {
	ctx            context.Context
	audioHandler   *handler.AudioHandler
	commandHandler *handler.CommandHandler
	executor       executor.Executor
	agent          llms.Agent
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Get API key from environment variable
	apiKey := os.Getenv("OPENAI_API_KEY")

	// Initialize executor and register tools
	exec := executor.NewDefaultExecutor()
	if err := vscode.RegisterVSCodeLifecycle(exec); err != nil {
		fmt.Printf("Warning: Failed to register VSCode tools: %v\n", err)
	}
	if err := spotify.RegisterSpotifyTool(exec); err != nil {
		fmt.Printf("Warning: Failed to register Spotify tools: %v\n", err)
	}
	if err := browser.RegisterBrowserTool(exec); err != nil {
		fmt.Printf("Warning: Failed to register Browser tools: %v\n", err)
	}

	// Initialize agent with executor
	agent := openaillms.NewOpenAIAgent(apiKey, exec)

	// Initialize audio handler
	audioHandler := handler.NewAudioHandler(apiKey)

	// Initialize command handler
	commandHandler := handler.NewCommandHandler(audioHandler, agent)

	return &App{
		audioHandler:   audioHandler,
		commandHandler: commandHandler,
		executor:       exec,
		agent:          agent,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	// Clean up audio handler resources
	if a.audioHandler != nil {
		a.audioHandler.Close()
	}

	// Clean up executor instances (VSCode, Spotify, etc.)
	if a.executor != nil {
		// Try to close VSCode instance if exists
		if _, ok := a.executor.LoadInstance("vscode"); ok {
			destroyCall := executor.ToolCall{
				Name:      "destroy_vscode",
				Arguments: []byte(`{}`),
			}
			a.executor.ExecuteTool(ctx, destroyCall)
		}

		// Try to stop Spotify server if exists
		if _, ok := a.executor.LoadInstance("spotify_server"); ok {
			stopCall := executor.ToolCall{
				Name:      "spotify_stop_server",
				Arguments: []byte(`{}`),
			}
			a.executor.ExecuteTool(ctx, stopCall)
		}
	}
}

// ==================== LLM Command Processing Methods ====================

// ProcessTextCommand processes text command with LLM and tool execution
// func (a *App) ProcessTextCommand(text string) (string, error) {
// 	return a.commandHandler.ProcessTextCommand(a.ctx, text)
// }

// ProcessVoiceCommandAuto records audio with VAD, transcribes, and processes with LLM
func (a *App) ProcessVoiceCommandAuto() (string, error) {
	return a.commandHandler.ProcessVoiceCommand(a.ctx)
}

// ResetConversation clears conversation history
func (a *App) ResetConversation() {
	a.commandHandler.ResetConversation()
}

// GetConversationHistory returns conversation history as JSON string
func (a *App) GetConversationHistory() []llms.Message {
	return a.commandHandler.GetConversationHistory()
}

// SetSystemPrompt updates the system prompt for LLM
// func (a *App) SetSystemPrompt(prompt string) {
// 	a.commandHandler.SetSystemPrompt(prompt)
// }

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
// func (a *App) TranscribeAudioData(audioData []byte) (string, error) {
// 	return a.audioHandler.TranscribeAudioData(a.ctx, audioData)
// }

// ==================== Complete Workflow Methods ====================

// ExecuteVoiceCommandWorkflow executes complete voice command workflow
// Returns: transcript, commandResult, error
// func (a *App) ExecuteVoiceCommandWorkflow() (string, string, error) {
// 	workflow, err := a.audioHandler.ExecuteVoiceCommandWorkflow(a.ctx)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	return workflow.Transcript, workflow.CommandResult, nil
// }

// ExecuteVoiceCommandWorkflowWithLanguage executes workflow with language specification
// func (a *App) ExecuteVoiceCommandWorkflowWithLanguage(language string) (string, string, error) {
// 	opts := llms.TranscribeOptions{
// 		Language: language,
// 	}
// 	workflow, err := a.audioHandler.ExecuteVoiceCommandWorkflowWithOptions(a.ctx, opts)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	return workflow.Transcript, workflow.CommandResult, nil
// }

// RecordTranscribeAndProcess records, transcribes, and processes the command
// Legacy method - kept for backward compatibility
// Deprecated: Use ExecuteVoiceCommandWorkflow instead
// func (a *App) RecordTranscribeAndProcess() (transcription string, result string, err error) {
// 	workflow, err := a.audioHandler.ExecuteVoiceCommandWorkflow(a.ctx)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	return workflow.Transcript, workflow.CommandResult, nil
// }

// ==================== Command Processing Methods ====================

// ProcessVoiceCommand processes voice recognition commands (placeholder)
// Deprecated: Use ProcessTextCommand with LLM integration instead
func (a *App) ProcessVoiceCommand(command string) string {
	return fmt.Sprintf("Command received: '%s' (Use ProcessTextCommand for LLM integration)", command)
}

// ==================== Status and State Methods ====================

// GetRecordingStatus gets current recording status
// func (a *App) GetRecordingStatus() string {
// 	return a.audioHandler.GetRecordingStatus()
// }

// IsRecording checks if currently recording
func (a *App) IsRecording() bool {
	return a.audioHandler.IsRecording()
}

// GetLastTranscript returns the last transcribed text
func (a *App) GetLastTranscript() string {
	return a.audioHandler.GetLastTranscript()
}

// GetASRServiceName returns the name of current ASR service
// func (a *App) GetASRServiceName() string {
// 	return a.audioHandler.GetASRServiceName()
// }

// ==================== Configuration Methods ====================

// SetAPIKey updates the OpenAI API key
func (a *App) SetAPIKey(apiKey string) {
	// Update ASR service
	newASR := openaillms.NewOpenAIASR(apiKey)
	a.audioHandler.SetASRService(newASR)

	// Recreate agent with new API key
	a.agent = openaillms.NewOpenAIAgent(apiKey, a.executor)
	a.commandHandler = handler.NewCommandHandler(a.audioHandler, a.agent)
}

// ==================== Tool Management Methods ====================

// GetAvailableTools returns list of available tools
func (a *App) GetAvailableTools() []string {
	schemas := a.executor.ToolSchemas()
	tools := make([]string, 0, len(schemas))
	for _, schema := range schemas {
		tools = append(tools, schema.Name)
	}
	return tools
}

// GetToolDescription returns description for a specific tool
func (a *App) GetToolDescription(toolName string) string {
	schemas := a.executor.ToolSchemas()
	for _, schema := range schemas {
		if schema.Name == toolName {
			return schema.Description
		}
	}
	return ""
}
