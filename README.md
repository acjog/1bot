# Agent Gateway (Rust)

基于 Rust + Axum + Tokio 实现的多 Agent 通信网关，作为 Orchestrator（高阶编排器）与 Remote Bot（Worker）之间的中间人桥接层。 Gateway 可以代理Remote Bot所有技能。

---

## 特性

- **A2A 协议接入** — JSON-RPC 2.0 over HTTPS
- **SSE 实时进度推送** — Orchestrator 通过 SSE 接收任务进度流
- **WebSocket 长连接** — Remote Bot 主动出站连接，无需公网地址
- **服务注册与发现** — 自动维护 Agent Card、心跳检测、过期清理
- **路由策略** — 基于技能 + 最低负载的路由算法
- **超时与容错** — 任务超时自动失败，Agent 心跳超时自动摘除
- **多租户 Namespace 隔离** — 不同租户的 Agent 和任务完全隔离
- **mTLS 双向证书认证** — Remote Bot 接入时验证客户端证书，防止未授权 Agent 接入
- **OAuth2 鉴权 (JWT)** — Orchestrator 接入时验证 Bearer Token，支持租户身份识别

## 架构概览

```
Orchestrator
   │  ── HTTP POST ─→ Gateway   :8443  (公网/DMZ)
   │  ←── SSE Stream ─ Gateway
   │
   ▼
Gateway
   │  ←── WebSocket ─→ Remote Bot   :8080  (内网，主动出站)
   ▼
   双向消息：任务下发 + 进度回传
```

---

## 快速开始

### 1. 编译

```bash
cd gateway
cargo build --release
```

### 2. 启动 Gateway

```bash
cargo run --release
```

默认端口：
- `8443` — Orchestrator 接入（HTTPS + JSON-RPC + SSE）
- `8080` — Remote Bot 注册（WebSocket）

### 3. 启动一个 Mock Nanobot 测试

```bash
cargo run --example mock_nanobot -- ws://127.0.0.1:8080/agent/register nanobot-text-001
```

### 4. 模拟 Orchestrator 下发任务

```bash
# 创建任务
curl -X POST http://127.0.0.1:8443/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "skill": "summarize",
    "instruction": "请摘要：人工智能正在改变世界，深度学习模型..."
  }'



# 响应
# {"task_id": "t-xxxx", "status": "queued", "assigned_agent": "nanobot-text-001"}
```

### 5. 订阅任务进度（SSE）
curl -N 即 --no-buffer，意思是禁用输出缓冲。

```bash
curl -N http://127.0.0.1:8443/tasks/t-a0f85faa-3942-4282-bb1d-afd5e775f45b/stream
```

curl -N http://1bot.net/tasks/t-1fb56ffe-f519-449f-bbdd-265234c3f269/stream

输出示例：
```
event: queued
data: {"event":"queued","task_id":"t-xxxx","agent_id":"nanobot-text-001"}

event: progress
data: {"event":"progress","task_id":"t-xxxx","pct":30,"message":"starting..."}

event: progress
data: {"event":"progress","task_id":"t-xxxx","pct":80,"message":"processing..."}

event: completed
data: {"event":"completed","task_id":"t-xxxx","artifact":{"text":"..."}}
```

### 6. 列出在线 Agent

```bash
curl http://127.0.0.1:8443/agents
```

---

## Agent Card（能力发现）

Gateway 作为 A2A Server，通过标准的 `/.well-known/agent.json` 端点暴露自身能力，便于其他 Agent 发现和接入。

### 获取 Agent Card

```bash
curl http://127.0.0.1:8443/.well-known/agent.json | jq .
```

### 响应示例

`skills` 字段是动态聚合的，反映当前所有在线 Remote Bot 的技能。以下为示例（假设已有两个 Nanobot 在线）：

