# Audio Handler 使用示例

本目录包含了 AudioHandler 的完整使用示例。

## 🚀 快速开始

### 1. 设置 OpenAI API Key

在运行示例之前，您需要设置 `OPENAI_API_KEY` 环境变量。

#### Windows (PowerShell)
```powershell
$env:OPENAI_API_KEY="your-api-key-here"
```

#### Windows (CMD)
```cmd
set OPENAI_API_KEY=your-api-key-here
```

#### Linux/Mac (bash/zsh)
```bash
export OPENAI_API_KEY=your-api-key-here
```

### 2. 配置代理（可选，国内用户推荐）⭐

如果您遇到网络连接问题（如 "dial tcp ... connectex: A connection attempt failed"），需要配置代理。

#### Windows (PowerShell)
```powershell
# HTTP 代理
$env:HTTP_PROXY="http://127.0.0.1:7890"
$env:HTTPS_PROXY="http://127.0.0.1:7890"

# 或 SOCKS5 代理（如果您的代理支持）
$env:HTTP_PROXY="socks5://127.0.0.1:7890"
$env:HTTPS_PROXY="socks5://127.0.0.1:7890"
```

#### Windows (CMD)
```cmd
set HTTP_PROXY=http://127.0.0.1:7890
set HTTPS_PROXY=http://127.0.0.1:7890
```

#### Linux/Mac
```bash
export HTTP_PROXY=http://127.0.0.1:7890
export HTTPS_PROXY=http://127.0.0.1:7890
```

**常用代理端口**:
- **Clash**: `7890` (HTTP) / `7891` (SOCKS5)
- **V2Ray**: `10809` (HTTP) / `10808` (SOCKS5)
- **Shadowsocks**: `1080` (SOCKS5)

**⚠️ 重要**: 
1. 将 `127.0.0.1:7890` 替换为您实际的代理地址和端口
2. 确保您的代理软件正在运行
3. 测试代理是否工作：`curl -x http://127.0.0.1:7890 https://api.openai.com`

### 3. 运行示例

```bash
cd examples
go run .
```

## 🔧 完整配置示例

```powershell
# 1. 设置 API Key
$env:OPENAI_API_KEY="sk-proj-xxxxx..."

# 2. 设置代理（国内用户必需）
$env:HTTP_PROXY="http://127.0.0.1:7890"
$env:HTTPS_PROXY="http://127.0.0.1:7890"

# 3. 运行
cd D:\Code_Project\golang\ai-voice-ctrl\examples
go run .
```

## 📚 示例列表

文件 `audio_handler_usage.go` 包含 12 个完整示例：

### 基础示例

1. **Example_BasicVoiceInput** - 基本语音输入
2. **Example_VoiceInputWithLanguage** - 指定语言（中文）✅ 当前运行
3. **Example_CompleteWorkflow** - 完整工作流

### 高级控制

4. **Example_ManualRecording** - 手动控制录音
5. **Example_FixedDurationRecording** - 固定时长录音
6. **Example_TranscribeFromFile** - 从文件转录

### 配置和定制

7. **Example_CustomVADConfiguration** - 自定义 VAD 配置
8. **Example_SwitchASRService** - 切换 ASR 服务
9. **Example_UpdateVADAtRuntime** - 运行时更新 VAD

### 状态和监控

10. **Example_CheckRecordingStatus** - 检查录音状态
11. **Example_AccessLastRecording** - 访问最后的录音
12. **Example_WorkflowWithDetails** - 详细的转录结果

## ⚠️ 常见问题

### 1. "错误: ASR service not configured"

**原因**: 未设置 `OPENAI_API_KEY` 环境变量

**解决方案**: 按照上面的步骤设置环境变量

### 2. "dial tcp ... connectex: A connection attempt failed" ⭐

**原因**: 无法连接到 OpenAI API（网络问题）

**解决方案**: 
1. **配置代理**（推荐）：设置 `HTTP_PROXY` 和 `HTTPS_PROXY` 环境变量
2. 检查网络连接
3. 检查防火墙设置
4. 尝试禁用 IPv6（如果使用）

