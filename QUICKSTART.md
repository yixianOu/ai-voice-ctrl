# 🎤 AI 语音控制助手 - 快速开始

## ⚡ 5 分钟上手

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

## 📂 录音文件说明

- 保存格式：`WAV`（44.1kHz，16-bit，立体声）
- 保存目录：`/tmp/voice_<时间戳>.wav`
- 可以使用 `aplay` 或其它播放器试听
- 录音结束后可将文件传入任意 STT（语音转文本）服务处理

---
综合来看，**`github.com/gen2brain/malgo`** 是最推荐的选择，原因如下：

## 为什么选择 malgo？

### ✅ 优势

1. **零外部依赖** - 基于 miniaudio（单头文件 C 库），编译时自动包含，无需预装任何音频库
2. **真正跨平台** - Windows/macOS/Linux 开箱即用，只需系统有 C 编译器（Go 默认就有 cgo）
3. **API 现代简洁** - 接口设计清晰，录音/播放代码都很直观
4. **活跃维护** - 最近仍在更新，社区支持较好
5. **性能好** - miniaudio 底层优化良好，延迟低

### ❌ PortAudio 的劣势

虽然 PortAudio 是老牌方案，但：
- 需要**手动安装原生库**（对新手不友好）
- 不同平台安装方式不同，增加部署复杂度
- Go binding 维护相对不活跃

---

## 快速上手示例

### 1. 安装
```bash
go get -u github.com/gen2brain/malgo
```

### 2. 录音到 WAV 文件（完整示例）

```go
package main

import (
    "encoding/binary"
    "fmt"
    "os"
    "os/signal"
    "github.com/gen2brain/malgo"
)

func main() {
    // 初始化音频上下文
    ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
    if err != nil {
        panic(err)
    }
    defer ctx.Uninit()

    // 配置录音参数
    deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
    deviceConfig.Capture.Format = malgo.FormatS16    // 16位采样
    deviceConfig.Capture.Channels = 1                 // 单声道
    deviceConfig.SampleRate = 44100                   // 44.1kHz
    deviceConfig.Alsa.NoMMap = 1

    // 用于存储录音数据
    var recordedData []byte

    // 数据回调函数
    onRecvFrames := func(outputSample, inputSamples []byte, framecount uint32) {
        recordedData = append(recordedData, inputSamples...)
    }

    // 启动录音设备
    captureCallbacks := malgo.DeviceCallbacks{
        Data: onRecvFrames,
    }
    device, err := malgo.InitDevice(ctx.Context, deviceConfig, captureCallbacks)
    if err != nil {
        panic(err)
    }

    err = device.Start()
    if err != nil {
        panic(err)
    }
    defer device.Uninit()

    fmt.Println("开始录音... 按 Ctrl+C 停止")

    // 等待中断信号
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    <-c

    // 停止录音
    device.Stop()
    fmt.Println("录音结束，正在保存...")

    // 保存为 WAV 文件
    saveWAV("output.wav", recordedData, 44100, 1, 16)
    fmt.Println("已保存到 output.wav")
}

// 简单的 WAV 文件写入函数
func saveWAV(filename string, data []byte, sampleRate, channels, bitsPerSample uint32) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()

    dataSize := uint32(len(data))
    
    // WAV 头
    f.WriteString("RIFF")
    binary.Write(f, binary.LittleEndian, uint32(36+dataSize))
    f.WriteString("WAVE")
    
    // fmt 子块
    f.WriteString("fmt ")
    binary.Write(f, binary.LittleEndian, uint32(16))        // 子块大小
    binary.Write(f, binary.LittleEndian, uint16(1))         // 音频格式（PCM）
    binary.Write(f, binary.LittleEndian, uint16(channels))
    binary.Write(f, binary.LittleEndian, sampleRate)
    binary.Write(f, binary.LittleEndian, sampleRate*channels*bitsPerSample/8) // 字节率
    binary.Write(f, binary.LittleEndian, uint16(channels*bitsPerSample/8))    // 块对齐
    binary.Write(f, binary.LittleEndian, uint16(bitsPerSample))
    
    // data 子块
    f.WriteString("data")
    binary.Write(f, binary.LittleEndian, dataSize)
    f.Write(data)
    
    return nil
}
```

### 3. 运行测试
```bash
go run main.go
# 说话几秒后按 Ctrl+C，会生成 output.wav
```

---

## 其他库对比

| 库 | 优势 | 劣势 | 推荐度 |
|---|---|---|---|
| **malgo** | 零依赖、跨平台、现代 API | 相对较新 | ⭐⭐⭐⭐⭐ |
| portaudio | 成熟稳定、功能全面 | 需手动安装原生库 | ⭐⭐⭐ |
| go-wav | 仅处理文件格式 | 不能录音，需配合其他库 | ⭐⭐ |

---

## 总结

**直接用 `malgo`** - 对于你的需求（跨平台录音，无需 shell 调用），它是最简单、最可靠的方案。上面的示例代码可直接运行，无需任何额外配置。

需要我帮你创建一个完整的项目示例或解答其他问题吗？
---
