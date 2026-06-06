//! 共享应用状态

use crate::registry::Registry;
use crate::task::TaskStore;

pub struct AppState {
    pub registry: Registry,
    pub task_store: TaskStore,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            registry: Registry::new(),
            task_store: TaskStore::new(),
        }
    }
}
