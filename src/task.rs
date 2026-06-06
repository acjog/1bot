//! Task 任务管理 - 维护任务状态、进度流广播

use chrono::{DateTime, Utc};
use dashmap::DashMap;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::broadcast;

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum TaskStatus {
    Created,
    Queued,
    Running,
    Completed,
    Failed,
    Canceled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskRecord {
    pub task_id: String,
    pub skill: String,
    pub instruction: String,
    pub constraints: serde_json::Value,
    pub status: TaskStatus,
    pub assigned_agent: Option<String>,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub artifact: Option<serde_json::Value>,
    pub error: Option<String>,
}

/// 任务进度事件 — 用于 SSE 推送给 BotB
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "event", rename_all = "snake_case")]
pub enum TaskEvent {
    Queued { task_id: String, agent_id: String },
    Running { task_id: String },
    Progress {
        task_id: String,
        pct: u32,
        message: Option<String>,
    },
    Completed {
        task_id: String,
        artifact: serde_json::Value,
    },
    Failed { task_id: String, error: String },
    Canceled { task_id: String },
}

/// 任务存储 — 包含状态和事件广播通道
pub struct TaskEntry {
    pub record: parking_lot::RwLock<TaskRecord>,
    /// SSE 事件广播：BotB 订阅 Receiver，Gateway 通过 Sender 推事件
    pub event_tx: broadcast::Sender<TaskEvent>,
}

pub struct TaskStore {
    tasks: DashMap<String, Arc<TaskEntry>>,
}

impl TaskStore {
    pub fn new() -> Self {
        Self {
            tasks: DashMap::new(),
        }
    }

    pub fn create(&self, skill: String, instruction: String, constraints: serde_json::Value) -> Arc<TaskEntry> {
        let task_id = format!("t-{}", uuid::Uuid::new_v4());
        let now = Utc::now();
        let (tx, _rx) = broadcast::channel(64);
        let entry = Arc::new(TaskEntry {
            record: parking_lot::RwLock::new(TaskRecord {
                task_id: task_id.clone(),
                skill,
                instruction,
                constraints,
                status: TaskStatus::Created,
                assigned_agent: None,
                created_at: now,
                updated_at: now,
                artifact: None,
                error: None,
            }),
            event_tx: tx,
        });
        self.tasks.insert(task_id, entry.clone());
        entry
    }

    pub fn get(&self, task_id: &str) -> Option<Arc<TaskEntry>> {
        self.tasks.get(task_id).map(|r| r.clone())
    }

    /// 更新状态并广播事件
    pub fn transition(&self, task_id: &str, status: TaskStatus, agent_id: Option<String>) {
        if let Some(entry) = self.tasks.get(task_id) {
            let mut rec = entry.record.write();
            rec.status = status;
            rec.updated_at = Utc::now();
            if let Some(aid) = agent_id {
                rec.assigned_agent = Some(aid);
            }
        }
    }

    pub fn set_artifact(&self, task_id: &str, artifact: serde_json::Value) {
        if let Some(entry) = self.tasks.get(task_id) {
            let mut rec = entry.record.write();
            rec.artifact = Some(artifact);
            rec.status = TaskStatus::Completed;
            rec.updated_at = Utc::now();
        }
    }

    pub fn set_error(&self, task_id: &str, error: String) {
        if let Some(entry) = self.tasks.get(task_id) {
            let mut rec = entry.record.write();
            rec.error = Some(error);
            rec.status = TaskStatus::Failed;
            rec.updated_at = Utc::now();
        }
    }

    pub fn broadcast(&self, task_id: &str, event: TaskEvent) {
        if let Some(entry) = self.tasks.get(task_id) {
            let _ = entry.event_tx.send(event); // 没人订阅时忽略错误
        }
    }

    /// 检查超时任务，标记为失败
    pub fn timeout_running_tasks(&self, timeout_secs: i64) -> usize {
        let now = Utc::now();
        let mut count = 0;
        for entry in self.tasks.iter() {
            let rec = entry.record.read();
            if rec.status == TaskStatus::Running
                && (now - rec.updated_at).num_seconds() > timeout_secs
            {
                let task_id = rec.task_id.clone();
                drop(rec);
                self.set_error(&task_id, "task timeout".into());
                self.broadcast(
                    &task_id,
                    TaskEvent::Failed {
                        task_id: task_id.clone(),
                        error: "task timeout".into(),
                    },
                );
                count += 1;
            }
        }
        count
    }
}
