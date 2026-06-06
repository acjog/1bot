# 多 Agent 协作架构设计文档

## 基于中间人网关的 BotB (高阶模型) 指挥 BotA (Nanobot) 方案

---

## 1. 架构概述

本方案采用 **Agent Gateway（中间人网关）** 作为通信枢纽，实现高阶模型 BotB 对低阶 Nanobot (BotA) 的实时指挥调度。BotA 不需要外网地址，只需能出站连接网关即可。

### 1.1 核心角色

| 角色 | 运行模型 | 职责 |
|------|---------|------|
| **BotB (Orchestrator)** | 高阶大模型 | 理解用户意图、任务规划、任务分解、质量审核、结果整合 |
| **BotA (Worker/Nanobot)** | 低阶小模型 | 执行具体子任务（代码生成、文本摘要、数据检索、分类标注等） |
| **Agent Gateway** | 无模型（纯路由） | 服务注册、能力发现、消息路由、负载均衡、鉴权 |

### 1.2 设计原则

- **BotA 零入站** — Nanobot 只需出站连接网关，无需公网 IP 或开放端口
- **协议标准化** — 遵循 A2A (Agent-to-Agent) 协议进行 Agent 间通信
- **实时性** — 通过 SSE (Server-Sent Events) 实现任务进度实时推送
- **弹性伸缩** — BotA 可动态扩缩容，网关自动发现和负载均衡
- **故障隔离** — 单个 BotA 故障不影响整体系统

---

## 2. 架构拓扑

```
┌─────────────────────────────────────────────────────────────────┐
│                         用户请求入口                              │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                   BotB (高阶模型 Orchestrator)                    │
│                                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐       │
│  │ 意图理解  │  │ 任务分解  │  │ 质量审核  │  │ 结果整合  │       │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘       │
└─────────────────────────┬───────────────────────────────────────┘
                          │ A2A 协议 (HTTPS + JSON-RPC)
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Agent Gateway (中间人网关)                     │
│                                                                  │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐  │
│  │ 服务注册   │ │ 能力发现   │ │ 消息路由   │ │ 负载均衡   │  │
│  │ Registry   │ │ Discovery  │ │ Router     │ │ LB         │  │
│  └────────────┘ └────────────┘ └────────────┘ └────────────┘  │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐  │
│  │ 鉴权认证   │ │ 限流熔断   │ │ 日志监控   │ │ SSE 管理   │  │
│  │ Auth       │ │ RateLimit  │ │ Observ.    │ │ Stream Mgr │  │
│  └────────────┘ └────────────┘ └────────────┘ └────────────┘  │
└──────────┬─────────────┬─────────────┬──────────────────────────┘
           │             │             │
     长连接注册      长连接注册      长连接注册
     (WebSocket)    (WebSocket)    (WebSocket)
           │             │             │
           ▼             ▼             ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│  BotA 实例1   │ │  BotA 实例2   │ │  BotA 实例3   │
│  (Nanobot)   │ │  (Nanobot)   │ │  (Nanobot)   │
│              │ │              │ │              │
│  技能:       │ │  技能:       │ │  技能:       │
│  - 文本摘要  │ │  - 代码生成  │ │  - 数据检索  │
│  - 翻译     │ │  - 代码审查  │ │  - 分类标注  │
│              │ │              │ │              │
│  ┌────────┐ │ │  ┌────────┐ │ │  ┌────────┐ │
│  │MCP Tools│ │ │  │MCP Tools│ │ │  │MCP Tools│ │
│  └────────┘ │ │  └────────┘ │ │  └────────┘ │
└──────────────┘ └──────────────┘ └──────────────┘
   (内网/本地)      (内网/本地)      (内网/本地)
```

---

## 3. 通信协议设计

### 3.1 协议栈

