# Audio Handler 架构图

## 总体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        Frontend (Wails UI)                       │
│                     React/TypeScript Components                  │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      │ Wails Bindings
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                         app.go                                   │
│                   Application Entry Point                        │
└─────────────────────┬───────────────────────────────────────────┘
                      │
                      │ Method Calls
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                  handler/audio_handler.go                        │
│           ╔═══════════════════════════════════════╗             │
│           ║      ORCHESTRATOR / DISPATCHER        ║             │
│           ║    (协调转发中心 - 不实现具体逻辑)      ║             │
│           ╚═══════════════════════════════════════╝             │
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  Layer 1: Audio Capture Methods                        │    │
│  │  • StartRecording() / StopRecording()                  │    │
│  │  • RecordAudio(seconds)                                │    │
│  │  • RecordAudioWithVAD()                                │    │
│  └────────────────────────────────────────────────────────┘    │
│                           │                                      │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  Layer 2: Speech Recognition Methods                   │    │
│  │  • TranscribeAudioData(ctx, data)                      │    │
│  │  • TranscribeAudioDataWithOptions(ctx, data, opts)     │    │
│  │  • RecordAndTranscribe(ctx)                            │    │
│  │  • RecordAndTranscribeWithOptions(ctx, opts)           │    │
│  └────────────────────────────────────────────────────────┘    │
│                           │                                      │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  Layer 3: Complete Workflow Methods                    │    │
│  │  • ExecuteVoiceCommandWorkflow(ctx)                    │    │
│  │  • ExecuteVoiceCommandWorkflowWithOptions(ctx, opts)   │    │
│  │  → Returns: VoiceCommandWorkflow (structured result)   │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
└──────────────┬─────────────────────────┬────────────────────────┘
               │                         │
               │                         │
        ┌──────▼──────┐          ┌──────▼──────────────────────┐
        │ Audio Tools │          │    ASR Service Interface    │
        │             │          │  (llms.ASRService)          │
        └──────┬──────┘          └──────┬──────────────────────┘
               │                         │
               │                         │
┌──────────────▼─────────────┐  ┌───────▼──────────────────────┐
│  tools/audio/capture.go    │  │   llms/openai/asr.go         │
│  • AudioRecorder           │  │   • OpenAIASR (Whisper)      │
│  • VAD Detection           │  │                              │
│  • Malgo Integration       │  │   (Future: Google, Azure,    │
└────────────────────────────┘  │    Local Whisper, etc.)      │
                                └──────────────────────────────┘
┌────────────────────────────┐
│  tools/audio/encoder.go    │
│  • WAV Encoding            │
│  • PCM Processing          │
└────────────────────────────┘
```

## 数据流图

### 简单转录流程 (RecordAndTranscribe)

```
User speaks
    │
    ▼
┌───────────────────────┐
│  RecordAndTranscribe  │  ← Frontend calls
└───────┬───────────────┘
        │
        ▼
┌───────────────────────┐
│ RecordAudioWithVAD()  │  ← Delegate to AudioRecorder
└───────┬───────────────┘
        │
        ▼
  [Audio Data: []byte]
        │
        ▼
┌───────────────────────┐
│TranscribeAudioData()  │  ← Delegate to ASRService
└───────┬───────────────┘
        │
        ▼
  [Transcript: string]
        │
        ▼
   Return to Frontend
```

### 完整工作流 (ExecuteVoiceCommandWorkflow)

```
User speaks
    │
    ▼
