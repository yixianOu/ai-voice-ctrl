# 基于 Wails + React 的 AI 电脑助手 - Go 后端设计文档

**技术栈：** Go 1.21+ | Wails v2 | React 18 | 通义千问 Qwen-Max

---

## 1. 整体架构设计

```
┌─────────────────────────────────────────────────────────┐
│              前端层 (React + TypeScript)                 │
│        聊天界面 | 历史记录 | 状态显示 | 快捷操作          │
└─────────────────────────────────────────────────────────┘
                      ↓ Wails Binding
┌─────────────────────────────────────────────────────────┐
│                 应用服务层 (app.go)                       │
│           统一入口 | 会话管理 | 结果聚合                  │
└─────────────────────────────────────────────────────────┘
        ↓                ↓                ↓
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  LLM 服务层   │  │  编排服务层   │  │  配置管理层   │
│  意图理解     │  │  任务拆解     │  │  环境检测     │
│  参数提取     │  │  执行调度     │  │  API 密钥     │
└──────────────┘  └──────────────┘  └──────────────┘
                       ↓
        ┌──────────────┴──────────────┐
        ↓                             ↓
┌──────────────────┐        ┌──────────────────┐
│   执行器管理器    │        │   上下文管理器    │
│  注册 | 路由     │        │  会话 | 变量     │
└──────────────────┘        └──────────────────┘
        ↓
┌─────────────────────────────────────────────────────────┐
│                    执行器层 (Executors)                   │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────┐ │
│  │系统控制  │  │音乐播放  │  │文件操作  │  │浏览器   │ │
│  └──────────┘  └──────────┘  └──────────┘  └─────────┘ │
└─────────────────────────────────────────────────────────┘
        ↓
┌─────────────────────────────────────────────────────────┐
│              系统接口层 (System Interface)                │
│     D-Bus | Shell Command | OS API | Process Control    │
└─────────────────────────────────────────────────────────┘
```

---

## 2. Go 项目结构

```
feiniuyun-oyx-yzy/
├── main.go                          # Wails 应用入口
├── app.go                           # 应用服务层（前端调用入口）
├── wails.json                       # Wails 配置
├── go.mod
├── go.sum
│
├── backend/
│   ├── llm/                         # LLM 服务层
│   │   ├── client.go               # 通义千问客户端
│   │   ├── function_defs.go        # Function Calling 定义
│   │   ├── prompt_builder.go       # Prompt 构建器
│   │   └── intent_parser.go        # 意图解析器
│   │
│   ├── orchestrator/                # 编排服务层
│   │   ├── manager.go              # 编排管理器
│   │   ├── task_planner.go         # 任务规划器
│   │   ├── executor.go             # 任务执行器
│   │   └── chain.go                # 任务链管理
│   │
│   ├── executors/                   # 执行器层
│   │   ├── interface.go            # 执行器接口定义
│   │   ├── registry.go             # 执行器注册表
│   │   │
│   │   ├── system/                 # 系统控制执行器
│   │   │   ├── app_launcher.go    # 应用启动
│   │   │   ├── volume_control.go  # 音量控制
│   │   │   ├── process_manager.go # 进程管理
│   │   │   └── system_info.go     # 系统信息
│   │   │
│   │   ├── music/                  # 音乐播放执行器
│   │   │   ├── player.go          # 播放器控制
│   │   │   ├── mpris_client.go    # MPRIS D-Bus 客户端
│   │   │   └── playlist.go        # 播放列表管理
│   │   │
│   │   ├── file/                   # 文件操作执行器
│   │   │   ├── file_ops.go        # 文件 CRUD
│   │   │   ├── search.go          # 文件搜索
│   │   │   └── content_edit.go    # 内容编辑
│   │   │
│   │   └── browser/                # 浏览器控制执行器
│   │       ├── launcher.go        # 浏览器启动
│   │       └── url_opener.go      # URL 打开
│   │
│   ├── context/                     # 上下文管理层
│   │   ├── session.go              # 会话管理
│   │   ├── variables.go            # 变量存储
│   │   └── history.go              # 历史记录
│   │
│   ├── system/                      # 系统接口层
│   │   ├── dbus_client.go          # D-Bus 客户端
│   │   ├── shell_executor.go       # Shell 命令执行
│   │   ├── process_util.go         # 进程工具
│   │   └── platform_detector.go    # 平台检测
│   │
│   ├── models/                      # 数据模型
│   │   ├── intent.go               # 意图模型
│   │   ├── task.go                 # 任务模型
│   │   ├── result.go               # 结果模型
│   │   └── message.go              # 消息模型
│   │
│   └── config/                      # 配置管理
│       ├── config.go               # 配置加载
│       ├── env.go                  # 环境变量
│       └── validator.go            # 配置验证
│
└── frontend/                        # React 前端（Wails 管理）
    └── src/
```

