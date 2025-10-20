package main

import (
	"ai-voice-ctrl/backend/handler"
	"ai-voice-ctrl/backend/llms"
	openaillms "ai-voice-ctrl/backend/llms/openai"
	tools "ai-voice-ctrl/backend/tools/audio"
	"context"
	"fmt"
	"os"
	"time"
)

// getAPIKey retrieves and validates the OpenAI API key from environment
func getAPIKey() (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set.\n" +
			"Please set it using:\n" +
			"  Linux/Mac: export OPENAI_API_KEY=your-api-key\n" +
			"  Windows (PowerShell): $env:OPENAI_API_KEY=\"your-api-key\"\n" +
			"  Windows (CMD): set OPENAI_API_KEY=your-api-key")
	}
	return apiKey, nil
}

// getProxyURL retrieves proxy URL from environment (optional)
func getProxyURL() string {
	// Check multiple proxy environment variables
	proxyURL := os.Getenv("HTTP_PROXY")
	if proxyURL == "" {
		proxyURL = os.Getenv("http_proxy")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("HTTPS_PROXY")
	}
	if proxyURL == "" {
		proxyURL = os.Getenv("https_proxy")
	}
	return proxyURL
}

// Example 1: Basic voice input with default configuration
func Example_BasicVoiceInput() {
	apiKey, err := getAPIKey()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Create handler with proxy support if configured
	var audioHandler *handler.AudioHandler
	proxyURL := getProxyURL()
	if proxyURL != "" {
		fmt.Printf("Using proxy: %s\n", proxyURL)
		asrService := openaillms.NewOpenAIASRWithProxy(apiKey, proxyURL)
		config := handler.AudioHandlerConfig{
			ASRService: asrService,
			EnableVAD:  true,
		}
		audioHandler = handler.NewAudioHandlerWithConfig(config)
	} else {
		audioHandler = handler.NewAudioHandler(apiKey)
	}
	defer audioHandler.Close()

	ctx := context.Background()

	// Record and transcribe with one call
	fmt.Println("Start speaking...")
	fmt.Println("Tip: You have 1 second to start speaking (grace period)")
	fmt.Println("     Recording will auto-stop when you finish speaking")
	text, err := audioHandler.RecordAndTranscribe(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("You said: %s\n", text)
}

// Example 2: Voice input with language specification
func Example_VoiceInputWithLanguage() {
	apiKey, err := getAPIKey()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	// Create handler with proxy support if configured
	var audioHandler *handler.AudioHandler
	proxyURL := getProxyURL()
	if proxyURL != "" {
		fmt.Printf("使用代理: %s\n", proxyURL)
		asrService := openaillms.NewOpenAIASRWithProxy(apiKey, proxyURL)
		config := handler.AudioHandlerConfig{
			ASRService: asrService,
			EnableVAD:  true,
		}
		audioHandler = handler.NewAudioHandlerWithConfig(config)
	} else {
		audioHandler = handler.NewAudioHandler(apiKey)
	}
	defer audioHandler.Close()

	ctx := context.Background()

	// Specify Chinese language for better accuracy
	opts := llms.TranscribeOptions{
		Language:    "zh",
		Temperature: 0.2,
		// Don't set ResponseFormat, use SDK default
	}

	fmt.Println("开始说话...")
	fmt.Println("提示：您有 1 秒的准备时间（录音不会立即停止）")
	fmt.Println("     说完后保持安静 0.7 秒自动结束")
	fmt.Println("提示：如果录音立即停止，可能是环境太安静，请尝试大声说话")
	response, err := audioHandler.RecordAndTranscribeWithOptions(ctx, opts)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		fmt.Println("\n建议:")
		fmt.Println("1. 确保麦克风权限已授予")
		fmt.Println("2. 检查麦克风是否正常工作")
		fmt.Println("3. 尝试大声说话（环境可能太安静）")
		fmt.Println("4. 可以运行 Example_FixedDurationRecording() 测试固定时长录音")
		return
	}

	fmt.Printf("识别结果: %s\n", response.Text)
	fmt.Printf("语言: %s\n", response.Language)
	fmt.Printf("时长: %.2f 秒\n", response.Duration)
}

