# 1bot —— 本地模型网关架构设计

> 定位：**一台本地的模型路由器 + 一套配置注入器**。
> 不是聊天客户端，不是模型管理面板。数据面必须极可靠（它是所有 agent 的唯一出口），控制面可以糙（一天开两次）。
>
> 本文档只描述**架构与决策**。原型现状、实测细节、技术选型清单、分期计划见 [`参考.md`](./参考.md)。

---

## 一、架构总览

```
┌──────────────── 控制面（Control Plane）────────────────┐
│  myapp-ui    Qt Widgets 原生界面（主打）                 │
│  myapp tui   纯 Go 终端界面（无桌面 / SSH / 保底）       │
│  Provider·Account·Group 编辑  ·  用量  ·  会话  ·  诊断   │
│                                                          │
│  ★ 独立二进制，与网关进程分离（见 2.5）                  │
└────────────────────────┬────────────────────────────────┘
                         │ HTTP JSON（控制面 API，带鉴权）
┌────────────────────────▼────────────────────────────────┐
│              网关进程（1 个，常驻，CGO_ENABLED=0）         │
│                                                          │
│  ┌── Ingress（库）────────────────────────────────┐     │
│  │  HTTP 解析 · SSE 写出 · 超时/取消(499)          │     │
│  └────────────────────────────────────────────────┘     │
│  ┌── Core（库）───────────────────────────────────┐     │
│  │  IR        协议互译（入站/出站）                 │     │
│  │  Router    候选池 → 过滤 → 排序 → 尝试          │     │
│  │  Quota     窗口与额度（source of truth）        │     │
│  │  Meter     用量记账                             │     │
│  │  State     冷却 / 并发 / 亲和 / 飞行中           │     │
│  └────────────────────────────────────────────────┘     │
│  ┌── PluginHost ──────────────────────────────────┐     │
│  │  插件子进程生命周期：启动/健康检查/重启/回收      │     │
│  └────────────────────────────────────────────────┘     │
└───────────────────────────┬──────────────────────────────┘
                            │ stdio JSON-RPC（管道）或 Unix socket
         ┌──────────────────┼──────────────────┐
         ▼                  ▼                  ▼
  ┌────────────┐     ┌────────────┐     ┌────────────┐
  │ plugin-    │     │ plugin-    │     │ plugin-    │
  │ codebuddy  │     │ codex      │     │ claude     │
  │ (ACP)      │     │(app-server)│     │(stream-json│
  │            │     │            │     │  + MCP)    │
  └─────┬──────┘     └─────┬──────┘     └─────┬──────┘
        ▼                  ▼                  ▼
   codebuddy --acp    codex app-server    claude -p
        │                  │                  │
        └──────────────────┴──────────────────┘
                           ▼
                      上游厂商

┌──────────────── 旁路（Sidecar，独立进程/命令）────────────┐
│  Wiring：改 Agent 配置（短命命令，改完就退）                │
│  Capture：抓本机请求做观察（不参与转发，可选）              │
└──────────────────────────────────────────────────────────┘
```

**插件之间不通信。** 所有调度、额度、fallback 都在 Core，插件只负责"跟一个上游说话"。

### 1.1 当前实现落点（CodeBuddy 插件）

架构图「网关 → 插件 → 厂商 CLI」落到当前代码，是三级进程拓扑：

```
gateway.exe（网关进程，CGO_ENABLED=0）
  ├── openai.Server    OpenAI 入站（HTTP → SSE）
  ├── transport        Transport 接口（跨层抽象）
  └── plugin.Host      PluginHost：启动 / 驱动插件子进程
          │  stdio JSON-RPC（统一插件协议，见第四章）
          ▼
plugin-codebuddy.exe（插件进程）
  └── acp.Client       驱动 codebuddy --acp（厂商知识全在这里）
          │  ACP
          ▼
codebuddy --acp（厂商 CLI）
```

对应包结构：