```json
{
  "name": "Agent Gateway",
  "description": "多 Agent 通信网关，作为 Orchestrator 与 Remote Bot（Nanobot Worker）之间的桥接层。",
  "url": "http://127.0.0.1:8443",
  "version": "0.1.0",
  "protocol": "a2a",
  "protocolVersion": "0.2.0",
  "capabilities": {
    "streaming": true,
    "pushNotifications": false,
    "stateTransitionHistory": false
  },
  "skills": [
    {
      "id": "summarize",
      "name": "Text Summarization",
      "description": "对输入文本进行摘要",
      "inputModes": ["text"],
      "outputModes": ["text"]
    },
    {
      "id": "translate",
      "name": "Translation",
      "description": "文本翻译",
      "inputModes": ["text"],
      "outputModes": ["text"]
    }
  ],
  "endpoints": {
    "orchestrator": {
      "base_url": "http://127.0.0.1:8443",
      "auth": "OAuth2 Bearer JWT",
      "routes": {
        "agent_card": "GET /.well-known/agent.json",
        "create_task": "POST /tasks",
        "get_task": "GET /tasks/:task_id",
        "sse_stream": "GET /tasks/:task_id/stream",
        "list_agents": "GET /agents",
        "jsonrpc": "POST /jsonrpc"
      }
    },
    "remote_bot": {
      "base_url": "ws://127.0.0.1:8080",
      "auth": "mTLS 客户端证书",
      "routes": {
        "register": "GET /agent/register (WebSocket Upgrade)"
      }
    }
  },
  "_meta": {
    "online_agents": 2,
    "total_concurrency": 6
  }
}
```

### Skills 说明

Agent Card 中的 `skills` 是 **所有在线 Remote Bot 技能的动态聚合**，即 Gateway 对外代理的完整能力集合。其他 Agent 可以通过该列表了解 Gateway 能处理哪些类型的任务。

| 字段 | 说明 |
|------|------|
| `skills` | 当前所有在线 Remote Bot 技能的去重聚合，动态更新 |
| `_meta.online_agents` | 当前在线 Remote Bot 数量 |
| `_meta.total_concurrency` | 所有 Remote Bot 的总并发容量 |

> **注意**：当没有 Remote Bot 在线时，`skills` 为空数组。随着 Remote Bot 的接入和离线，`skills` 实时变化。

### 静态参考文件

项目根目录下的 `agent-card.json` 是完整的静态参考模板，包含所有字段和示例值。

---

---

## A2A 协议接入（快速示例）

也可以通过标准 JSON-RPC 接入：

```bash
curl -X POST http://127.0.0.1:8443/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tasks/send",
    "params": {
      "skill": "file-process",
      "instruction": "你是什么模型"
    },
    "id": "1"
  }'
```

支持的方法：
- `tasks/send` — 创建任务
- `tasks/get` — 查询任务状态
- `agents/list` — 列出在线 Agent

---

## 配置

支持环境变量或 `config.toml` 两种方式，优先级：环境变量 > 配置文件 > 默认值。

### 环境变量

```bash
GATEWAY_BOT_B_ADDR=0.0.0.0:8443
GATEWAY_BOT_A_ADDR=0.0.0.0:8080
GATEWAY_TASK_TIMEOUT_SECS=60
GATEWAY_HEARTBEAT_TIMEOUT_SECS=30

# mTLS（Remote Bot 端）
GATEWAY_MTLS_ENABLED=true
GATEWAY_MTLS_CERT_PATH=certs/server.crt
GATEWAY_MTLS_KEY_PATH=certs/server.key
GATEWAY_MTLS_CA_PATH=certs/ca.crt

# OAuth2（Orchestrator 端）
GATEWAY_OAUTH2_ENABLED=true
GATEWAY_OAUTH2_JWT_SECRET=your-secret-key
GATEWAY_OAUTH2_ISSUER=gateway-auth
```

### config.toml