// Example 2b: Voice input with fixed duration (recommended for testing)
func Example_VoiceInputFixedDuration() {
	apiKey, err := getAPIKey()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	// Create handler with proxy support if configured
	var audioHandler *handler.AudioHandler
	proxyURL := getProxyURL()
	if proxyURL != "" {
		fmt.Printf("使用代理: %s\n", proxyURL)
		asrService := openaillms.NewOpenAIASRWithProxy(apiKey, proxyURL)
		config := handler.AudioHandlerConfig{
			ASRService: asrService,
			EnableVAD:  false, // Disable VAD for fixed duration recording
		}
		audioHandler = handler.NewAudioHandlerWithConfig(config)
	} else {
		// Create without VAD
		config := handler.AudioHandlerConfig{
			OpenAIAPIKey: apiKey,
			EnableVAD:    false,
		}
		audioHandler = handler.NewAudioHandlerWithConfig(config)
	}
	defer audioHandler.Close()

	ctx := context.Background()

	// Record for 5 seconds
	fmt.Println("开始录音 5 秒，请说话...")
	audioData, err := audioHandler.RecordAudio(5)
	if err != nil {
		fmt.Printf("录音错误: %v\n", err)
		return
	}

	fmt.Printf("录音完成，音频大小: %d 字节\n", len(audioData))

	// Save WAV file for debugging
	debugFile := "debug_recording.wav"
	if err := os.WriteFile(debugFile, audioData, 0644); err != nil {
		fmt.Printf("警告: 无法保存调试文件: %v\n", err)
	} else {
		fmt.Printf("调试: 已保存音频到 %s\n", debugFile)
	}

	// Check WAV header
	if len(audioData) >= 44 {
		fmt.Printf("调试: WAV 头信息:\n")
		fmt.Printf("  RIFF: %s\n", string(audioData[0:4]))
		fmt.Printf("  文件大小: %d\n", uint32(audioData[4])|uint32(audioData[5])<<8|uint32(audioData[6])<<16|uint32(audioData[7])<<24)
		fmt.Printf("  WAVE: %s\n", string(audioData[8:12]))
		fmt.Printf("  fmt : %s\n", string(audioData[12:16]))
		fmt.Printf("  音频格式: %d\n", uint16(audioData[20])|uint16(audioData[21])<<8)
		fmt.Printf("  声道数: %d\n", uint16(audioData[22])|uint16(audioData[23])<<8)
		sampleRate := uint32(audioData[24]) | uint32(audioData[25])<<8 | uint32(audioData[26])<<16 | uint32(audioData[27])<<24
		fmt.Printf("  采样率: %d Hz\n", sampleRate)
		fmt.Printf("  位深度: %d bits\n", uint16(audioData[34])|uint16(audioData[35])<<8)

		// Calculate duration
		dataSize := uint32(audioData[40]) | uint32(audioData[41])<<8 | uint32(audioData[42])<<16 | uint32(audioData[43])<<24
		duration := float64(dataSize) / float64(sampleRate) / 2.0 // 2 bytes per sample (16-bit)
		fmt.Printf("  数据大小: %d 字节\n", dataSize)
		fmt.Printf("  计算时长: %.2f 秒\n", duration)
	}

	// Transcribe with Chinese language
	opts := llms.TranscribeOptions{
		Language:    "zh",
		Temperature: 0.2,
		// Don't set ResponseFormat, use SDK default for better compatibility
	}

	fmt.Println("正在识别...")
	response, err := audioHandler.TranscribeAudioDataWithOptions(ctx, audioData, opts)
	if err != nil {
		fmt.Printf("识别错误: %v\n", err)
		fmt.Println("\n请检查上面的 WAV 头信息，并尝试用音频播放器打开 debug_recording.wav 测试是否正常")
		return
	}

	fmt.Printf("识别结果: %s\n", response.Text)
	fmt.Printf("语言: %s\n", response.Language)
	fmt.Printf("时长: %.2f 秒\n", response.Duration)
}

/* we don't need to use these examples */

// Example 3: Complete workflow with detailed results
func Example_CompleteWorkflow() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	audioHandler := handler.NewAudioHandler(apiKey)
	defer audioHandler.Close()

	ctx := context.Background()

	fmt.Println("Start speaking...")
	workflow, err := audioHandler.ExecuteVoiceCommandWorkflow(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print detailed results
	fmt.Printf("\n=== Voice Command Workflow Results ===\n")
	fmt.Printf("Recording Duration: %v\n", workflow.RecordingEndTime.Sub(workflow.RecordingStartTime))
	fmt.Printf("Transcription Time: %v\n", workflow.TranscriptionTime.Sub(workflow.RecordingEndTime))
	fmt.Printf("Audio Size: %d bytes\n", len(workflow.AudioData))
	fmt.Printf("Transcript: %s\n", workflow.Transcript)
	fmt.Printf("Command Result: %s\n", workflow.CommandResult)
}

