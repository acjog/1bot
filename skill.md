# Agent Gateway - 接口协议文档

> 本文档描述 Agent Gateway 支持的所有接口协议，供 Orchestrator 和 Remote Bot 接入参考。

---

## 概述

Agent Gateway 是多 Agent 通信网关，作为 Orchestrator（高阶编排器）与 Remote Bot（Worker）之间的桥接层。

- **Orchestrator 端**：`:8443`，HTTP/HTTPS + JSON-RPC + SSE
- **Remote Bot 端**：`:8080`，WebSocket 全双工长连接

---

## Orchestrator 协议（HTTP/SSE）

### 认证

启用 OAuth2 时，所有请求需携带 JWT Token：

```
Authorization: Bearer <JWT Token>
```

JWT Payload 结构：

```json
{
  "sub": "tenant-001",
  "namespace": "production",
  "exp": 1700000000,
  "iss": "gateway-auth"
}
```

### 端点列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/.well-known/agent.json` | 获取 Agent Card（能力发现） |
| GET | `/.well-known/skill.md` | 获取本协议文档 |
| POST | `/tasks` | 创建任务（RESTful） |
| GET | `/tasks/:task_id` | 查询任务状态 |
| GET | `/tasks/:task_id/stream` | SSE 订阅任务进度流 |
| POST | `/jsonrpc` | JSON-RPC 2.0 统一入口 |
| GET | `/agents` | 列出在线 Agent |
| GET | `/health` | 健康检查 |

---

### 1. 创建任务

**POST /tasks**

```json
{
  "skill": "summarize",
  "instruction": "请对以下文本进行摘要：...",
  "constraints": {"max_length": 100}
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|:----:|------|
| `skill` | string | ✅ | 要求的技能 ID |
| `instruction` | string | ✅ | 任务指令/输入内容 |
| `constraints` | object | ❌ | 约束条件（默认 `{}`） |

**响应（200）：**

```json
{
  "task_id": "t-a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "queued",
  "assigned_agent": "nanobot-text-001"
}
```

---

### 2. 查询任务状态

**GET /tasks/:task_id**

**响应（200）：**

```json
{
  "task_id": "t-xxx",
  "skill": "summarize",
  "instruction": "...",
  "constraints": {},
  "status": "completed",
  "assigned_agent": "nanobot-text-001",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:45Z",
  "artifact": {"text": "摘要结果..."},
  "error": null
}
```

任务状态枚举：`created` | `queued` | `running` | `completed` | `failed` | `canceled`

---

### 3. SSE 订阅任务进度流

**GET /tasks/:task_id/stream**

响应 `Content-Type: text/event-stream`：

```
event: queued
data: {"event":"queued","task_id":"t-xxx","agent_id":"nanobot-text-001"}

event: progress
data: {"event":"progress","task_id":"t-xxx","pct":50,"message":"processing..."}

event: completed
data: {"event":"completed","task_id":"t-xxx","artifact":{"text":"结果..."}}
```

**SSE 事件类型：**

| 事件名 | 触发时机 | data 字段 |
|--------|---------|-----------|
| `queued` | 任务已入队 | `task_id`, `agent_id` |
| `running` | Agent 开始处理 | `task_id` |
| `progress` | 进度更新 | `task_id`, `pct`(0-100), `message` |
| `completed` | 任务完成 | `task_id`, `artifact` |
| `failed` | 任务失败 | `task_id`, `error` |
| `canceled` | 任务被取消 | `task_id` |

---

### 4. JSON-RPC 2.0 统一入口

**POST /jsonrpc**

支持的方法：

#### tasks/send

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/send",
  "params": {
    "skill": "summarize",
    "instruction": "任务内容",
    "constraints": {}
  },
  "id": "req-001"
}
```