```toml
bot_b_addr = "0.0.0.0:8443"
bot_a_addr = "0.0.0.0:8080"
task_timeout_secs = 60
heartbeat_timeout_secs = 30
bot_b_rate_limit = 100

# mTLS 双向证书认证（Remote Bot 接入）
mtls_enabled = true
mtls_cert_path = "certs/server.crt"
mtls_key_path = "certs/server.key"
mtls_ca_path = "certs/ca.crt"

# OAuth2 鉴权（Orchestrator 接入）
oauth2_enabled = true
oauth2_jwt_secret = "your-secret-key"
oauth2_issuer = "gateway-auth"
```

### mTLS 证书生成

```bash
# 生成 CA
openssl req -x509 -newkey rsa:4096 -keyout certs/ca.key -out certs/ca.crt -days 365 -nodes -subj "/CN=Gateway CA"

# 生成 Gateway 服务端证书
openssl req -newkey rsa:4096 -keyout certs/server.key -out certs/server.csr -nodes -subj "/CN=gateway"
openssl x509 -req -in certs/server.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/server.crt -days 365

# 生成 Remote Bot 客户端证书
openssl req -newkey rsa:4096 -keyout certs/client.key -out certs/client.csr -nodes -subj "/CN=nanobot-001"
openssl x509 -req -in certs/client.csr -CA certs/ca.crt -CAkey certs/ca.key -CAcreateserial -out certs/client.crt -days 365
```

### OAuth2 JWT Token

Orchestrator 请求时需携带 `Authorization: Bearer <token>`，JWT Payload 结构：

```json
{
  "sub": "tenant-001",
  "namespace": "production",
  "exp": 1700000000,
  "iss": "gateway-auth"
}
```

- `sub` — 租户/用户标识
- `namespace` — 命名空间，用于多租户隔离（只能访问同 namespace 下的 Agent）
- `exp` — 过期时间（Unix 时间戳）
- `iss` — 签发者（需与配置中的 `oauth2_issuer` 匹配）

---

## 模块结构

```
src/
├── main.rs              # 入口，启动两个 server + 后台监控 + mTLS 服务
├── config.rs            # 配置加载（含 mTLS / OAuth2 配置）
├── auth.rs              # 认证模块：OAuth2 中间件 + mTLS 配置构建
├── error.rs             # 错误类型 + IntoResponse
├── state.rs             # AppState 共享状态
├── registry.rs          # Agent 注册表 + Agent Card + Namespace 隔离
├── task.rs              # 任务存储 + 事件广播
├── router.rs            # 路由策略（已合并到 registry）
└── handlers/
    ├── mod.rs
    ├── common.rs        # health、agent_card、skill_md
    ├── remote_bot.rs    # Remote Bot WebSocket handler（支持 namespace 注册）
    └── orchestrator.rs  # Orchestrator HTTP/SSE/JSON-RPC handler（支持 JWT 鉴权）
```

---

## 协议详细文档

> 完整协议文档也可通过 `GET /.well-known/skill.md` 在线获取。

---

### Remote Bot 协议（WebSocket）

Remote Bot 通过 WebSocket 连接 Gateway 的 `:8080/agent/register` 端点，建立全双工长连接。

#### 连接地址

```
ws://127.0.0.1:8080/agent/register       # 普通模式
wss://127.0.0.1:8080/agent/register      # mTLS 模式（需携带客户端证书）
```

#### Remote Bot → Gateway（上行消息，A2A JSON-RPC 2.0）

##### 1. x-agent/register（注册，必须是连接后的第一条消息）

