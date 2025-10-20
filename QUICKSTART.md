# 🎤 AI 语音控制助手 - 快速开始

## ⚡ 5 分钟上手

### 1️⃣ 安装系统依赖（Linux）

```bash
sudo apt update
sudo apt install -y alsa-utils pulseaudio-utils playerctl xdg-utils
```

确认麦克风设备：

```bash
arecord -l
```

### 2️⃣ 安装前端依赖

```bash
cd frontend
pnpm install
```

### 3️⃣ 启动应用

```bash
cd ..
wails dev
```

或使用脚本：

```bash
./start.sh
```

### 4️⃣ 使用录音

1. 在界面中点击「🎤 开始录音」
2. 对着麦克风讲话
3. 点击「🔴 停止录音」结束
4. 成功后会提示录音文件路径（位于 `/tmp/voice_*.wav`）

> 当前仅实现音频录制，尚未集成语音识别（STT）。

### 5️⃣ 测试命令

界面底部提供测试按钮，可直接调用后端命令：
- 「打开火狐」
- 「音量50%」
- 「播放音乐」
- 「下一首」

如果要手动执行命令，可在前端输入框中调用 `ProcessVoiceCommand`。

---

## 🧪 验证麦克风

可以在终端直接测试录音链路：

```bash
arecord -q -f cd -t wav /tmp/test.wav
# Ctrl+C 停止后播放
aplay /tmp/test.wav
```

如果能正常录音播放，Wails 应用也能使用系统麦克风。

---

## � 常见问题

| 问题 | 解决方案 |
|------|-----------|
| 启动提示 `arecord: command not found` | 安装 `alsa-utils` |
| 录音失败，返回「设备忙」 | 确认没有其他程序占用麦克风（如 Zoom、OBS） |
| `StopRecording` 报错 | 检查是否已停止或录音进程异常退出 |
| 没生成音频文件 | 确认 `/tmp` 目录有写权限，或修改代码中的文件保存路径 |

---

## 📂 录音文件说明

- 保存格式：`WAV`（44.1kHz，16-bit，立体声）
- 保存目录：`/tmp/voice_<时间戳>.wav`
- 可以使用 `aplay` 或其它播放器试听
- 录音结束后可将文件传入任意 STT（语音转文本）服务处理
