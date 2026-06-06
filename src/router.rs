//! 路由策略 - 已合并到 Registry::pick_agent_for_skill
//! 此模块预留给未来扩展（如亲和性路由、优先级路由）

use crate::registry::{AgentHandle, Registry};
use std::sync::Arc;

#[allow(dead_code)]
pub enum RoutingStrategy {
    /// 最低负载（默认）
    LeastLoad,
    /// 上下文亲和性（同 contextId 路由到同一 Agent）
    Affinity { context_id: String },
    /// 轮询
    RoundRobin,
}

#[allow(dead_code)]
pub fn route(
    registry: &Registry,
    skill_id: &str,
    _strategy: RoutingStrategy,
) -> Option<Arc<AgentHandle>> {
    // 当前只实现 LeastLoad
    registry.pick_agent_for_skill(skill_id)
}
