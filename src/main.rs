//! Agent Gateway - Main entry point
//!
//! 启动 HTTP/WebSocket 服务，监听两类连接：
//! - Orchestrator 通过 HTTPS + SSE 接入（公网/DMZ 暴露）
//! - Remote Bot (Nanobot) 通过 WebSocket 主动连接注册（内网）

mod config;
mod error;
mod registry;
mod router;
mod task;
mod handlers;
mod state;
mod auth;

use std::sync::Arc;
use axum::{
    middleware,
    routing::{get, post},
    Router,
};
use tower_http::trace::TraceLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use crate::auth::JwtConfig;
use crate::state::AppState;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // 初始化日志
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "agent_gateway=debug,tower_http=info".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    // 加载配置
    let cfg = config::Config::load()?;
    tracing::info!("Loaded config: orchestrator_addr={}, remote_bot_addr={}", cfg.bot_b_addr, cfg.bot_a_addr);

    // 初始化共享状态
    let state = Arc::new(AppState::new());

    // 启动两个 server: Orchestrator 接入端 + Remote Bot 注册端
    let orchestrator_app = build_orchestrator_router(state.clone(), &cfg);
    let remote_bot_app = build_remote_bot_router(state.clone());

    let orchestrator_listener = tokio::net::TcpListener::bind(&cfg.bot_b_addr).await?;
    tracing::info!("Orchestrator endpoint listening on {}", cfg.bot_b_addr);

    // 启动后台监控任务（心跳检测、超时清理）
    let bg_state = state.clone();
    tokio::spawn(async move {
        background_monitor(bg_state).await;
    });

    // Remote Bot 端：根据配置决定是否启用 mTLS
    if cfg.mtls_enabled {
        tracing::info!("Remote Bot mTLS enabled, cert={}, key={}, ca={}",
            cfg.mtls_cert_path, cfg.mtls_key_path, cfg.mtls_ca_path);
        let tls_config = crate::auth::build_mtls_config(
            &cfg.mtls_cert_path,
            &cfg.mtls_key_path,
            &cfg.mtls_ca_path,
        )?;
        let tls_acceptor = tokio_rustls::TlsAcceptor::from(Arc::new(tls_config));
        let remote_bot_listener = tokio::net::TcpListener::bind(&cfg.bot_a_addr).await?;
        tracing::info!("Remote Bot endpoint (mTLS) listening on {}", cfg.bot_a_addr);

        let (r1, r2) = tokio::join!(
            axum::serve(orchestrator_listener, orchestrator_app),
            serve_mtls(remote_bot_listener, tls_acceptor, remote_bot_app),
        );
        r1?;
        r2?;
    } else {
        let remote_bot_listener = tokio::net::TcpListener::bind(&cfg.bot_a_addr).await?;
        tracing::info!("Remote Bot endpoint listening on {}", cfg.bot_a_addr);

        // 并发运行两个 server
        let (r1, r2) = tokio::join!(
            axum::serve(orchestrator_listener, orchestrator_app),
            axum::serve(remote_bot_listener, remote_bot_app),
        );
        r1?;
        r2?;
    }

    Ok(())
}

/// Orchestrator 接入路由：标准 A2A 协议（JSON-RPC over HTTPS）+ SSE
fn build_orchestrator_router(state: Arc<AppState>, cfg: &config::Config) -> Router {
    let router = Router::new()
        .route("/.well-known/agent.json", get(handlers::common::agent_card))
        .route("/.well-known/skill.md", get(handlers::common::skill_md))
        .route("/jsonrpc", post(handlers::orchestrator::handle_jsonrpc))
        .route("/tasks", post(handlers::orchestrator::create_task))
        .route("/tasks/:task_id", get(handlers::orchestrator::get_task))
        .route("/tasks/:task_id/stream", get(handlers::orchestrator::sse_stream))
        .route("/agents", get(handlers::orchestrator::list_agents))
        .route("/health", get(handlers::common::health));

    // 根据配置决定是否启用 OAuth2 中间件
    let router = if cfg.oauth2_enabled {
        tracing::info!("OAuth2 enabled for Orchestrator endpoint");
        let jwt_cfg = JwtConfig {
            secret: cfg.oauth2_jwt_secret.clone(),
            issuer: cfg.oauth2_issuer.clone(),
        };
        router
            .layer(axum::Extension(jwt_cfg))
            .layer(middleware::from_fn(auth::oauth2_middleware))
    } else {
        router
    };

    router
        .layer(TraceLayer::new_for_http())
        .with_state(state)
}

/// Remote Bot 注册路由：WebSocket 长连接
fn build_remote_bot_router(state: Arc<AppState>) -> Router {
    Router::new()
        .route("/agent/register", get(handlers::remote_bot::ws_register))
        .route("/health", get(handlers::common::health))
        .layer(TraceLayer::new_for_http())
        .with_state(state)
}

/// 后台监控：心跳超时检测、任务超时清理
async fn background_monitor(state: Arc<AppState>) {
    let mut ticker = tokio::time::interval(std::time::Duration::from_secs(10));
    loop {
        ticker.tick().await;
        let removed = state.registry.evict_stale_agents(30); // 30s 心跳超时
        if removed > 0 {
            tracing::warn!("Evicted {} stale agents", removed);
        }
        let timed_out = state.task_store.timeout_running_tasks(60); // 60s 任务超时
        if timed_out > 0 {
            tracing::warn!("Marked {} tasks as failed (timeout)", timed_out);
        }
    }
}

/// mTLS 服务：接受 TLS 连接并转发给 axum
async fn serve_mtls(
    listener: tokio::net::TcpListener,
    tls_acceptor: tokio_rustls::TlsAcceptor,
    app: Router,
) -> std::result::Result<(), std::io::Error> {
    use hyper_util::rt::TokioIo;
    use tower::Service;

    loop {
        let (tcp_stream, addr) = match listener.accept().await {
            Ok(conn) => conn,
            Err(e) => {
                tracing::warn!("TCP accept error: {}", e);
                continue;
            }
        };

        let acceptor = tls_acceptor.clone();
        let mut app = app.clone();

        tokio::spawn(async move {
            let tls_stream = match acceptor.accept(tcp_stream).await {
                Ok(s) => s,
                Err(e) => {
                    tracing::warn!("TLS handshake failed from {}: {}", addr, e);
                    return;
                }
            };
            tracing::debug!("mTLS connection established from {}", addr);

            let io = TokioIo::new(tls_stream);
            let hyper_svc = hyper::service::service_fn(move |req| {
                let mut app = app.clone();
                async move { app.call(req).await }
            });

            if let Err(e) = hyper_util::server::conn::auto::Builder::new(
                hyper_util::rt::TokioExecutor::new(),
            )
            .serve_connection(io, hyper_svc)
            .await
            {
                tracing::warn!("Connection error from {}: {}", addr, e);
            }
        });
    }
}
