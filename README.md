# 1bot

本地模型网关（概念验证）。把编程 CLI（CodeBuddy / Codex / ZCode）统一封装在
OpenAI 兼容的 HTTP 接口后面，让任何 OpenAI 客户端用一套协议驱动不同上游。

## 架构

数据流（一次请求）：

```
OpenAI 客户端
  -> internal/openai        解析 / 翻译 / SSE
    -> internal/transport   中立请求 / 事件类型
      -> internal/plugin    插件宿主（子进程生命周期、JSON-RPC）
        -> 插件进程          plugin/*（驱动具体 CLI）
          -> codebuddy --acp / codex app-server / zcode app-server
```

网关进程只负责 HTTP、协议翻译和插件调度；认证、模型调用、上下文由上游 CLI
自己完成。网关与插件解耦：进程间遵循 `internal/proto` 的 stdio 线协议，网关内部通过
`internal/transport.Transport`（以及可选能力接口）调用插件宿主。

## 目录

```
cmd/gateway/        网关入口（main.go）
internal/
  openai/           OpenAI 兼容 HTTP 层
    server.go         /v1/chat/completions、/v1/models、/healthz、鉴权、请求日志
    translate.go      OpenAI 消息 <-> 中立消息、工具声明与参数归一化
    stream.go         SSE 流式输出 / 非流式单条输出
  transport/        中立类型：Request / Message / Event / ToolCall / ToolResult / Part
  proto/            网关↔插件 的线协议（stdio NDJSON JSON-RPC）
  plugin/           网关侧插件宿主（host.go，实现 Transport）
plugin/
  acp/              CodeBuddy 插件客户端（ACP 协议 over stdio）
    client.go         initialize 握手、session、turn、反向工具请求
    rpc.go            JSON-RPC 帧收发的底层实现
  codex/             Codex 插件客户端（app-server 变体）
  zcode/             ZCode 插件客户端（app-server 变体）
  cmd/
    plugin-codebuddy/    CodeBuddy 插件入口 + 客户端工具 MCP 桥
    plugin-codex/        Codex 插件入口
    plugin-zcode/        ZCode 插件入口
build.bat / build.sh   构建脚本
go.mod                模块 1bot（Go 1.22）
```

## 模块职责

- **`internal/openai`**：实现 OpenAI Chat Completions 协议。支持 `stream` 流式
  （SSE，`text/event-stream`）与非流式两种返回；把 OpenAI 的 `messages`、`tools`、
  `tool_choice` 翻译成中立结构再交给 transport；支持可选的 Bearer API Key 鉴权；
  可把入站请求以 JSONL 落盘（`-log`）。
- **`internal/transport`**：定义双方的中立契约——`Transport` 接口、`Message`、
  `Event`（`text` / `thought` / `tool_call`）、`ToolCall` / `ToolResult`、以及
  `Part`（`text` / `tool_call` / `tool_result`）。工具调用与结果通过 Part 在
  往返中保留，而不是被拍平为散文。
- **`internal/proto`**：网关与插件之间的线协议，形状刻意贴近 ACP，便于 CodeBuddy
  插件适配上游。已实现 `turn`（核心→插件跑一轮）、`event`（插件→核心的
  流式增量通知）、`models`（查询模型目录）；`cancel` 仅预留。每帧一行 NDJSON。
- **`internal/plugin`**：网关侧插件宿主。启动插件子进程，通过 stdio JSON-RPC
  驱动它，并把整个过程封装成 `transport.Transport`。按请求 id 匹配响应、按
  turn id 分发事件。每轮 300 秒超时。
- **`plugin/acp`**：驱动 `codebuddy --acp`。在 `initialize` 握手时声明
  `fs`（读/写文件）与 `terminal` 能力，使 CodeBuddy 把这些工具**反向请求**给网关，
  而非自己执行；反向请求被还原成 `transport.EventToolCall`，其结果由下一轮 turn
  带回。`session/prompt` 长连接跨轮保持，便于上下文与工具结果续接。