#### tasks/get

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/get",
  "params": {"taskId": "t-xxx"},
  "id": "req-002"
}
```

#### agents/list

```json
{
  "jsonrpc": "2.0",
  "method": "agents/list",
  "params": {},
  "id": "req-003"
}
```

---

### 5. 列出在线 Agent

**GET /agents**

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

## Remote Bot 协议（WebSocket）

### 连接地址

```
ws://127.0.0.1:8080/agent/register       # 普通模式
wss://127.0.0.1:8080/agent/register      # mTLS 模式（需携带客户端证书）
```

### 认证

启用 mTLS 时，Remote Bot 需携带客户端证书进行双向认证。

---

### Remote Bot → Gateway（上行消息，A2A JSON-RPC 2.0）

#### 1. x-agent/register（注册，连接后第一条消息）

```json
{
  "jsonrpc": "2.0",
  "method": "x-agent/register",
  "params": {
    "card": {
      "name": "nanobot-text-001",
      "version": "1.0.0",
      "description": "文本处理 Nanobot",
      "skills": [
        {
          "id": "summarize",
          "name": "Text Summarization",
          "description": "对输入文本进行摘要",
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
| `params.card.name` | string | ✅ | Agent 唯一标识 |
| `params.card.version` | string | ✅ | Agent 版本号 |
| `params.card.skills` | array | ✅ | 支持的技能列表 |
| `params.card.skills[].id` | string | ✅ | 技能 ID（Gateway 按此路由） |
| `params.card.max_concurrency` | u32 | ❌ | 最大并发数（默认 5） |
| `params.namespace` | string | ❌ | 命名空间（默认 `"default"`） |

#### 2. tasks/status（进度上报）

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/status",
  "params": {
    "id": "t-xxx",
    "status": "running",
    "progress": 50,
    "message": "正在处理..."
  }
}
```

#### 3. tasks/status（任务完成）

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/status",
  "params": {
    "id": "t-xxx",
    "status": "completed",
    "artifact": {
      "text": "处理结果...",
      "metadata": {"model": "qwen-7b", "tokens_used": 256}
    }
  }
}
```

#### 4. tasks/status（任务失败）

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/status",
  "params": {
    "id": "t-xxx",
    "status": "failed",
    "error": "模型推理超时"
  }
}
```

#### 5. Pong（心跳响应，作为 x-agent/ping 的 JSON-RPC response）

```json
{
  "jsonrpc": "2.0",
  "result": {"status": "ok"},
  "id": "ping"
}
```

---

### Gateway → Remote Bot（下行消息，A2A JSON-RPC 2.0）

#### 1. tasks/send（任务分派）

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/send",
  "params": {
    "id": "t-xxx",
    "message": {
      "role": "user",
      "parts": [{"type": "text", "text": "请对以下文本进行摘要..."}]
    },
    "metadata": {
      "skill": "summarize",
      "constraints": {"max_length": 100}
    }
  },
  "id": "t-xxx"
}
```

#### 2. tasks/cancel（取消任务）

```json
{
  "jsonrpc": "2.0",
  "method": "tasks/cancel",
  "params": {"id": "t-xxx"},
  "id": "t-xxx"
}
```

#### 3. x-agent/ping（心跳探活）

```json
{
  "jsonrpc": "2.0",
  "method": "x-agent/ping",
  "params": {},
  "id": "ping"
}
```

Gateway 每 10 秒下发一次，Remote Bot 必须回复 JSON-RPC response（id 为 `"ping"`），超时 30s 未响应将被摘除。

---

## 交互时序

```
Orchestrator                            Gateway                                 Remote Bot
 │                                        │                                      │
 │                                        │←─── WebSocket Connect ──────────────│
 │                                        │←─── x-agent/register ──────────────│
 │                                        │                                      │
 │── POST /tasks ────────────────────────→│                                      │
 │                                        │──── tasks/send ────────────────────→│
 │←── {"task_id":"t-xxx","status":"queued"}│                                      │
 │                                        │                                      │
 │── GET /tasks/t-xxx/stream ────────────→│                                      │
 │←── event: queued ─────────────────────│                                      │
 │                                        │←─── tasks/status (running,pct=50) ──│
 │←── event: progress (pct=50) ──────────│                                      │
 │                                        │←─── tasks/status (completed) ───────│
 │←── event: completed ──────────────────│                                      │
```

---

## 错误响应

所有 HTTP 端点在出错时返回统一格式：

```json
{
  "error": "错误描述信息"
}
```

常见 HTTP 状态码：

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权（JWT 无效或过期） |
| 404 | 任务不存在 |
| 503 | 无可用 Agent |