---

## 3. 核心模块设计

### 3.1 应用服务层 (`app.go`)

**职责：**
- 作为 Wails 绑定的统一入口
- 协调各个子系统
- 管理应用生命周期
- 聚合返回结果给前端

**核心方法：**
```go
type App struct {
    ctx          context.Context
    llm          *llm.Client
    orchestrator *orchestrator.Manager
    session      *context.Session
    config       *config.Config
}

// 前端调用的主方法
func (a *App) ExecuteCommand(userInput string) *models.ExecuteResult
func (a *App) GetHistory() []models.Message
func (a *App) ClearHistory()
func (a *App) GetSystemStatus() *models.SystemStatus
func (a *App) SaveConfig(cfg *config.Config) error
```

**设计要点：**
- 所有前端调用必须通过这一层
- 统一错误处理和日志记录
- 维护全局状态（会话、配置）
- 返回结构化数据（JSON 友好）

---

### 3.2 LLM 服务层 (`backend/llm/`)

#### 3.2.1 `client.go` - 通义千问客户端

**职责：**
- 封装通义千问 API 调用
- 处理请求重试和超时
- 管理 API 密钥和限流

**核心功能：**
```go
type Client struct {
    apiKey     string
    endpoint   string
    httpClient *http.Client
    rateLimiter *rate.Limiter
}

// 调用 LLM Function Calling
func (c *Client) CallWithFunctions(
    userInput string, 
    functions []FunctionDef,
    history []Message,
) (*FunctionCall, error)

// 流式调用（可选）
func (c *Client) StreamCall(
    userInput string,
    onToken func(string),
) error
```

**实现要点：**
- 使用 HTTP 客户端调用阿里云 DashScope API
- 支持请求超时（5-10 秒）
- 实现指数退避重试（最多 3 次）
- 错误分类：网络错误、API 错误、解析错误

#### 3.2.2 `function_defs.go` - Function Calling 定义

**职责：**
- 定义所有可调用的函数签名
- 为 LLM 提供清晰的函数描述

**设计要点：**
```go
type FunctionDef struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// 预定义所有函数
var SystemFunctions = []FunctionDef{
    {
        Name: "open_application",
        Description: "打开指定的应用程序",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "app_name": map[string]string{
                    "type": "string",
                    "description": "应用名称，如 firefox, chrome, vscode",
                    "enum": []string{"firefox", "chrome", "vscode", "terminal"},
                },
            },
            "required": []string{"app_name"},
        },
    },
    {
        Name: "control_volume",
        Description: "控制系统音量",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "action": map[string]interface{}{
                    "type": "string",
                    "enum": []string{"set", "increase", "decrease", "mute"},
                },
                "value": map[string]interface{}{
                    "type": "integer",
                    "minimum": 0,
                    "maximum": 100,
                    "description": "音量值（仅 action=set 时需要）",
                },
            },
            "required": []string{"action"},
        },
    },
    // ... 其他函数定义
}
```

**关键函数列表：**
1. `open_application` - 打开应用
2. `control_volume` - 音量控制
3. `control_music` - 音乐播放控制
4. `create_file` - 创建文件
5. `write_file` - 写入文件
6. `search_file` - 搜索文件
7. `open_url` - 打开网址
8. `execute_chain` - 执行命令链（组合任务）