- **`plugin/cmd/plugin-codebuddy`**：插件进程入口。监听 stdio 上的 proto 帧，
  调用 `acp.Client` 跑 turn，并把文本/思考/工具调用以 `event` 回传。同时内嵌一个
  **loopback HTTP MCP 服务器**：客户端声明的 function tools 经此桥接给 CodeBuddy
  ——`tools/list` 返回工具清单，`tools/call` 把调用注入在途 turn 并阻塞，直到客户端
  在后续 turn 回传结果。工具集合变化时重建 ACP session 以携带最新工具。

## 运行方式

构建：

```
go build -o gateway.exe ./cmd/gateway
go build -o plugin-codebuddy.exe ./plugin/cmd/plugin-codebuddy
```

启动网关（指定插件、模型、监听地址等）：

```
gateway.exe -plugin plugin-codebuddy.exe -model <model-id> -y
```

客户端调用：

```
curl -N http://127.0.0.1:8787/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"<model-id>","stream":true,"messages":[{"role":"user","content":"hi"}]}'
```

网关参数（`cmd/gateway/main.go`）：
`-plugin`（插件可执行体）、`-bin`（交给插件的 CLI 名）、`-model`、`-addr`、
`-cwd`、`-y`（跳过权限确认）、`-v`（打印协议帧）、`-no-fs`（不声明 fs/terminal
能力，工具改由 API 客户端执行）、`-api-key`（HTTP 鉴权）、`-log`（入站请求日志）。

## 自定义插件：stdin/stdout JSON-RPC 协议

插件可以用任意语言实现，不需要链接 Go 或 Qt。下文描述的是 **网关 ↔ 插件**
协议，不是插件 ↔ 上游 CLI 的 ACP / app-server 协议；插件负责二者的转换。
类型以 `internal/proto/proto.go`、`internal/proto/models.go`、
`internal/transport/transport.go` 和 `internal/transport/models.go` 为准。

### 1. 启动与传输

- 网关启动插件可执行文件，并保持进程常驻，通过 **stdin 发请求、stdout 收响应和事件**。
- UTF-8 NDJSON：每个完整 JSON 对象独占一行，以换行结束；写完必须 flush。
  字符串中的换行使用 JSON 转义，不要输出多行缩进 JSON。
- **stdout 只能输出协议帧**。调试日志、CLI 启动提示、安装输出必须写 stderr；
  上游 CLI 的 stdout 必须由插件消费，不能直接透传。
- 不使用 HTTP、SSE、`Content-Length` 头或 LSP 分帧；没有网关级
  `initialize` 握手、就绪通知或能力协商。上游初始化由插件自行完成。
- 当前宿主读取单行的 Scanner 上限为 8 MiB；插件应将单帧控制在此上限以下。
- stdin EOF 时释放会话、关闭并回收上游 CLI，然后退出。宿主关闭 stdin 后等待
  约 3 秒，超时会强制结束插件；Windows 下还会尝试结束其进程树。

网关按条件传递下列参数，插件应解析其支持的参数：

| 参数 | 当前传递条件 / 含义 |
|---|---|
| `--bin <CLI>` | 网关 `-bin` 不等于默认值 `codebuddy` 时传入；未传时插件使用自己的默认 CLI |
| `--model <ID>` | 启动模型非空时传入，作为默认模型 |
| `--cwd <目录>` | 工作目录非空时传入 |
| `-v` | 开启协议诊断日志，日志写 stderr |
| `-y`、`-no-fs` | 仅启用对应网关开关时传入，属于 CodeBuddy 相关选项，不要求所有插件支持 |

在 `ui-control` 的“插件”页配置插件程序与 CLI 路径，再在“网关管理”中选择。
相对插件路径以网关程序目录为基准。CLI 安装命令只安装依赖，**不会生成兼容协议的插件**。

### 2. 帧结构与方法

请求包含 `jsonrpc: "2.0"`、`id`、`method` 和可选 `params`。
响应原样回传请求的 `id`，并且只包含 `result` 或 `error` 之一。
当前网关生成递增的整数 ID；不要将响应 ID 转成字符串。

