# Audio Handler 快速上手指南

## 🚀 5 分钟快速开始

### 1. 最简单的使用方式

```go
package main

import (
    "ai-voice-ctrl/backend/handler"
    "context"
    "fmt"
    "os"
)

func main() {
    // 获取 API Key
    apiKey := os.Getenv("OPENAI_API_KEY")
    
    // 创建 handler
    audioHandler := handler.NewAudioHandler(apiKey)
    defer audioHandler.Close()
    
    // 录制并转录（一行搞定！）
    ctx := context.Background()
    fmt.Println("请说话...")
    text, err := audioHandler.RecordAndTranscribe(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("你说: %s\n", text)
}
```

就这么简单！🎉

---

## 📖 常见使用场景

### 场景 1: 中文语音输入

```go
// 指定中文，提高识别准确度
opts := llms.TranscribeOptions{
    Language: "zh",
    Temperature: 0.2,
}

fmt.Println("请说话...")
response, err := audioHandler.RecordAndTranscribeWithOptions(ctx, opts)
fmt.Printf("识别结果: %s\n", response.Text)
```

### 场景 2: 获取详细的转录信息

```go
opts := llms.TranscribeOptions{
    Language: "zh",
    ResponseFormat: "verbose_json",
    TimestampGranularities: []string{"segment"},
}

response, err := audioHandler.RecordAndTranscribeWithOptions(ctx, opts)

// 查看详细信息
fmt.Printf("语言: %s\n", response.Language)
fmt.Printf("时长: %.2f 秒\n", response.Duration)
for i, seg := range response.Segments {
    fmt.Printf("[%d] %.2f-%.2f: %s\n", i+1, seg.Start, seg.End, seg.Text)
}
```

### 场景 3: 从文件转录

```go
// 读取音频文件
audioData, err := os.ReadFile("recording.wav")
if err != nil {
    panic(err)
}

// 转录
text, err := audioHandler.TranscribeAudioData(ctx, audioData)
fmt.Printf("文件内容: %s\n", text)
```

### 场景 4: 固定时长录制

```go
// 录制 5 秒钟
fmt.Println("录制 5 秒...")
audioData, err := audioHandler.RecordAudio(5)

// 转录
text, err := audioHandler.TranscribeAudioData(ctx, audioData)
fmt.Printf("结果: %s\n", text)
```

### 场景 5: 手动控制录制

```go
// 开始录制
fmt.Println("开始录制...")
audioHandler.StartRecording()

// 录制一段时间
time.Sleep(3 * time.Second)

// 停止并获取数据
audioData, err := audioHandler.StopRecording()
fmt.Printf("录制了 %d 字节\n", len(audioData))

// 转录
text, err := audioHandler.TranscribeAudioData(ctx, audioData)
fmt.Printf("结果: %s\n", text)
```

---

## ⚙️ 高级配置

### 自定义 VAD 参数

```go
import tools "ai-voice-ctrl/backend/tools/audio"

// 创建自定义 VAD 配置
customVAD := tools.VADConfig{
    SilenceThreshold:  0.01,  // 降低阈值（安静环境）
    SilenceDuration:   500 * time.Millisecond,
    MinSpeechDuration: 200 * time.Millisecond,
    MaxRecordDuration: 30 * time.Second,
}

// 使用自定义配置创建 handler
config := handler.AudioHandlerConfig{
    VADConfig:    customVAD,
    OpenAIAPIKey: apiKey,
    EnableVAD:    true,
}
audioHandler := handler.NewAudioHandlerWithConfig(config)
```

### 运行时更新 VAD

```go
// 创建新的 VAD 配置
newVAD := tools.VADConfig{
    SilenceThreshold: 0.015,
    // ... 其他参数
}

// 更新配置
err := audioHandler.UpdateVADConfig(newVAD)
if err != nil {
    fmt.Printf("更新失败: %v\n", err)
}
```

### 切换 ASR 服务

```go
import openaillms "ai-voice-ctrl/backend/llms/openai"

// 创建新的 ASR 服务
newASR := openaillms.NewOpenAIASR("new-api-key")

// 切换服务
audioHandler.SetASRService(newASR)

fmt.Printf("当前服务: %s\n", audioHandler.GetASRServiceName())
```