**配置代理示例**:
```powershell
# 如果您使用 Clash
$env:HTTP_PROXY="http://127.0.0.1:7890"
$env:HTTPS_PROXY="http://127.0.0.1:7890"

# 测试代理是否工作
curl -x http://127.0.0.1:7890 https://api.openai.com
```

### 3. "错误: 401 Unauthorized"

**原因**: API Key 无效或过期

**解决方案**: 
- 检查 API Key 是否正确
- 在 OpenAI 控制台验证 Key 是否有效
- 确保账户有足够的额度

### 4. 录音无声音

**原因**: 
- 麦克风权限未授予
- 使用了错误的麦克风设备
- VAD 阈值设置过高

**解决方案**:
- 检查系统麦克风权限
- 尝试使用不同的 VAD 配置（Example 7）
- 使用固定时长录音测试（Example 5）

### 5. 识别不准确

**解决方案**:
- 指定正确的语言参数（`Language: "zh"` 用于中文）
- 调整 Temperature 参数（0-1，越低越保守）
- 确保录音环境安静
- 说话清晰，避免过快

## 🌐 代理配置详解

### 如何找到代理端口？

**Clash**:
1. 打开 Clash
2. 点击 "常规" 或 "General"
3. 查看 "端口" 或 "Port"，通常是 `7890`

**V2Ray**:
1. 打开 V2Ray/V2RayN
2. 查看 "参数设置" → "本地监听端口"
3. HTTP 代理通常是 `10809`

**Shadowsocks**:
1. 打开 Shadowsocks
2. 右键托盘图标 → "选项设置"
3. 查看 "本地代理" 端口，通常是 `1080`

### 验证代理是否工作

```powershell
# 测试代理连接
curl -x http://127.0.0.1:7890 https://api.openai.com/v1/models

# 如果成功，会返回 JSON 数据
# 如果失败，检查代理地址和端口是否正确
```

### 临时 vs 永久配置

**临时配置**（当前终端有效）:
```powershell
$env:HTTP_PROXY="http://127.0.0.1:7890"
```

**永久配置**（所有终端有效）:
1. Win + R → `sysdm.cpl`
2. 高级 → 环境变量
3. 新建用户变量：
   - 变量名: `HTTP_PROXY`
   - 变量值: `http://127.0.0.1:7890`
4. 同样添加 `HTTPS_PROXY`

## 🎯 运行特定示例

如果您想运行特定的示例，可以修改 `main()` 函数：

```go
func main() {
    // 运行示例 1
    Example_BasicVoiceInput()
    
    // 或运行示例 3
    Example_CompleteWorkflow()
    
    // 或运行多个示例
    Example_BasicVoiceInput()
    Example_VoiceInputWithLanguage()
}
```

## 📖 更多文档

- [AudioHandler 快速上手](../AUDIO_HANDLER_QUICKSTART.md)
- [AudioHandler 架构设计](../AUDIO_HANDLER_ARCHITECTURE.md)
- [AudioHandler 设计理念](../AUDIO_HANDLER_DESIGN.md)
- [重构总结](../AUDIO_HANDLER_REFACTOR.md)

## 🆘 获取帮助

如果遇到问题：

1. **网络问题** → 配置代理（见上方详细说明）
2. **API Key 问题** → 检查环境变量设置
3. **录音问题** → 检查麦克风权限
4. 查看上面的常见问题
5. 阅读相关文档

## 🎉 开始探索

### 快速开始（国内用户）

```powershell
# 1. 设置 API Key
$env:OPENAI_API_KEY="sk-proj-xxxxx..."

# 2. 设置代理（重要！）
$env:HTTP_PROXY="http://127.0.0.1:7890"
$env:HTTPS_PROXY="http://127.0.0.1:7890"

# 3. 进入目录
cd D:\Code_Project\golang\ai-voice-ctrl\examples

# 4. 运行示例
go run .
```

看到 "开始说话..." 提示后，开始录音并说话！

祝您使用愉快！🚀