// Example 4: Manual recording control
func Example_ManualRecording() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	audioHandler := handler.NewAudioHandler(apiKey)
	defer audioHandler.Close()

	ctx := context.Background()

	// Start recording manually
	fmt.Println("Recording started (manual mode)...")
	err := audioHandler.StartRecording()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Record for 5 seconds
	time.Sleep(5 * time.Second)

	// Stop recording
	audioData, err := audioHandler.StopRecording()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Recorded %d bytes\n", len(audioData))

	// Transcribe the recorded audio
	text, err := audioHandler.TranscribeAudioData(ctx, audioData)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Transcript: %s\n", text)
}

// Example 5: Fixed-duration recording
func Example_FixedDurationRecording() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	audioHandler := handler.NewAudioHandler(apiKey)
	defer audioHandler.Close()

	ctx := context.Background()

	// Record for exactly 3 seconds
	fmt.Println("Recording for 3 seconds...")
	audioData, err := audioHandler.RecordAudio(3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Transcribe
	text, err := audioHandler.TranscribeAudioData(ctx, audioData)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Transcript: %s\n", text)
}

// Example 6: Transcribe audio from file
func Example_TranscribeFromFile() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	audioHandler := handler.NewAudioHandler(apiKey)
	defer audioHandler.Close()

	ctx := context.Background()

	// Read audio file
	audioData, err := os.ReadFile("sample_audio.wav")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	// Transcribe
	text, err := audioHandler.TranscribeAudioData(ctx, audioData)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Transcript: %s\n", text)
}

// Example 7: Custom VAD configuration
func Example_CustomVADConfiguration() {
	apiKey := os.Getenv("OPENAI_API_KEY")

	// Create custom VAD config for quieter environment
	customVADConfig := tools.VADConfig{
		SilenceThreshold:  0.01, // Lower threshold for quiet environments
		SilenceDuration:   500 * time.Millisecond,
		MinSpeechDuration: 200 * time.Millisecond,
		MaxRecordDuration: 20 * time.Second,
	}

	config := handler.AudioHandlerConfig{
		VADConfig:    customVADConfig,
		EnableVAD:    true,
		OpenAIAPIKey: apiKey,
	}

	audioHandler := handler.NewAudioHandlerWithConfig(config)
	defer audioHandler.Close()

	ctx := context.Background()

	fmt.Println("Start speaking (custom VAD settings)...")
	text, err := audioHandler.RecordAndTranscribe(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Transcript: %s\n", text)
}

// Example 8: Switch ASR service at runtime
func Example_SwitchASRService() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	audioHandler := handler.NewAudioHandler(apiKey)
	defer audioHandler.Close()

	ctx := context.Background()

	// Use default OpenAI Whisper
	fmt.Printf("Current ASR service: %s\n", audioHandler.GetASRServiceName())

	// Record audio
	audioData, err := audioHandler.RecordAudioWithVAD()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Switch to a different ASR service (example)
	// Note: You would implement CustomASR to match llms.ASRService interface
	// customASR := NewCustomASRService()
	// audioHandler.SetASRService(customASR)

	// Or use different OpenAI configuration
	newASR := openaillms.NewOpenAIASRWithConfig(apiKey, "whisper-1")
	audioHandler.SetASRService(newASR)
	fmt.Printf("Switched to: %s\n", audioHandler.GetASRServiceName())

	// Transcribe with new service
	text, err := audioHandler.TranscribeAudioData(ctx, audioData)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Transcript: %s\n", text)
}

// Example 9: Check recording status
func Example_CheckRecordingStatus() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	audioHandler := handler.NewAudioHandler(apiKey)
	defer audioHandler.Close()

	// Check initial status
	fmt.Printf("Is recording? %v\n", audioHandler.IsRecording())
	fmt.Printf("Status: %s\n", audioHandler.GetRecordingStatus())

	// Start recording in background
	go func() {
		audioHandler.StartRecording()
		time.Sleep(3 * time.Second)
		audioHandler.StopRecording()
	}()

	// Monitor status
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		fmt.Printf("[%ds] Status: %s, Recording: %v\n",
			i+1,
			audioHandler.GetRecordingStatus(),
			audioHandler.IsRecording())
	}
}