---

## 🔍 状态查询

### 检查录制状态

```go
if audioHandler.IsRecording() {
    fmt.Println("正在录制中...")
}

status := audioHandler.GetRecordingStatus()
fmt.Printf("状态: %s\n", status)
```

### 获取历史数据

```go
// 获取上次录制的音频
lastAudio := audioHandler.GetLastWavData()
fmt.Printf("上次录制: %d 字节\n", len(lastAudio))

// 获取上次的转录结果
lastText := audioHandler.GetLastTranscript()
fmt.Printf("上次转录: %s\n", lastText)

// 保存到文件
os.WriteFile("last_recording.wav", lastAudio, 0644)
```

---

## 🎯 完整工作流

### 使用工作流获取详细信息

```go
fmt.Println("开始录制...")
workflow, err := audioHandler.ExecuteVoiceCommandWorkflow(ctx)
if err != nil {
    panic(err)
}

// 打印详细结果
fmt.Printf("\n=== 工作流结果 ===\n")
fmt.Printf("录制时长: %v\n", workflow.RecordingEndTime.Sub(workflow.RecordingStartTime))
fmt.Printf("转录耗时: %v\n", workflow.TranscriptionTime.Sub(workflow.RecordingEndTime))
fmt.Printf("音频大小: %d 字节\n", len(workflow.AudioData))
fmt.Printf("转录结果: %s\n", workflow.Transcript)

// 保存音频
os.WriteFile("workflow_audio.wav", workflow.AudioData, 0644)
```

---

## 🛠️ 错误处理

### 基础错误处理

```go
text, err := audioHandler.RecordAndTranscribe(ctx)
if err != nil {
    // 检查错误类型
    if asrErr, ok := err.(*llms.ASRError); ok {
        fmt.Printf("ASR 错误: %s (代码: %s)\n", asrErr.Message, asrErr.Code)
    } else {
        fmt.Printf("其他错误: %v\n", err)
    }
    return
}
```

### 带超时的录制

```go
// 创建带超时的 context
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

text, err := audioHandler.RecordAndTranscribe(ctx)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        fmt.Println("超时了！")
    } else {
        fmt.Printf("错误: %v\n", err)
    }
}
```

### 可取消的录制

```go
ctx, cancel := context.WithCancel(context.Background())

// 在另一个 goroutine 中取消
go func() {
    time.Sleep(5 * time.Second)
    fmt.Println("取消录制...")
    cancel()
}()

text, err := audioHandler.RecordAndTranscribe(ctx)
if err != nil {
    if ctx.Err() == context.Canceled {
        fmt.Println("已取消")
    }
}
```

---

## 📦 Wails 集成示例

### 在 app.go 中使用

```go
// app.go
package main

import (
    "ai-voice-ctrl/backend/handler"
    "context"
    "os"
)

type App struct {
    ctx          context.Context
    audioHandler *handler.AudioHandler
}

func NewApp() *App {
    apiKey := os.Getenv("OPENAI_API_KEY")
    return &App{
        audioHandler: handler.NewAudioHandler(apiKey),
    }
}

// Wails 方法: 录制并转录
func (a *App) RecordAndTranscribe() (string, error) {
    return a.audioHandler.RecordAndTranscribe(a.ctx)
}

// Wails 方法: 带语言选项的录制
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

// Wails 方法: 检查状态
func (a *App) IsRecording() bool {
    return a.audioHandler.IsRecording()
}

func (a *App) GetRecordingStatus() string {
    return a.audioHandler.GetRecordingStatus()
}

// 清理资源
func (a *App) shutdown(ctx context.Context) {
    a.audioHandler.Close()
}
```

### 前端调用 (React/TypeScript)