```
┌─────────────────────────────────────┐
│  A2A (Agent-to-Agent Protocol)      │  ← Agent 间通信
├─────────────────────────────────────┤
│  JSON-RPC 2.0                       │  ← 消息格式
├─────────────────────────────────────┤
│  HTTPS / WebSocket / SSE            │  ← 传输层
├─────────────────────────────────────┤
│  TLS 1.3                            │  ← 加密
└─────────────────────────────────────┘
```

### 3.2 通信链路

| 链路 | 协议 | 方向 | 说明 |
|------|------|------|------|
| BotB → Gateway | HTTPS + JSON-RPC | BotB 发起 | 下发任务、查询状态 |
| Gateway → BotB | SSE | Gateway 推送 | 实时进度流、结果回传 |
| BotA → Gateway | WebSocket | BotA 主动连接 | 注册上线 + 接收任务 |
| BotA → Gateway | WebSocket | BotA 推送 | 回传执行结果和进度 |

### 3.3 BotA 注册流程

```
BotA 启动
    │
    ▼
连接 Gateway WebSocket 端点
ws://gateway.internal:8080/agent/register
    │
    ▼
发送 Agent Card (能力声明)
{
  "name": "nanobot-text-001",
  "version": "1.0.0",
  "skills": [
    {"id": "summarize", "name": "文本摘要", "inputModes": ["text"]},
    {"id": "translate", "name": "翻译", "inputModes": ["text"]}
  ],
  "capabilities": {"streaming": true},
  "maxConcurrency": 5
}
    │
    ▼
Gateway 确认注册，维持心跳
    │
    ▼
等待接收任务...
```

### 3.4 任务下发与执行流程

```
1. BotB → Gateway:  POST /tasks/create
   {
     "method": "tasks/send",
     "params": {
       "taskId": "t-20260604-001",
       "skill": "summarize",
       "message": {
         "role": "user",
         "parts": [{"type": "text", "text": "请摘要以下文档..."}]
       },
       "constraints": {
         "maxTokens": 200,
         "language": "zh-CN"
       }
     }
   }

2. Gateway 路由: 根据 skill 匹配可用 BotA，按负载均衡选择

3. Gateway → BotA (via WebSocket):
   {
     "type": "task_assign",
     "taskId": "t-20260604-001",
     "instruction": "请摘要以下文档...",
     "constraints": {"maxTokens": 200}
   }

4. BotA 执行中，实时回传进度:
   → {"type": "progress", "taskId": "t-001", "status": "running", "pct": 30}
   → {"type": "progress", "taskId": "t-001", "status": "running", "pct": 70}

5. Gateway 转发给 BotB (via SSE):
   event: task_progress
   data: {"taskId": "t-001", "status": "running", "pct": 30}
   
   event: task_progress
   data: {"taskId": "t-001", "status": "running", "pct": 70}

6. BotA 完成:
   → {"type": "result", "taskId": "t-001", "artifact": {"text": "摘要结果..."}}

7. Gateway → BotB (via SSE):
   event: task_completed
   data: {"taskId": "t-001", "artifact": {"text": "摘要结果..."}}

8. BotB 审核结果:
   - 合格 → 整合输出
   - 不合格 → 重新下发（附修改指令）
```

---

## 4. Agent Gateway 核心模块

### 4.1 服务注册与发现 (Registry)

```yaml
# 网关维护的 Agent 注册表
agents:
  nanobot-text-001:
    status: online
    connected_at: "2026-06-04T16:00:00Z"
    skills: [summarize, translate]
    load: 2/5  # 当前任务数/最大并发
    last_heartbeat: "2026-06-04T16:59:00Z"
    
  nanobot-code-001:
    status: online
    connected_at: "2026-06-04T15:30:00Z"
    skills: [code_generate, code_review]
    load: 1/3
    last_heartbeat: "2026-06-04T16:59:02Z"
```

### 4.2 路由策略