| 方法 | 方向 | 是否有顶层 ID | 用途 |
|---|---|---|---|
| `turn` | 网关 → 插件 | 有 | 执行一轮对话，最后返回一次结果或错误 |
| `event` | 插件 → 网关 | 无 | 文本、思考或工具调用通知，用 `params.id` 关联 turn |
| `models` | 网关 → 插件 | 有 | 返回当前插件/账号可发现的模型目录，或明确不支持 |
| `cancel` | 预留 | 尚无已实现契约 | 当前宿主不发送，不能依赖它取消上游调用 |

不要把上游的 `session/update`、`thread/start` 等方法直接写到协议 stdout。
当前宿主也不提供通用的“插件反向 RPC 请求”处理；工具调用须使用下面的事件流程。

### 3. turn 请求与文本返回

stdin 示例（每个代码块中的每一行都是独立帧）：

```json
{"jsonrpc":"2.0","id":1,"method":"turn","params":{"session":"session-a","model":"model-a","messages":[{"role":"user","parts":[{"kind":"text","text":"你好"}]}],"tools":[],"toolChoice":"auto"}}
```

`params` 字段：

| 字段 | 类型 | 含义 |
|---|---|---|
| `session` | string，可省略 | 不透明会话标识，可用来复用上游上下文；不同会话应隔离 |
| `model` | string，可省略 | 请求模型；未提供时使用插件默认值 |
| `messages` | array | 完整会话消息，不是简单的最后一条 prompt |
| `tools` | array，可省略 | 客户端声明的工具，元素为 `name`、可选 `description`、`parameters`（JSON Schema 对象） |
| `toolChoice` | JSON，可省略 | 当前 HTTP 层支持 `"auto"` / `"none"`；省略或 null 按默认策略处理 |

每条消息是 `{"role":"...","parts":[...]}`，role 保留
`system` / `developer` / `user` / `assistant` / `tool`。
Part 有三种：`text` 使用 `text`，`tool_call` 使用 `call`，
`tool_result` 使用 `result`，结构见下一节。不要直接按 OpenAI 的
`messages[].content` 读取该协议。

stdout 示例：

```json
{"jsonrpc":"2.0","method":"event","params":{"id":1,"kind":"text","text":"你好！"}}
{"jsonrpc":"2.0","id":1,"result":{"stopReason":"end_turn"}}
```

- 可发送多条 `text` 增量；`thought` 事件结构相同，只把 kind 改为 `thought`。
  当前 Chat Completions 层仅在 verbose 日志中处理 thought，不把它作为正文返回。
- 正文放在 event 中，**不是放在最终 result 中**。即使 HTTP 客户端使用非流式请求，
  插件也使用相同事件协议，由网关收集结果。
- 最后必须且只能返回一次结果：`end_turn` 表示正常结束，`tool_use` 表示需要
  客户端执行工具。不要在最终结果之后继续发送该 turn 的事件。
- 新适配器应处理请求级 model；如不能切换，应明确校验/报错，不要无声地换成其他模型。
  现有 CodeBuddy、ZCode 的请求级模型切换仍有限制，模型目录不代表切换能力已实现。

### 4. 工具调用与下一轮回传

stdin：客户端声明工具。

```json
{"jsonrpc":"2.0","id":2,"method":"turn","params":{"session":"session-a","messages":[{"role":"user","parts":[{"kind":"text","text":"查询上海天气"}]}],"tools":[{"name":"weather","description":"查询天气","parameters":{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}}],"toolChoice":"auto"}}
```

stdout：插件输出完整工具调用，并结束当前 turn。

```json
{"jsonrpc":"2.0","method":"event","params":{"id":2,"kind":"tool_call","call":{"id":"call-1","name":"weather","arguments":{"city":"上海"}}}}
{"jsonrpc":"2.0","id":2,"result":{"stopReason":"tool_use"}}
```

`call.arguments` 是原始 JSON 值，通常为对象，**不是二次编码的 JSON 字符串**。
每个调用输出一条完整事件；当前协议没有工具参数分片合并机制。
`params.id` 是 turn 请求 ID，`call.id` 是工具调用 ID，两者用途不同。

工具由 API 客户端执行；下一次 HTTP 请求经网关转换为新 turn，携带历史与工具结果：

