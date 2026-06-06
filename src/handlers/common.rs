//! 通用 handler

use std::sync::Arc;

use axum::{extract::State, http::StatusCode, response::IntoResponse, Json};
use serde_json::json;

use crate::state::AppState;

pub async fn health() -> impl IntoResponse {
    (StatusCode::OK, Json(json!({"status": "ok"})))
}

/// 返回 skill.md 协议文档
pub async fn skill_md() -> impl IntoResponse {
    let content = include_str!("../../skill.md");
    (
        StatusCode::OK,
        [("content-type", "text/markdown; charset=utf-8")],
        content,
    )
}

/// A2A 标准 Agent Card 端点
/// skills 动态聚合所有在线 Remote Bot 的技能，对外暴露 Gateway 代理的完整能力
pub async fn agent_card(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let agents = state.registry.list();
    let online_agents = agents.len();
    let total_concurrency: u32 = agents.iter().map(|a| a.card.max_concurrency).sum();

    // 聚合所有在线 Remote Bot 的技能（去重），作为 Gateway 对外暴露的 skills
    let mut skills = Vec::new();
    let mut seen = std::collections::HashSet::new();
    for agent in &agents {
        for skill in &agent.card.skills {
            if seen.insert(skill.id.clone()) {
                skills.push(json!({
                    "id": skill.id,
                    "name": skill.name,
                    "description": skill.description,
                    "inputModes": skill.input_modes,
                    "outputModes": skill.output_modes
                }));
            }
        }
    }

    let card = json!({
        "name": "Agent Gateway",
        "description": "多 Agent 通信网关，作为 Orchestrator 与 Remote Bot（Worker）之间的桥接层。提供任务路由、进度推送、服务发现、多租户隔离、反向代理Remote 等能力。",
        "url": "http://127.0.0.1:8443",
        "version": env!("CARGO_PKG_VERSION"),
        "protocol": "a2a",
        "protocolVersion": "0.2.0",
        "capabilities": {
            "streaming": true,
            "pushNotifications": false,
            "stateTransitionHistory": false
        },
        "skills": skills,
        "defaultInputModes": ["text"],
        "defaultOutputModes": ["text"],
        "provider": {
            "organization": "Agent Gateway",
            "url": "https://github.com/agent-gateway"
        },
        "authentication": {
            "schemes": ["bearer"],
            "credentials": null
        },
        "supportsAuthenticatedExtendedCard": false,
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
                    "jsonrpc": "POST /jsonrpc",
                    "health": "GET /health"
                }
            },
            "remote_bot": {
                "base_url": "ws://127.0.0.1:8080",
                "auth": "mTLS 客户端证书",
                "routes": {
                    "register": "GET /agent/register (WebSocket Upgrade)",
                    "health": "GET /health"
                }
            }
        },
        "_meta": {
            "online_agents": online_agents,
            "total_concurrency": total_concurrency
        }
    });

    (StatusCode::OK, Json(card))
}