| 策略 | 规则 | 适用场景 |
|------|------|---------|
| **技能匹配** | 根据 task.skill 匹配具备该能力的 BotA | 默认策略 |
| **最低负载** | 同技能 BotA 中选当前任务最少的 | 常规负载均衡 |
| **亲和性** | 相同 context 的任务路由到同一 BotA | 多轮对话/上下文相关 |
| **优先级** | 高优任务抢占低优任务的资源 | 紧急任务 |

### 4.3 容错机制

```
任务超时 (默认 60s)
    │
    ├─ BotA 无响应 → 标记该 BotA 为 unhealthy → 重新路由到其他 BotA
    │
    ├─ BotA 返回错误 → 通知 BotB，由 BotB 决定重试或降级
    │
    └─ Gateway 自身故障 → 多实例部署 + 客户端自动重连
```

### 4.4 限流与熔断

```yaml
rate_limiting:
  per_bot_b: 100 tasks/min     # 单个 BotB 的发送限制
  per_bot_a: 10 tasks/concurrent # 单个 BotA 的并发限制
  global: 1000 tasks/min        # 全局限制

circuit_breaker:
  failure_threshold: 5          # 连续失败 5 次触发熔断
  recovery_timeout: 30s         # 熔断后 30s 尝试恢复
  half_open_requests: 2         # 半开状态试探请求数
```

---

## 5. 安全设计

### 5.1 认证体系

```
┌───────────┐        ┌───────────┐        ┌───────────┐
│   BotB    │        │  Gateway  │        │   BotA    │
└─────┬─────┘        └─────┬─────┘        └─────┬─────┘
      │                     │                     │
      │  OAuth2 Bearer Token│                     │
      │────────────────────>│                     │
      │                     │                     │
      │                     │  mTLS 双向证书认证    │
      │                     │<────────────────────│
      │                     │                     │
```

| 链路 | 认证方式 | 说明 |
|------|---------|------|
| BotB → Gateway | OAuth 2.0 Client Credentials | BotB 持有 client_id + secret |
| BotA → Gateway | mTLS 双向证书 | BotA 启动时加载客户端证书，Gateway 验证 |
| Gateway 内部 | JWT 签名 | 任务消息携带签名，防篡改 |

### 5.2 数据安全

- 传输层全部 TLS 1.3 加密
- 敏感字段（用户数据、API Key）在网关层脱敏后再下发 BotA
- 任务结果在网关缓存不超过 5 分钟，过期自动清除
- 审计日志记录所有任务的下发、执行、结果

---

## 6. Task 生命周期

```
                    ┌─────────┐
                    │ created │  BotB 创建任务
                    └────┬────┘
                         │
                         ▼
                    ┌─────────┐
                    │ queued  │  Gateway 路由队列中
                    └────┬────┘
                         │
                         ▼
                    ┌─────────┐
          ┌────────│ running │  BotA 执行中 (实时推送进度)
          │        └────┬────┘
          │             │
          │    ┌────────┼────────┐
          │    │        │        │
          │    ▼        ▼        ▼
          │ ┌───────┐ ┌──────┐ ┌────────┐
          │ │failed │ │done  │ │canceled│
          │ └───┬───┘ └──┬───┘ └────────┘
          │     │         │
          │     ▼         ▼
          │ ┌────────────────┐
          └>│ BotB 审核判定   │
            └───────┬────────┘
                    │
              ┌─────┼─────┐
              │           │
              ▼           ▼
         ┌────────┐  ┌────────┐
         │ retry  │  │ accept │
         │ (重做)  │  │ (采纳) │
         └────────┘  └────────┘
```

---

## 7. 部署方案

### 7.1 最小可用部署

