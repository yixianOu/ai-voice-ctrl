# App.go 迁移指南

## 📋 更新概述

`app.go` 已经更新以适配重构后的 `AudioHandler` 架构。所有原有的方法都已保留并增强，确保向后兼容性。

## 🔄 主要变更

### 1. 核心变量重命名

```go
// 之前 ❌
type App struct {
    ctx          context.Context
    voiceHandler *handler.VoiceHandler
}

// 现在 ✅
type App struct {
    ctx          context.Context
    audioHandler *handler.AudioHandler  // 重命名并使用新的 AudioHandler
}
```

### 2. 添加资源清理

```go
// 新增 shutdown 方法
func (a *App) shutdown(ctx context.Context) {
    if a.audioHandler != nil {
        a.audioHandler.Close()  // 确保资源被正确释放
    }
}
```

### 3. Context 传递

所有异步操作现在都使用 `a.ctx` 进行 context 传递：

```go
// 之前 ❌
func (a *App) RecordAndTranscribe() (string, error) {
    return a.voiceHandler.RecordAndTranscribe()  // 没有 context
}

// 现在 ✅
func (a *App) RecordAndTranscribe() (string, error) {
    return a.audioHandler.RecordAndTranscribe(a.ctx)  // 支持超时和取消
}
```

## 📊 API 对照表

### 音频录制方法（完全兼容）

| 方法 | 变更 | 说明 |
|------|------|------|
| `StartRecording()` | ✅ 兼容 | 手动开始录制 |
| `StopRecording()` | ✅ 兼容 | 手动停止录制 |
| `RecordAudio(seconds)` | ✅ 兼容 | 定时录制 |
| `RecordAudioWithVAD()` | ✅ 兼容 | VAD 自动录制 |

### 语音识别方法（增强）

| 方法 | 变更 | 说明 |
|------|------|------|
| `RecordAndTranscribe()` | ✅ 兼容 + 增强 | 添加了 context 支持 |
| `RecordAndTranscribeWithLanguage(lang)` | ✨ 新增 | 指定语言的转录 |
| `TranscribeAudioData(data)` | ✨ 新增 | 转录已有音频数据 |

### 工作流方法

| 方法 | 变更 | 说明 |
|------|------|------|
| `RecordTranscribeAndProcess()` | ⚠️ 弃用 | 保留但标记为过时 |
| `ExecuteVoiceCommandWorkflow()` | ✨ 新增 | 推荐使用的新方法 |
| `ExecuteVoiceCommandWorkflowWithLanguage(lang)` | ✨ 新增 | 带语言参数的工作流 |

### 状态查询方法

| 方法 | 变更 | 说明 |
|------|------|------|
| `GetRecordingStatus()` | ✅ 兼容 | 获取录制状态 |
| `IsRecording()` | ✨ 新增 | 检查是否正在录制 |
| `GetLastTranscript()` | ✨ 新增 | 获取上次转录结果 |
| `GetASRServiceName()` | ✨ 新增 | 获取当前 ASR 服务名称 |

### 配置方法

| 方法 | 变更 | 说明 |
|------|------|------|
| `SetAPIKey(key)` | ✅ 增强 | 现在使用 ASR 服务切换 |

## 🎯 新增功能详解

### 1. 带语言参数的转录

```go
// 前端调用示例
text, err := app.RecordAndTranscribeWithLanguage("zh")
```

**优势**:
- 提高中文识别准确度
- 支持多语言应用
- 减少前端传参复杂度

### 2. 转录已有音频数据

```go
// 前端上传音频文件后转录
text, err := app.TranscribeAudioData(audioBytes)
```

**用途**:
- 处理从文件上传的音频
- 处理从网络接收的音频
- 批量处理音频文件

### 3. 完整工作流方法

```go
// 执行完整的语音命令工作流
transcript, commandResult, err := app.ExecuteVoiceCommandWorkflow()
```

**特点**:
- 一次调用完成录制、转录、处理
- 返回结构化结果
- 包含时间追踪信息

### 4. 状态查询增强

```go
// 检查是否正在录制
isRecording := app.IsRecording()

// 获取上次转录结果
lastText := app.GetLastTranscript()

// 查看当前 ASR 服务
serviceName := app.GetASRServiceName()  // "OpenAI Whisper"
```

## 📝 方法分类

### Layer 1: 音频录制（基础操作）
```
StartRecording()
StopRecording()
RecordAudio(seconds)
RecordAudioWithVAD()
```

### Layer 2: 语音识别（协调操作）
```
RecordAndTranscribe()
RecordAndTranscribeWithLanguage(language)
TranscribeAudioData(audioData)
```

### Layer 3: 完整工作流（高级操作）
```
ExecuteVoiceCommandWorkflow()
ExecuteVoiceCommandWorkflowWithLanguage(language)
```

### 辅助方法
```
GetRecordingStatus()
IsRecording()
GetLastTranscript()
GetASRServiceName()
SetAPIKey(apiKey)
ProcessVoiceCommand(command)  // 占位符，待实现
```

## 🔄 迁移步骤

### 前端代码无需修改的情况

以下前端调用**完全兼容**，无需修改：

```typescript
// ✅ 这些调用仍然有效
await StartRecording()
await StopRecording()
await RecordAudio(5)
await RecordAudioWithVAD()
await RecordAndTranscribe()
await GetRecordingStatus()
```