#### 3.2.3 `intent_parser.go` - 意图解析器

**职责：**
- 将 LLM 返回的 Function Call 转换为内部任务模型
- 验证参数完整性
- 处理解析失败的降级策略

**核心逻辑：**
```go
type IntentParser struct {
    validator *validator.Validate
}

// 解析 LLM 响应为任务
func (p *IntentParser) Parse(
    functionCall *FunctionCall,
) (*models.Task, error) {
    // 1. 映射函数名到执行器类型
    executorType := p.mapFunctionToExecutor(functionCall.Name)
    
    // 2. 验证参数
    if err := p.validateParameters(functionCall); err != nil {
        return nil, err
    }
    
    // 3. 构建任务对象
    task := &models.Task{
        ID:           uuid.New().String(),
        Type:         executorType,
        Action:       functionCall.Name,
        Parameters:   functionCall.Arguments,
        Status:       "pending",
        CreatedAt:    time.Now(),
    }
    
    return task, nil
}

// 降级策略：基于规则匹配
func (p *IntentParser) FallbackParse(userInput string) (*models.Task, error) {
    // 正则匹配常见命令
    // 例如：匹配 "打开 XXX"、"音量 XX%"
}
```

---

### 3.3 编排服务层 (`backend/orchestrator/`)

#### 3.3.1 `manager.go` - 编排管理器

**职责：**
- 接收任务并决定执行策略
- 管理执行器生命周期
- 协调复杂任务的执行

**核心方法：**
```go
type Manager struct {
    registry  *executors.Registry
    executor  *Executor
    planner   *TaskPlanner
    context   *context.Session
}

// 执行单个任务
func (m *Manager) ExecuteTask(task *models.Task) (*models.TaskResult, error)

// 执行任务链
func (m *Manager) ExecuteChain(tasks []*models.Task) ([]*models.TaskResult, error)

// 取消任务
func (m *Manager) CancelTask(taskID string) error
```

#### 3.3.2 `task_planner.go` - 任务规划器

**职责：**
- 将复杂命令拆解为多个子任务
- 分析任务依赖关系
- 构建执行 DAG

**关键功能：**
```go
type TaskPlanner struct {
    llmClient *llm.Client
}

// 规划任务执行顺序
func (p *TaskPlanner) Plan(userIntent string) ([]*models.Task, error) {
    // 1. 调用 LLM 进行任务分解
    // 2. 构建依赖图
    // 3. 拓扑排序确定执行顺序
    // 4. 返回有序任务列表
}

// 检测任务依赖
func (p *TaskPlanner) DetectDependencies(tasks []*models.Task) map[string][]string
```

**示例场景：**
```
用户输入："打开 VS Code 并创建一个 main.go 文件"

规划结果：
Task 1: open_application(app_name="vscode")
Task 2: create_file(filename="main.go", wait_for=Task1)
```

#### 3.3.3 `executor.go` - 任务执行器

**职责：**
- 实际执行任务
- 调用具体的执行器
- 处理执行异常

**执行流程：**
```go
type Executor struct {
    registry *executors.Registry
    timeout  time.Duration
}

func (e *Executor) Execute(task *models.Task) (*models.TaskResult, error) {
    // 1. 从注册表获取执行器
    executor, err := e.registry.Get(task.Type)
    if err != nil {
        return nil, err
    }
    
    // 2. 设置超时
    ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
    defer cancel()
    
    // 3. 执行任务
    result := executor.Execute(ctx, task)
    
    // 4. 记录日志
    e.logExecution(task, result)
    
    return result, nil
}
```

---

### 3.4 执行器层 (`backend/executors/`)

#### 3.4.1 `interface.go` - 执行器接口

**统一接口定义：**
```go
// 执行器接口
type Executor interface {
    // 获取执行器名称
    Name() string
    
    // 获取执行器类型
    Type() ExecutorType
    
    // 执行任务
    Execute(ctx context.Context, task *models.Task) *models.TaskResult
    
    // 验证任务参数
    Validate(task *models.Task) error
    
    // 支持的操作列表
    SupportedActions() []string
}

// 执行器类型
type ExecutorType string

const (
    TypeSystem  ExecutorType = "system"
    TypeMusic   ExecutorType = "music"
    TypeFile    ExecutorType = "file"
    TypeBrowser ExecutorType = "browser"
)
```