```json
{
  "jsonrpc": "2.0",
  "method": "x-agent/register",
  "params": {
    "card": {
      "name": "nanobot-text-001",
      "version": "1.0.0",
      "description": "文本处理 Nanobot，支持摘要和翻译",
      "skills": [
        {
          "id": "summarize",
          "name": "Text Summarization",
          "description": "对输入文本进行摘要",
          "input_modes": ["text"],
          "output_modes": ["text"]
        },
        {
          "id": "translate",
          "name": "Translation",
          "description": "文本翻译",
          "input_modes": ["text"],
          "output_modes": ["text"]
        }
      ],
      "capabilities": {
        "streaming": true,
        "push_notifications": false
      },
      "max_concurrency": 3
    },
    "namespace": "production"
  },
  "id": "register-1"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `method` | string | ✅ | 固定值 `"x-agent/register"` |
| `params.card.name` | string | ✅ | Agent 唯一标识（同时作为 agent_id） |
| `params.card.version` | string | ✅ | Agent 版本号 |
| `params.card.description` | string | ❌ | Agent 描述 |
| `params.card.skills` | array | ✅ | 支持的技能列表 |
| `params.card.skills[].id` | string | ✅ | 技能 ID（Gateway 按此路由任务） |
| `params.card.skills[].name` | string | ✅ | 技能名称 |
| `params.card.skills[].description` | string | ❌ | 技能描述 |
| `params.card.skills[].input_modes` | array | ❌ | 输入模式（如 `["text", "image"]`） |
| `params.card.skills[].output_modes` | array | ❌ | 输出模式 |
| `params.card.capabilities.streaming` | bool | ❌ | 是否支持流式输出（默认 false） |
| `params.card.capabilities.push_notifications` | bool | ❌ | 是否支持推送通知（默认 false） |
| `params.card.max_concurrency` | u32 | ❌ | 最大并发任务数（默认 5） |
| `params.namespace` | string | ❌ | 命名空间（默认 `"default"`，用于多租户隔离） |

##### 2. tasks/status（任务进度上报）

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/status",
  "params": {
    "id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "status": "running",
    "progress": 50,
    "message": "正在处理第 5/10 段..."
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `method` | string | ✅ | 固定值 `"tasks/status"` |
| `params.id` | string | ✅ | 任务 ID |
| `params.status` | string | ✅ | 状态：`"running"` |
| `params.progress` | u32 | ✅ | 进度百分比（0-100） |
| `params.message` | string | ❌ | 进度描述信息 |

##### 3. tasks/status（任务完成，返回结果）

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/status",
  "params": {
    "id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "status": "completed",
    "artifact": {
      "text": "这是摘要结果：人工智能正在深刻改变各行各业...",
      "metadata": {
        "model": "qwen-7b",
        "tokens_used": 256
      }
    }
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `method` | string | ✅ | 固定值 `"tasks/status"` |
| `params.id` | string | ✅ | 任务 ID |
| `params.status` | string | ✅ | 状态：`"completed"` |
| `params.artifact` | object | ✅ | 任务产出物（结构自定义） |

##### 4. tasks/status（任务失败）

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/status",
  "params": {
    "id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "status": "failed",
    "error": "模型推理超时：GPU 显存不足"
  }
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `method` | string | ✅ | 固定值 `"tasks/status"` |
| `params.id` | string | ✅ | 任务 ID |
| `params.status` | string | ✅ | 状态：`"failed"` |
| `params.error` | string | ✅ | 错误描述 |

##### 5. Pong（心跳响应，作为 x-agent/ping 的 JSON-RPC response）

```json
{
  "jsonrpc": "2.0",
  "result": {"status": "ok"},
  "id": "ping"
}
```

收到 Gateway 下发的 `x-agent/ping` 后必须回复对应的 JSON-RPC response，否则 30 秒无响应将被摘除。

---

#### Gateway → Remote Bot（下行消息，A2A JSON-RPC 2.0）

##### 1. tasks/send（任务分派）

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/send",
  "params": {
    "id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "message": {
      "role": "user",
      "parts": [{"type": "text", "text": "请对以下文本进行摘要：人工智能正在改变世界，深度学习模型在自然语言处理领域取得了突破性进展..."}]
    },
    "metadata": {
      "skill": "summarize",
      "constraints": {"max_length": 100, "language": "zh"}
    }
  },
  "id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `method` | string | ✅ | 固定值 `"tasks/send"` |
| `params.id` | string | ✅ | 任务 ID（全局唯一） |
| `params.message.role` | string | ✅ | 消息角色（`"user"`） |
| `params.message.parts` | array | ✅ | 消息内容部分 |
| `params.message.parts[].type` | string | ✅ | 内容类型（`"text"`） |
| `params.message.parts[].text` | string | ✅ | 任务指令/输入内容 |
| `params.metadata.skill` | string | ✅ | 要求的技能 ID |
| `params.metadata.constraints` | object | ❌ | 约束条件（默认 `{}`） |
| `id` | string | ✅ | JSON-RPC 请求 ID（与 task_id 相同） |

##### 2. tasks/cancel（取消任务）

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/cancel",
  "params": {
    "id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  },
  "id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `method` | string | ✅ | 固定值 `"tasks/cancel"` |
| `params.id` | string | ✅ | 要取消的任务 ID |

##### 3. x-agent/ping（心跳探活）

```json
{
  "jsonrpc": "2.0",
  "method": "x-agent/ping",
  "params": {},
  "id": "ping"
}
```

Gateway 每 10 秒下发一次，Remote Bot 必须回复 JSON-RPC response：`{"jsonrpc":"2.0","result":{"status":"ok"},"id":"ping"}`。

---

#### Remote Bot 完整交互时序示例

```
Remote Bot                              Gateway
 │                                        │
 │──── WebSocket Connect ────────────────→│
 │                                        │
 │──── x-agent/register ────────────────→│  注册成功
 │                                        │
 │←─── x-agent/ping ────────────────────│  心跳探活（每10s）
 │──── response {id:"ping"} ───────────→│
 │                                        │
 │←─── tasks/send ──────────────────────│  收到任务
 │                                        │
 │──── tasks/status (running,pct=10) ──→│  上报进度 10%
 │──── tasks/status (running,pct=50) ──→│  上报进度 50%
 │──── tasks/status (running,pct=100) ─→│  上报进度 100%
 │                                        │
 │──── tasks/status (completed) ───────→│  返回结果
 │                                        │
 │←─── x-agent/ping ────────────────────│  继续心跳...
 │──── response {id:"ping"} ───────────→│