| 包 | 进程 | 职责 |
|---|---|---|
| `cmd/gateway` | gateway.exe | 组装 + 启动 + 信号处理 |
| `internal/openai` | gateway.exe | Ingress：OpenAI 入站 → SSE |
| `internal/transport` | gateway.exe | 唯一跨层接口 |
| `internal/plugin` | gateway.exe | PluginHost：插件子进程生命周期 |
| `internal/proto` | 两端共用 | 统一插件协议（见第四章） |
| `cmd/plugin-codebuddy` | plugin-codebuddy.exe | 插件进程入口 |
| `internal/acp` | plugin-codebuddy.exe | 驱动 codebuddy --acp |

关键：`openai` 只依赖 `transport.Transport`。网关从「直接 exec codebuddy」改成「exec 插件」时，`openai` 层一行未动——这就是 Transport 抽象生效的证明。

---

## 二、进程边界的划分原则

> **进程边界按「生命周期 + 失败域」划分；agent 之间的差异按「适配器」划分。**

10 个 agent 的配置格式不同是**代码多样性**问题（用 adapter/插件解决），不是**并发/隔离**问题（不需要 10 个进程）。

### 2.1 插件必须独立进程 ✅

| 收益 | 说明 |
|---|---|
| **崩溃隔离** | CodeBuddy 子进程 OOM / ACP 不兼容 → 只挂那个插件，网关和其他 provider 照常 |
| **语言无关** | ACP 插件用 Go，Codex 插件用 Node 都行；Core 只认 JSON-RPC |
| **热重启** | 升级某个适配器不用重启网关，不断在途请求 |
| **厂商知识外置** | Core 代码里不出现任何 `codebuddy` / `codex` 字样 |

### 2.2 Ingress 与 Core 合并为一个进程 ⚠️

**决策：拆进程收益 < 成本，改为库边界隔离。**

拆分能拿到的（都很弱）：协议解析崩溃不波及 Core（但 HTTP 解析是最成熟的代码）；接入可独立伸缩（本地工具 1 个实例就够）。

拆分要付出的（都很痛）：
- **流式 SSE 多跳一跳**：token 链路变成 `插件 → Core → 接入 → 客户端`（合并后 `插件 → Core → 客户端`）
- 背压、超时、取消要跨三个进程传递，语义极易出错
- 连接生命周期与请求生命周期分属两个进程，排障难度翻倍

真正需要拆的场景是多协议接入、独立伸缩、路由热更新不掉连接——**本地网关一个都不占**。

### 2.3 Wiring 是旁路，不进网关进程

| 维度 | Wiring | Gateway |
|---|---|---|
| 触发频率 | 用户点一下 / 切方案 / 换端口 | 每秒几十次 |
| 失败后果 | **改坏用户文件**（最严重） | 一次请求失败（可重试） |
| 依赖知识 | N 个 Agent 的配置格式 + 路径 | HTTP + 插件协议 |
| 需要权限 | 写用户 home、可能动 Keychain | 只需网络 |
| 升级影响 | 随时可重启 | **重启会断所有 agent 的请求** |

**Wiring 是短命命令，不是常驻进程**：

```
myapp serve          → 网关进程（常驻，永不写用户配置）
myapp wire claude    → 短命进程（改完就退）
myapp wire --all
myapp unwire codex
myapp rewire         → 网关换端口后批量改
myapp wire --watch   → 可选：单个守护进程做漂移检测（1 个管所有 agent，不是每 agent 一个）
```

每个 agent 的格式知识是一个 **adapter 实现**（编译期多样性），不是进程：

```go
type Adapter interface {
    ID() string                                      // "claude" / "codex" / "codebuddy"
    Detect() (bool, error)                           // 扫路径/注册表，判断是否装了
    Read() (Config, error)                           // 读当前配置
    Plan(cfg Config, target Target) ([]Edit, error)  // 算出改哪些字节区间
    Apply(edits []Edit) error                        // 原子写 + 备份
    Revert(stash Stash) error                        // 断开时原样还原
}
```

**硬约束：网关进程永远不写用户的 agent 配置文件。** 它只读取 `applied.json` 来展示状态。

### 2.4 两条方向相反的数据路径（易混淆）

| | 方向 | 谁在上游 |
|---|---|---|
| **Wiring** | 让 agent 指向网关 | 网关在中间，agent 是客户端 |
| **插件（ACP 子进程）** | 网关驱动 agent | 网关是客户端，agent 在上游 |

一个产品里两者可以共存，但**不是同一个概念**，不要混成一套抽象。

