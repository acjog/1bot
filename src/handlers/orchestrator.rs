//! Orchestrator 接入 - HTTPS + JSON-RPC + SSE

use std::{convert::Infallible, sync::Arc, time::Duration};

use axum::{
    extract::{Path, State},
    response::sse::{Event, KeepAlive, Sse},
    Json,
};
use futures::stream::Stream;
use serde::{Deserialize, Serialize};
use serde_json::json;
use tokio_stream::StreamExt;

use crate::{
    auth::Claims,
    error::{GatewayError, Result},
    registry::DownMessage,
    state::AppState,
    task::{TaskEvent, TaskStatus},
};

/// JSON-RPC 2.0 请求
#[derive(Debug, Deserialize)]
pub struct JsonRpcRequest {
    pub jsonrpc: String,
    pub method: String,
    pub params: serde_json::Value,
    pub id: serde_json::Value,
}

#[derive(Debug, Serialize)]
pub struct JsonRpcResponse {
    pub jsonrpc: &'static str,
    pub result: serde_json::Value,
    pub id: serde_json::Value,
}

/// A2A 标准 JSON-RPC 入口（统一入口，方便 A2A 客户端接入）
pub async fn handle_jsonrpc(
    State(state): State<Arc<AppState>>,
    claims: Option<axum::Extension<Claims>>,
    Json(req): Json<JsonRpcRequest>,
) -> Result<Json<JsonRpcResponse>> {
    let namespace = claims
        .map(|c| c.0.namespace.clone())
        .unwrap_or_else(|| "default".to_string());
    if req.jsonrpc != "2.0" {
        return Err(GatewayError::InvalidRequest("jsonrpc must be 2.0".into()));
    }
    let result = match req.method.as_str() {
        "tasks/send" => {
            let p: CreateTaskRequest = serde_json::from_value(req.params)?;
            let task = create_task_inner(&state, p, &namespace).await?;
            serde_json::to_value(task)?
        }
        "tasks/get" => {
            let task_id = req
                .params
                .get("taskId")
                .and_then(|v| v.as_str())
                .ok_or_else(|| GatewayError::InvalidRequest("taskId required".into()))?;
            let entry = state
                .task_store
                .get(task_id)
                .ok_or_else(|| GatewayError::TaskNotFound(task_id.into()))?;
            let record = entry.record.read().clone();
            serde_json::to_value(&record)?
        }
        "agents/list" => {
            let agents: Vec<_> = state
                .registry
                .list()
                .into_iter()
                .map(|h| {
                    json!({
                        "id": h.id,
                        "card": h.card,
                        "load": h.load(),
                        "max_concurrency": h.card.max_concurrency,
                        "connected_at": h.connected_at,
                    })
                })
                .collect();
            json!(agents)
        }
        m => {
            return Err(GatewayError::InvalidRequest(format!("unknown method: {}", m)));
        }
    };

    Ok(Json(JsonRpcResponse {
        jsonrpc: "2.0",
        result,
        id: req.id,
    }))
}

/// 简化的 RESTful 任务创建（非标准 A2A，方便测试）
#[derive(Debug, Deserialize)]
pub struct CreateTaskRequest {
    pub skill: String,
    pub instruction: String,
    #[serde(default)]
    pub constraints: serde_json::Value,
}

#[derive(Debug, Serialize)]
pub struct CreateTaskResponse {
    pub task_id: String,
    pub status: TaskStatus,
    pub assigned_agent: String,
}

pub async fn create_task(
    State(state): State<Arc<AppState>>,
    claims: Option<axum::Extension<Claims>>,
    Json(req): Json<CreateTaskRequest>,
) -> Result<Json<CreateTaskResponse>> {
    let namespace = claims
        .map(|c| c.0.namespace.clone())
        .unwrap_or_else(|| "default".to_string());
    let resp = create_task_inner(&state, req, &namespace).await?;
    Ok(Json(resp))
}

async fn create_task_inner(
    state: &Arc<AppState>,
    req: CreateTaskRequest,
    namespace: &str,
) -> Result<CreateTaskResponse> {
    // 1. 在指定 namespace 中选择 Remote Bot
    let agent = state
        .registry
        .pick_agent_for_skill_in_ns(&req.skill, namespace)
        .ok_or_else(|| GatewayError::NoAvailableAgent(req.skill.clone()))?;

    // 2. 创建任务记录
    let entry = state.task_store.create(
        req.skill.clone(),
        req.instruction.clone(),
        req.constraints.clone(),
    );
    let task_id = entry.record.read().task_id.clone();

    // 3. 状态变为 queued
    state.task_store.transition(&task_id, TaskStatus::Queued, Some(agent.id.clone()));
    state.task_store.broadcast(
        &task_id,
        TaskEvent::Queued {
            task_id: task_id.clone(),
            agent_id: agent.id.clone(),
        },
    );

    // 4. 通过 WebSocket 通道发给 Remote Bot（A2A: tasks/send）
    agent.inc_load();
    if let Err(e) = agent.tx.send(DownMessage::TaskSend {
        task_id: task_id.clone(),
        skill: req.skill,
        instruction: req.instruction,
        constraints: req.constraints,
    }) {
        agent.dec_load();
        state.task_store.set_error(&task_id, format!("dispatch failed: {}", e));
        return Err(GatewayError::AgentUnavailable(agent.id.clone()));
    }

    Ok(CreateTaskResponse {
        task_id,
        status: TaskStatus::Queued,
        assigned_agent: agent.id.clone(),
    })
}