┌────────────────────────────────┐
│ExecuteVoiceCommandWorkflow     │
└────────┬───────────────────────┘
         │
         ├─ [Step 1] Record Audio
         │      │
         │      ▼
         │  ┌────────────────────┐
         │  │ RecordAudioWithVAD │ ───→ tools/audio/capture.go
         │  └────────┬───────────┘
         │           │
         │           ▼
         │      [AudioData]
         │
         ├─ [Step 2] Transcribe
         │      │
         │      ▼
         │  ┌─────────────────────┐
         │  │ TranscribeAudioData │ ───→ llms/openai/asr.go
         │  └────────┬────────────┘
         │           │
         │           ▼
         │      [Transcript]
         │
         ├─ [Step 3] Process Command (Future)
         │      │
         │      ▼
         │  ┌───────────────────┐
         │  │ ProcessCommand    │ ───→ (To be implemented)
         │  └────────┬──────────┘        • Prompt engineering
         │           │                   • LLM API call
         │           ▼                   • Function calling
         │      [CommandResult]          • Local execution
         │
         └─ Combine all results
                │
                ▼
       ┌────────────────────────┐
       │ VoiceCommandWorkflow   │
       │ • AudioData            │
       │ • Transcript           │
       │ • TranscriptDetails    │
       │ • CommandResult        │
       │ • Timestamps           │
       └────────┬───────────────┘
                │
                ▼
         Return to Frontend
```

## 组件职责

```
┌────────────────────────────────────────────────────────────┐
│                     AudioHandler                           │
│                   (协调转发中心)                           │
├────────────────────────────────────────────────────────────┤
│ 职责:                                                      │
│ ✓ 协调各组件的调用顺序                                     │
│ ✓ 管理工作流状态                                           │
│ ✓ 提供统一的 API 接口                                      │
│ ✓ 处理错误和上下文传递                                     │
│                                                            │
│ 不做的事:                                                  │
│ ✗ 不实现具体的音频采集逻辑                                 │
│ ✗ 不实现具体的语音识别逻辑                                 │
│ ✗ 不直接操作音频设备                                       │
│ ✗ 不直接调用 API                                          │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│              tools/audio (AudioRecorder)                   │
│                  (音频采集专家)                            │
├────────────────────────────────────────────────────────────┤
│ 职责:                                                      │
│ ✓ 音频设备管理 (malgo)                                     │
│ ✓ 实时音频采集                                             │
│ ✓ VAD (Voice Activity Detection)                          │
│ ✓ PCM 数据处理                                             │
│ ✓ WAV 格式编码                                             │
└────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────┐
│              llms/asr (ASRService)                         │
│                 (语音识别专家)                             │
├────────────────────────────────────────────────────────────┤
│ 职责:                                                      │
│ ✓ 语音转文本 (Speech-to-Text)                             │
│ ✓ API 调用 (OpenAI, Google, Azure, etc.)                 │
│ ✓ 响应解析                                                 │
│ ✓ 错误处理                                                 │
│ ✓ 支持多种选项 (语言、温度、格式等)                       │
└────────────────────────────────────────────────────────────┘
```

## 依赖关系图

```
                     ┌─────────────────┐
                     │   Frontend      │
                     │   (Wails UI)    │
                     └────────┬────────┘
                              │
                              │ depends on
                              ▼
                     ┌─────────────────┐
                     │     app.go      │
                     └────────┬────────┘
                              │
                              │ depends on
                              ▼
                     ┌─────────────────┐
                     │ AudioHandler    │
                     └────────┬────────┘
                              │
                 ┌────────────┴────────────┐
                 │                         │
         depends on                 depends on
                 │                         │
                 ▼                         ▼
      ┌──────────────────┐    ┌──────────────────────┐
      │ tools/audio      │    │ llms.ASRService      │
      │ (Concrete)       │    │ (Interface)          │
      └──────────────────┘    └──────┬───────────────┘
                                      │
                              implements by
                                      │
                         ┌────────────┴──────────┐
                         │                       │
                         ▼                       ▼
              ┌─────────────────┐    ┌──────────────────┐
              │ OpenAIASR       │    │  GoogleASR       │
              │ (Whisper)       │    │  (Future)        │
              └─────────────────┘    └──────────────────┘
                         │                       │
                         ▼                       ▼
              ┌─────────────────┐    ┌──────────────────┐
              │ AzureASR        │    │  LocalWhisperASR │
              │ (Future)        │    │  (Future)        │
              └─────────────────┘    └──────────────────┘