### 2.5 控制面 UI：Qt Widgets，且必须二进制隔离

**决策：控制面采用 Qt Widgets 原生界面**，目标是 Windows 7 等老系统也能使用——那里没有 WebView2，老浏览器也跑不动现代 SPA，而 Qt 是原生绘制。

Qt 需要 CGO，因此**必须在二进制层面隔离**：

| 二进制 | CGO | 构成 | 部署形态 |
|---|---|---|---|
| `myapp` | `CGO_ENABLED=0` | Ingress + Core + PluginHost（数据面 + 控制面 API） | 静态单文件；容器 / 无桌面 / Linux 老发行版均可 |
| `myapp-ui` | `CGO_ENABLED=1` | Qt Widgets 界面 | 桌面可选安装；崩溃或缺失不影响网关服务 |

三条硬约束：

1. **Qt 的任何代价不得外溢到网关二进制** —— Qt5 生命周期风险、几十 MB DLL、许可证合规全部局限在 `myapp-ui`。
2. **`myapp tui` 必须保留** —— 纯 Go 终端界面，覆盖 SSH / 无桌面 / Qt 装不上的场景。它和 Qt UI 一样，只是控制面 HTTP API 的一个客户端。
3. 控制面**永远通过网络调用网关**，不得链进同一个进程——这是 ADR-2、ADR-6 成立的前提。

> 前置条件与已知风险（Qt 版本陷阱、Go 对 Win7 的支持边界、Go+Qt 绑定选型、发布清单）见 [`参考.md`](./参考.md)。

---

## 三、通信方式

**框架统一通信方式：`stdio JSON-RPC`（Core ↔ 插件）/ `Unix socket`（可选）。**

| 链路 | 方式 |
|---|---|
| Core ↔ 插件 | **stdio JSON-RPC** —— 管道，每行一条 JSON |
| 控制面 → Core | HTTP JSON（带鉴权） |
| Capture（旁路观察，可选） | 只读共享内存 ring buffer（特例） |

三点理由：

1. **崩溃语义** —— 插件挂了，OS 关闭 fd，对端收到 EOF，天然检测；共享内存会留下持锁死锁。
2. **零成本基础设施** —— 背压由管道/内核缓冲提供，无需手写 ring buffer 与唤醒机制；`strace`/`nc` 可直接观测。
3. **保住 `CGO_ENABLED=0`** —— stdio/socket 是纯 Go；共享内存需要 cgo/syscall，会毁掉静态单二进制。

IPC 开销占请求延迟的 **0.01%–0.05%**（上游 LLM 调用 100–500 ms，IPC 往返 10–30 µs），不是瓶颈。

> 详细论证、与被否决方案（共享内存）的完整对比，见 [`参考.md`](./参考.md) 第七节。

---

## 四、插件协议

直接长成 **ACP 的形状**——CodeBuddy 插件几乎只做透传，少写一层转换。

```go
// Core → 插件（stdin，每行一条 JSON）
{"jsonrpc":"2.0","id":1,"method":"turn",
 "params":{"session":"s1","prompt":[{"type":"text","text":"..."}]}}
{"jsonrpc":"2.0","method":"cancel","params":{"id":1}}
{"jsonrpc":"2.0","id":2,"method":"quota","params":{}}      // 额度查询
{"jsonrpc":"2.0","id":3,"method":"models","params":{}}     // 模型列表

// 插件 → Core（stdout）
{"jsonrpc":"2.0","method":"event",
 "params":{"id":1,"kind":"text","text":"我"}}              // 通知：流式增量
{"jsonrpc":"2.0","method":"event",
 "params":{"id":1,"kind":"tool","name":"Read"}}
{"jsonrpc":"2.0","id":1,"result":{"stopReason":"end_turn"}}
{"jsonrpc":"2.0","id":2,"result":{"windows":[{"share":42,"resetsAt":"..."}]}}

// 反向请求（插件向 Core 要权限/工具能力）
{"jsonrpc":"2.0","id":9,"method":"permission","params":{"tool":"Bash"}}
```

`event.kind` 归一化：`text` / `thought` / `tool` / `usage` / `error`。

---

## 五、传输层抽象

路由层完全不需要知道底下是 HTTP 还是子进程：

