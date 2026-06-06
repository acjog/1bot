//! Remote Bot WebSocket 注册与通信

use std::sync::{atomic::AtomicU32, Arc};

use axum::{
    extract::{
        ws::{Message, WebSocket, WebSocketUpgrade},
        State,
    },
    response::IntoResponse,
};
use chrono::Utc;
use futures_util::{SinkExt, StreamExt};
use tokio::sync::mpsc;

use crate::{
    registry::{AgentHandle, DownMessage, JsonRpcMessage, UpMessage},
    state::AppState,
    task::{TaskEvent, TaskStatus},
};

/// Remote Bot 通过 WebSocket 接入 Gateway
pub async fn ws_register(
    ws: WebSocketUpgrade,
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    ws.on_upgrade(move |socket| handle_socket(socket, state))
}

async fn handle_socket(socket: WebSocket, state: Arc<AppState>) {
    let (mut sender, mut receiver) = socket.split();
    // 给该 Remote Bot 发消息的通道
    let (tx, mut rx) = mpsc::unbounded_channel::<DownMessage>();

    // 1. 等待第一条消息：必须是 Register
    let agent_id = match wait_for_register(&mut receiver, &state, tx.clone()).await {
        Some(id) => id,
        None => {
            tracing::warn!("WebSocket closed before register");
            return;
        }
    };

    let agent_id_for_writer = agent_id.clone();
    let agent_id_for_reader = agent_id.clone();
    let agent_id_for_pinger = agent_id.clone();
    let state_for_reader = state.clone();
    let state_for_writer = state.clone();

    // 2. Pinger 任务：Gateway 主动定时下发 Ping 探活（每 10 秒）
    let ping_tx = tx.clone();
    let mut pinger = tokio::spawn(async move {
        let mut ticker = tokio::time::interval(std::time::Duration::from_secs(10));
        loop {
            ticker.tick().await;
            if ping_tx.send(DownMessage::Ping).is_err() {
                tracing::debug!("Ping channel closed for agent {}", agent_id_for_pinger);
                break;
            }
            tracing::trace!("Sent Ping to agent {}", agent_id_for_pinger);
        }
    });

    // 3. Writer 任务：从 mpsc 通道读取消息，序列化为 A2A JSON-RPC 发送给 Remote Bot
    let mut writer = tokio::spawn(async move {
        while let Some(msg) = rx.recv().await {
            let json = msg.to_jsonrpc();
            if sender.send(Message::Text(json)).await.is_err() {
                tracing::warn!("Failed to send to Remote Bot {}, closing", agent_id_for_writer);
                break;
            }
        }
        // 退出时清理
        state_for_writer.registry.unregister(&agent_id_for_writer);
    });

    // 4. Reader 任务：处理 Remote Bot 发上来的消息
    let mut reader = tokio::spawn(async move {
        while let Some(msg) = receiver.next().await {
            let msg = match msg {
                Ok(m) => m,
                Err(e) => {
                    tracing::warn!("WebSocket recv error: {}", e);
                    break;
                }
            };
            match msg {
                Message::Text(text) => {
                    if let Err(e) = handle_up_message(&text, &agent_id_for_reader, &state_for_reader).await {
                        tracing::warn!("Handle up message error: {}", e);
                    }
                }
                Message::Ping(p) => {
                    tracing::trace!("Received Ping from {}", agent_id_for_reader);
                    // axum ws 自动回 Pong；这里只更新心跳
                    if let Some(h) = state_for_reader.registry.get(&agent_id_for_reader) {
                        h.touch();
                    }
                    // 防止 unused warning
                    let _ = p;
                }
                Message::Close(_) => {
                    tracing::info!("Remote Bot {} closed connection", agent_id_for_reader);
                    break;
                }
                _ => {}
            }
        }
        state_for_reader.registry.unregister(&agent_id_for_reader);
    });

    // 5. 任一任务结束 → 终止整个连接
    tokio::select! {
        _ = (&mut writer) => {
            reader.abort();
            pinger.abort();
        },
        _ = (&mut reader) => {
            writer.abort();
            pinger.abort();
        },
        _ = (&mut pinger) => {
            writer.abort();
            reader.abort();
        },
    }
    tracing::info!("Agent {} disconnected", agent_id);
}

/// 等待 Remote Bot 发来第一条 Register 消息（A2A JSON-RPC: x-agent/register）
async fn wait_for_register(
    receiver: &mut futures_util::stream::SplitStream<WebSocket>,
    state: &Arc<AppState>,
    tx: mpsc::UnboundedSender<DownMessage>,
) -> Option<String> {
    while let Some(msg) = receiver.next().await {
        let msg = msg.ok()?;
        if let Message::Text(text) = msg {
            let rpc: JsonRpcMessage = serde_json::from_str(&text).ok()?;
            let up = UpMessage::from_jsonrpc(&rpc)?;
            if let UpMessage::Register { card, namespace } = up {
                let id = card.name.clone();
                let handle = Arc::new(AgentHandle {
                    id: id.clone(),
                    namespace: namespace.clone(),
                    card,
                    connected_at: Utc::now(),
                    last_heartbeat: parking_lot::RwLock::new(Utc::now()),
                    current_load: AtomicU32::new(0),
                    tx,
                });
                state.registry.register(handle);
                tracing::info!("Agent {} registered in namespace '{}'", id, namespace);
                return Some(id);
            }
            tracing::warn!("First message was not x-agent/register: {}", text);
            return None;
        }
    }
    None
}

/// 处理 Remote Bot 上行消息（A2A JSON-RPC 2.0）
async fn handle_up_message(
    text: &str,
    agent_id: &str,
    state: &Arc<AppState>,
) -> anyhow::Result<()> {
    let rpc: JsonRpcMessage = serde_json::from_str(text)?;
    let up = UpMessage::from_jsonrpc(&rpc)
        .ok_or_else(|| anyhow::anyhow!("Unknown JSON-RPC message: {:?}", rpc.method))?;

    if let Some(h) = state.registry.get(agent_id) {
        h.touch();
    }

    match up {
        UpMessage::Progress { task_id, status, pct, message } => {
            tracing::debug!("Task {} progress: {}% ({})", task_id, pct, status);
            // 状态首次进入 running 时更新
            state.task_store.transition(&task_id, TaskStatus::Running, None);
            state.task_store.broadcast(
                &task_id,
                TaskEvent::Progress { task_id: task_id.clone(), pct, message },
            );
        }
        UpMessage::Result { task_id, artifact } => {
            tracing::info!("Task {} completed", task_id);
            state.task_store.set_artifact(&task_id, artifact.clone());
            state.task_store.broadcast(
                &task_id,
                TaskEvent::Completed { task_id: task_id.clone(), artifact },
            );
            // 释放该 Agent 的 load
            if let Some(h) = state.registry.get(agent_id) {
                h.dec_load();
            }
        }
        UpMessage::Error { task_id, error } => {
            tracing::warn!("Task {} failed: {}", task_id, error);
            state.task_store.set_error(&task_id, error.clone());
            state.task_store.broadcast(
                &task_id,
                TaskEvent::Failed { task_id: task_id.clone(), error },
            );
            if let Some(h) = state.registry.get(agent_id) {
                h.dec_load();
            }
        }
        UpMessage::Pong => {
            tracing::trace!("Pong from {}", agent_id);
        }
        UpMessage::Register { .. } => {
            tracing::warn!("Duplicate Register from {}", agent_id);
        }
    }

    Ok(())
}