### 推荐的前端更新

虽然不是必须的，但建议使用新方法：

```typescript
// 旧方法（仍可用）
const [transcript, result, err] = await RecordTranscribeAndProcess()

// 新方法（推荐）✨
const [transcript, result, err] = await ExecuteVoiceCommandWorkflow()

// 带语言参数（新功能）✨
const [transcript, result, err] = await ExecuteVoiceCommandWorkflowWithLanguage("zh")
```

## 🎨 前端集成示例

### React/TypeScript 组件示例

```typescript
import { 
    RecordAndTranscribe, 
    ExecuteVoiceCommandWorkflow,
    IsRecording,
    GetLastTranscript 
} from '../wailsjs/go/main/App';

function VoiceControl() {
    const [isRecording, setIsRecording] = useState(false);
    const [transcript, setTranscript] = useState('');
    const [result, setResult] = useState('');

    // 方法 1: 简单转录
    const handleSimpleRecord = async () => {
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

    // 方法 2: 完整工作流（推荐）
    const handleFullWorkflow = async () => {
        try {
            setIsRecording(true);
            const [text, cmdResult, error] = await ExecuteVoiceCommandWorkflow();
            if (error) throw error;
            setTranscript(text);
            setResult(cmdResult);
        } catch (error) {
            console.error('工作流失败:', error);
        } finally {
            setIsRecording(false);
        }
    };

    // 方法 3: 带语言参数
    const handleChineseRecord = async () => {
        try {
            setIsRecording(true);
            const text = await RecordAndTranscribeWithLanguage("zh");
            setTranscript(text);
        } catch (error) {
            console.error('录制失败:', error);
        } finally {
            setIsRecording(false);
        }
    };

    return (
        <div>
            <button onClick={handleSimpleRecord} disabled={isRecording}>
                简单录制
            </button>
            <button onClick={handleFullWorkflow} disabled={isRecording}>
                完整工作流
            </button>
            <button onClick={handleChineseRecord} disabled={isRecording}>
                中文录制
            </button>
            
            {transcript && <p>转录: {transcript}</p>}
            {result && <p>结果: {result}</p>}
        </div>
    );
}
```

## ⚠️ 重要注意事项

### 1. Context 超时控制

现在所有异步操作都支持 context 超时，但默认使用 app 的全局 context：

```go
// 在 app.go 中，如果需要自定义超时
func (a *App) RecordAndTranscribeWithTimeout(timeoutSeconds int) (string, error) {
    ctx, cancel := context.WithTimeout(a.ctx, time.Duration(timeoutSeconds)*time.Second)
    defer cancel()
    return a.audioHandler.RecordAndTranscribe(ctx)
}
```

### 2. 资源清理

新增的 `shutdown` 方法会在应用关闭时自动调用，确保音频设备被正确释放。

### 3. API Key 更新机制

`SetAPIKey` 现在会创建新的 ASR 服务实例：

```go
// 内部实现
func (a *App) SetAPIKey(apiKey string) {
    newASR := openaillms.NewOpenAIASR(apiKey)
    a.audioHandler.SetASRService(newASR)
}
```

## 🚀 未来扩展

### 预留的扩展点

```go
// ProcessVoiceCommand 目前是占位符
// 未来将集成 LLM 和 function calling
func (a *App) ProcessVoiceCommand(command string) string {
    // TODO: 
    // 1. 将 command 发送给 LLM (GPT-4, Claude, etc.)
    // 2. LLM 返回 function call
    // 3. 执行本地函数
    // 4. 返回执行结果
}
```

### 可能添加的新方法

```go
// 实时流式识别（未来）
func (a *App) StartStreamingRecognition() error

// 多轮对话支持（未来）
func (a *App) ContinueConversation(context string) (string, error)

// 自定义 prompt（未来）
func (a *App) ExecuteWithPrompt(promptTemplate string) (string, error)
```

## 📊 性能对比

| 操作 | 旧实现 | 新实现 | 改进 |
|------|--------|--------|------|
| 简单转录 | ~2s | ~2s | 无变化 |
| Context 支持 | ❌ | ✅ | 可取消/超时 |
| 资源管理 | 手动 | 自动 | 防止泄漏 |
| 错误信息 | 基础 | 详细 | 更易调试 |
| 扩展性 | 低 | 高 | 易于添加功能 |

## ✅ 兼容性检查表

- [x] 所有原有方法保留
- [x] 方法签名兼容
- [x] 返回值格式兼容
- [x] 错误处理兼容
- [x] 前端无需修改即可运行
- [x] 新功能可选使用
- [x] 向后兼容

## 📚 相关文档

- [AUDIO_HANDLER_QUICKSTART.md](./AUDIO_HANDLER_QUICKSTART.md) - AudioHandler 快速上手
- [AUDIO_HANDLER_ARCHITECTURE.md](./AUDIO_HANDLER_ARCHITECTURE.md) - 架构详解
- [AUDIO_HANDLER_REFACTOR.md](./AUDIO_HANDLER_REFACTOR.md) - 重构总结

---

**总结**: 这次更新完全向后兼容，前端代码无需修改即可继续工作。同时提供了更强大的新功能和更好的架构设计，建议在新功能开发时采用新的 API。
