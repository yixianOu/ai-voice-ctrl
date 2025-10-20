package llms

import (
	"context"

	"github.com/openai/openai-go/v2"
)

// ASRService defines the interface for Automatic Speech Recognition services
type ASRService interface {
	// Transcribe converts audio data to text
	// audioData: WAV format audio bytes
	// Returns the transcribed text or error
	Transcribe(ctx context.Context, audioData []byte) (string, error)

	// TranscribeWithOptions converts audio data to text with additional options
	TranscribeWithOptions(ctx context.Context, audioData []byte, opts TranscribeOptions) (TranscribeResponse, error)

	// GetServiceName returns the name of the ASR service
	GetServiceName() string
}

// TranscribeOptions contains optional parameters for transcription
type TranscribeOptions struct {
	// Language specifies the language of the audio (ISO-639-1 format, e.g., "en", "zh")
	// Leave empty for automatic detection
	Language string

	// Prompt provides context to improve transcription accuracy
	Prompt string

	// Temperature controls randomness (0.0 - 1.0)
	// Higher values make output more random, lower values more deterministic
	Temperature float32

	// ResponseFormat specifies the format of the response
	// Options: "json", "text", "srt", "vtt", "verbose_json"
	ResponseFormat openai.AudioResponseFormat

	// TimestampGranularities specifies the timestamp granularities to populate
	// Options: "word", "segment"
	TimestampGranularities []string
}

// TranscribeResponse contains the detailed transcription result
type TranscribeResponse struct {
	// Text is the transcribed text
	Text string

	// Language is the detected or specified language
	Language string

	// Duration is the audio duration in seconds
	Duration float64

	// Segments contains word-level or sentence-level segments (if available)
	Segments []Segment

	// Words contains word-level timestamps (if requested)
	Words []Word

	// RawResponse contains the raw response from the API
	RawResponse *openai.Transcription
}

// Segment represents a time-aligned text segment
type Segment struct {
	// ID is the segment identifier
	ID int

	// Text is the segment text
	Text string

	// Start is the start time in seconds
	Start float64

	// End is the end time in seconds
	End float64

	// Confidence is the confidence score (0.0 - 1.0)
	Confidence float64

	// Tokens are the token IDs for this segment
	Tokens []int

	// Temperature is the sampling temperature used for this segment
	Temperature float64

	// AvgLogprob is the average log probability of the tokens
	AvgLogprob float64

	// CompressionRatio is the compression ratio of this segment
	CompressionRatio float64

	// NoSpeechProb is the probability of no speech in this segment
	NoSpeechProb float64
}

// Word represents a word-level timestamp
type Word struct {
	// Word is the word text
	Word string

	// Start is the start time in seconds
	Start float64

	// End is the end time in seconds
	End float64
}

// ASRError represents an error from the ASR service
type ASRError struct {
	// Code is the error code
	Code string

	// Message is the error message
	Message string

	// StatusCode is the HTTP status code (if applicable)
	StatusCode int

	// Err is the underlying error
	Err error
}

func (e *ASRError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *ASRError) Unwrap() error {
	return e.Err
}
