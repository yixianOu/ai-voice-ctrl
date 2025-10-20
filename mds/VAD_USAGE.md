# Voice Activity Detection (VAD) 使用指南

## 🎯 什么是 VAD？

VAD (Voice Activity Detection - 语音活动检测) 是一种智能检测用户何时停止说话的技术。

**工作原理：**
1. 实时监测音频流的音量（RMS - Root Mean Square）
2. 当音量低于阈值超过设定时间（如 700ms），判定为"用户说完了"
3. 自动停止录音并返回音频数据

**优势：**
- ✅ 即时响应：用户说完立刻处理，无需等待
- ✅ 流畅体验：类似 Siri、Google Assistant 的交互方式
- ✅ 节省流量：只录制有效语音，不包含长时间静音

---

## 📚 使用方式

### 方式 1: 使用 VAD 的录音器（推荐）

```go
import (
    "ai-voice-ctrl/backend"
    "fmt"
    "log"
)

func main() {
    // 创建启用 VAD 的录音器
    recorder := backend.NewAudioRecorderWithVAD()
    defer recorder.Close()
    
    fmt.Println("开始说话...")
    
    // 自动检测：用户说完后自动停止
    wavData, err := recorder.RecordAudioWithVAD()
    if err != nil {
        log.Fatalf("录音失败: %v", err)
    }
    
    fmt.Printf("录制完成，音频大小: %d bytes\n", len(wavData))
    
    // 现在可以发送给 Whisper API 进行识别
}
```

### 方式 2: 自定义 VAD 配置

```go
import (
    "ai-voice-ctrl/backend"
    tools "ai-voice-ctrl/backend/tools/audio"
    "time"
)

func main() {
    // 自定义 VAD 参数
    vadConfig := tools.VADConfig{
        SilenceThreshold:  0.015,          // 静音阈值（更灵敏）
        SilenceDuration:   500 * time.Millisecond,  // 500ms 静音就停止
        MinSpeechDuration: 200 * time.Millisecond,  // 至少说 200ms
        MaxRecordDuration: 20 * time.Second,        // 最多录 20 秒
    }
    
    recorder := backend.NewAudioRecorderWithCustomVAD(vadConfig)
    defer recorder.Close()
    
    wavData, _ := recorder.RecordAudioWithVAD()
    // 处理音频数据...
}
```

### 方式 3: 手动控制（不使用 VAD）

```go
import (
    "ai-voice-ctrl/backend"
    "time"
)

func main() {
    // 创建普通录音器（不启用 VAD）
    recorder := backend.NewAudioRecorder()
    defer recorder.Close()
    
    // 手动控制开始和停止
    recorder.StartRecording()
    
    // 录制固定时长
    time.Sleep(5 * time.Second)
    
    wavData, _ := recorder.StopRecording()
    // 处理音频数据...
}
```

### 方式 4: 在 voice_handler 中使用

```go
// backend/voice_handler.go 已经集成了 VAD 功能

handler := backend.NewVoiceHandler()

// 使用 VAD 录音（自动停止）
wavData, err := handler.RecordAudioWithVAD()

// 或者固定时长录音
wavData, err := handler.RecordAudio(5) // 5秒
```

---

## ⚙️ VAD 配置参数详解

```go
type VADConfig struct {
    SilenceThreshold  float64       // 静音阈值 (0.0 - 1.0)
    SilenceDuration   time.Duration // 静音持续时间
    MinSpeechDuration time.Duration // 最小语音时长
    MaxRecordDuration time.Duration // 最大录音时长
}
```

### `SilenceThreshold` - 静音阈值

**含义**: 音频音量低于此值时视为静音

| 值 | 灵敏度 | 适用场景 |
|----|--------|---------|
| 0.01 | 非常灵敏 | 安静环境，清晰发音 |
| 0.02 | **推荐** | 正常办公室环境 |
| 0.03 | 较不灵敏 | 嘈杂环境 |
| 0.05 | 很不灵敏 | 非常嘈杂的环境 |

**如何调整**:
- 如果**过早停止**（用户还没说完就停了）→ 提高阈值（如 0.03）
- 如果**停不下来**（用户说完了还在录）→ 降低阈值（如 0.015）

### `SilenceDuration` - 静音持续时间

**含义**: 静音持续多久后判定为"说完了"

| 值 | 响应速度 | 适用场景 |
|----|---------|---------|
| 300-500ms | 很快 | 短句、命令词（"打开备忘录"）|
| 700ms | **推荐** | 正常对话 |
| 1000ms | 较慢 | 长句子、思考时有停顿 |

**推荐**:
- **命令式交互**: 500ms（快速响应）
- **自然对话**: 700ms（标准配置）
- **长文本输入**: 1000ms（给用户思考时间）

### `MinSpeechDuration` - 最小语音时长

**含义**: 至少录制多久才允许停止（避免误触）

| 值 | 说明 |
|----|------|
| 200ms | 最小值，防止短促噪音触发 |
| 300ms | **推荐**，正常说一个词的时间 |
| 500ms | 更保守，确保至少说了一句话 |

### `MaxRecordDuration` - 最大录音时长

**含义**: 强制停止的最长时间（防止无限录音）

| 值 | 说明 |
|----|------|
| 10s | 短命令 |
| 30s | **推荐**，正常对话 |
| 60s | 长文本、详细描述 |