```go
type Transport interface {
    Do(ctx context.Context, req *ir.Request, out chan<- ir.Event) error
    Close() error
}

type HTTPTransport struct { ... }        // API key / 兼容 API
type ProcessTransport struct { ... }     // ACP / app-server / stream-json
```

`Seat` 上多一个字段即可：

```go
type Seat struct {
    ...
    Transport TransportKind  // http | process
    Session   string         // process 用：对话 → 子进程的键
}
```

**路由、排序、额度、fallback 全部不变——它们只看 Seat，不看传输方式。**

### 选型规则

```
厂商校验请求内容/调用方身份？
  ├─ 是（Claude：非本尊会被计 Extra Usage）→ 必须子进程
  ├─ 否（ChatGPT/Codex：只校验凭证+header）→ HTTP 重放更快
  └─ 提供 ACP 等标准协议          → 优先用标准协议，零逆向
```

各上游的具体建议见 `参考.md`。

**ACP 的战略价值**：只实现一个 ACP client，就能接所有支持 ACP 的 agent。新增厂商 = 用户装个 CLI，核心零改动。按优先级降级：

```
ACP → CLI 私有 stream-json → HTTP 重放 token → 不支持
```

---

## 六、领域模型（最该先定，改起来最贵）

### 6.1 一级分界：Config 与 State

整个模型里最容易导致后期返工的一条线：**用户拥有的配置**与**系统观测的状态**必须分开存储。

| | 谁拥有 | 变化频率 | 写策略 | 丢了会怎样 |
|---|---|---|---|---|
| **Config** | 用户 | 一天两次 | 原子写 + 备份 + 回滚 | 严重，需恢复 |
| **State** | 系统观测 | 每几秒 | 直接覆盖 | 自愈，仅暂时不准 |

**判据**：任何字段问一句"删掉它，系统能不能自己恢复"——能的是 State，不能的是 Config。

```go
// Config：用户拥有，绝不自动覆盖
type Config struct {
    Providers []Provider
    Accounts  []Account
    Groups    []Group
    Routes    []Route
}

// State：系统拥有，可随时重建
type State struct {
    Observed  map[AccountID]Quota
    Cooldowns map[AccountID]time.Time
    InFlight  map[AccountID]int
    Affinity  map[ConversationID]AccountID
}
```

混在一起的后果：配置文件里出现每 5 秒变的字段 → 用户手改与系统更新互相覆盖 → 回滚会把过期额度数据带回来 → 原子写策略无法统一。

### 6.2 Provider = 接入方式，不是厂商

`Provider` 这个词最容易引错抽象。它应表示**接入方式**，厂商退化为纯展示字段。

因为同一厂商的不同接入方式，其额度形状与传输方式完全不同：

| 接入方式 | 传输 | 额度模型 | 需 OAuth 刷新 |
|---|---|---|---|
| Anthropic API key | HTTP | 按量计费，无窗口 | 否 |
| **Claude 订阅** | **子进程** | 5h + 周窗口 | **是** |
| OpenAI API key | HTTP | 无窗口 | 否 |
| **ChatGPT 订阅** | HTTP 重放 | 5h + 周窗口 | **是** |
| CodeBuddy 订阅 | ACP 子进程 | 厂商自有 | 是 |

若 Provider = 厂商，"同一厂商下既有 key 账号又有订阅账号"会逼你在 Provider 内部写分支。拆开则各自声明，没有 if-else。

```go
type Provider struct {
    ID         string          // "anthropic-api" / "claude-sub" / "codebuddy-acp"
    Vendor     string          // 仅用于 UI 分组显示
    Kind       ProviderKind    // api-key | subscription | local
    Transport  TransportKind   // http | acp | app-server | stream-json
    PluginID   string          // 由哪个插件实现
    Models     []ModelSpec     // 静态模型目录
    WindowSpec []WindowSpec    // 该方式的额度窗口如何定义
}

// WindowSpec 声明该接入方式有哪些额度窗口、额度从哪里观测。
// 例：Claude 订阅 = [5h 滚动窗, 168h 周窗]；API key 计费 = []（无窗口）。
type WindowSpec struct {
    Span   time.Duration // 窗口时长，如 5h / 168h
    Kind   WindowKind    // rolling（从首次请求起算）| fixed（固定时刻）| quota（按量额度）
    Source Source        // 额度如何观测：api | header | probe | estimate
    Model  string        // 空 = 全部模型；否则只对该模型计数
}
```

