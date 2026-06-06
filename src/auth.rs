//! OAuth2 鉴权中间件 — BotB 接入时验证 Bearer Token (JWT)

use axum::{
    extract::Request,
    http::{header, StatusCode},
    middleware::Next,
    response::{IntoResponse, Response},
    Json,
};
use jsonwebtoken::{decode, DecodingKey, Validation, Algorithm};
use serde::{Deserialize, Serialize};
use serde_json::json;

/// JWT Claims 结构
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Claims {
    /// 主体（租户/用户标识）
    pub sub: String,
    /// 命名空间（多租户隔离）
    pub namespace: String,
    /// 过期时间
    pub exp: usize,
    /// 签发者
    #[serde(default)]
    pub iss: String,
}

/// OAuth2 鉴权中间件
/// 从 Authorization: Bearer <token> 中提取 JWT，验证签名和有效期，
/// 将 Claims 注入请求扩展供后续 handler 使用
pub async fn oauth2_middleware(
    mut req: Request,
    next: Next,
) -> Response {
    // 从共享状态中获取配置
    let extensions = req.extensions();
    let jwt_secret = extensions
        .get::<JwtConfig>()
        .map(|c| c.secret.clone())
        .unwrap_or_default();
    let expected_issuer = extensions
        .get::<JwtConfig>()
        .map(|c| c.issuer.clone())
        .unwrap_or_default();

    // 提取 Authorization header
    let auth_header = req
        .headers()
        .get(header::AUTHORIZATION)
        .and_then(|v| v.to_str().ok());

    let token = match auth_header {
        Some(h) if h.starts_with("Bearer ") => &h[7..],
        _ => {
            return (
                StatusCode::UNAUTHORIZED,
                Json(json!({"error": "Missing or invalid Authorization header"})),
            )
                .into_response();
        }
    };

    // 验证 JWT
    let mut validation = Validation::new(Algorithm::HS256);
    if !expected_issuer.is_empty() {
        validation.set_issuer(&[&expected_issuer]);
    }

    match decode::<Claims>(token, &DecodingKey::from_secret(jwt_secret.as_bytes()), &validation) {
        Ok(token_data) => {
            // 将 claims 注入请求扩展，后续 handler 可通过 Extension<Claims> 获取
            req.extensions_mut().insert(token_data.claims);
            next.run(req).await
        }
        Err(e) => {
            tracing::warn!("JWT validation failed: {}", e);
            (
                StatusCode::UNAUTHORIZED,
                Json(json!({"error": format!("Invalid token: {}", e)})),
            )
                .into_response()
        }
    }
}

/// JWT 配置 — 作为 Extension 注入到请求中
#[derive(Debug, Clone)]
pub struct JwtConfig {
    pub secret: String,
    pub issuer: String,
}

// ============ mTLS 配置构建 ============

use rustls::pki_types::{CertificateDer, PrivateKeyDer};
use std::fs::File;
use std::io::BufReader;

/// 构建 mTLS ServerConfig：
/// - 加载 Gateway 服务端证书和私钥
/// - 加载 CA 证书用于验证 BotA 客户端证书
pub fn build_mtls_config(
    cert_path: &str,
    key_path: &str,
    ca_path: &str,
) -> anyhow::Result<rustls::ServerConfig> {
    // 加载服务端证书链
    let cert_file = File::open(cert_path)?;
    let mut cert_reader = BufReader::new(cert_file);
    let certs: Vec<CertificateDer<'static>> = rustls_pemfile::certs(&mut cert_reader)
        .collect::<std::result::Result<Vec<_>, _>>()?;

    // 加载服务端私钥
    let key_file = File::open(key_path)?;
    let mut key_reader = BufReader::new(key_file);
    let key: PrivateKeyDer<'static> = rustls_pemfile::private_key(&mut key_reader)?
        .ok_or_else(|| anyhow::anyhow!("No private key found in {}", key_path))?;

    // 加载 CA 证书（用于验证客户端证书）
    let ca_file = File::open(ca_path)?;
    let mut ca_reader = BufReader::new(ca_file);
    let ca_certs: Vec<CertificateDer<'static>> = rustls_pemfile::certs(&mut ca_reader)
        .collect::<std::result::Result<Vec<_>, _>>()?;

    let mut root_store = rustls::RootCertStore::empty();
    for cert in ca_certs {
        root_store.add(cert)?;
    }

    // 构建要求客户端证书的验证器
    let client_verifier = rustls::server::WebPkiClientVerifier::builder(root_store.into())
        .build()?;

    let config = rustls::ServerConfig::builder()
        .with_client_cert_verifier(client_verifier)
        .with_single_cert(certs, key)?;

    Ok(config)
}