pub async fn get_task(
    State(state): State<Arc<AppState>>,
    Path(task_id): Path<String>,
) -> Result<Json<serde_json::Value>> {
    let entry = state
        .task_store
        .get(&task_id)
        .ok_or_else(|| GatewayError::TaskNotFound(task_id))?;
    let record = entry.record.read().clone();
    Ok(Json(serde_json::to_value(&record)?))
}

/// SSE 端点：Orchestrator 订阅任务进度流
pub async fn sse_stream(
    State(state): State<Arc<AppState>>,
    Path(task_id): Path<String>,
) -> Result<Sse<impl Stream<Item = std::result::Result<Event, Infallible>>>> {
    let entry = state
        .task_store
        .get(&task_id)
        .ok_or_else(|| GatewayError::TaskNotFound(task_id.clone()))?;

    // 先订阅，再读取当前状态，避免在读取和订阅之间丢失事件
    let rx = entry.event_tx.subscribe();

    // 构造当前状态的初始事件，确保即使订阅晚了也能看到已有状态
    let record = entry.record.read().clone();
    let initial_event = match record.status {
        TaskStatus::Created => None,
        TaskStatus::Queued => Some(TaskEvent::Queued {
            task_id: record.task_id.clone(),
            agent_id: record.assigned_agent.clone().unwrap_or_default(),
        }),
        TaskStatus::Running => Some(TaskEvent::Running {
            task_id: record.task_id.clone(),
        }),
        TaskStatus::Completed => Some(TaskEvent::Completed {
            task_id: record.task_id.clone(),
            artifact: record.artifact.clone().unwrap_or(json!(null)),
        }),
        TaskStatus::Failed => Some(TaskEvent::Failed {
            task_id: record.task_id.clone(),
            error: record.error.clone().unwrap_or_default(),
        }),
        TaskStatus::Canceled => Some(TaskEvent::Canceled {
            task_id: record.task_id.clone(),
        }),
    };

    // 初始状态流
    let initial_stream = futures::stream::iter(initial_event.into_iter().map(|evt| {
        let json = serde_json::to_string(&evt).unwrap_or_default();
        let event_name = match &evt {
            TaskEvent::Queued { .. } => "queued",
            TaskEvent::Running { .. } => "running",
            TaskEvent::Progress { .. } => "progress",
            TaskEvent::Completed { .. } => "completed",
            TaskEvent::Failed { .. } => "failed",
            TaskEvent::Canceled { .. } => "canceled",
        };
        Ok(Event::default().event(event_name).data(json))
    }));

    // 后续实时事件流
    let live_stream = tokio_stream::wrappers::BroadcastStream::new(rx).filter_map(|res| {
        let evt = res.ok()?;
        let json = serde_json::to_string(&evt).ok()?;
        let event_name = match &evt {
            TaskEvent::Queued { .. } => "queued",
            TaskEvent::Running { .. } => "running",
            TaskEvent::Progress { .. } => "progress",
            TaskEvent::Completed { .. } => "completed",
            TaskEvent::Failed { .. } => "failed",
            TaskEvent::Canceled { .. } => "canceled",
        };
        Some(Ok(Event::default().event(event_name).data(json)))
    });

    // 合并：先推初始状态，再接实时流
    let stream = initial_stream.chain(live_stream);

    Ok(Sse::new(stream).keep_alive(KeepAlive::new().interval(Duration::from_secs(15))))
}

/// 列出所有在线 Agent
pub async fn list_agents(State(state): State<Arc<AppState>>) -> Json<serde_json::Value> {
    let agents: Vec<_> = state
        .registry
        .list()
        .into_iter()
        .map(|h| {
            json!({
                "id": h.id,
                "skills": h.card.skills.iter().map(|s| &s.id).collect::<Vec<_>>(),
                "load": format!("{}/{}", h.load(), h.card.max_concurrency),
                "connected_at": h.connected_at,
                "last_heartbeat": *h.last_heartbeat.read(),
            })
        })
        .collect();
    Json(json!({ "agents": agents }))
}