### 6.3 Account 与 Credential：所有权与刷新权

**Account 必须是一等实体**，不能塞进 `Provider.keys []string`——否则"同一 provider 下两个订阅账号各自不同的 5h/周窗口"无法表达。

Credential 最关键的是**刷新权归属**：两个进程同时持有同一 refresh token 并轮流刷新，会互相把对方踢下线。

```go
type Credential struct {
    Kind      CredKind    // static-key | oauth | handed-over
    Access    string
    Refresh   string      // 可空
    ExpiresAt time.Time
    Origin    Origin      // 本机文件 / keychain / 自有 store
    Owner     Ownership   // ours | borrowed
}
```

`Owner` 决定一整套行为边界：

| | `ours`（用户在 UI 登录） | `borrowed`（读本机 CLI） |
|---|---|---|
| 过期时刷新 | ✅ 自己刷 | ❌ 不刷，交还原生 CLI |
| 刷新后写回 | 写自有 store | **绝不碰用户的文件** |
| 被第三方改动 | 无所谓 | 视为漂移，只读降级 |

从本机 CLI 读来的订阅是 `borrowed`，遵循 **read, never changed**。

### 6.3.1 Account 的完整定义

Account 分两层：**本体（Config）** 与 **运行时状态（State）**——这是 6.1 分界的直接应用。

#### 本体（Config，用户拥有、落盘、可回滚）

```go
type Account struct {
    // —— 身份 ——
    ID       string     // 稳定主键，见下方"ID 规则"
    Label    string     // 展示名，用户随便改
    Provider ProviderID // 属于哪个接入方式（决定传输、额度窗口形状、插件）
    UID      string     // 上游身份：OAuth sub / 账号 id / key 指纹

    // —— 凭据 ——
    Credential Credential // 见 6.3；Access/Refresh 是敏感字段

    // —— 能力 ——
    Models []ModelRef // 覆盖 Provider 目录；空 = 继承全部（6.4 第二层）

    // —— 调度元数据 ——
    Tags     []string // 供 Group / Route 筛选
    Disabled bool     // 手动停用，立即退出路由池
}
```

#### 运行时状态（State，系统观测、可重建、不落 Config）

```go
type AccountState struct {
    Quota    Quota     // 窗口观测值（带 Source，见 6.5）
    Cooldown time.Time // 冷却到这一刻才可再用
    InFlight int       // 正在进行的请求数
    Healthy  bool      // 最近一次探测/请求是否成功
}
```

`Quota` / `Cooldown` / `InFlight` / `Healthy` 每几秒就变，是**观测值**，放 `State.Observed[AccountID]`，**绝不放进 `Account` 结构**。

#### 四个设计点

1. **ID 规则**：派生 ID `provider/norm(uid)`，不用随机 UUID——State 重建时能凭它自动对上号，无需映射表。隐私权衡：若 uid 是手机号/邮箱，它会散落到日志与 State 里；「可读性」与「不泄露 PII」二选一，须现在定（magpie 用 `sha256(uid)[:16]`）。
2. **凭据放哪**：`ours` 存自有 store；`borrowed` **不落盘**，只存 `Origin`，access_token 仅缓存在内存、过期重读文件。
3. **UID 与 Label 分开**：UID 上游给定、不可改（对账/去重用），Label 用户起、纯展示。
4. **Account → Seat**：Account 不参与排序；由 Account × 可用 Model × 可用 Transport 在运行时派生出 `Seat`（见 6.9）。一个多模型账号会派生出多个 Seat，共享额度、能力不同。

#### 实例

