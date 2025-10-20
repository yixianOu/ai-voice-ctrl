# README

## About

This is the official Wails React-TS template.

You can configure the project by editing `wails.json`. More information about the project settings can be found
here: https://wails.io/docs/reference/project-config

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.

---

## 🏗️ 完整架构建议

```mermaid
flowchart TB
    subgraph Wails[Wails 桌面应用]
        subgraph UI[React WebView]
            View["用户界面<br/>- 展示执行结果"]
        end

        subgraph Backend[Go Backend]
            Recorder["AudioRecorder (malgo)"]
            Relay["后端调度 / LLM 中转"]
            Adapter["VSCode / VLC 控制适配层"]
        end
    end

    subgraph STTService[语音识别服务]
        STT["Whisper / Azure Speech"]
    end

    subgraph LLM[大模型理解层]
        GPT["LLM<br/>(GPT-4 / Claude)"]
    end

    subgraph Clients[外部应用]
        VLC["VLC<br/>HTTP API"]
        VSCode["VS Code<br/>CLI"]
    end

    View -->|录音控制| Recorder
    Recorder --> Relay
    Relay --> STT
    STT --> Relay
    Relay --> GPT
    GPT --> Relay

    View <-->|界面请求| Relay
    Relay --> Adapter
    Adapter --> VLC
    Adapter --> VSCode
    Relay --> View
```