```yaml
# docker-compose.yml 示例

services:
  gateway:
    image: agent-gateway:latest
    ports:
      - "443:443"       # BotB 接入 (HTTPS)
      - "8080:8080"     # BotA 注册 (WebSocket, 内网)
    environment:
      - AUTH_PROVIDER=oauth2
      - LOG_LEVEL=info
    deploy:
      replicas: 2       # 高可用至少 2 实例

  bot-a-text:
    image: nanobot-text:latest
    environment:
      - GATEWAY_URL=ws://gateway:8080/agent/register
      - AGENT_NAME=nanobot-text-001
      - MAX_CONCURRENCY=5
    deploy:
      replicas: 3       # 按需扩缩容

  bot-a-code:
    image: nanobot-code:latest
    environment:
      - GATEWAY_URL=ws://gateway:8080/agent/register
      - AGENT_NAME=nanobot-code-001
      - MAX_CONCURRENCY=3
    deploy:
      replicas: 2

  redis:
    image: redis:7-alpine
    # 用于 Gateway 集群间的状态同步和任务队列
```

### 7.2 网络拓扑

```
                    ┌─── 公网 ───┐
                    │            │
              ┌─────┴─────┐     │
              │ BotB (云端) │     │  可选：BotB 也可在内网
              └─────┬─────┘     │
                    │           │
              ┌─────┴─────┐    │
              │ Gateway    │◄───┘  Gateway 对外暴露 443
              │ (DMZ/边界) │
              └─────┬─────┘
                    │
            ┌───── 内网 ─────┐
            │       │        │
        ┌───┴──┐ ┌──┴───┐ ┌─┴────┐
        │BotA-1│ │BotA-2│ │BotA-3│   BotA 全部在内网
        └──────┘ └──────┘ └──────┘   只需出站到 Gateway
```

---

## 8. 监控与可观测性

### 8.1 关键指标

| 指标 | 含义 | 告警阈值 |
|------|------|---------|
| `task_latency_p99` | 任务端到端延迟 P99 | > 10s |
| `bot_a_online_count` | 在线 BotA 数量 | < 预期的 50% |
| `task_failure_rate` | 任务失败率 | > 5% |
| `gateway_queue_depth` | 网关待分配任务队列深度 | > 100 |
| `bot_a_heartbeat_miss` | BotA 心跳丢失次数 | 连续 > 3 次 |

### 8.2 分布式追踪

每个任务携带 `trace_id`，贯穿 BotB → Gateway → BotA 全链路：

```json
{
  "trace_id": "tr-20260604-abc123",
  "spans": [
    {"service": "BotB", "action": "task_create", "duration_ms": 5},
    {"service": "Gateway", "action": "route", "duration_ms": 2},
    {"service": "BotA-text-001", "action": "execute", "duration_ms": 3200},
    {"service": "Gateway", "action": "deliver", "duration_ms": 1},
    {"service": "BotB", "action": "review", "duration_ms": 800}
  ]
}
```

---

## 9. 扩展能力

### 9.1 BotA 热插拔

- 新 BotA 实例启动 → 自动连接 Gateway → 注册技能 → 即刻可用
- BotA 下线 → 心跳超时 → Gateway 自动摘除 → 进行中任务自动重路由

### 9.2 技能版本管理

```json
{
  "skills": [
    {
      "id": "summarize",
      "version": "2.0",
      "description": "支持多语言摘要",
      "deprecated_versions": ["1.0"]
    }
  ]
}
```

### 9.3 多租户隔离

Gateway 支持 namespace 隔离，不同团队的 BotB/BotA 互不可见：

```
Gateway
  ├── namespace: team-alpha
  │   ├── BotB-alpha → BotA-alpha-1, BotA-alpha-2
  │   
  └── namespace: team-beta
      ├── BotB-beta → BotA-beta-1, BotA-beta-2
```

---

## 10. 技术选型建议

| 组件 | 推荐方案 | 备选 |
|------|---------|------|
| Gateway 框架 | Go (高性能网络) | Rust / Node.js |
| 消息格式 | JSON-RPC 2.0 (A2A 标准) | gRPC (高吞吐场景) |
| BotA 连接 | WebSocket | gRPC 双向流 |
| BotB 接入 | HTTPS + SSE | HTTPS + WebSocket |
| 状态存储 | Redis Cluster | etcd (强一致性需求) |
| 服务发现 | 网关内置 Registry | Consul / etcd |
| 监控 | Prometheus + Grafana | DataDog |
| 追踪 | OpenTelemetry + Jaeger | Zipkin |
| 证书管理 | cert-manager (K8s) | Vault PKI |