```json
{"jsonrpc":"2.0","id":3,"method":"turn","params":{"session":"session-a","messages":[{"role":"user","parts":[{"kind":"text","text":"查询上海天气"}]},{"role":"assistant","parts":[{"kind":"tool_call","call":{"id":"call-1","name":"weather","arguments":{"city":"上海"}}}]},{"role":"tool","parts":[{"kind":"tool_result","result":{"callId":"call-1","content":"晴，22℃","isError":false}}]}],"tools":[{"name":"weather","parameters":{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}}],"toolChoice":"auto"}}
```

插件用 `result.callId` 匹配原来的调用并续接上游，再输出 id 为 3 的事件与最终结果。
`content` 是字符串；`isError` 可省略，默认 false。
不要在网关所在机器擅自执行客户端声明的工具；不支持客户端工具的插件应返回错误。

### 5. models 请求

stdin：

```json
{"jsonrpc":"2.0","id":4,"method":"models"}
```

stdout：

```json
{"jsonrpc":"2.0","id":4,"result":{"models":[{"id":"model-a","name":"模型 A"},{"id":"model-b","name":"模型 B"}]}}
```

- `models` 必须是数组；确实没有模型时返回 `[]`，不能省略或返回 null。
- 每项 `id` 必须非空，`name` 可省略。重复 ID 会被网关去重。
- 返回上游实际发现的目录，不能把启动参数 `--model` 冒充完整目录；
  上游有分页时，插件负责收集所有页。
- 不能枚举时返回错误，而不是空列表：

```json
{"jsonrpc":"2.0","id":4,"error":{"code":-32601,"message":"plugin does not expose a model catalog"}}
```

网关把结果转为 OpenAI `GET /v1/models` 的 `object: "list"` / `data`；
`owned_by` 由插件文件名推导，不从 UI 显示名称读取。
`-32601` 映射为 HTTP 501，查询超时为 504，其他查询错误为 502。
当前 Codex 调用 `model/list`；CodeBuddy 解析 ACP 的
`models.availableModels`（后端不提供时返回不支持）；ZCode 暂返回不支持。

### 6. 错误、并发与实现检查

业务或上游错误示例：

```json
{"jsonrpc":"2.0","id":5,"error":{"code":-32000,"message":"upstream unavailable"}}
```

建议新插件对未知的带 ID 方法返回 `-32601`，对无效参数返回 `-32602`；
错误信息不得包含 API key、令牌或完整敏感请求。现有插件对其他未知方法可能直接忽略，
新实现不应依赖这种行为。

当前宿主对 turn 等待最长 300 秒，对 models 等待最长 30 秒。
**等待超时不等于上游执行已取消**；取消协议目前未接通。
宿主可以发出多个请求并按 ID 分发；现有三个插件入口串行处理，
长 turn 可能使 models 排队并超时。若新插件并发处理查询和 turn，
必须串行化 stdout 写入（例如加锁），避免帧交错，同时保护上游会话状态。

接入前至少验证：

1. 启动后可接收连续多行请求，stdout 没有任何非 JSON 日志，每帧及时 flush。
2. 文本事件、工具事件都能按 turn ID 关联，每次请求恰好一个最终结果或错误。
3. 能处理工具结果续接；不支持的工具、模型选择明确报错。
4. models 返回有效目录或明确不支持；未知方法不会令调用方无限等待。
5. stdin EOF 后关闭上游子进程；并发写入不会破坏 NDJSON。
6. 初始化和调用失败可诊断，但不在 stdout 或错误消息中泄露凭据。

可参考 `plugin/cmd/plugin-codex/main.go` 的请求分发、
`plugin/cmd/plugin-codebuddy/main.go` 的工具桥接，以及
`internal/plugin/host.go` / `internal/plugin/models.go` 的宿主行为。

## 当前实现边界

- 网关启动时**只选一个插件**，不按 HTTP `model` 自动路由到不同上游。
- 一次只跑**一个在途 turn**（ACP 单进程串行）；会话依据首条 user 消息派生的
  key 保持连续，新会话自动开启新 key。
- `tool_choice` 目前仅支持 `auto` / `none`。
- 不内置额度、多账户、fallback、控制面 API；这些是后续演进方向。
- 单二进制（网关） + 独立插件二进制，插件进程与上游 CLI 进程分离，便于隔离与扩展。
