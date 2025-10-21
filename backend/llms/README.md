# LLMs Package 设计文档

## 概述

`llms`包负责与大语言模型的交互，包括ASR（语音识别）和Function Calling（函数调用）。

## 架构设计

```
backend/
├── llms/                          # LLM服务层
│   ├── asr_interface.go           # ASR接口定义
│   ├── agent_interface.go         # Agent接口定义（新增）
│   └── openai/                    # OpenAI实现
│       ├── asr.go                 # Whisper ASR实现
│       └── agent.go               # Function Calling Agent（新增）
├── handler/                       # 业务处理层
│   ├── audio_handler.go           # 音频处理（录音+ASR）
│   └── command_handler.go         # 命令处理（文本+LLM+工具）（新增）
└── executor/                      # 工具执行层
    ├── executor.go                # 工具接口定义
    ├── aggregator.go              # 工具注册与执行
    └── vscode/                    # VSCode工具实现
```

## 1. 音频处理流程

### 1.1 完整流程

```
语音输入 → 录音 → WAV数据 → ASR转录 → 文本输出
```

### 1.2 代码实现

#### 方式一：分步操作

```go
// Step 1: 录音（自动检测静音停止）
audioHandler := handler.NewAudioHandler(apiKey)
wavData, err := audioHandler.RecordAudioWithVAD()

// Step 2: 转录
text, err := audioHandler.TranscribeAudioData(ctx, wavData)
```

#### 方式二：一体化操作（推荐）

```go
// 录音+转录一步完成
audioHandler := handler.NewAudioHandler(apiKey)
text, err := audioHandler.RecordAndTranscribe(ctx)
```

#### 方式三：完整工作流

```go
// 包含时间戳等详细信息
workflow, err := audioHandler.ExecuteVoiceCommandWorkflow(ctx)
// workflow.Transcript 包含转录文本
// workflow.AudioData 包含原始音频
// workflow.RecordingStartTime, RecordingEndTime 等
```

### 1.3 核心组件

- **AudioHandler** (`handler/audio_handler.go`)
  - 协调录音器和ASR服务
  - 提供多种录音模式（VAD/固定时长/手动）
  - 管理录音状态和历史记录

- **ASRService** (`llms/asr_interface.go`)
  - 定义ASR服务接口
  - `Transcribe()`: 基础转录
  - `TranscribeWithOptions()`: 高级选项（语言、时间戳等）

- **OpenAIASR** (`llms/openai/asr.go`)
  - 基于OpenAI Whisper的ASR实现
  - 支持多种音频格式
  - 支持代理配置

## 2. Function Calling 设计

### 2.1 架构模式

```
用户请求(文本) 
    ↓
LLM Agent (OpenAI GPT-4 + Function Calling)
    ↓
Tool Call 决策
    ↓
Executor 执行工具
    ↓
结果反馈给 LLM
    ↓
最终响应
```

### 2.2 核心接口

#### Agent接口 (`llms/agent_interface.go`)

```go
type Agent interface {
    // 处理用户消息，返回助手响应
    Chat(ctx context.Context, message string) (string, error)
    
    // 处理多轮对话
    ChatWithHistory(ctx context.Context, messages []Message) (string, error)
    
    // 重置对话历史
    Reset()
    
    // 获取对话历史
    GetHistory() []Message
}

type Message struct {
    Role    string // "system", "user", "assistant", "tool"
    Content string
    ToolCalls []ToolCall  // assistant消息可能包含工具调用
    ToolCallID string      // tool消息需要引用tool call ID
}
```

#### Agent实现 (`llms/openai/agent.go`)

```go
type OpenAIAgent struct {
    client   *openai.Client
    executor executor.Executor
    model    string
    messages []openai.ChatCompletionMessage
}

func NewOpenAIAgent(apiKey string, exec executor.Executor) *OpenAIAgent {
    // 注入executor实例
}

func (a *OpenAIAgent) Chat(ctx context.Context, message string) (string, error) {
    // 1. 准备tools schemas from executor
    // 2. 发送请求到OpenAI with function calling
    // 3. 循环处理tool calls
    // 4. 执行工具通过executor.ExecuteTool()
    // 5. 将结果反馈给OpenAI
    // 6. 返回最终响应
}
```

### 2.3 集成方式

#### CommandHandler (`handler/command_handler.go`)

```go
type CommandHandler struct {
    audioHandler *AudioHandler  // 处理语音输入
    agent        llms.Agent      // 处理LLM+工具调用
}

// 语音命令处理（录音 → ASR → LLM → 工具执行）
func (h *CommandHandler) ProcessVoiceCommand(ctx context.Context) (string, error) {
    // 1. 录音+转录
    text, err := h.audioHandler.RecordAndTranscribe(ctx)
    
    // 2. LLM处理+工具调用
    response, err := h.agent.Chat(ctx, text)
    
    return response, nil
}

// 文本命令处理（直接LLM → 工具执行）
func (h *CommandHandler) ProcessTextCommand(ctx context.Context, text string) (string, error) {
    return h.agent.Chat(ctx, text)
}
```

### 2.4 使用示例

```go
// 初始化
exec := executor.NewDefaultExecutor()
vscode.RegisterVSCodeLifecycle(exec)

agent := openai_llms.NewOpenAIAgent(apiKey, exec)
audioHandler := handler.NewAudioHandler(apiKey)

cmdHandler := handler.NewCommandHandler(audioHandler, agent)

// 语音控制VSCode
response, err := cmdHandler.ProcessVoiceCommand(ctx)
// 用户说："打开test.md文件"
// LLM自动调用vscode_open_file工具
// 返回："已为您打开test.md文件"

// 文本控制VSCode
response, err := cmdHandler.ProcessTextCommand(ctx, "创建workspace /tmp/demo")
// LLM自动调用create_vscode工具
```

## 3. 依赖注入模式

遵循"外部注入"原则，executor实例通过构造函数传入：

```go
// ✅ 正确：依赖注入
agent := openai_llms.NewOpenAIAgent(apiKey, executor)

// ❌ 错误：内部创建
// agent内部不应该自己new executor
```

## 4. 多轮对话支持

Agent自动管理对话历史：

```go
agent := openai_llms.NewOpenAIAgent(apiKey, exec)

// 第一轮
agent.Chat(ctx, "创建workspace /tmp/test")

// 第二轮（上下文保留）
agent.Chat(ctx, "在里面写一个hello.md文件")

// 查看历史
history := agent.GetHistory()

// 重置对话
agent.Reset()
```

## 5. 错误处理

```go
// 工具执行失败时，LLM会收到错误信息并尝试修正
response, err := agent.Chat(ctx, "打开不存在的文件")
// LLM可能返回："文件不存在，我已为您创建该文件"
```
