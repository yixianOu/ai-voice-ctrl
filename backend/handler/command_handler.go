// Package handler implements the command handler coordinating voice/text input with LLM agent and tool execution.
package handler

import (
	"context"
	"fmt"

	"ai-voice-ctrl/backend/llms"
)

// CommandHandler coordinates voice/text input with LLM agent and tool execution
type CommandHandler struct {
	audioHandler *AudioHandler
	agent        llms.Agent
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(audioHandler *AudioHandler, agent llms.Agent) *CommandHandler {
	return &CommandHandler{
		audioHandler: audioHandler,
		agent:        agent,
	}
}

// ProcessVoiceCommand handles voice input: record -> transcribe -> LLM -> execute tools
func (h *CommandHandler) ProcessVoiceCommand(ctx context.Context) (string, error) {
	// Step 1: Record and transcribe audio
	text, err := h.audioHandler.RecordAndTranscribe(ctx)
	if err != nil {
		return "", fmt.Errorf("voice input failed: %w", err)
	}

	// Step 2: Process with LLM and execute tools
	response, err := h.agent.Chat(ctx, text)
	if err != nil {
		return "", fmt.Errorf("command processing failed: %w", err)
	}

	return response, nil
}

// ProcessTextCommand handles text input directly: LLM -> execute tools
func (h *CommandHandler) ProcessTextCommand(ctx context.Context, text string) (string, error) {
	response, err := h.agent.Chat(ctx, text)
	if err != nil {
		return "", fmt.Errorf("command processing failed: %w", err)
	}

	return response, nil
}

// ProcessVoiceCommandWithOptions handles voice input with transcription options
func (h *CommandHandler) ProcessVoiceCommandWithOptions(ctx context.Context, opts llms.TranscribeOptions) (string, error) {
	// Record and transcribe with options
	transcriptResp, err := h.audioHandler.RecordAndTranscribeWithOptions(ctx, opts)
	if err != nil {
		return "", fmt.Errorf("voice input failed: %w", err)
	}

	// Process with LLM
	response, err := h.agent.Chat(ctx, transcriptResp.Text)
	if err != nil {
		return "", fmt.Errorf("command processing failed: %w", err)
	}

	return response, nil
}

// ResetConversation clears conversation history
func (h *CommandHandler) ResetConversation() {
	h.agent.Reset()
}

// GetConversationHistory returns conversation history
func (h *CommandHandler) GetConversationHistory() []llms.Message {
	return h.agent.GetHistory()
}

// SetSystemPrompt updates the system prompt for LLM
func (h *CommandHandler) SetSystemPrompt(prompt string) {
	h.agent.SetSystemPrompt(prompt)
}