---

## 11. 通信协议选型决策

### 11.1 协议对比总览

| 协议 | 通信方向 | 连接模式 | 延迟 | 复杂度 | 一句话描述 |
|------|:-------:|---------|:----:|:-----:|-----------|
| **HTTP** | 请求→响应 | 一来一回断开 | 高 | 最低 | 写信：寄一封等一封回 |
| **SSE** | 服务端→客户端 (单向推) | 一次请求，持续接收 | 低 | 低 | 听广播：打开就能收，但不能回话 |
| **WebSocket** | 全双工 (双向随时) | 持久连接，双方同时收发 | 极低 | 中 | 打电话：拨通后随时说话 |
| **gRPC 双向流** | 全双工 | 基于 HTTP/2 多路复用 | 极低 | 高 | 对讲机阵列：多通道并行 |

### 11.2 本架构的协议选择逻辑

#### BotA ↔ Gateway：选择 WebSocket

**需求分析：**
- Gateway 需要向 BotA **下发任务**（Gateway → BotA）
- BotA 需要向 Gateway **回传进度和结果**（BotA → Gateway）
- BotA 在内网，只能**出站连接**，不能被外部主动访问
- 需要**长连接保持**，用于注册、心跳、实时通信

**为什么选 WebSocket：**

| 候选方案 | 能否满足 | 不选的原因 |
|---------|:-------:|-----------|
| HTTP 轮询 | 勉强能 | 延迟高，频繁建连开销大，不实时 |
| SSE | 不能 | 单向推送，BotA 无法通过 SSE 向 Gateway 发送数据 |
| **WebSocket** | **完美** | **双向通信 + BotA 主动出站建连 + 持久连接 + 低延迟** |
| gRPC 双向流 | 能 | 过重，需要 proto 定义，对轻量 Nanobot 不友好 |

**决策：WebSocket 是唯一满足「BotA 只能出站 + 双向通信」约束的轻量方案。**

```
BotA 启动 → 主动连接 Gateway 的 ws://gateway:8080/agent/register
            ↕ 此连接同时用于：
            ← Gateway 下发任务
            → BotA 回传进度/结果
            ↕ 定期心跳保活
```

#### Gateway → BotB：选择 SSE

**需求分析：**
- BotB 需要**实时接收**任务进度和结果（Gateway → BotB 推送）
- BotB 下发新任务时用**普通 HTTP POST** 即可（BotB → Gateway）
- 通信模式本质是「BotB 发命令，然后听结果」— 下发和接收可以分开

**为什么选 SSE 而不是 WebSocket：**

| 维度 | SSE | WebSocket |
|------|-----|-----------|
| 方向 | 服务端→客户端 (够用) | 双向 (过剩) |
| 实现复杂度 | 极低，标准 HTTP | 需要协议升级、帧处理 |
| 代理/CDN 兼容 | 完美兼容（就是 HTTP 长连接） | 部分代理需特殊配置 |
| 断线重连 | 浏览器原生自动重连 | 需自行实现 |
| 发送命令 | 另开 POST 请求 | 同一连接发 |

**决策：BotB 的推送需求是单向的（只接收），SSE 更简单可靠，发命令用 POST 即可。**

```
BotB 下发任务:  POST https://gateway/tasks/create  (普通 HTTP)
BotB 听取进度:  GET  https://gateway/tasks/stream   (SSE 长连接)
```

#### BotB → Gateway 任务下发：选择 HTTPS + JSON-RPC

**需求分析：**
- 符合 A2A 协议标准（A2A 定义的传输层就是 HTTPS + JSON-RPC 2.0）
- 任务下发是低频操作（相比进度推送），不需要持久连接
- 需要可靠的请求-响应语义（确认任务已被接收）

