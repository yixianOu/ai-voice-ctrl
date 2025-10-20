package llms

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"ai-voice-ctrl/backend/llms"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

const (
	// Default model
	defaultModel = "whisper-1"
)

// OpenAIASR implements the ASRService interface using OpenAI Whisper API
type OpenAIASR struct {
	client openai.Client
	model  string
}

// TODO: Integrate into one configuration method
// NewOpenAIASR creates a new OpenAI ASR service instance
func NewOpenAIASR(apiKey string) *OpenAIASR {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &OpenAIASR{
		client: client,
		model:  defaultModel,
	}
}

// NewOpenAIASRWithProxy creates a new OpenAI ASR service with proxy support
func NewOpenAIASRWithProxy(apiKey string, proxyURL string) *OpenAIASR {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}

	// Add proxy if specified
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
			httpClient := &http.Client{Transport: transport}
			opts = append(opts, option.WithHTTPClient(httpClient))
		}
	}

	client := openai.NewClient(opts...)

	return &OpenAIASR{
		client: client,
		model:  defaultModel,
	}
}

// NewOpenAIASRWithConfig creates a new OpenAI ASR service with custom configuration
func NewOpenAIASRWithConfig(apiKey string, model string, opts ...option.RequestOption) *OpenAIASR {
	if model == "" {
		model = defaultModel
	}

	allOpts := append([]option.RequestOption{option.WithAPIKey(apiKey)}, opts...)
	client := openai.NewClient(allOpts...)

	return &OpenAIASR{
		client: client,
		model:  model,
	}
}

// TODO: Integrate into one configuration method
// Transcribe implements ASRService.Transcribe
func (o *OpenAIASR) Transcribe(ctx context.Context, audioData []byte) (string, error) {
	if len(audioData) == 0 {
		return "", &llms.ASRError{
			Code:    "INVALID_INPUT",
			Message: "audio data is empty",
		}
	}

	// Create transcription request
	// Don't specify ResponseFormat, let SDK use default
	transcription, err := o.client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		File:  openai.File(bytes.NewReader(audioData), "audio.wav", "audio/wav"),
		Model: openai.AudioModel(o.model),
	})

	if err != nil {
		return "", o.wrapError(err)
	}

	return transcription.Text, nil
}

// TranscribeWithOptions implements ASRService.TranscribeWithOptions
func (o *OpenAIASR) TranscribeWithOptions(ctx context.Context, audioData []byte, opts llms.TranscribeOptions) (llms.TranscribeResponse, error) {
	if len(audioData) == 0 {
		return llms.TranscribeResponse{}, &llms.ASRError{
			Code:    "INVALID_INPUT",
			Message: "audio data is empty",
		}
	}

	// Build transcription parameters
	// Use openai.File to properly set filename and content-type for format detection
	params := openai.AudioTranscriptionNewParams{
		File:  openai.File(bytes.NewReader(audioData), "audio.wav", "audio/wav"),
		Model: openai.AudioModel(o.model),
	}

	// Add optional parameters using field methods
	if opts.Language != "" {
		params.Language = openai.String(opts.Language)
	}
	if opts.Prompt != "" {
		params.Prompt = openai.String(opts.Prompt)
	}
	if opts.Temperature > 0 {
		params.Temperature = openai.Float(float64(opts.Temperature))
	}

	// Set response format only if explicitly specified
	// Default (empty) works best with the SDK
	if opts.ResponseFormat != "" {
		responseFormat := opts.ResponseFormat
		// Map string to enum type (handle case differences)
		switch string(responseFormat) {
		case "json":
			params.ResponseFormat = openai.AudioResponseFormatJSON
		case "text":
			params.ResponseFormat = openai.AudioResponseFormatText
		case "srt", "SRT":
			params.ResponseFormat = openai.AudioResponseFormatSRT
		case "vtt", "VTT":
			params.ResponseFormat = openai.AudioResponseFormatVTT
		case "verbose_json":
			// Map verbose_json to json for compatibility
			params.ResponseFormat = openai.AudioResponseFormatJSON
		}
	}
	// If ResponseFormat is not set, SDK will use default (which works best)

	// Add timestamp granularities if specified
	if len(opts.TimestampGranularities) > 0 {
		params.TimestampGranularities = opts.TimestampGranularities
	}

	// Call OpenAI API
	transcription, err := o.client.Audio.Transcriptions.New(ctx, params)
	if err != nil {
		return llms.TranscribeResponse{}, o.wrapError(err)
	}

	// Convert to our response format
	return o.convertResponse(transcription), nil
}

// convertResponse converts OpenAI transcription to our response format
func (o *OpenAIASR) convertResponse(trans *openai.Transcription) llms.TranscribeResponse {
	resp := llms.TranscribeResponse{
		Text:        trans.Text,
		RawResponse: trans,
	}

	// Note: The openai-go v2 library currently only returns Text field in the standard response
	// For more detailed information (segments, words, duration, language), you may need to use
	// the verbose_json format and parse the raw JSON, or wait for library updates

	return resp
}

// wrapError wraps OpenAI errors into ASRError
func (o *OpenAIASR) wrapError(err error) *llms.ASRError {
	if err == nil {
		return nil
	}

	// Try to extract OpenAI API error details
	if apiErr, ok := err.(*openai.Error); ok {
		return &llms.ASRError{
			Code:       apiErr.Code,
			Message:    fmt.Sprintf("OpenAI API error: %s", apiErr.Message),
			StatusCode: apiErr.StatusCode,
			Err:        err,
		}
	}

	// Generic error
	return &llms.ASRError{
		Code:    "REQUEST_FAILED",
		Message: "failed to transcribe audio",
		Err:     err,
	}
}

// GetServiceName implements ASRService.GetServiceName
func (o *OpenAIASR) GetServiceName() string {
	return "OpenAI Whisper"
}

// SetModel updates the model to use for transcription
func (o *OpenAIASR) SetModel(model string) {
	o.model = model
}