---

## 🎮 实战场景配置

### 场景 1: 语音命令（如 "打开备忘录"）

```go
vadConfig := tools.VADConfig{
    SilenceThreshold:  0.02,
    SilenceDuration:   500 * time.Millisecond,  // 快速响应
    MinSpeechDuration: 200 * time.Millisecond,
    MaxRecordDuration: 10 * time.Second,
}
```

### 场景 2: 自然对话

```go
vadConfig := tools.DefaultVADConfig() // 使用默认配置即可
// SilenceThreshold:  0.02
// SilenceDuration:   700ms
// MinSpeechDuration: 300ms
// MaxRecordDuration: 30s
```

### 场景 3: 长文本输入

```go
vadConfig := tools.VADConfig{
    SilenceThreshold:  0.02,
    SilenceDuration:   1000 * time.Millisecond,  // 给更多思考时间
    MinSpeechDuration: 500 * time.Millisecond,
    MaxRecordDuration: 60 * time.Second,         // 允许更长录音
}
```

### 场景 4: 嘈杂环境

```go
vadConfig := tools.VADConfig{
    SilenceThreshold:  0.04,                     // 提高阈值，过滤背景噪音
    SilenceDuration:   800 * time.Millisecond,
    MinSpeechDuration: 400 * time.Millisecond,
    MaxRecordDuration: 30 * time.Second,
}
```

---

## 🔧 工作流程

```
1. 用户点击按钮 / 唤醒词触发
   ↓
2. StartRecording() - 开始录音
   ↓
3. 实时 VAD 检测循环：
   │
   ├─ 计算音频帧的 RMS (音量)
   │
   ├─ 音量 > 阈值？
   │  ├─ 是 → 更新 lastSoundTime（有声音）
   │  └─ 否 → 检查静音时长
   │           │
   │           ├─ 静音时长 >= SilenceDuration？
   │           │  └─ 是 → 触发停止信号
   │           │
   │           └─ 录音时长 >= MaxRecordDuration？
   │              └─ 是 → 强制停止
   │
   ↓
4. StopRecording() - 停止录音
   ↓
5. 编码为 WAV 格式
   ↓
6. 返回音频数据（[]byte）
   ↓
7. 发送给 Whisper API 识别
```

---

## 📊 性能优化

### 内存使用
- PCM 数据存储在内存中（`[]int16`）
- 16kHz 单声道，1 秒 ≈ 32KB
- 30 秒录音 ≈ 1MB 内存

### 实时性能
- VAD 检测在音频回调中进行（< 1ms）
- 停止延迟 = SilenceDuration (通常 500-1000ms)

---

## ⚠️ 注意事项

1. **麦克风权限**: 确保应用有麦克风访问权限
2. **CGO 依赖**: malgo 需要 CGO，Windows 需要 MinGW
3. **线程安全**: 内部使用 `sync.Mutex` 保护共享数据
4. **资源释放**: 使用后调用 `Close()` 释放资源

---

## 🚀 快速开始

```go
package main

import (
    "ai-voice-ctrl/backend"
    "fmt"
    "log"
)

func main() {
    // 最简单的方式：使用默认 VAD 配置
    recorder := backend.NewAudioRecorderWithVAD()
    defer recorder.Close()
    
    fmt.Println("🎤 开始说话，说完后自动停止...")
    
    wavData, err := recorder.RecordAudioWithVAD()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("✅ 录音完成！大小: %d bytes\n", len(wavData))
    
    // TODO: 发送给 Whisper API
    // text, err := transcribe(wavData, apiKey)
}
```

---

## 🔗 相关文件

- `backend/tools/audio/audio_recorder.go` - VAD 核心实现
- `backend/audio_recorder.go` - 后端包装器
- `backend/voice_handler.go` - 语音处理器
- `backend/tools/audio/encoder.go` - WAV 编码和 Whisper API

---

## 📝 示例：完整的语音识别流程

```go
package main

import (
    "ai-voice-ctrl/backend"
    tools "ai-voice-ctrl/backend/tools/audio"
    "fmt"
    "log"
    "os"
)

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("请设置 OPENAI_API_KEY 环境变量")
    }
    
    // 1. 创建 VAD 录音器
    recorder := backend.NewAudioRecorderWithVAD()
    defer recorder.Close()
    
    fmt.Println("🎤 请说话...")
    
    // 2. 录音（自动检测结束）
    wavData, err := recorder.RecordAudioWithVAD()
    if err != nil {
        log.Fatalf("录音失败: %v", err)
    }
    
    fmt.Println("✅ 录音完成")
    fmt.Println("🔄 正在识别...")
    
    // 3. 发送给 Whisper API
    text, err := tools.Transcribe(wavData, apiKey)
    if err != nil {
        log.Fatalf("识别失败: %v", err)
    }
    
    // 4. 显示结果
    fmt.Printf("📝 识别结果: %s\n", text)
    
    // 5. 处理命令
    // processCommand(text)
}
```

---

## 🎉 总结

VAD 功能让您的语音助手能够：
- ✅ 自然地检测用户何时说完
- ✅ 立即响应，无需等待固定时长
- ✅ 提供类似 Siri 的流畅体验
- ✅ 节省 API 调用成本（只传输有效语音）

使用 `NewAudioRecorderWithVAD()` 即可开始！🚀
