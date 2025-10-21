# AI 语音控制桌面应用架构说明

## 1. 总体架构概览
- **Frontend (Wails UI)**：基于 React/TypeScript 的桌面界面，通过 Wails 提供的绑定调用 Go 后端。
- **app.go**：应用入口，负责初始化处理器（`AudioHandler`），并向前端暴露录音、转写、工作流等方法。
- **handler/audio_handler.go**：整体调度中心（Orchestrator），负责串联音频采集、语音识别以及完整工作流逻辑。
- **tools/audio/**：音频采集与处理模块，负责设备管理、VAD、音频编码等细节实现。
- **llms/**：语音识别服务接口与具体实现（例如 OpenAI Whisper），对外提供 ASR 能力。

## 2. AudioHandler 分层职责
- **第 1 层：音频采集方法**
  - `StartRecording`、`StopRecording`、`RecordAudio`、`RecordAudioWithVAD`
  - 调用 `tools/audio/capture.go` 中的 `AudioRecorder` 实现，完成音频流采集。
- **第 2 层：语音识别方法**
  - `TranscribeAudioData`、`TranscribeAudioDataWithOptions`
  - `RecordAndTranscribe`、`RecordAndTranscribeWithOptions`
  - 将音频数据委托给 ASR 服务（`llms.ASRService` 接口）。
- **第 3 层：完整工作流方法**
  - `ExecuteVoiceCommandWorkflow`、`ExecuteVoiceCommandWorkflowWithOptions`
  - 先录音、再转写，未来扩展指令解析与执行，最终返回结构化结果 `VoiceCommandWorkflow`。

## 3. 数据流
1. 用户在前端触发录音 (`RecordAndTranscribe`)。
2. `AudioHandler` 调用 `RecordAudioWithVAD` 采集原始 PCM/WAV 数据。
3. 采集到的音频数据传递给 `TranscribeAudioData`，交由 ASR 服务转换为文本。
4. 转写结果回传至前端；若执行完整工作流，则继续进入指令处理与返回阶段。

## 4. 工作流步骤
1. **Record Audio**：调用 `RecordAudioWithVAD`，由 `tools/audio/capture.go` 实现音频采集。
2. **Transcribe**：调用 `TranscribeAudioData`，由 `llms/openai/asr.go`（OpenAI Whisper）负责语音转文本。
3. **Process Command (Future)**：预留功能，用于调用 LLM 进行语义理解、Function Calling、本地执行等。
4. **Combine Results**：综合音频数据、转写文本、命令执行结果，组装成 `VoiceCommandWorkflow` 返回。

## 5. 组件角色说明
- **AudioHandler**：
  - 负责协调各模块调用顺序与状态管理。
  - 提供统一 API，屏蔽底层实现细节。
  - 不直接实现音频采集或语音识别逻辑。
- **tools/audio (AudioRecorder)**：
  - 管理音频设备（基于 malgo）。
  - 执行实时采集、VAD 检测、PCM 处理、WAV 编码等。
- **llms/asr (ASRService 接口及实现)**：
  - 统一定义语音识别能力。
  - 调用外部服务（OpenAI、未来的 Google/Azure 等）。
  - 处理 API 响应与错误，并支持语言、格式等配置项。

## 6. 依赖关系
- 前端依赖 `app.go`，通过 Wails 调用后端方法。
- `app.go` 依赖 `AudioHandler` 执行核心逻辑。
- `AudioHandler` 同时依赖音频工具模块和 ASR 服务接口。
- `tools/audio` 提供具体实现；`llms.ASRService` 由 OpenAI Whisper 等实现类满足，未来可新增更多 ASR 供应商。

## 7. 扩展点
- **音频采集**：可替换或扩展更多音频库、加入降噪/回声消除等能力。
- **语音识别**：通过实现 `llms.ASRService` 接口接入更多供应商或本地模型。
- **命令处理**：在工作流第三步扩展 LLM Function Calling、本地命令执行、结果回显流程。
- **错误处理与监控**：在 `AudioHandler` 中集中记录日志、指标，保证链路稳定。

该文档用于快速了解 AI 语音控制桌面应用中各层职责、数据流和扩展方向。