```

## 接口抽象

```
┌──────────────────────────────────────────────────────────────┐
│                    llms.ASRService                            │
│                     (Interface)                               │
├──────────────────────────────────────────────────────────────┤
│ + Transcribe(ctx, audioData) (string, error)                 │
│ + TranscribeWithOptions(ctx, audioData, opts)                │
│     (TranscribeResponse, error)                              │
│ + GetServiceName() string                                    │
└──────────────────────────────────────────────────────────────┘
                              △
                              │ implements
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
        │                     │                     │
┌───────┴──────┐   ┌──────────┴────────┐   ┌───────┴────────┐
│  OpenAIASR   │   │   GoogleASR       │   │   AzureASR     │
│              │   │   (Future)        │   │   (Future)     │
└──────────────┘   └───────────────────┘   └────────────────┘

Benefits of Interface Abstraction:
✓ 依赖倒置 (Dependency Inversion)
✓ 可替换性 (Substitutability)
✓ 可测试性 (Testability) - 可以注入 Mock
✓ 扩展性 (Extensibility) - 轻松添加新服务
```

## 配置注入模式

```
┌──────────────────────────────────────────────────────────┐
│              AudioHandlerConfig                           │
│              (Configuration Object)                       │
├──────────────────────────────────────────────────────────┤
│ • VADConfig        tools.VADConfig                        │
│ • ASRService       llms.ASRService     [Injectable]      │
│ • OpenAIAPIKey     string                                 │
│ • EnableVAD        bool                                   │
└──────────┬───────────────────────────────────────────────┘
           │
           │ injected into
           ▼
┌──────────────────────────────────────────────────────────┐
│              AudioHandler                                 │
├──────────────────────────────────────────────────────────┤
│ - recorder       *tools.AudioRecorder                     │
│ - asrService     llms.ASRService       [Injected]        │
│ - vadConfig      tools.VADConfig                          │
│ - lastWavData    []byte                                   │
│ - lastTranscript string                                   │
└──────────────────────────────────────────────────────────┘

Usage:
    config := handler.AudioHandlerConfig{
        VADConfig:    customVADConfig,
        ASRService:   myCustomASR,    // Injectable!
        EnableVAD:    true,
    }
    handler := handler.NewAudioHandlerWithConfig(config)
```

## 未来扩展: LLM 集成

```
                  Current Architecture
┌─────────────────────────────────────────────────────┐
│  RecordAudioWithVAD → TranscribeAudioData           │
│                           │                         │
│                           ▼                         │
│                    [Transcript Text]                │
│                           │                         │
│                           ▼                         │
│                  processCommandPlaceholder()        │
│                   (返回占位符字符串)                │
└─────────────────────────────────────────────────────┘

                   Future Architecture
┌─────────────────────────────────────────────────────┐
│  RecordAudioWithVAD → TranscribeAudioData           │
│                           │                         │
│                           ▼                         │
│                    [Transcript Text]                │
│                           │                         │
│                           ▼                         │
│              ┌─────────────────────────┐            │
│              │   ProcessCommand()      │            │
│              │   ┌─────────────────┐   │            │
│              │   │ 1. Build Prompt │   │            │
│              │   │ 2. Call LLM API │   │            │
│              │   │ 3. Parse Result │   │            │
│              │   │ 4. Function Call│   │            │
│              │   │ 5. Execute      │   │            │
│              │   └─────────────────┘   │            │
│              └────────┬────────────────┘            │
│                       │                             │
│                       ▼                             │
│              [Structured Result]                    │
└─────────────────────────────────────────────────────┘

Integration Points:
• backend/prompts/        ← Prompt templates
• backend/llms/chat/      ← LLM API clients
• backend/executor/       ← Function execution
```

---

这个架构设计确保了:
1. ✅ **职责清晰**: 每个组件都有明确的职责
2. ✅ **松耦合**: 通过接口降低组件间的依赖
3. ✅ **高内聚**: 相关功能组织在一起
4. ✅ **易扩展**: 可以轻松添加新功能而不影响现有代码
5. ✅ **可测试**: 依赖注入使得单元测试变得简单
