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