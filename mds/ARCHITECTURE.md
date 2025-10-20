# 📚 重构后的架构说明

## 🎯 重构目标

✅ **简化调用链** - 删除不必要的中间层  
✅ **直接使用核心实现** - `backend/handler` 直接调用 `tools/audio`  
✅ **更清晰的职责分离** - Handler 专注于业务逻辑  

---

## 📂 新的文件结构

```
ai-voice-ctrl/
├── app.go                           # Wails 应用入口
├── main.go                          # 程序入口
│
├── backend/
│   └── handler/
│       └── voice_handler.go         # 🔥 语音处理器（业务逻辑层）
│
└── backend/tools/audio/
    ├── capture.go                   # 🎤 音频录制 + VAD（核心实现）
    └── encoder.go                   # 🔊 WAV 编码 + Whisper API
```

**已删除的文件：**
- ❌ `backend/audio_recorder.go` - 不再需要的中间封装层
- ❌ `backend/voice_handler.go` - 已移动到 `backend/handler/`
- ❌ `backend/tools/audio/audio_recorder.go` - 已合并到 `capture.go`

---

## 🔗 调用关系

### 旧架构（3层）：
```
Frontend → app.go → voice_handler.go → audio_recorder.go → tools/audio → malgo
                      (backend)          (backend)           (核心)
```

### 新架构（2层）：
```
Frontend → app.go → handler/voice_handler.go → tools/audio → malgo
                      (业务逻辑)                 (核心实现)
```

**优势：**
- ✅ 减少一层封装，代码更简洁
- ✅ 职责更清晰：handler 处理业务，tools 处理技术
- ✅ 更容易维护和测试

---

## 🎯 VoiceHandler 的作用

`backend/handler/voice_handler.go` 是**业务逻辑层**，负责：

### 1️⃣ **录音管理**
```go
handler.StartRecording()        // 开始录音
handler.StopRecording()         // 停止录音
handler.RecordAudioWithVAD()    // VAD 自动录音
```

### 2️⃣ **API 集成**
```go
handler.RecordAndTranscribe()   // 录音 + 转文字
```

### 3️⃣ **命令处理**
```go
handler.ProcessCommand(text)    // 处理语音命令
handler.RecordTranscribeAndProcess()  // 一站式处理
```

### 4️⃣ **状态管理**
```go
handler.IsRecording()           // 是否录音中
handler.GetRecordingStatus()    // 获取状态
handler.GetLastWavData()        // 获取最后录音
```

### 5️⃣ **配置管理**
```go
handler.SetAPIKey(apiKey)       // 设置 API Key
```

---

## 💻 使用示例

### 示例 1: 最简单的语音识别

```go
package main

import (
    "ai-voice-ctrl/backend/handler"
    "fmt"
    "log"
    "os"
)

func main() {
    // 创建语音处理器
    apiKey := os.Getenv("OPENAI_API_KEY")
    vh := handler.NewVoiceHandler(apiKey)
    defer vh.Close()
    
    fmt.Println("🎤 请说话...")
    
    // 一键完成：录音 + 识别
    text, err := vh.RecordAndTranscribe()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("📝 识别结果: %s\n", text)
}
```

### 示例 2: 完整的语音助手

```go
package main

import (
    "ai-voice-ctrl/backend/handler"
    "fmt"
    "log"
    "os"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    vh := handler.NewVoiceHandler(apiKey)
    defer vh.Close()
    
    for {
        fmt.Println("\n🎤 请说出命令...")
        
        // 录音 + 识别 + 处理
        transcription, result, err := vh.RecordTranscribeAndProcess()
        if err != nil {
            log.Printf("❌ 错误: %v\n", err)
            continue
        }
        
        fmt.Printf("📝 您说: %s\n", transcription)
        fmt.Printf("🤖 结果: %s\n", result)
    }
}
```

### 示例 3: 在 Wails 应用中使用

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
    voiceHandler *handler.VoiceHandler
}

func NewApp() *App {
    apiKey := os.Getenv("OPENAI_API_KEY")
    return &App{
        voiceHandler: handler.NewVoiceHandler(apiKey),
    }
}

// 前端调用：录音并识别
func (a *App) RecordAndTranscribe() (string, error) {
    return a.voiceHandler.RecordAndTranscribe()
}

// 前端调用：完整流程
func (a *App) RecordTranscribeAndProcess() (string, string, error) {
    return a.voiceHandler.RecordTranscribeAndProcess()
}

// 前端调用：获取状态
func (a *App) GetRecordingStatus() string {
    return a.voiceHandler.GetRecordingStatus()
}
```

### 示例 4: 前端 TypeScript 调用

```typescript
// frontend/src/components/VoiceControl.tsx

import { RecordAndTranscribe, GetRecordingStatus } from '../../wailsjs/go/main/App';