```go
Account{
    ID:       "claude-sub/alan@example.com",
    Label:    "Claude Max · 主号",
    Provider: "claude-sub",
    UID:      "user_29f3...",        // OAuth sub
    Credential: Credential{
        Kind:      "oauth",
        Access:    "sk-ant-oat01-...",
        Refresh:   "sk-ant-ort01-...",
        ExpiresAt: /* 4 小时后 */,
        Origin:    "~/.claude/.credentials.json", // 从 Claude Code 读的
        Owner:     "borrowed",        // ← 绝不刷新、绝不写回
    },
    Models:   nil,                    // 继承目录全部
    Tags:     []string{"max", "us"},
    Disabled: false,
}

// 对应 State（在 State.Observed 里，不在 Account 里）
AccountState{
    Quota: Quota{Windows: []Window{
        {Share: 38, ResetsAt: /* ... */, Span: 5 * time.Hour, Source: "header"},
        {Share: 72, ResetsAt: /* ... */, Span: 7 * 24 * time.Hour, Source: "header"},
    }},
    Cooldown: time.Time{},
    InFlight: 2,
    Healthy:  true,
}
```

### 6.4 Model 的三层归属

```
第 1 层  目录    Provider.Models   该接入方式支持哪些（静态）
第 2 层  可用    Account.Models    这个账号实际能用的子集（订阅等级不同）
第 3 层  计数    Window.Model      哪些模型消耗这个窗口（有些模型免费）
```

第 3 层最易忽略：订阅里常有不计额度的模型。若 Window 不带 Model 过滤，路由会把白嫖机会算成消耗。

```go
type ModelSpec struct {
    ID      string   // "claude-sub/claude-opus-4" —— 全局扁平命名空间
    Caps    CapSet   // streaming / tools / vision / thinking
    Context int
    Free    bool     // 订阅上不扣额度
}
```

**Model 是全局扁平命名空间 `provider/model`，定死。** 别名只用于路由选择，不做标识。

### 6.5 Quota 挂 Account，且带 Source

```go
type Quota struct {
    Windows []Window // 按 Span 降序（周 > 5h）；观测值
    Known   bool     // 是否观测到过；false = 未知，路由需谨慎降权
}

type Window struct {
    Share    float64     // 0-100
    ResetsAt time.Time
    Span     time.Duration
    Source   Source      // api | header | probe | estimate
    Model    string      // 只对某些模型计数时
}
```

`Source` 可靠性递减：`api`（官方额度端点，可信）> `header`（响应头推断）> `probe`（探测请求）> `estimate`（猜的）。

额度数据从来不可靠，必须知道"这是官方报的还是我猜的"，猜的降权。

**`Known == false`（从未观测到）与 `Windows == nil`（该方式无窗口，如 API key 计费）语义不同，不能混。**

### 6.6 Group：嵌套语义与 Policy 合并

```go
type Group struct {
    ID       string
    Label    string
    Members  []Member    // Account 引用 或 嵌套 Group
    Strategy StrategyID  // 排序策略
    Policy   Policy      // 并发 / RPM / 错误归类
}

type Member struct {
    Kind MemberKind // account | group
    ID   string
}

type Policy struct {
    Concurrency int        // 最大并发
    RPM         int        // 每分钟请求数
    OnError     ErrorClass // 429 / 5xx / 超时 如何归类与冷却
}
```

`Group.Members` 可以是 Account 引用或嵌套 Group。**选人时按递归下降**：

1. 先过滤掉完全不可用的 member（冷却中 / 无额度 / 不支持该模型）
2. 由本层 Strategy 挑一个
3. 若挑到的是子 Group，递归进入，用它自己的 Strategy 继续选

这样既保留内层 Policy，又不会选到死路。

**Policy 嵌套时取交集**（最严格的生效）。否则上层 `RPM=100`、下层 `RPM=10`，上层的宽松配置会把下层压垮。

**排序只能由 Strategy 给**，Group 本身不参与排序决策。

### 6.7 Route

```go
type Selector struct {
    Kind  SelectorKind // model 精确 | alias | prefix
    Value string       // 匹配值
}

type Route struct {
    Selector Selector
    Priority int      // 多条匹配时的仲裁
    Target   GroupID
}
```

多条 Route 同时匹配时必须明确仲裁顺序（按 Priority 或最长前缀优先），**不能依赖配置文件中的先后顺序**。

### 6.8 记账：额度 ≠ 钱（双轨）

Meter 要分清两套单位，UI 混着显示会误导用户：

| | 单位 | 用途 |
|---|---|---|
| **订阅消耗** | 窗口百分比 | 路由决策、额度展示 |
| **真实花费** | 货币 | 只对本计费率账号有意义 |