**决策：遵循 A2A 标准，简单可靠。**

### 11.3 协议分层总结

```
┌─────────────────────────────────────────────────────────────┐
│                    通信协议选型全景                             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   BotB                  Gateway                  BotA       │
│   ┌────┐               ┌──────┐               ┌────┐      │
│   │    │── POST(命令) ─→│      │── WS(下发) ──→│    │      │
│   │    │← SSE(进度/结果)│      │← WS(进度/结果)│    │      │
│   └────┘               └──────┘               └────┘      │
│                                                             │
│   选型理由:             选型理由:             选型理由:        │
│   • 发命令=低频POST     • 双向路由转发        • 只能出站连接    │
│   • 听进度=单向SSE      • 协议桥接            • 双向通信需求    │
│   • A2A 标准兼容        • 心跳+注册管理       • 轻量持久连接    │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│  选型口诀:                                                    │
│  • 「只需服务端推」 → SSE (Gateway→BotB)                       │
│  • 「双方都要主动发」 → WebSocket (BotA↔Gateway)               │
│  • 「一次性命令」 → HTTP POST (BotB→Gateway)                   │
└─────────────────────────────────────────────────────────────┘
```

### 11.4 为什么不全部用 WebSocket？

虽然 WebSocket 能覆盖所有场景，但**不选它做 BotB 接入**的理由是：

| 考量 | 说明 |
|------|------|
| **简单性** | SSE 就是普通 HTTP 响应，不需要协议升级、帧解析、Ping/Pong 心跳 |
| **兼容性** | SSE 能穿透所有 HTTP 代理、CDN、WAF；WebSocket 部分环境需特殊配置 |
| **标准对齐** | A2A 协议推荐的实时推送方式就是 SSE |
| **职责清晰** | 发命令(POST) 和听结果(SSE) 分离，代码更清晰，调试更方便 |
| **断线恢复** | SSE 有内置的 `Last-Event-ID` 机制，断线后自动从断点续传 |

**原则：能用简单方案解决的，不上复杂方案。WebSocket 只在必须双向的地方用。**

### 11.5 Nanobot 原生 A2A 支持

Nanobot (HKUDS 开源) 已原生支持 A2A 协议，安装方式：

```bash
pip install "nanobot-ai[a2a]"
```

配置启用：
```json
{
  "channels": {
    "a2a": {
      "enabled": true,
      "port": 8080,
      "publicHost": "https://my-agent.example.com"
    }
  }
}
```

在网关模式下的适配方式：
- Nanobot 的 A2A Server 可以不对外暴露（仅内网）
- 通过 WebSocket 主动连接 Gateway 注册
- Gateway 做 A2A ↔ WebSocket 的协议桥接
- 或者在同内网部署时，Gateway 直接通过 HTTP 调用 Nanobot 的 A2A 端点

---

## 12. 总结

本架构通过 Agent Gateway 解耦了 BotB 与 BotA 的直接依赖：

- **BotA (Nanobot) 不需要外网地址**，只需出站连接网关
- **实时通信** 通过 WebSocket (BotA↔Gateway) + SSE (Gateway→BotB) 实现
- **协议选型清晰** — WebSocket 用于必须双向的链路，SSE 用于单向推送，HTTP POST 用于命令下发
- **遵循 A2A 协议标准**，具备互操作性和可扩展性
- **Nanobot 原生支持 A2A**，`pip install "nanobot-ai[a2a]"` 即可接入
- **弹性伸缩**，BotA 按需增减，网关自动感知
- **生产级可靠性**，熔断、重试、监控、审计一应俱全

---

*文档版本: v1.1 | 更新时间: 2026-06-04 | 协议标准: A2A v1.0 + MCP*
*v1.1 变更: 新增第 11 章「通信协议选型决策」，补充 Nanobot A2A 原生支持说明*
