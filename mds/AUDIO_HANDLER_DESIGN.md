# Audio Handler Architecture Design

## 📋 概述

`audio_handler.go` 是一个**协调转发中心**（Orchestrator/Dispatcher），负责协调音频采集、语音识别和命令处理的完整工作流程。

## 🏗️ 架构设计原则

### 1. 单一职责原则 (SRP)
- **AudioHandler**: 协调调度，不直接实现具体功能
- **tools/audio**: 专注音频采集和编码
- **llms/asr**: 专注语音识别服务

### 2. 依赖倒置原则 (DIP)
- 依赖于接口 `llms.ASRService` 而非具体实现
- 可以轻松切换不同的 ASR 提供商（OpenAI, Google, Azure, 本地模型等）

### 3. 开闭原则 (OCP)
- 对扩展开放：可以添加新的 ASR 服务、新的工作流
- 对修改封闭：核心协调逻辑稳定

## 📦 组件层次结构

```
Frontend (Wails UI)
    ↓
app.go (Application Entry)
    ↓
handler/audio_handler.go (Coordinator/Dispatcher)
    ↓
    ├─→ tools/audio (Audio Capture)
    │   ├─ capture.go (Recording with VAD)
    │   └─ encoder.go (WAV encoding)
    │
    └─→ llms/asr (Speech Recognition)
        ├─ asr_interface.go (Interface)
        └─ openai/asr.go (OpenAI Whisper)
```

## 🎯 核心功能分层

### Layer 1: 音频采集方法
底层音频录制功能的简单封装：

```go
// 手动模式：开始/停止录制
StartRecording() error
StopRecording() ([]byte, error)

// 定时录制：固定时长
RecordAudio(seconds int) ([]byte, error)

// 自动模式：VAD 自动检测
RecordAudioWithVAD() ([]byte, error)
```

**设计意图**: 提供灵活的录制方式，适配不同场景。

### Layer 2: 语音识别方法
协调音频数据与 ASR 服务：

```go
// 转录已有音频数据
TranscribeAudioData(ctx, audioData) (string, error)

// 转录已有音频（带选项）
TranscribeAudioDataWithOptions(ctx, audioData, opts) (TranscribeResponse, error)

// 录制并转录（一步到位）
RecordAndTranscribe(ctx) (string, error)

// 录制并转录（带选项）
RecordAndTranscribeWithOptions(ctx, opts) (TranscribeResponse, error)
```

**设计意图**: 
- 分离"已有数据转录"和"录制后转录"两种场景
- 支持从文件、网络等来源获取音频数据

### Layer 3: 完整工作流方法
端到端的业务流程：

```go
// 执行完整的语音命令工作流
ExecuteVoiceCommandWorkflow(ctx) (*VoiceCommandWorkflow, error)

// 执行工作流（带选项）
ExecuteVoiceCommandWorkflowWithOptions(ctx, opts) (*VoiceCommandWorkflow, error)
```

**设计意图**: 
- 封装完整业务逻辑
- 返回详细的工作流结果，包含时间戳、中间数据等
- 为前端提供清晰的调用入口

## 🔄 工作流设计

### VoiceCommandWorkflow 结构

```go
type VoiceCommandWorkflow struct {
    AudioData          []byte                    // 录制的音频数据
    Transcript         string                    // 转录文本
    TranscriptDetails  *llms.TranscribeResponse  // 详细转录结果（分段、词级时间戳等）
    CommandResult      string                    // 命令处理结果
    RecordingStartTime time.Time                 // 录制开始时间
    RecordingEndTime   time.Time                 // 录制结束时间
    TranscriptionTime  time.Time                 // 转录完成时间
}
```

**设计优势**:
- 完整的可追溯性（时间戳）
- 保留中间数据用于调试和分析
- 结构化返回，便于前端展示

## 🔌 可扩展性设计

### 1. ASR 服务切换

```go
// 切换到不同的 ASR 服务
handler.SetASRService(googleASR)
handler.SetASRService(azureASR)
handler.SetASRService(localWhisperASR)
```

### 2. 自定义配置

```go
config := handler.AudioHandlerConfig{
    VADConfig:    customVADConfig,
    ASRService:   myCustomASR,
    EnableVAD:    true,
}
handler := handler.NewAudioHandlerWithConfig(config)
```