同一个请求，"消耗 1% 周窗"与"花了 $0.01"不等价——前者的实际价值取决于该账号的订阅价格与剩余时间。

```go
type UsageRecord struct {
    At        time.Time
    AccountID string
    Model     string
    Usage     Usage    // {Input, CacheRead, CacheWrite, Output}
    Credited  float64  // 记入哪个窗口、多少份额（订阅视角）
    Charged   float64  // 实际花费（按量视角，可空）
}
```

### 6.9 调度契约（抄契约，不抄策略）

```go
type Seat struct {
    ID          string
    Kind        Kind          // api-key | subscription | local
    Used        float64       // 0-100
    Renewals    []Window      // 按 span 降序
    Pace        float64       // 剩余额度 / 到重置的小时数
    Cooldown    time.Time
    InFlight    int
    Concurrency int
    Tags        []string
}

type Strategy interface {
    Order(pool Pool, seats []Seat) []int   // 只返回顺序
}
```

Seat 由 Account + 模型可用性 + 传输可用性在**运行时派生**，不落 Config。

**核心契约**：池的**过滤**（冷却/并发/限额/能力）由 Core 做；**Strategy 只负责排序**；结果非法就 **fail-open** 回退原顺序。

第一版策略只需两条：
1. 有额度 > 快用尽 > 已用尽（分层）
2. 同层内 `Renewals[0]` 最早的先 —— **时间价值：趁快过期的先吃**

不要 v1 做意图路由。

### 6.10 枚举汇总

上面各类型引用的枚举，集中在此钉死。全部用 `string` 底层类型（而非 int）：日志与 JSON 里可读，且新增取值不破坏旧数据。

```go
// —— 接入方式 ——
type ProviderKind  string // api-key | subscription | local
type TransportKind string // http | acp | app-server | stream-json

// —— 凭据 ——
type CredKind  string // static-key | oauth | handed-over
type Ownership string // ours | borrowed

// —— 额度 ——
type WindowKind string // rolling（首次请求起算）| fixed（固定时刻）| quota（按量）
type Source     string // api（官方额度端点）| header（响应头）| probe（探测）| estimate（猜）

// —— 路由 ——
type MemberKind   string // account | group
type SelectorKind string // model（精确）| alias | prefix

// —— 能力 ——
type CapSet struct {
    Streaming bool
    Tools     bool
    Vision    bool
    Thinking  bool
}

// —— 错误归类（Policy.OnError 与 State 冷却共用）——
type ErrorClass string // retryable | rate-limited | auth | fatal

// —— 用量（6.8）——
type Usage struct {
    Input      int
    CacheRead  int
    CacheWrite int
    Output     int
}
```

补充说明：

- `Seat.Kind`（6.9）复用 `ProviderKind`，语义一致：候选的来源类型。
- `Source` 可靠性排序 `api > header > probe > estimate`（见 6.5），路由按此对额度降权。
- `ErrorClass` 决定两件事：① 是否触发重试/冷却；② 冷却时长。`auth` 类（凭据失效）**不进冷却**，而是把账号标记为 `borrowed` 待重读、或 `ours` 待刷新。

---

## 七、协议层：别一上来做四种

v1 只做 **2×2**：

```
入站：OpenAI Chat  +  Anthropic Messages
出站：OpenAI Chat  +  Anthropic Messages
```

= 4 个转换器，其中 2 个恒等映射。已能覆盖 Claude Code、Codex、绝大多数 SDK。

IR 里这五个字段**从第一天就要有**：

| 字段 | 为什么不能省 |
|---|---|
| `Part[]`（text / tool_use / tool_result / thinking / image） | 各协议对"工具结果"的位置和形状不同 |
| `ToolCallID` 双向映射 | OpenAI 允许任意字符串，Anthropic 有格式约束 |
| `StopReason` 归一 | `end_turn` / `stop` / `tool_use` / `tool_calls` / `max_tokens` |
| `Usage` 四段 | `input` / `cache_read` / `cache_write` / `output` |
| `Event[]` 规范化 | 流式要做重试/fallback，必须能重放事件序列 |

**坑**：thinking 块的签名跨协议不可移植，Anthropic → OpenAI 必须降级成纯文本或丢弃，不能伪装保留。