```

---

---

### Orchestrator 协议（HTTP/SSE）

Orchestrator 通过 HTTP 连接 Gateway 的 `:8443` 端口，支持 RESTful API 和 JSON-RPC 两种方式。

#### 端点列表

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/tasks` | 创建任务（RESTful） |
| GET | `/tasks/:task_id` | 查询任务状态 |
| GET | `/tasks/:task_id/stream` | SSE 订阅任务进度流 |
| POST | `/jsonrpc` | JSON-RPC 2.0 统一入口 |
| GET | `/agents` | 列出在线 Agent |
| GET | `/health` | 健康检查 |

#### 认证（启用 OAuth2 时）

所有请求需携带 JWT Token：

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

#### 1. 创建任务（RESTful）

**请求：**

```bash
curl -X POST http://127.0.0.1:8443/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "skill": "summarize",
    "instruction": "请对以下文本进行摘要：人工智能正在改变世界...",
    "constraints": {"max_length": 100}
  }'
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `skill` | string | ✅ | 要求的技能 ID |
| `instruction` | string | ✅ | 任务指令/输入内容 |
| `constraints` | object | ❌ | 约束条件（默认 `{}`） |

**响应（200 OK）：**

```json
{
  "task_id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "queued",
  "assigned_agent": "nanobot-text-001"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `task_id` | string | 任务 ID（用于后续查询和订阅） |
| `status` | string | 初始状态（`queued`） |
| `assigned_agent` | string | 被分配的 Agent ID |

---

#### 2. 查询任务状态

**请求：**

```bash
curl http://127.0.0.1:8443/tasks/t-a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

**响应（200 OK）：**

```json
{
  "task_id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "skill": "summarize",
  "instruction": "请对以下文本进行摘要...",
  "constraints": {"max_length": 100},
  "status": "completed",
  "assigned_agent": "nanobot-text-001",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:45Z",
  "artifact": {
    "text": "摘要结果..."
  },
  "error": null
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `task_id` | string | 任务 ID |
| `skill` | string | 技能 ID |
| `instruction` | string | 任务指令 |
| `constraints` | object | 约束条件 |
| `status` | string | 状态：`created` / `queued` / `running` / `completed` / `failed` / `canceled` |
| `assigned_agent` | string\|null | 分配的 Agent ID |
| `created_at` | string | 创建时间（ISO 8601） |
| `updated_at` | string | 最后更新时间 |
| `artifact` | object\|null | 任务产出物（完成时有值） |
| `error` | string\|null | 错误信息（失败时有值） |

---

#### 3. SSE 订阅任务进度流

**请求：**

```bash
curl -N http://127.0.0.1:8443/tasks/t-a1b2c3d4-e5f6-7890-abcd-ef1234567890/stream
```

**响应（`Content-Type: text/event-stream`）：**

```
event: queued
data: {"event":"queued","task_id":"t-a1b2c3d4-e5f6-7890-abcd-ef1234567890","agent_id":"nanobot-text-001"}

event: progress
data: {"event":"progress","task_id":"t-a1b2c3d4-e5f6-7890-abcd-ef1234567890","pct":10,"message":"processing 10%..."}

event: progress
data: {"event":"progress","task_id":"t-a1b2c3d4-e5f6-7890-abcd-ef1234567890","pct":50,"message":"processing 50%..."}

event: progress
data: {"event":"progress","task_id":"t-a1b2c3d4-e5f6-7890-abcd-ef1234567890","pct":100,"message":"processing 100%..."}

event: completed
data: {"event":"completed","task_id":"t-a1b2c3d4-e5f6-7890-abcd-ef1234567890","artifact":{"text":"摘要结果..."}}
```

**SSE 事件类型：**

| 事件名 | 触发时机 | data 字段 |
|--------|---------|-----------|
| `queued` | 任务已入队，等待 Agent 处理 | `task_id`, `agent_id` |
| `running` | Agent 开始处理 | `task_id` |
| `progress` | 进度更新 | `task_id`, `pct`(0-100), `message` |
| `completed` | 任务完成 | `task_id`, `artifact` |
| `failed` | 任务失败 | `task_id`, `error` |
| `canceled` | 任务被取消 | `task_id` |

> **注意**：SSE 连接建立时会先推送当前任务状态作为初始事件，即使任务已完成也能看到结果。

---

#### 4. JSON-RPC 2.0 统一入口

##### tasks/send — 创建任务

**请求：**

```bash
curl -X POST http://127.0.0.1:8443/jsonrpc \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tasks/send",
    "params": {
      "skill": "translate",
      "instruction": "Translate to English: 今天天气真好",
      "constraints": {"target_lang": "en"}
    },
    "id": "req-001"
  }'
```

**响应：**

```json
{
  "jsonrpc": "2.0",
  "result": {
    "task_id": "t-b2c3d4e5-f6a7-8901-bcde-f23456789012",
    "status": "queued",
    "assigned_agent": "nanobot-text-001"
  },
  "id": "req-001"
}
```

##### tasks/get — 查询任务

**请求：**

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/get",
  "params": {
    "taskId": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  },
  "id": "req-002"
}
```

**响应：**

```json
{
  "jsonrpc": "2.0",
  "result": {
    "task_id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "skill": "summarize",
    "instruction": "...",
    "constraints": {},
    "status": "completed",
    "assigned_agent": "nanobot-text-001",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:45Z",
    "artifact": {"text": "摘要结果..."},
    "error": null
  },
  "id": "req-002"
}
```

##### agents/list — 列出在线 Agent

**请求：**

```json
{
  "jsonrpc": "2.0",
  "method": "agents/list",
  "params": {},
  "id": "req-003"
}
```

**响应：**

```json
{
  "jsonrpc": "2.0",
  "result": [
    {
      "id": "nanobot-text-001",
      "card": {
        "name": "nanobot-text-001",
        "version": "1.0.0",
        "description": "文本处理 Nanobot",
        "skills": [
          {"id": "summarize", "name": "Text Summarization", "description": "", "input_modes": ["text"], "output_modes": ["text"]},
          {"id": "translate", "name": "Translation", "description": "", "input_modes": ["text"], "output_modes": ["text"]}
        ],
        "capabilities": {"streaming": true, "push_notifications": false},
        "max_concurrency": 3
      },
      "load": 1,
      "max_concurrency": 3,
      "connected_at": "2025-01-15T10:00:00Z"
    }
  ],
  "id": "req-003"
}
```

---

#### 5. 列出在线 Agent（RESTful）

**请求：**

```bash
curl http://127.0.0.1:8443/agents
```

**响应：**

```json
{
  "agents": [
    {
      "id": "nanobot-text-001",
      "skills": ["summarize", "translate"],
      "load": "1/3",
      "connected_at": "2025-01-15T10:00:00Z",
      "last_heartbeat": "2025-01-15T10:30:50Z"
    }
  ]
}
```

---

#### Orchestrator 完整交互时序示例

```
Orchestrator                            Gateway                                 Remote Bot
 │                                        │                                      │
 │── POST /tasks ────────────────────────→│                                      │
 │                                        │──── tasks/send (JSON-RPC) ──────────→│
 │←── {"task_id":"t-xxx","status":"queued"}│                                      │
 │                                        │                                      │
 │── GET /tasks/t-xxx/stream ────────────→│                                      │
 │←── event: queued ─────────────────────│                                      │
 │                                        │←─── tasks/status (running,pct=10) ──│
 │←── event: progress (pct=10) ──────────│                                      │
 │                                        │←─── tasks/status (running,pct=50) ──│
 │←── event: progress (pct=50) ──────────│                                      │
 │                                        │←─── tasks/status (running,pct=100) ─│
 │←── event: progress (pct=100) ─────────│                                      │
 │                                        │←─── tasks/status (completed) ───────│
 │←── event: completed ──────────────────│                                      │
 │                                        │                                      │
 │── GET /tasks/t-xxx ───────────────────→│                                      │
 │←── {"status":"completed","artifact":{}}│                                      │
```

---

---

### 端到端完整示例

以下是一个完整的端到端测试流程：

```bash
# 终端 1：启动 Gateway
cargo run --release

# 终端 2：启动 Mock Nanobot（注册 summarize 和 translate 技能）
cargo run --example mock_nanobot -- ws://127.0.0.1:8080/agent/register nanobot-text-001

# 终端 3：创建任务
curl -s -X POST http://127.0.0.1:8443/tasks \
  -H "Content-Type: application/json" \
  -d '{"skill":"summarize","instruction":"请摘要：AI正在改变世界"}' | jq .
# 输出: {"task_id":"t-xxxx","status":"queued","assigned_agent":"nanobot-text-001"}

# 终端 4：订阅 SSE 进度流（用上一步返回的 task_id）
curl -N http://127.0.0.1:8443/tasks/t-xxxx/stream
# 输出:
# event: queued
# data: {"event":"queued","task_id":"t-xxxx","agent_id":"nanobot-text-001"}
#
# event: progress
# data: {"event":"progress","task_id":"t-xxxx","pct":10,"message":"processing 10%..."}
# ...（每 6 秒一次，直到 100%）
#
# event: completed
# data: {"event":"completed","task_id":"t-xxxx","artifact":{"text":"[mock summary] AI正在改变世界"}}

# 查询最终结果
curl -s http://127.0.0.1:8443/tasks/t-xxxx | jq .
```

---

## 后续扩展

- [x] mTLS 双向证书认证（Remote Bot 接入）
- [x] OAuth2 鉴权（Orchestrator 接入）
- [x] 多租户 namespace 隔离
- [ ] Redis 持久化（多 Gateway 实例共享状态）
- [ ] OpenTelemetry 分布式追踪
- [ ] Prometheus 指标导出
- [ ] 任务持久化与恢复

## License

MIT
