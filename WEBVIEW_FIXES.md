# WebView 问题解决方案

## 问题说明

在 Linux 上使用 Wails WebView 时，可能遇到以下问题：
1. 输入框无法编辑（浏览器正常，客户端不行）
2. 启动时出现警告和错误信息

## 已实施的解决方案

### 1. 输入框编辑问题

**原因**：WebKit WebView 在某些情况下会禁用文本输入。

**解决方案**：
- 在 `VoiceControl.tsx` 中移除了 `disabled={isProcessing}` 属性
- 在 `VoiceControl.css` 中添加了 WebView 兼容性 CSS：
  ```css
  -webkit-user-select: text;
  user-select: text;
  -webkit-appearance: none;
  cursor: text;
  ```

### 2. 启动警告和错误

#### 错误 1: "Overriding existing handler for signal 10"

**原因**：WebKit 使用信号 10 进行垃圾回收，与系统信号冲突。

**解决方案**：
设置环境变量 `JSC_SIGNAL_FOR_GC=10`

#### 错误 2: "Unknown message from front end: runtime:ready"

**原因**：Wails v2 的已知问题，WebView 和 DevServer 之间的消息处理竞态。

**解决方案**：
- 在 `main.go` 中添加了自定义 logger
- 在 `wails.json` 中设置了 `debounceMS: 100`

#### 错误 3: "[ExternalAssetHandler] Proxy error: context canceled"

**原因**：DevServer 热重载时的正常清理行为。

**解决方案**：这是正常的，可以忽略（在生产构建中不会出现）。

### 3. GTK 输入法支持

**解决方案**：设置环境变量 `GTK_IM_MODULE=ibus`

## 使用方法

### 方法 1：使用启动脚本（推荐）

```fish
./dev.fish
```

### 方法 2：手动设置环境变量

```fish
set -x JSC_SIGNAL_FOR_GC 10
set -x GTK_IM_MODULE ibus
wails dev
```

### 方法 3：在浏览器中开发（最稳定）

启动 wails dev 后，在浏览器中访问：http://localhost:34115

这样可以避免所有 WebView 相关问题，同时仍然可以调用 Go 后端方法。

## 生产构建

生产构建不会有这些开发时的问题：

```bash
wails build
```

## 额外优化

如果仍然遇到问题，可以尝试：

1. **禁用 GPU 加速**：
   ```fish
   set -x WEBKIT_DISABLE_COMPOSITING_MODE 1
   ```

2. **使用不同的 GTK 主题**：
   某些 GTK 主题可能与 WebView 不兼容

3. **更新依赖**：
   ```bash
   sudo pacman -S webkit2gtk # Arch/Manjaro
   sudo apt install libwebkit2gtk-4.0-dev # Ubuntu/Debian
   ```

## 参考

- [Wails Issue #1769](https://github.com/wailsapp/wails/issues/1769) - WebView input issues
- [Wails Issue #1891](https://github.com/wailsapp/wails/issues/1891) - Signal 10 warning
- [Wails Docs - Linux](https://wails.io/docs/reference/options#linux)