#### 3.4.2 `registry.go` - 执行器注册表

**职责：**
- 管理所有执行器实例
- 提供执行器查询和路由

```go
type Registry struct {
    executors map[ExecutorType]Executor
    mu        sync.RWMutex
}

// 注册执行器
func (r *Registry) Register(executor Executor) error

// 获取执行器
func (r *Registry) Get(execType ExecutorType) (Executor, error)

// 列出所有执行器
func (r *Registry) List() []Executor

// 初始化所有执行器
func (r *Registry) InitializeAll() error {
    // 注册所有内置执行器
    r.Register(system.NewAppLauncher())
    r.Register(system.NewVolumeControl())
    r.Register(music.NewPlayer())
    r.Register(file.NewFileOperator())
    r.Register(browser.NewBrowserController())
}
```

#### 3.4.3 系统控制执行器 (`executors/system/`)

##### `app_launcher.go` - 应用启动器

**核心功能：**
```go
type AppLauncher struct {
    appPaths map[string]string // 应用名 -> 可执行文件路径映射
}

func (a *AppLauncher) Execute(ctx context.Context, task *models.Task) *models.TaskResult {
    appName := task.Parameters["app_name"].(string)
    
    // 1. 查找应用路径
    path, exists := a.appPaths[appName]
    if !exists {
        // 使用 xdg-open 或 which 命令查找
        path = a.findAppPath(appName)
    }
    
    // 2. 启动应用
    cmd := exec.CommandContext(ctx, path)
    if err := cmd.Start(); err != nil {
        return &models.TaskResult{
            Success: false,
            Error:   err.Error(),
        }
    }
    
    // 3. 等待应用窗口出现（可选）
    time.Sleep(1 * time.Second)
    
    return &models.TaskResult{
        Success: true,
        Message: fmt.Sprintf("已打开 %s", appName),
        Data:    map[string]interface{}{"pid": cmd.Process.Pid},
    }
}
```

**实现要点：**
- 支持常见应用别名（firefox, chrome, vscode）
- 使用 `xdg-open` 作为后备方案
- 检测应用是否已运行（避免重复启动）
- 支持应用参数传递

##### `volume_control.go` - 音量控制

**实现方案：**
```go
type VolumeControl struct {
    backend string // "pactl" | "amixer"
}

func (v *VolumeControl) Execute(ctx context.Context, task *models.Task) *models.TaskResult {
    action := task.Parameters["action"].(string)
    
    switch action {
    case "set":
        value := task.Parameters["value"].(int)
        return v.setVolume(value)
    case "increase":
        return v.adjustVolume(+10)
    case "decrease":
        return v.adjustVolume(-10)
    case "mute":
        return v.toggleMute()
    }
}

func (v *VolumeControl) setVolume(value int) *models.TaskResult {
    // 使用 pactl
    cmd := exec.Command("pactl", "set-sink-volume", "@DEFAULT_SINK@", 
                       fmt.Sprintf("%d%%", value))
    if err := cmd.Run(); err != nil {
        return &models.TaskResult{Success: false, Error: err.Error()}
    }
    
    return &models.TaskResult{
        Success: true,
        Message: fmt.Sprintf("音量已设为 %d%%", value),
    }
}
```

**Linux 音量控制方案：**
- **首选：** PulseAudio (`pactl`)
- **备选：** ALSA (`amixer`)
- **检测：** 启动时检测可用工具

#### 3.4.4 音乐播放执行器 (`executors/music/`)

##### `mpris_client.go` - MPRIS D-Bus 客户端

**MPRIS 协议：**
- Linux 标准音乐播放器控制协议
- 通过 D-Bus 通信
- 支持：Spotify、网易云、VLC 等

