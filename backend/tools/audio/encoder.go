package tools

import (
	"bytes"
	"fmt"
)

const (
	SAMPLE_RATE       = 16000 // Whisper recommended sampling rate
	CHANNELS          = 1     // Mono audio
	SECONDS_TO_RECORD = 5     // Record for 5 seconds
)

// EncodeToWAV encodes raw PCM data in int16 format into byte slices in WAV format (exported version)
func EncodeToWAV(pcmData []int16, sampleRate, channels uint32) ([]byte, error) {
	return encodeToWAV(pcmData, sampleRate, channels)
}

// encodeToWAV Encode raw PCM data in int16 format into byte slices in WAV format
func encodeToWAV(pcmData []int16, sampleRate, channels uint32) ([]byte, error) {
	if len(pcmData) == 0 {
		return nil, fmt.Errorf("empty PCM data")
	}

	// Manually create WAV file to ensure correct format
	// WAV file structure:
	// - RIFF header (12 bytes)
	// - fmt chunk (24 bytes)
	// - data chunk (8 bytes + audio data)

	bitsPerSample := uint16(16)
	byteRate := sampleRate * uint32(channels) * uint32(bitsPerSample/8)
	blockAlign := uint16(channels) * bitsPerSample / 8
	dataSize := uint32(len(pcmData) * 2) // 2 bytes per int16 sample
	fileSize := 36 + dataSize            // Total file size minus 8 bytes

	buf := new(bytes.Buffer)

	// RIFF header
	buf.WriteString("RIFF")
	writeUint32(buf, fileSize)
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	writeUint32(buf, 16)               // fmt chunk size
	writeUint16(buf, 1)                // Audio format (1 = PCM)
	writeUint16(buf, uint16(channels)) // Number of channels
	writeUint32(buf, sampleRate)       // Sample rate
	writeUint32(buf, byteRate)         // Byte rate
	writeUint16(buf, blockAlign)       // Block align
	writeUint16(buf, bitsPerSample)    // Bits per sample

	// data chunk
	buf.WriteString("data")
	writeUint32(buf, dataSize)

	// Write PCM data (little-endian int16)
	for _, sample := range pcmData {
		writeInt16(buf, sample)
	}

	return buf.Bytes(), nil
}

// Helper functions to write little-endian values
func writeUint32(buf *bytes.Buffer, val uint32) {
	buf.WriteByte(byte(val))
	buf.WriteByte(byte(val >> 8))
	buf.WriteByte(byte(val >> 16))
	buf.WriteByte(byte(val >> 24))
}

func writeUint16(buf *bytes.Buffer, val uint16) {
	buf.WriteByte(byte(val))
	buf.WriteByte(byte(val >> 8))
}

func writeInt16(buf *bytes.Buffer, val int16) {
	buf.WriteByte(byte(val))
	buf.WriteByte(byte(val >> 8))
}
