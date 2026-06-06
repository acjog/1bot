//! 全局配置

use serde::Deserialize;

#[derive(Debug, Deserialize, Clone)]
pub struct Config {
    /// Orchestrator 接入地址（公网/DMZ）
    pub bot_b_addr: String,
    /// Remote Bot 注册地址（内网）
    pub bot_a_addr: String,
    /// 任务超时（秒）
    pub task_timeout_secs: u64,
    /// Agent 心跳超时（秒）
    pub heartbeat_timeout_secs: u64,
    /// 单个 Orchestrator 限流（每分钟）
    pub bot_b_rate_limit: u32,

    // --- mTLS 配置（Remote Bot 接入） ---
    /// 是否启用 mTLS（Remote Bot 端）
    #[serde(default)]
    pub mtls_enabled: bool,
    /// Gateway 服务端证书路径
    #[serde(default)]
    pub mtls_cert_path: String,
    /// Gateway 服务端私钥路径
    #[serde(default)]
    pub mtls_key_path: String,
    /// CA 证书路径（用于验证 Remote Bot 客户端证书）
    #[serde(default)]
    pub mtls_ca_path: String,

    // --- OAuth2 配置（Orchestrator 接入） ---
    /// 是否启用 OAuth2 鉴权
    #[serde(default)]
    pub oauth2_enabled: bool,
    /// JWT 签名密钥（HS256）
    #[serde(default)]
    pub oauth2_jwt_secret: String,
    /// JWT 签发者
    #[serde(default)]
    pub oauth2_issuer: String,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            bot_b_addr: "0.0.0.0:8443".into(),
            bot_a_addr: "0.0.0.0:8080".into(),
            task_timeout_secs: 60,
            heartbeat_timeout_secs: 30,
            bot_b_rate_limit: 100,
            mtls_enabled: false,
            mtls_cert_path: String::new(),
            mtls_key_path: String::new(),
            mtls_ca_path: String::new(),
            oauth2_enabled: false,
            oauth2_jwt_secret: String::new(),
            oauth2_issuer: String::new(),
        }
    }
}

impl Config {
    pub fn load() -> anyhow::Result<Self> {
        let cfg = config::Config::builder()
            .add_source(config::File::with_name("config").required(false))
            .add_source(config::Environment::with_prefix("GATEWAY").separator("_"))
            .set_default("bot_b_addr", "0.0.0.0:8443")?
            .set_default("bot_a_addr", "0.0.0.0:8080")?
            .set_default("task_timeout_secs", 60)?
            .set_default("heartbeat_timeout_secs", 30)?
            .set_default("bot_b_rate_limit", 100)?
            .set_default("mtls_enabled", false)?
            .set_default("mtls_cert_path", "")?
            .set_default("mtls_key_path", "")?
            .set_default("mtls_ca_path", "")?
            .set_default("oauth2_enabled", false)?
            .set_default("oauth2_jwt_secret", "")?
            .set_default("oauth2_issuer", "")?
            .build()?;
        Ok(cfg.try_deserialize()?)
    }
}
