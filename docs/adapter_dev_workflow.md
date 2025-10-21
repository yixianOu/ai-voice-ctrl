# 外部应用适配层开发与测试流程

## 1. 规划阶段
- 明确需要支持的应用与操作清单（例如 VLC 播放、VS Code 文件操作）。
- 为每类操作定义统一接口签名与错误约定，确认输入输出格式。
- 列出平台依赖（CLI、HTTP API、配置文件），确保开发环境已安装并可手动调用。

## 2. 基础骨架搭建
- 在 Go 后端定义面向 LLM 的工具接口（函数名称、参数结构、返回值）。
- 设计 Function Calling 使用的 JSON Schema，确保可被 OpenAI/Anthropic 解析。
- 预先划分命令类别（播放音乐、编辑文件等），与外部应用操作一一映射。

## 3. 适配函数实现
- 在 Go 中实现各个工具函数，内部调用现有的外部应用控制逻辑
- 确保函数返回结构化结果或错误信息，方便直接回填到 LLM 对话中。
- 统一采用 `ToolRegistry` 模式：`map[string]ToolDefinition` 存储函数名称、描述、参数 Schema 与执行函数。
- 工具函数签名推荐 `func(ctx context.Context, payload json.RawMessage) (Result, error)`，先解码参数再执行业务。
  
## 4. 单元测试
- 为 Go 工具函数编写单元测试，覆盖参数校验、错误路径、外部调用代理等逻辑。

## 5. 日志与观测
- 在适配层统一记录执行日志（命令、参数、返回状态），方便后续调试与 LLM 对接。

## 6. LLM 集成
- 构造 Function Calling 请求：包含用户输入、工具列表（JSON Schema）以及历史上下文。
- 当 LLM 返回 `tool_calls` 时，调用对应的 Go 函数并把结果作为 `tool` 消息再次发送。
- 封装一个统一的调度逻辑，管理多轮工具调用、异常重试与用户回显。

## 7. 联调与回归
- 编写回归清单：Schema 更新、函数签名、外部应用状态、LLM 调度重试。
- 在正式接入 LLM 前，先用假数据模拟 Function Calling 循环，验证端到端流程。

如果用function call的话，那么工具调用的接口要怎么定义？
**定义步骤**  
- **梳理业务动作**：先列出每个外部操作（播放音乐、暂停、打开文件等），每个动作对应一个 "工具函数"。  
- **命名规范**：函数名用蛇形或小驼峰，例如 `play_music`, `open_vscode_file`，保持语义清晰。  
- **参数 Schema**：为每个工具写 JSON Schema：  
  - `type: object`；  
  - `properties` 描述每个字段的类型、含义；  
  - `required` 指出必填参数；  
  - 可选字段写在 `properties` 里但不列入 `required`。  
- **返回结构**：函数返回一个结构化结果（建议 Go 里返回 `struct{ Success bool; Data any; Message string }`），LLM 读取 `Success`/`Message` 即可决定下一步。  
- **错误处理**：函数遇到异常时返回成功为 `false`、附带可读错误文本；避免直接 panic，以免中断整个 Function Calling 流程。  
- **Go 端实现**：  
  - 定义统一接口 `type Tool func(ctx context.Context, payload json.RawMessage) (ToolResult, error)`；  
  - 在工具注册表中按名称存放实现；  
  - 调度层收到 `tool_calls` 后用名称查找实现，先解析参数 JSON，再调用具体函数。  
- **工具注册**：准备 `map[string]ToolDefinition`（含 Schema、描述），在向 LLM 发起请求时填入 `tools` 字段；只暴露需要的函数，控制最小权限面。  
- **测试**：  
  - 集成测试模拟 LLM 返回 `tool_calls` 的 JSON，验证调度层可以找到函数并返回结果。  

# Function Call 工具使用速览

面向上层调用者，我们仅需记住“三步走”：

1. **准备工具 (`ToolDefinition`)**  
  - 为每个能力定义名称、描述、JSON Schema，以及 `ToolExecutor`。  
  - `ToolExecutor` 签名固定为 `func(ctx context.Context, payload json.RawMessage) (ToolResult, error)`，负责解析参数并返回 `ToolResult`。

2. **注册到工具表 (`ToolRegistry`)**  
  - 每个应用实现 `ToolRegistry`，例如 `VSCodeToolRegistry`，统一返回 `map[string]ToolDefinition`。  
  - 上层只需要把 registry 交给执行器，具体路径校验、CLI 检查等由 registry 自己处理。

3. **通过执行器调用 (`Executor`)**  
  - 执行器汇总所有工具，提供给 LLM 注册使用。  
  - 当收到 `ToolCall` 时，执行器用名称定位 `ToolExecutor`，执行后把 `ToolResult` 回传给 LLM。

> 口诀：工具负责“干活”，Registry 负责“登记”，Executor 负责“转发给 LLM 并回调”。

## 典型串联

1. `NewVSCodeToolRegistry(workspace)` 生成 VS Code 工具集合；
2. `executor.RegisterRegistry("vscode", registry)` 把 VS Code 工具挂载到执行器；
3. 执行器将全部工具 Schema 提供给 LLM；
4. LLM 调用 `vscode_open_file` → 执行器路由到实际逻辑 → 用 `ToolResult` 反馈执行情况。

## 常见扩展点

- **新增工具**：在对应 registry 中追加一个 `ToolDefinition` 即可；
- **新增应用**：实现新的 `ToolRegistry` 并注册；
- **能力控制**：执行器可按需启用/禁用某个 registry，实现权限或平台隔离。

工具的生命周期管理，code cli功能测试，llm按需创建工具，工具注册的key为method_name_PID（但是方法是一样的），关闭则对PID发送关闭命令。要注册到executor的是对象实例？
llm按需创建工具，能否优雅管理同类工具的多个实例？工具注册的key为method_name_PID，关闭则对PID发送关闭命令。是否可行？实现麻烦吗？llm的function call响应是什么？请帮我修改文档