function VoiceControl() {
    const [status, setStatus] = useState('idle');
    const [result, setResult] = useState('');
    
    const handleRecord = async () => {
        try {
            setStatus('recording');
            
            // 调用后端：录音 + 识别
            const text = await RecordAndTranscribe();
            
            setResult(text);
            setStatus('idle');
        } catch (err) {
            console.error('Error:', err);
            setStatus('error');
        }
    };
    
    return (
        <div>
            <button onClick={handleRecord} disabled={status === 'recording'}>
                🎤 {status === 'recording' ? '录音中...' : '开始录音'}
            </button>
            <p>识别结果: {result}</p>
        </div>
    );
}
```

---

## 🔧 直接使用 tools/audio（高级用法）

如果您需要更精细的控制，也可以直接使用 `tools/audio`：

```go
package main

import (
    tools "ai-voice-ctrl/backend/tools/audio"
    "fmt"
    "os"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    
    // 方式 1: 直接创建录音器
    recorder := tools.NewAudioRecorderWithVAD(tools.DefaultVADConfig())
    defer recorder.Close()
    
    // 方式 2: 录音
    wavData, _ := recorder.RecordAudioWithVAD()
    
    // 方式 3: 调用 Whisper API
    text, _ := tools.Transcribe(wavData, apiKey)
    
    fmt.Println("识别结果:", text)
}
```

---

## 📊 API 对比

### VoiceHandler API（推荐使用）

| 方法 | 说明 | 返回值 |
|-----|------|--------|
| `NewVoiceHandler(apiKey)` | 创建处理器（默认 VAD） | `*VoiceHandler` |
| `RecordAndTranscribe()` | 录音 + 识别 | `(text string, err)` |
| `RecordTranscribeAndProcess()` | 录音 + 识别 + 处理 | `(text, result, err)` |
| `StartRecording()` | 开始录音 | `error` |
| `StopRecording()` | 停止录音 | `([]byte, error)` |
| `RecordAudioWithVAD()` | VAD 录音 | `([]byte, error)` |
| `RecordAudio(seconds)` | 固定时长录音 | `([]byte, error)` |
| `ProcessCommand(text)` | 处理命令 | `string` |
| `GetRecordingStatus()` | 获取状态 | `string` |
| `SetAPIKey(key)` | 设置 API Key | - |
| `Close()` | 释放资源 | - |

### Tools/Audio API（底层 API）

| 方法 | 说明 | 返回值 |
|-----|------|--------|
| `NewAudioRecorder()` | 创建录音器（无 VAD） | `*AudioRecorder` |
| `NewAudioRecorderWithVAD(config)` | 创建录音器（有 VAD） | `*AudioRecorder` |
| `RecordAudioWithVAD()` | VAD 录音 | `([]byte, error)` |
| `RecordAudio(duration)` | 固定时长录音 | `([]byte, error)` |
| `EncodeToWAV(pcm, rate, channels)` | 编码为 WAV | `([]byte, error)` |
| `Transcribe(wav, apiKey)` | Whisper 识别 | `(string, error)` |

---

## 🎉 总结

### VoiceHandler 的价值：

1. **业务逻辑封装** - 提供高层 API，隐藏技术细节
2. **状态管理** - 管理录音状态和 API Key
3. **一站式服务** - `RecordTranscribeAndProcess()` 一键完成所有步骤
4. **易于测试** - 业务逻辑和技术实现分离
5. **便于扩展** - 可以轻松添加命令处理、错误重试等功能

### 为什么放在 handler 包：

- ✅ 符合分层架构原则
- ✅ 与其他 handler（如 `audio_recognition.go`）保持一致
- ✅ 清晰表明这是业务处理层
- ✅ 便于添加更多 handler（如 `command_handler.go`）

### 推荐使用场景：

| 场景 | 使用方式 |
|------|---------|
| Wails 应用 | 使用 `handler.VoiceHandler` |
| 简单脚本 | 直接使用 `tools/audio` |
| 需要命令处理 | 使用 `handler.VoiceHandler` |
| 需要自定义 VAD | 两者都可以 |

**最佳实践：**
```go
// ✅ 推荐：使用 VoiceHandler
vh := handler.NewVoiceHandler(apiKey)
text, _ := vh.RecordAndTranscribe()

// ✅ 也可以：直接使用 tools（高级用法）
recorder := tools.NewAudioRecorderWithVAD(tools.DefaultVADConfig())
wavData, _ := recorder.RecordAudioWithVAD()
text, _ := tools.Transcribe(wavData, apiKey)
```

---

## 🚀 快速开始

```bash
# 1. 设置环境变量
export OPENAI_API_KEY="your-api-key"

# 2. 运行应用
go run .

# 3. 或者运行 Wails 应用
wails dev
```

现在您的应用架构更清晰、更简洁了！🎉
