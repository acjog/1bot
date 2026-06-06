//! 错误类型定义

use axum::{
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde_json::json;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum GatewayError {
    #[error("Agent not found: {0}")]
    AgentNotFound(String),

    #[error("Task not found: {0}")]
    TaskNotFound(String),

    #[error("No available agent for skill: {0}")]
    NoAvailableAgent(String),

    #[error("Agent unavailable: {0}")]
    AgentUnavailable(String),

    #[error("Invalid request: {0}")]
    InvalidRequest(String),

    #[error("Rate limit exceeded")]
    RateLimitExceeded,

    #[error("Internal error: {0}")]
    Internal(#[from] anyhow::Error),

    #[error("Serialization error: {0}")]
    Serde(#[from] serde_json::Error),
}

impl IntoResponse for GatewayError {
    fn into_response(self) -> Response {
        let (status, code) = match &self {
            Self::AgentNotFound(_) => (StatusCode::NOT_FOUND, "agent_not_found"),
            Self::TaskNotFound(_) => (StatusCode::NOT_FOUND, "task_not_found"),
            Self::NoAvailableAgent(_) => (StatusCode::SERVICE_UNAVAILABLE, "no_available_agent"),
            Self::AgentUnavailable(_) => (StatusCode::SERVICE_UNAVAILABLE, "agent_unavailable"),
            Self::InvalidRequest(_) => (StatusCode::BAD_REQUEST, "invalid_request"),
            Self::RateLimitExceeded => (StatusCode::TOO_MANY_REQUESTS, "rate_limit"),
            Self::Internal(_) | Self::Serde(_) => (StatusCode::INTERNAL_SERVER_ERROR, "internal"),
        };
        let body = Json(json!({
            "error": {
                "code": code,
                "message": self.to_string(),
            }
        }));
        (status, body).into_response()
    }
}

pub type Result<T> = std::result::Result<T, GatewayError>;