### 3. 未来扩展点：LLM 命令处理

**预留接口**:
```go
// TODO: 将在未来实现
func (ah *AudioHandler) ProcessCommand(
    ctx context.Context, 
    transcript string, 
    functions []FunctionDefinition
) (CommandResult, error)
```

**未来架构**:
```
transcript → Prompt Engineering → LLM (GPT-4) → Function Call → Local Execution
```

## 📊 使用场景示例

### 场景 1: 简单语音输入
```go
// 前端调用
handler := handler.NewAudioHandler(apiKey)
text, err := handler.RecordAndTranscribe(ctx)
// 用户说话 → 自动检测结束 → 返回文本
```

### 场景 2: 带语言指定的转录
```go
opts := llms.TranscribeOptions{
    Language: "zh",  // 中文
    Temperature: 0.2,
}
response, err := handler.RecordAndTranscribeWithOptions(ctx, opts)
```

### 场景 3: 完整工作流（含时间戳）
```go
workflow, err := handler.ExecuteVoiceCommandWorkflow(ctx)
fmt.Printf("Recording took: %v\n", workflow.RecordingEndTime.Sub(workflow.RecordingStartTime))
fmt.Printf("Transcript: %s\n", workflow.Transcript)
```

### 场景 4: 从文件转录
```go
audioData, _ := os.ReadFile("audio.wav")
text, err := handler.TranscribeAudioData(ctx, audioData)
```

## 🎯 状态管理

### 内部状态
- `lastWavData`: 最后一次录制的音频数据
- `lastTranscript`: 最后一次转录的文本

### 状态查询方法
```go
IsRecording() bool              // 是否正在录制
GetRecordingStatus() string     // 录制状态描述
GetLastWavData() []byte         // 获取上次音频
GetLastTranscript() string      // 获取上次转录
GetASRServiceName() string      // 当前 ASR 服务名称
```

## 🔐 资源管理

### 生命周期
```go
// 创建
handler := handler.NewAudioHandler(apiKey)

// 使用
workflow, err := handler.ExecuteVoiceCommandWorkflow(ctx)

// 关闭（释放音频设备）
defer handler.Close()
```

## 🚀 未来路线图

### Phase 1: ✅ 当前阶段
- [x] 音频采集协调
- [x] 语音识别集成
- [x] 基础工作流

### Phase 2: 🔄 规划中（LLM 集成）
- [ ] Prompt 工程模块
- [ ] Function Calling 定义
- [ ] LLM API 集成（OpenAI GPT-4, Claude, 本地模型）
- [ ] 命令执行器

### Phase 3: 📋 待规划
- [ ] 对话上下文管理
- [ ] 多轮对话支持
- [ ] 命令历史记录
- [ ] 用户偏好学习

## 🎨 设计模式应用

### 1. Facade Pattern (外观模式)
`AudioHandler` 为复杂的音频处理和识别系统提供简化接口。

### 2. Strategy Pattern (策略模式)
通过 `llms.ASRService` 接口支持多种 ASR 策略。

### 3. Template Method Pattern (模板方法)
`ExecuteVoiceCommandWorkflow` 定义了固定的工作流骨架。

## 📝 代码规范

### 方法命名
- `Record*`: 音频采集相关
- `Transcribe*`: 语音识别相关
- `Execute*Workflow`: 完整工作流
- `Get*`: 状态查询
- `Set*`/`Update*`: 配置修改

### 错误处理
- 所有错误都包装了上下文信息
- 使用 `fmt.Errorf("context: %w", err)` 保留错误链

### 注释规范
- 每个 Layer 用分隔符标记
- 每个方法都有用途说明
- TODO 标记未来实现点

## 🔗 相关文档

- [VAD_USAGE.md](./VAD_USAGE.md) - VAD 使用指南
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 整体架构文档
- [REFACTOR_SUMMARY.md](./REFACTOR_SUMMARY.md) - 重构总结

---

**设计理念**: 让 AudioHandler 成为一个**纯粹的协调者**，它知道如何组合各个组件，但不实现具体的音频处理或识别逻辑。这种设计使得系统高度模块化、可测试、可扩展。
