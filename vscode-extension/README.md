# AI Voice Control Bridge Extension

轻量级 VSCode 扩展，为 Go 后端提供 HTTP 桥接接口。

## 功能

- 打开/关闭文件
- 刷新文件内容
- 关闭 VSCode 窗口
- 执行任意 VSCode 命令

## 安装

```bash
cd vscode-extension
npm install
npm run compile
```

在 VSCode 中按 F5 启动调试，或打包安装：

```bash
npx vsce package
code --install-extension ai-voice-ctrl-bridge-0.1.0.vsix
```

## 配置

- `aiVoiceCtrlBridge.port`: HTTP 服务端口 (默认 9527)
- `aiVoiceCtrlBridge.autoStart`: 启动时自动开启服务 (默认 true)

## API

### 打开文件
```json
POST http://localhost:9527
{
  "action": "open_file",
  "params": {
    "path": "/path/to/file.txt",
    "line": 10
  }
}
```

### 关闭文件
```json
POST http://localhost:9527
{
  "action": "close_file",
  "params": {
    "path": "/path/to/file.txt"
  }
}
```

### 刷新文件
```json
POST http://localhost:9527
{
  "action": "refresh_file",
  "params": {
    "path": "/path/to/file.txt"
  }
}
```

### 关闭窗口
```json
POST http://localhost:9527
{
  "action": "close_window",
  "params": {}
}
```

## Go 后端集成

```go
tool, _ := executor.NewVSCodeHTTPTool("/workspace", 9527)
tool.RegisterHTTPLifecycle(exec)
```