```typescript
// Frontend code
import { RecordAndTranscribe, IsRecording } from '../wailsjs/go/main/App';

function VoiceInput() {
    const [isRecording, setIsRecording] = useState(false);
    const [transcript, setTranscript] = useState('');

    const handleRecord = async () => {
        try {
            setIsRecording(true);
            const text = await RecordAndTranscribe();
            setTranscript(text);
        } catch (error) {
            console.error('录制失败:', error);
        } finally {
            setIsRecording(false);
        }
    };

    return (
        <div>
            <button onClick={handleRecord} disabled={isRecording}>
                {isRecording ? '录制中...' : '开始录制'}
            </button>
            {transcript && <p>识别结果: {transcript}</p>}
        </div>
    );
}
```

---

## 🎨 最佳实践

### ✅ DO - 推荐做法

```go
// ✅ 使用 context
ctx := context.Background()
text, err := handler.RecordAndTranscribe(ctx)

// ✅ 及时关闭资源
defer handler.Close()

// ✅ 错误处理
if err != nil {
    log.Printf("错误: %v", err)
    return
}

// ✅ 指定语言提高准确度
opts := llms.TranscribeOptions{
    Language: "zh",
}
```

### ❌ DON'T - 避免的做法

```go
// ❌ 忘记关闭资源
handler := handler.NewAudioHandler(apiKey)
// 忘记调用 handler.Close()

// ❌ 忽略错误
text, _ := handler.RecordAndTranscribe(ctx)  // 不要忽略错误

// ❌ 不使用 context
// handler.RecordAndTranscribe()  // 旧 API，已废弃

// ❌ 在循环中创建 handler
for i := 0; i < 10; i++ {
    h := handler.NewAudioHandler(apiKey)  // 资源泄漏！
    // ...
}
```

---

## 🐛 常见问题

### Q1: 录制时没有声音？

```go
// 检查录制状态
fmt.Println("状态:", audioHandler.GetRecordingStatus())

// 检查 VAD 阈值是否太高
config := tools.VADConfig{
    SilenceThreshold: 0.01,  // 降低阈值
    // ...
}
audioHandler.UpdateVADConfig(config)
```

### Q2: 识别准确度低？

```go
// 指定语言
opts := llms.TranscribeOptions{
    Language: "zh",        // 明确指定语言
    Temperature: 0.2,      // 降低温度，更确定性
}
```

### Q3: 录制时间太短或太长？

```go
// 调整 VAD 参数
config := tools.VADConfig{
    SilenceDuration:   800 * time.Millisecond,  // 增加静音检测时长
    MinSpeechDuration: 500 * time.Millisecond,  // 增加最小语音时长
    MaxRecordDuration: 60 * time.Second,        // 增加最大录制时长
}
```

### Q4: 如何测试而不真实录制？

```go
// 使用 mock ASR 服务
type MockASR struct{}

func (m *MockASR) Transcribe(ctx context.Context, data []byte) (string, error) {
    return "mock transcript", nil
}

func (m *MockASR) TranscribeWithOptions(ctx context.Context, data []byte, opts llms.TranscribeOptions) (llms.TranscribeResponse, error) {
    return llms.TranscribeResponse{Text: "mock transcript"}, nil
}

func (m *MockASR) GetServiceName() string {
    return "Mock ASR"
}

// 注入 mock
audioHandler.SetASRService(&MockASR{})
```

---

## 📚 更多资源

- [完整架构设计](./AUDIO_HANDLER_ARCHITECTURE.md) - 架构图和设计理念
- [重构总结](./AUDIO_HANDLER_REFACTOR.md) - 重构前后对比
- [详细设计文档](./AUDIO_HANDLER_DESIGN.md) - 深入的设计说明
- [使用示例](./examples/audio_handler_usage.go) - 12 个完整示例

---

## 💡 小贴士

1. **性能优化**: 使用 `RecordAndTranscribe` 而不是分步调用，可以减少一次数据复制
2. **资源管理**: 始终使用 `defer handler.Close()` 确保资源释放
3. **错误处理**: 检查 `*llms.ASRError` 类型获取详细的 API 错误信息
4. **调试**: 使用 `GetLastWavData()` 保存音频文件进行调试
5. **多语言**: 始终指定 `Language` 参数以提高准确度

Happy coding! 🚀
