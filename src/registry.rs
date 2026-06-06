//! Agent 注册表 - 维护所有在线 Remote Bot 的状态

use chrono::{DateTime, Utc};
use dashmap::DashMap;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::mpsc;

/// Agent Card - A2A 协议定义的能力声明
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AgentCard {
    pub name: String,
    pub version: String,
    pub description: Option<String>,
    pub skills: Vec<Skill>,
    pub capabilities: Capabilities,
    #[serde(default = "default_concurrency")]
    pub max_concurrency: u32,
}

fn default_concurrency() -> u32 {
    5
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Skill {
    pub id: String,
    pub name: String,
    #[serde(default)]
    pub description: String,
    #[serde(default)]
    pub input_modes: Vec<String>,
    #[serde(default)]
    pub output_modes: Vec<String>,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct Capabilities {
    #[serde(default)]
    pub streaming: bool,
    #[serde(default)]
    pub push_notifications: bool,
}

/// A2A JSON-RPC 2.0 请求/通知（Gateway ↔ Remote Bot 双向使用）
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct JsonRpcMessage {
    pub jsonrpc: String,
    pub method: Option<String>,
    pub params: Option<serde_json::Value>,
    pub id: Option<serde_json::Value>,
    /// JSON-RPC response 字段
    pub result: Option<serde_json::Value>,
    pub error: Option<serde_json::Value>,
}

impl JsonRpcMessage {
    /// 创建一个 JSON-RPC 请求
    pub fn request(method: &str, params: serde_json::Value, id: serde_json::Value) -> Self {
        Self {
            jsonrpc: "2.0".into(),
            method: Some(method.into()),
            params: Some(params),
            id: Some(id),
            result: None,
            error: None,
        }
    }

    /// 创建一个 JSON-RPC 通知（无 id）
    pub fn notification(method: &str, params: serde_json::Value) -> Self {
        Self {
            jsonrpc: "2.0".into(),
            method: Some(method.into()),
            params: Some(params),
            id: None,
            result: None,
            error: None,
        }
    }

    /// 创建一个 JSON-RPC 成功响应
    pub fn response(id: serde_json::Value, result: serde_json::Value) -> Self {
        Self {
            jsonrpc: "2.0".into(),
            method: None,
            params: None,
            id: Some(id),
            result: Some(result),
            error: None,
        }
    }

    /// 判断是否为请求（有 method 且有 id）
    pub fn is_request(&self) -> bool {
        self.method.is_some() && self.id.is_some()
    }

    /// 判断是否为通知（有 method 但无 id）
    pub fn is_notification(&self) -> bool {
        self.method.is_some() && self.id.is_none()
    }

    /// 判断是否为响应（有 result 或 error）
    pub fn is_response(&self) -> bool {
        self.result.is_some() || self.error.is_some()
    }
}

/// Gateway 发给 Remote Bot 的下行消息（通过 WebSocket，A2A JSON-RPC 2.0）
///
/// 方法列表：
/// - `tasks/send` — 任务分派
/// - `tasks/cancel` — 取消任务
/// - `x-agent/ping` — 心跳探活
#[derive(Debug, Clone)]
pub enum DownMessage {
    TaskSend {
        task_id: String,
        skill: String,
        instruction: String,
        constraints: serde_json::Value,
    },
    TaskCancel {
        task_id: String,
    },
    Ping,
}

impl DownMessage {
    /// 序列化为 A2A JSON-RPC 2.0 格式
    pub fn to_jsonrpc(&self) -> String {
        let msg = match self {
            DownMessage::TaskSend { task_id, skill, instruction, constraints } => {
                JsonRpcMessage::request(
                    "tasks/send",
                    serde_json::json!({
                        "id": task_id,
                        "message": {
                            "role": "user",
                            "parts": [{"type": "text", "text": instruction}]
                        },
                        "metadata": {
                            "skill": skill,
                            "constraints": constraints
                        }
                    }),
                    serde_json::json!(task_id),
                )
            }
            DownMessage::TaskCancel { task_id } => {
                JsonRpcMessage::request(
                    "tasks/cancel",
                    serde_json::json!({"id": task_id}),
                    serde_json::json!(task_id),
                )
            }
            DownMessage::Ping => {
                JsonRpcMessage::request(
                    "x-agent/ping",
                    serde_json::json!({}),
                    serde_json::json!("ping"),
                )
            }
        };
        serde_json::to_string(&msg).unwrap_or_default()
    }
}

/// Remote Bot 发给 Gateway 的上行消息（通过 WebSocket，A2A JSON-RPC 2.0）
///
/// 方法列表：
/// - `x-agent/register` — 注册 Agent Card
/// - `tasks/status` — 进度/状态通知
/// - `x-agent/pong` — 心跳响应
#[derive(Debug, Clone)]
pub enum UpMessage {
    Register {
        card: AgentCard,
        namespace: String,
    },
    Progress {
        task_id: String,
        status: String,
        pct: u32,
        message: Option<String>,
    },
    Result {
        task_id: String,
        artifact: serde_json::Value,
    },
    Error {
        task_id: String,
        error: String,
    },
    Pong,
}

impl UpMessage {
    /// 从 JSON-RPC 消息解析为 UpMessage
    pub fn from_jsonrpc(msg: &JsonRpcMessage) -> Option<Self> {
        // 处理响应消息（如 pong 响应）
        if msg.is_response() {
            // x-agent/ping 的响应视为 Pong
            if let Some(id) = &msg.id {
                if id.as_str() == Some("ping") {
                    return Some(UpMessage::Pong);
                }
            }
            return None;
        }

        let method = msg.method.as_deref()?;
        let params = msg.params.as_ref();

        match method {
            "x-agent/register" => {
                let p = params?;
                let card: AgentCard = serde_json::from_value(p["card"].clone()).ok()?;
                let namespace = p["namespace"].as_str()
                    .unwrap_or("default")
                    .to_string();
                Some(UpMessage::Register { card, namespace })
            }
            "x-agent/pong" => Some(UpMessage::Pong),
            "tasks/status" => {
                let p = params?;
                let task_id = p["id"].as_str()?.to_string();
                let status = p["status"].as_str().unwrap_or("running").to_string();

                match status.as_str() {
                    "completed" => {
                        let artifact = p.get("artifact")
                            .or_else(|| p.get("artifacts"))
                            .cloned()
                            .unwrap_or(serde_json::json!({}));
                        Some(UpMessage::Result { task_id, artifact })
                    }
                    "failed" => {
                        let error = p["error"].as_str()
                            .unwrap_or("unknown error")
                            .to_string();
                        Some(UpMessage::Error { task_id, error })
                    }
                    _ => {
                        // running / working 等中间状态视为进度
                        let pct = p["progress"].as_u64()
                            .or_else(|| p["pct"].as_u64())
                            .unwrap_or(0) as u32;
                        let message = p["message"].as_str().map(|s| s.to_string());
                        Some(UpMessage::Progress { task_id, status, pct, message })
                    }
                }
            }
            _ => None,
        }
    }
}

/// 在线 Agent 句柄 — 持有指向该 Remote Bot WebSocket 写端的发送通道
pub struct AgentHandle {
    pub id: String,
    /// 命名空间（多租户隔离）
    pub namespace: String,
    pub card: AgentCard,
    pub connected_at: DateTime<Utc>,
    pub last_heartbeat: parking_lot::RwLock<DateTime<Utc>>,
    pub current_load: std::sync::atomic::AtomicU32,
    /// 给该 Remote Bot 发消息的通道（mpsc 单向，写端在 ws_register handler）
    pub tx: mpsc::UnboundedSender<DownMessage>,
}

impl AgentHandle {
    pub fn touch(&self) {
        *self.last_heartbeat.write() = Utc::now();
    }

    pub fn load(&self) -> u32 {
        self.current_load.load(std::sync::atomic::Ordering::Relaxed)
    }

    pub fn inc_load(&self) {
        self.current_load.fetch_add(1, std::sync::atomic::Ordering::Relaxed);
    }

    pub fn dec_load(&self) {
        self.current_load.fetch_sub(1, std::sync::atomic::Ordering::Relaxed);
    }

    pub fn is_full(&self) -> bool {
        self.load() >= self.card.max_concurrency
    }

    pub fn supports_skill(&self, skill_id: &str) -> bool {
        self.card.skills.iter().any(|s| s.id == skill_id)
    }
}

/// Agent 注册表 - 线程安全
pub struct Registry {
    agents: DashMap<String, Arc<AgentHandle>>,
}

impl Registry {
    pub fn new() -> Self {
        Self {
            agents: DashMap::new(),
        }
    }

    pub fn register(&self, handle: Arc<AgentHandle>) {
        tracing::info!(
            "Agent registered: id={}, skills={:?}",
            handle.id,
            handle.card.skills.iter().map(|s| &s.id).collect::<Vec<_>>()
        );
        self.agents.insert(handle.id.clone(), handle);
    }

    pub fn unregister(&self, id: &str) {
        if self.agents.remove(id).is_some() {
            tracing::info!("Agent unregistered: id={}", id);
        }
    }

    pub fn get(&self, id: &str) -> Option<Arc<AgentHandle>> {
        self.agents.get(id).map(|r| r.clone())
    }

    pub fn list(&self) -> Vec<Arc<AgentHandle>> {
        self.agents.iter().map(|r| r.clone()).collect()
    }

    /// 选择一个支持指定 skill 的 Remote Bot — 最低负载策略（带 namespace 隔离）
    pub fn pick_agent_for_skill(&self, skill_id: &str) -> Option<Arc<AgentHandle>> {
        self.pick_agent_for_skill_in_ns(skill_id, "default")
    }

    /// 在指定 namespace 中选择支持指定 skill 的 Remote Bot
    pub fn pick_agent_for_skill_in_ns(&self, skill_id: &str, namespace: &str) -> Option<Arc<AgentHandle>> {
        self.agents
            .iter()
            .filter(|h| h.namespace == namespace && h.supports_skill(skill_id) && !h.is_full())
            .min_by_key(|h| h.load())
            .map(|r| r.clone())
    }

    /// 心跳超时检测，移除心跳超时的 Agent
    pub fn evict_stale_agents(&self, timeout_secs: i64) -> usize {
        let now = Utc::now();
        let stale: Vec<String> = self
            .agents
            .iter()
            .filter(|h| {
                let last = *h.last_heartbeat.read();
                (now - last).num_seconds() > timeout_secs
            })
            .map(|h| h.id.clone())
            .collect();
        let count = stale.len();
        for id in stale {
            self.unregister(&id);
        }
        count
    }
}