**实现方案：**
```go
type MPRISClient struct {
    conn       *dbus.Conn
    playerName string // "org.mpris.MediaPlayer2.spotify"
}

func (m *MPRISClient) Play() error {
    obj := m.conn.Object(m.playerName, "/org/mpris/MediaPlayer2")
    call := obj.Call("org.mpris.MediaPlayer2.Player.Play", 0)
    return call.Err
}

func (m *MPRISClient) Pause() error {
    obj := m.conn.Object(m.playerName, "/org/mpris/MediaPlayer2")
    call := obj.Call("org.mpris.MediaPlayer2.Player.Pause", 0)
    return call.Err
}

func (m *MPRISClient) Next() error {
    obj := m.conn.Object(m.playerName, "/org/mpris/MediaPlayer2")
    call := obj.Call("org.mpris.MediaPlayer2.Player.Next", 0)
    return call.Err
}

func (m *MPRISClient) GetMetadata() (map[string]interface{}, error) {
    // 获取当前播放信息（歌名、艺术家等）
}
```

**备选方案：**
- 使用 `playerctl` 命令行工具
- 更简单但功能略少

##### `player.go` - 播放器控制器

**职责：**
- 自动检测可用播放器
- 统一控制接口
- 处理播放器未运行的情况

```go
type Player struct {
    mprisClient *MPRISClient
    fallbackCmd string // "playerctl"
}

func (p *Player) Execute(ctx context.Context, task *models.Task) *models.TaskResult {
    action := task.Parameters["action"].(string)
    
    // 1. 检测播放器
    if err := p.detectPlayer(); err != nil {
        return &models.TaskResult{
            Success: false,
            Error:   "未找到可用的音乐播放器",
        }
    }
    
    // 2. 执行操作
    switch action {
    case "play":
        return p.play()
    case "pause":
        return p.pause()
    case "next":
        return p.next()
    case "previous":
        return p.previous()
    }
}

func (p *Player) detectPlayer() error {
    // 1. 尝试 MPRIS (D-Bus)
    if p.mprisClient.IsAvailable() {
        return nil
    }
    
    // 2. 尝试 playerctl
    if _, err := exec.LookPath("playerctl"); err == nil {
        p.fallbackCmd = "playerctl"
        return nil
    }
    
    return fmt.Errorf("no player found")
}
```

#### 3.4.5 文件操作执行器 (`executors/file/`)

##### `file_ops.go` - 文件操作

**核心功能：**
```go
type FileOperator struct {
    workDir string // 工作目录
}

func (f *FileOperator) Execute(ctx context.Context, task *models.Task) *models.TaskResult {
    action := task.Parameters["action"].(string)
    
    switch action {
    case "create":
        return f.createFile(task.Parameters)
    case "write":
        return f.writeFile(task.Parameters)
    case "append":
        return f.appendFile(task.Parameters)
    case "delete":
        return f.deleteFile(task.Parameters)
    case "search":
        return f.searchFiles(task.Parameters)
    }
}

func (f *FileOperator) createFile(params map[string]interface{}) *models.TaskResult {
    filename := params["filename"].(string)
    path := filepath.Join(f.workDir, filename)
    
    // 创建目录（如果不存在）
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return &models.TaskResult{Success: false, Error: err.Error()}
    }
    
    // 创建文件
    file, err := os.Create(path)
    if err != nil {
        return &models.TaskResult{Success: false, Error: err.Error()}
    }
    defer file.Close()
    
    // 写入初始内容（如果有）
    if content, ok := params["content"].(string); ok {
        file.WriteString(content)
    }
    
    return &models.TaskResult{
        Success: true,
        Message: fmt.Sprintf("已创建文件: %s", filename),
        Data:    map[string]interface{}{"path": path},
    }
}

func (f *FileOperator) writeFile(params map[string]interface{}) *models.TaskResult {
    filename := params["filename"].(string)
    content := params["content"].(string)
    mode := params["mode"].(string) // "overwrite" | "append"
    
    path := filepath.Join(f.workDir, filename)
    
    var flag int
    if mode == "append" {
        flag = os.O_APPEND | os.O_WRONLY | os.O_CREATE
    } else {
        flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
    }
    
    file, err := os.OpenFile(path, flag, 0644)
    if err != nil {
        return &models.TaskResult{Success: false, Error: err.Error()}
    }
    defer file.Close()
    
    if _, err := file.WriteString(content); err != nil {
        return &models.TaskResult{Success: false, Error: err.Error()}
    }
    
    return &models.TaskResult{
        Success: true,
        Message: fmt.Sprintf("已写入文件: %s", filename),
    }
}
```