// Example 10: Access last recording data
func Example_AccessLastRecording() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	audioHandler := handler.NewAudioHandler(apiKey)
	defer audioHandler.Close()

	ctx := context.Background()

	// Record and transcribe
	text, err := audioHandler.RecordAndTranscribe(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Transcript: %s\n", text)

	// Access the audio data that was just recorded
	lastAudio := audioHandler.GetLastWavData()
	lastTranscript := audioHandler.GetLastTranscript()

	fmt.Printf("Last audio size: %d bytes\n", len(lastAudio))
	fmt.Printf("Last transcript: %s\n", lastTranscript)

	// Save to file
	err = os.WriteFile("last_recording.wav", lastAudio, 0644)
	if err != nil {
		fmt.Printf("Error saving file: %v\n", err)
		return
	}

	fmt.Println("Audio saved to last_recording.wav")
}

// Example 11: Update VAD configuration at runtime
func Example_UpdateVADAtRuntime() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	audioHandler := handler.NewAudioHandler(apiKey)
	defer audioHandler.Close()

	ctx := context.Background()

	// First recording with default VAD
	fmt.Println("Recording with default VAD...")
	text1, _ := audioHandler.RecordAndTranscribe(ctx)
	fmt.Printf("Result 1: %s\n", text1)

	// Update VAD for more sensitive detection
	newVADConfig := tools.VADConfig{
		SilenceThreshold:  0.015,
		SilenceDuration:   500 * time.Millisecond,
		MinSpeechDuration: 250 * time.Millisecond,
		MaxRecordDuration: 25 * time.Second,
	}

	err := audioHandler.UpdateVADConfig(newVADConfig)
	if err != nil {
		fmt.Printf("Error updating VAD: %v\n", err)
		return
	}

	// Second recording with updated VAD
	fmt.Println("Recording with updated VAD...")
	text2, _ := audioHandler.RecordAndTranscribe(ctx)
	fmt.Printf("Result 2: %s\n", text2)
}

// Example 12: Workflow with detailed transcription
func Example_WorkflowWithDetails() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	audioHandler := handler.NewAudioHandler(apiKey)
	defer audioHandler.Close()

	ctx := context.Background()

	// Request detailed transcription with segments
	opts := llms.TranscribeOptions{
		Language:               "en",
		ResponseFormat:         "json", // Use "json" instead of "verbose_json"
		TimestampGranularities: []string{"segment", "word"},
	}

	fmt.Println("Start speaking...")
	workflow, err := audioHandler.ExecuteVoiceCommandWorkflowWithOptions(ctx, opts)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\n=== Detailed Transcription Results ===\n")
	fmt.Printf("Full Text: %s\n", workflow.Transcript)

	if workflow.TranscriptDetails != nil {
		fmt.Printf("Language: %s\n", workflow.TranscriptDetails.Language)
		fmt.Printf("Duration: %.2f seconds\n", workflow.TranscriptDetails.Duration)

		// Print segments
		if len(workflow.TranscriptDetails.Segments) > 0 {
			fmt.Printf("\nSegments:\n")
			for i, seg := range workflow.TranscriptDetails.Segments {
				fmt.Printf("  [%d] %.2f-%.2f: %s (confidence: %.2f%%)\n",
					i+1, seg.Start, seg.End, seg.Text, seg.Confidence*100)
			}
		}

		// Print words
		if len(workflow.TranscriptDetails.Words) > 0 {
			fmt.Printf("\nWords:\n")
			for i, word := range workflow.TranscriptDetails.Words {
				fmt.Printf("  [%d] %.2f-%.2f: %s\n",
					i+1, word.Start, word.End, word.Word)
			}
		}
	}
}

func main() {
	fmt.Println("=== AI Voice Control 示例 ===")
	fmt.Println()

	// Check API key first
	_, err := getAPIKey()
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		fmt.Println()
		fmt.Println("设置完成后，请重新运行此程序。")
		return
	}

	fmt.Println("✅ API Key 已配置")

	// Check proxy
	proxyURL := getProxyURL()
	if proxyURL != "" {
		fmt.Printf("✅ 代理已配置: %s\n", proxyURL)
	} else {
		fmt.Println("⚠️  未配置代理（国内用户可能需要）")
	}
	fmt.Println()

	// Run the fixed-duration example (more reliable for testing)
	Example_BasicVoiceInput()
}
