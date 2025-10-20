package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const (
	SAMPLE_RATE       = 16000 // Whisper recommended sampling rate
	CHANNELS          = 1     // Mono audio
	SECONDS_TO_RECORD = 5     // Record for 5 seconds
)

// bufferSeeker wraps bytes.Buffer to implement io.WriteSeeker
type bufferSeeker struct {
	buf *bytes.Buffer
	pos int64
}

func newBufferSeeker() *bufferSeeker {
	return &bufferSeeker{
		buf: new(bytes.Buffer),
		pos: 0,
	}
}

func (bs *bufferSeeker) Write(p []byte) (n int, err error) {
	n, err = bs.buf.Write(p)
	bs.pos += int64(n)
	return
}

func (bs *bufferSeeker) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart: // From the beginning of the file
		abs = offset
	case io.SeekCurrent: // From the current position
		abs = bs.pos + offset
	case io.SeekEnd: // From the end of the file
		abs = int64(bs.buf.Len()) + offset
	default:
		return 0, fmt.Errorf("bufferSeeker.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, fmt.Errorf("bufferSeeker.Seek: negative position")
	}
	bs.pos = abs
	return abs, nil
}

func (bs *bufferSeeker) Bytes() []byte {
	return bs.buf.Bytes()
}

// encodeToWAV Encode raw PCM data in int16 format into byte slices in WAV format
func encodeToWAV(pcmData []int16) ([]byte, error) {
	// Step 1: Create a seekable buffer
	buf := newBufferSeeker()

	// Step 2:Creating a WAV Encoder
	// Parameters: writer, sample rate, bit depth, channels, audio format (1 for PCM)
	encoder := wav.NewEncoder(buf, SAMPLE_RATE, 16, CHANNELS, 1)

	// Step 3: Creating an audio.IntBuffer to adapt to the encoder
	audioBuf := &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: CHANNELS,
			SampleRate:  SAMPLE_RATE,
		},
		Data:           make([]int, len(pcmData)),
		SourceBitDepth: 16,
	}

	// Step 4: Convert int16 to int (required by the go-audio library)
	for i, v := range pcmData {
		audioBuf.Data[i] = int(v)
	}

	// Step 5: Write data and close the encoder
	if err := encoder.Write(audioBuf); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// transcribe 函数负责将 WAV 音频数据发送给 OpenAI Whisper API
func transcribe(wavData []byte, apiKey string) (string, error) {
	apiURL := "https://api.openai.com/v1/audio/transcriptions"

	// 创建一个 buffer 来存储 multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 创建一个名为 'file' 的 form-data part
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("创建 form file 失败: %w", err)
	}
	// 将 WAV 数据写入 part
	if _, err := io.Copy(part, bytes.NewReader(wavData)); err != nil {
		return "", fmt.Errorf("写入文件数据失败: %w", err)
	}

	// 添加 'model' 字段
	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return "", fmt.Errorf("写入 model 字段失败: %w", err)
	}

	// 关闭 multipart writer, 这会写入结尾的 boundary
	if err := writer.Close(); err != nil {
		return "", err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 检查是否有 API 错误
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 解析 JSON 响应
	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析 JSON 响应失败: %w", err)
	}

	return result.Text, nil
}