**安全考虑：**
- 限制可操作的目录范围（沙箱）
- 文件大小限制
- 危险操作二次确认（删除）
- 路径遍历攻击防护

---

### 3.5 上下文管理层 (`backend/context/`)

#### `session.go` - 会话管理

**职责：**
- 维护对话历史
- 存储会话变量（如"当前文件"）
- 支持多轮对话理解

```go
type Session struct {
    ID        string
    Messages  []models.Message
    Variables map[string]interface{} // 存储上下文变量
    CreatedAt time.Time
    UpdatedAt time.Time
    mu        sync.RWMutex
}

// 添加消息
func (s *Session) AddMessage(role, content string)

// 获取历史（用于 LLM 上下文）
func (s *Session) GetHistory(limit int) []models.Message

// 设置变量（如：当前文件路径）
func (s *Session) SetVariable(key string, value interface{})

// 获取变量（用于指代消解：如"它"）
func (s *Session) GetVariable(key string) (interface{}, bool)

// 清空会话
func (s *Session) Clear()
```

**上下文变量示例：**
```go
// 用户："创建文件 test.txt"
session.SetVariable("current_file", "test.txt")

// 用户："在里面写入 Hello World"（没有指定文件名）
currentFile := session.GetVariable("current_file") // "test.txt"
```

---

### 3.6 系统接口层 (`backend/system/`)

#### `dbus_client.go` - D-Bus 客户端

**D-Bus 用途：**
- 与系统服务通信
- 控制音乐播放器（MPRIS）
- 获取系统通知
- 控制桌面环境

```go
type DBusClient struct {
    sessionConn *dbus.Conn
    systemConn  *dbus.Conn
}

func NewDBusClient() (*DBusClient, error) {
    session, err := dbus.SessionBus()
    if err != nil {
        return nil, err
    }
    
    system, err := dbus.SystemBus()
    if err != nil {
        return nil, err
    }
    
    return &DBusClient{
        sessionConn: session,
        systemConn:  system,
    }, nil
}

// 调用 D-Bus 方法
func (d *DBusClient) Call(
    busName, objectPath, method string, 
    args ...interface{},
) ([]interface{}, error)

// 监听 D-Bus 信号
func (d *DBusClient) Listen(
    busName, signal string,
    handler func(signal *dbus.Signal),
) error
```

#### `shell_executor.go` - Shell 命令执行

**安全的 Shell 执行器：**
```go
type ShellExecutor struct {
    allowedCommands map[string]bool // 白名单
    timeout         time.Duration
}

func (s *ShellExecutor) Execute(command string, args ...string) (string, error) {
    // 1. 检查命令是否在白名单
    if !s.allowedCommands[command] {
        return "", fmt.Errorf("command not allowed: %s", command)
    }
    
    // 2. 创建命令
    ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
    defer cancel()
    
    cmd := exec.CommandContext(ctx, command, args...)
    
    // 3. 执行并捕获输出
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("command failed: %v, output: %s", err, output)
    }
    
    return string(output), nil
}

// 安全的命令白名单
var AllowedCommands = []string{
    "xdg-open",
    "pactl",
    "amixer",
    "playerctl",
    "which",
    "ls",
    "cat",
    // ... 其他安全命令
}
```

**防护措施：**
- 命令白名单
- 参数过滤（防止注入）
- 超时控制
- 资源限制

---

## 4. 数据流设计

### 4.1 命令执行流程

```
1. 前端发送用户输入
   ↓
2. App.ExecuteCommand() 接收
   ↓
3. LLM Client 解析意图
   ↓ (Function Call)
4. Intent Parser 转换为 Task
   ↓
5. Orchestrator 规划执行
   ↓
6. Executor Registry 路由到具体执行器
   ↓
7. 执行器调用系统接口
   ↓
8. 返回执行结果
   ↓
9. 聚合结果返回前端
```

