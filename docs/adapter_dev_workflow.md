# 外部应用适配层开发与测试流程

## 1. 规划阶段
- 明确需要支持的应用与操作清单（例如 VLC 播放、VS Code 文件操作）。
- 为每类操作定义统一接口签名与错误约定，确认输入输出格式。
- 列出平台依赖（CLI、HTTP API、配置文件），确保开发环境已安装并可手动调用。

## 2. 基础骨架搭建
- 在 `backend` 中只保留接口定义与调用入口，实际实现下沉至 Node sidecar。
- 设计 JSON-RPC / HTTP 消息格式，约定命令名称、参数、错误码。
- 准备 TypeScript 项目脚手架（如 `pnpm init`），创建基础目录结构与共享类型定义。

## 3. Node Sidecar 实现
- 按应用拆分模块，如 `controllers/vlc.ts`、`controllers/vscode.ts`，封装 CLI/HTTP 调用。
- 用 TypeScript 编写统一的命令路由器，将来自 Go 的请求分发给具体控制器。
- 处理参数校验、超时、错误分类，并把结果序列化回传给 Go。

## 4. 单元与集成测试
- 使用 Jest / Vitest 为 TypeScript 控制器编写单元测试，模拟命令执行与错误场景。
- 提供可选的集成测试脚本（标记 `npm run test:integration`），在本地真实调用外部应用时执行。
- 实现一个命令行入口（如 `node scripts/drive-command.ts`），便于手工触发 sidecar 指令。

## 5. 日志与观测
- 在适配层统一记录执行日志（命令、参数、返回状态），方便后续调试与 LLM 对接。
- Node 侧输出结构化日志（JSON），Go 侧转储或上报，便于集中分析。

## 6. 与 Go 后端对接
- Go 后端提供 `SidecarClient`，负责管理 Node 进程、心跳与重试。
- 所有已有的 Go 方法改为调用 `SidecarClient.Call(ctx, command, payload)`，保持向前兼容。
- 预留 `ExecuteCommand` 统一入口，未来可直接映射到 MCP Handler。

## 7. 联调与回归
- 使用 Wails 前端 / CLI 脚本调用 Go 后端方法，确认 sidecar 通路正常。
- 编写回归清单：命令路由、错误回传、外部应用状态变化、重启 sidecar 行为。
- 在 LLM 调度上线前，用固定 JSON 指令驱动 `SidecarClient` 验证端到端流程。
