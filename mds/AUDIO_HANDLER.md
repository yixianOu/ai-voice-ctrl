我来详细分析 audio_handler.go 文件的作用和实现的方法。

📋 文件作用
audio_handler.go 是整个语音控制系统的核心协调器/调度中心，它扮演以下角色：

🎯 核心职责
协调者 (Orchestrator)

协调音频采集（tools/audio）和语音识别（llms/asr）
不直接实现底层功能，只负责调度
转发中心 (Dispatcher)

接收前端请求
分发到对应的服务层
返回统一的响应
状态管理器 (State Manager)

维护录音状态
缓存最后的录音数据和识别结果
配置管理器 (Config Manager)

管理 VAD 配置
管理 ASR 服务切换
🏗️ 架构定位
📦 结构体定义
1. AudioHandler - 主结构体
2. AudioHandlerConfig - 配置结构
3. VoiceCommandWorkflow - 工作流结果
🔧 实现的方法（共 18 个）
📊 方法分类统计
分类	数量	方法
构造函数	3	NewAudioHandler, NewAudioHandlerWithConfig, DefaultAudioHandlerConfig
音频采集	4	StartRecording, StopRecording, RecordAudio, RecordAudioWithVAD
语音识别	4	TranscribeAudioData, TranscribeAudioDataWithOptions, RecordAndTranscribe, RecordAndTranscribeWithOptions
完整工作流	2	ExecuteVoiceCommandWorkflow, ExecuteVoiceCommandWorkflowWithOptions
状态查询	4	IsRecording, GetRecordingStatus, GetLastWavData, GetLastTranscript
配置管理	3	SetASRService, GetASRServiceName, UpdateVADConfig
资源管理	1	Close
命令处理	1	processCommandPlaceholder (内部方法)
🎯 详细方法说明
1️⃣ 构造函数（3个）
NewAudioHandler(apiKey string) *AudioHandler
作用: 使用默认配置创建 AudioHandler
使用场景: 快速启动，最简单的方式
调用链:

NewAudioHandlerWithConfig(config AudioHandlerConfig) *AudioHandler
作用: 使用自定义配置创建 AudioHandler
核心逻辑:

✅ 初始化 ASR 服务（优先使用传入的，否则创建 OpenAI）
✅ 检查 VADConfig 是否为零值，如果是则使用默认值
✅ 根据 EnableVAD 决定创建带 VAD 或不带 VAD 的录音器
DefaultAudioHandlerConfig(apiKey string) AudioHandlerConfig
作用: 返回默认配置
默认值:

VADConfig: DefaultVADConfig()
EnableVAD: true
OpenAIAPIKey: 传入的 key
2️⃣ 音频采集方法（4个）
这些方法都是简单转发到 recorder，并缓存结果到 lastWavData。

StartRecording() error
作用: 手动开始录音
使用场景: GUI 按钮控制
特点:

✅ 非阻塞
✅ 需要手动调用 StopRecording()
StopRecording() ([]byte, error)
作用: 手动停止录音，返回 WAV 数据
特点:

✅ 返回完整的 WAV 文件（包含文件头）
✅ 自动缓存到 lastWavData
RecordAudio(seconds int) ([]byte, error)
作用: 录制固定时长的音频
使用场景:

✅ 测试音频采集
✅ 录制固定长度样本
RecordAudioWithVAD() ([]byte, error)
作用: 使用 VAD 自动检测录音结束
使用场景:

⭐ 最常用的方法
✅ 智能对话系统
✅ 语音助手
工作流程:

3️⃣ 语音识别方法（4个）
这些方法协调 recorder 和 asrService。

TranscribeAudioData(ctx context.Context, audioData []byte) (string, error)
作用: 转录已有的音频数据（简单版）
返回: 纯文本
使用场景: 转录文件、网络接收的音频

TranscribeAudioDataWithOptions(ctx, audioData, opts) (TranscribeResponse, error)
作用: 转录音频数据（带选项）
返回: 详细响应（包含分段、时间戳等）
使用场景:

✅ 需要时间戳
✅ 需要分段信息
✅ 字幕生成
RecordAndTranscribe(ctx context.Context) (string, error)
作用: ⭐ 一步完成录音+识别（简单版）
工作流程:

使用场景:

⭐ 最常用的方法
✅ 快速语音输入
✅ 语音命令
RecordAndTranscribeWithOptions(ctx, opts) (TranscribeResponse, error)
作用: 录音+识别（带选项）
使用场景: 需要详细信息时

4️⃣ 完整工作流方法（2个）
这些方法返回完整的工作流结果，包含时间统计。

ExecuteVoiceCommandWorkflow(ctx) (*VoiceCommandWorkflow, error)
作用: 执行端到端工作流
步骤:

返回值:

使用场景:

✅ 性能分析
✅ 调试流程
✅ 监控时延
ExecuteVoiceCommandWorkflowWithOptions(ctx, opts) (*VoiceCommandWorkflow, error)
作用: 带选项的完整工作流
区别: 返回的 TranscriptDetails 包含详细信息

5️⃣ 状态查询方法（4个）
简单转发到 recorder 或返回缓存数据。

IsRecording() bool
作用: 检查是否正在录音
使用场景: UI 状态显示

GetRecordingStatus() string
作用: 获取录音状态字符串
返回值: "recording" 或 "idle"

GetLastWavData() []byte
作用: 获取最后一次录音的 WAV 数据
使用场景:

✅ 保存录音文件
✅ 重新识别
GetLastTranscript() string
作用: 获取最后一次识别的文本

6️⃣ 配置管理方法（3个）
SetASRService(service llms.ASRService)
作用: 运行时切换 ASR 服务
使用场景:

✅ 在 OpenAI / Google / Azure 之间切换
✅ 服务降级
GetASRServiceName() string
作用: 获取当前 ASR 服务名称
返回值: "OpenAI Whisper", "Google Speech-to-Text" 等

UpdateVADConfig(config tools.VADConfig) error
作用: 更新 VAD 配置
注意: 会重新初始化 recorder

7️⃣ 命令处理方法（1个）
processCommandPlaceholder(transcript string) string (内部)
作用: 命令处理占位符
当前实现: 只返回简单的提示信息
未来: 将集成 LLM + Function Calling

8️⃣ 资源管理方法（1个）
Close() error
作用: 释放所有资源
重要: 必须调用！

🎯 方法调用关系图
💡 设计模式
1. 外观模式 (Facade Pattern)
AudioHandler 作为外观，隐藏了底层复杂性：

tools/audio 的音频处理细节
llms/asr 的 API 调用细节
2. 策略模式 (Strategy Pattern)
asrService 是接口，可以切换不同的实现：

OpenAI Whisper
Google Speech-to-Text
Azure Speech
本地 Whisper
3. 依赖注入 (Dependency Injection)
通过 AudioHandlerConfig 注入依赖，便于测试和扩展

🎉 总结
AudioHandler 的核心价值:

✅ 统一入口 - 前端只需要知道 AudioHandler
✅ 职责清晰 - 只负责协调，不实现具体功能
✅ 高度解耦 - 各层独立，易于替换
✅ 易于测试 - 依赖接口，可以 mock
✅ 便于扩展 - 添加新功能不影响现有代码
这个文件是整个语音控制系统的大脑，负责协调和调度所有组件！🧠