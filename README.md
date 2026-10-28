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
自己完成。网关与插件解耦——`internal/transport.Transport` 是它们之间唯一的契约。

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
  插件几乎原样透传。三种方法：`turn`（核心→插件跑一轮）、`event`（插件→核心的
  流式增量通知）、`cancel`（预留）。每帧一行 NDJSON。
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

## 当前实现边界

- 网关启动时**只选一个插件**，不按 HTTP `model` 自动路由到不同上游。
- 一次只跑**一个在途 turn**（ACP 单进程串行）；会话依据首条 user 消息派生的
  key 保持连续，新会话自动开启新 key。
- `tool_choice` 目前仅支持 `auto` / `none`。
- 不内置额度、多账户、fallback、控制面 API；这些是后续演进方向。
- 单二进制（网关） + 独立插件二进制，插件进程与上游 CLI 进程分离，便于隔离与扩展。