### 4.2 组合任务执行流程

```
用户输入："打开 VS Code 并创建 main.go"
   ↓
LLM 识别为组合任务
   ↓
Task Planner 拆解：
  Task 1: open_application(vscode)
  Task 2: create_file(main.go, depends_on=Task1)
   ↓
Executor 顺序执行：
  1. 执行 Task 1 → 成功
  2. 等待 2 秒（应用启动）
  3. 执行 Task 2 → 成功
   ↓
返回聚合结果：
  "已打开 VS Code，已创建 main.go"
```

---

## 5. 错误处理策略

### 5.1 错误分类

```go
type ErrorType string

const (
    ErrorTypeNetwork    ErrorType = "network"    // 网络错误（LLM API）
    ErrorTypeParsing    ErrorType = "parsing"    // 解析错误
    ErrorTypeExecution  ErrorType = "execution"  // 执行错误
    ErrorTypePermission ErrorType = "permission" // 权限错误
    ErrorTypeNotFound   ErrorType = "not_found"  // 资源未找到
)

type ExecuteError struct {
    Type    ErrorType
    Message string
    Cause   error
    Retry   bool // 是否可重试
}
```

### 5.2 错误恢复策略

| 错误类型 | 恢复策略 |
|---------|---------|
| LLM API 超时 | 重试 3 次，降级到规则匹配 |
| 应用未找到 | 提示用户安装，提供替代方案 |
| 音乐播放器未运行 | 提示启动播放器 |
| 文件权限错误 | 提示需要权限，提供 sudo 选项 |
| 命令执行失败 | 记录日志，返回详细错误信息 |

---

## 6. 性能优化

### 6.1 LLM 调用优化

1. **本地规则优先**
   ```go
   // 简单命令不调用 LLM
   if isSimpleCommand(userInput) {
       return parseByRules(userInput)
   }
   ```

2. **缓存高频命令**
   ```go
   type IntentCache struct {
       cache map[string]*models.Task
       ttl   time.Duration
   }
   ```

3. **批量调用**
   - 组合任务一次性规划

### 6.2 执行器优化

1. **连接池**
   ```go
   // D-Bus 连接复用
   type DBusPool struct {
       conns chan *dbus.Conn
   }
   ```

2. **并行执行**
   ```go
   // 无依赖任务并行
   func (e *Executor) ExecuteParallel(tasks []*models.Task) []*models.TaskResult
   ```

---

## 7. 安全设计

### 7.1 权限控制

```go
type Permission string

const (
    PermissionSystem Permission = "system"
    PermissionFile   Permission = "file"
    PermissionNetwork Permission = "network"
)

// 危险操作白名单
var DangerousActions = map[string]bool{
    "delete_file":      true,
    "system_shutdown":  true,
    "install_package":  true,
}

// 执行前检查
func (e *Executor) checkPermission(task *models.Task) error {
    if DangerousActions[task.Action] {
        // 需要用户确认
        return ErrNeedConfirmation
    }
    return nil
}
```

### 7.2 沙箱限制

```go
type Sandbox struct {
    allowedPaths []string
    maxFileSize  int64
}

// 检查文件路径
func (s *Sandbox) ValidatePath(path string) error {
    absPath, err := filepath.Abs(path)
    if err != nil {
        return err
    }
    
    for _, allowed := range s.allowedPaths {
        if strings.HasPrefix(absPath, allowed) {
            return nil
        }
    }
    
    return fmt.Errorf("path not allowed: %s", path)
}
```

---

## 8. 测试策略

### 8.1 单元测试

```go
// executors/system/app_launcher_test.go
func TestAppLauncher_Execute(t *testing.T) {
    launcher := NewAppLauncher()
    
    task := &models.Task{
        Type:   TypeSystem,
        Action: "open_application",
        Parameters: map[string]interface{}{
            "app_name": "firefox",
        },
    }
    
    result := launcher.Execute(context.Background(), task)
    assert.True(t, result.Success)
}
```

### 8.2 集成测试

```go
// 测试完整流程
func TestEndToEnd_OpenApplicationAndCreateFile(t *testing.T) {
    app := NewApp()
    
    result := app.ExecuteCommand("打开 VS Code 并创建文件 test.go")
    
    assert.True(t, result.Success)
    assert.Len(t, result.Actions, 2)
}
```

### 8.3 Mock LLM

```go
type MockLLMClient struct {
    responses map[string]*FunctionCall
}

func (m *MockLLMClient) CallWithFunctions(
    userInput string,
    functions []FunctionDef,
    history []Message,
) (*FunctionCall, error) {
    return m.responses[userInput], nil
}
```

---

## 9. 部署与运维

### 9.1 配置管理

```go
// config.yaml
llm:
  provider: "qwen"
  api_key: "${QWEN_API_KEY}"
  model: "qwen-max"
  timeout: 10s

executors:
  system:
    allowed_apps:
      - firefox
      - chrome
      - vscode
  file:
    work_dir: "/home/user/Documents"
    max_file_size: 10485760 # 10MB
  music:
    default_player: "spotify"

logging:
  level: "info"
  file: "/var/log/ai-assistant.log"
```

### 9.2 日志记录

```go
type Logger struct {
    logger *zap.Logger
}

func (l *Logger) LogExecution(task *models.Task, result *models.TaskResult) {
    l.logger.Info("task executed",
        zap.String("task_id", task.ID),
        zap.String("action", task.Action),
        zap.Bool("success", result.Success),
        zap.Duration("duration", result.Duration),
    )
}
```

---

## 10. 一周开发里程碑

| Day | Go 后端任务 | 验收标准 |
|-----|-----------|---------|
| **1** | 项目结构 + LLM 集成 | 能调用通义千问并解析 Function Call |
| **2** | 系统控制执行器 | 能打开应用、控制音量 |
| **3** | 音乐播放执行器 | 能控制播放器（MPRIS/playerctl） |
| **4** | 文件操作执行器 | 能创建、写入文件 |
| **5** | 任务编排 + 组合命令 | 能执行"打开 XX 并创建文件" |
| **6** | 错误处理 + 测试 | 稳定可演示 |
| **7** | 配置优化 + 文档 | 完整交付 |

---

## 11. 关键技术选型

### 11.1 依赖库

```go
// go.mod
require (
    github.com/wailsapp/wails/v2 v2.x.x         // Wails 框架
    github.com/godbus/dbus/v5 v5.x.x           // D-Bus 客户端
    github.com/alibabacloud-go/dashscope-go/v2 // 通义千问 SDK
    github.com/spf13/viper v1.x.x              // 配置管理
    go.uber.org/zap v1.x.x                     // 日志
    github.com/google/uuid v1.x.x              // UUID
    gopkg.in/yaml.v3 v3.x.x                    // YAML 解析
)
```

### 11.2 Linux 系统工具

| 功能 | 工具 | 备选 |
|------|------|------|
| 应用启动 | `xdg-open` | `gtk-launch` |
| 音量控制 | `pactl` | `amixer` |
| 音乐控制 | D-Bus MPRIS | `playerctl` |
| 进程管理 | `ps`, `kill` | - |
| 文件搜索 | `find` | `fd` |

---

## 12. 后续扩展方向

### V2 功能（2-4 周）

1. **语音输入集成**
   - 集成语音识别 SDK
   - 前端 WebRTC 录音

2. **更多执行器**
   - 浏览器控制（Selenium）
   - 邮件处理
   - 日程管理

3. **AI Agent 能力**
   - 自主规划
   - 反思与优化
   - 长期记忆

### V3 功能（1-3 月）

1. **插件系统**
   - 第三方执行器
   - 热加载

2. **多用户支持**
   - 权限管理
   - 配置隔离

3. **跨平台支持**
   - Windows 适配
   - macOS 适配

---

**文档版本:** v1.0  
**创建日期:** 2025-10-20  
**作者:** GitHub Copilot
