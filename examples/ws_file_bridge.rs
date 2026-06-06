//! examples/ws_file_bridge.rs
//!
//! WebSocket <-> 本地文件读写桥接器
//!
//! 该示例将 Gateway 的 WebSocket 协议转换为本地文件读写：
//! - 收到任务时，将任务信息写入 `tasks/<task_id>/request.json`
//! - 轮询等待 `tasks/<task_id>/response.json` 文件出现
//! - 读取 response.json 内容，通过 WebSocket 返回给 Gateway
//!
//! 这样，不支持 WebSocket 的本地程序可以通过文件系统与 Gateway 交互。
//!
//! 运行方式:
//! cargo run --example ws_file_bridge -- ws://127.0.0.1:8080/agent/register file-bot-001
//!
//! 文件交互流程:
//! 1. 桥接器收到任务 -> 写入 ./tasks/<task_id>/request.json
//! 2. 外部程序读取 request.json，处理后写入 ./tasks/<task_id>/response.json
//! 3. 桥接器检测到 response.json -> 读取内容 -> 通过 WebSocket 发送结果给 Gateway
//!
//! response.json 格式示例:
//! {
//!   "status": "completed",
//!   "artifact": { "text": "处理结果内容" }
//! }
//!
//! 也支持流式进度上报，写入 ./tasks/<task_id>/progress.jsonl（每行一个JSON）:
//! {"pct": 30, "message": "正在分析..."}
//! {"pct": 60, "message": "正在生成..."}

use std::env;
use std::path::PathBuf;
use std::time::Duration;

use futures_util::{SinkExt, StreamExt};
use serde_json::{json, Value};
use tokio::fs;
use tokio::sync::mpsc;
use tokio::time::sleep;
use tokio_tungstenite::{connect_async, tungstenite::Message};

/// 任务工作目录
const TASKS_DIR: &str = "./tasks";

/// 轮询间隔（毫秒）
const POLL_INTERVAL_MS: u64 = 500;

/// 任务超时时间（秒）
const TASK_TIMEOUT_SECS: u64 = 300;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args: Vec<String> = env::args().collect();
    let url = args
        .get(1)
        .cloned()
        .unwrap_or_else(|| "ws://127.0.0.1:8080/agent/register".into());
    let name = args
        .get(2)
        .cloned()
        .unwrap_or_else(|| "file-bridge-001".into());

    // 确保任务目录存在
    fs::create_dir_all(TASKS_DIR).await?;

    println!("=== WebSocket <-> 文件桥接器 ===");
    println!("连接地址: {}", url);
    println!("Bot 名称: {}", name);
    println!("任务目录: {}", std::fs::canonicalize(TASKS_DIR)?.display());
    println!("轮询间隔: {}ms", POLL_INTERVAL_MS);
    println!("任务超时: {}s", TASK_TIMEOUT_SECS);
    println!();

    let (ws_stream, _) = connect_async(&url).await?;
    let (mut ws_write, mut ws_read) = ws_stream.split();

    // 用于从轮询任务向主循环发送 WebSocket 消息
    let (tx, mut rx) = mpsc::channel::<String>(64);

    // 1. 发送 Register（A2A JSON-RPC: x-agent/register）
    let register = json!({
        "jsonrpc": "2.0",
        "method": "x-agent/register",
        "params": {
            "card": {
                "name": name,
                "version": "1.0.0",
                "description": "WebSocket-File Bridge: 将WS协议转换为本地文件读写",
                "skills": [
                    {
                        "id": "file_process",
                        "name": "File-based Processing",
                        "input_modes": ["text"],
                        "output_modes": ["text"]
                    }
                ],
                "capabilities": {"streaming": true, "push_notifications": false},
                "max_concurrency": 5
            },
            "namespace": "default"
        },
        "id": "register-1"
    });
    ws_write.send(Message::Text(register.to_string())).await?;
    println!("[桥接器] 已注册到 Gateway");
    println!("[桥接器] 外部程序请监听 {} 目录", TASKS_DIR);
    println!("---");

    // 2. 启动发送协程：从 channel 接收消息并通过 WebSocket 发送
    let send_task = tokio::spawn(async move {
        while let Some(msg) = rx.recv().await {
            if let Err(e) = ws_write.send(Message::Text(msg)).await {
                eprintln!("[发送] WebSocket 发送失败: {}", e);
                break;
            }
        }
    });

    // 3. 主循环：处理 WebSocket 下行消息
    while let Some(msg) = ws_read.next().await {
        let msg = match msg {
            Ok(m) => m,
            Err(e) => {
                eprintln!("[桥接器] WebSocket 读取错误: {}", e);
                break;
            }
        };

        match msg {
            Message::Text(text) => {
                let v: Value = match serde_json::from_str(&text) {
                    Ok(v) => v,
                    Err(_) => continue,
                };

                // 响应 Gateway Ping 探活（A2A: x-agent/ping）
                if v["method"] == "x-agent/ping" {
                    let id = v["id"].clone();
                    let pong = json!({
                        "jsonrpc": "2.0",
                        "result": {"status": "ok"},
                        "id": id
                    }).to_string();
                    let _ = tx.send(pong).await;
                    continue;
                }

                // 处理任务分配（A2A: tasks/send）
                if v["method"] == "tasks/send" {
                    let params = &v["params"];
                    let task_id = params["id"].as_str().unwrap_or("unknown").to_string();
                    let instruction = params["message"]["parts"][0]["text"]
                        .as_str()
                        .unwrap_or("")
                        .to_string();

                    println!("[桥接器] 收到任务: {}", task_id);
                    println!("[桥接器] 指令: {}", instruction);

                    // 将任务写入文件
                    let task_dir = PathBuf::from(TASKS_DIR).join(&task_id);
                    if let Err(e) = fs::create_dir_all(&task_dir).await {
                        eprintln!("[桥接器] 创建任务目录失败: {}", e);
                        continue;
                    }

                    let request_file = task_dir.join("request.json");
                    let request_content = json!({
                        "task_id": task_id,
                        "instruction": instruction,
                        "received_at": chrono::Utc::now().to_rfc3339(),
                        "raw_message": v
                    });

                    if let Err(e) = fs::write(
                        &request_file,
                        serde_json::to_string_pretty(&request_content).unwrap(),
                    )
                    .await
                    {
                        eprintln!("[桥接器] 写入请求文件失败: {}", e);
                        continue;
                    }

                    println!("[桥接器] 任务已写入: {}", request_file.display());

                    // 启动轮询协程，等待结果文件
                    let tx_clone = tx.clone();
                    tokio::spawn(poll_task_result(task_id, task_dir, tx_clone));
                }
            }
            Message::Ping(p) => {
                // WebSocket 协议层 Ping，直接回复
                let pong_msg = format!("__ws_pong__{}", String::from_utf8_lossy(&p));
                let _ = tx.send(pong_msg).await;
            }
            Message::Close(_) => {
                println!("[桥接器] Gateway 关闭连接");
                break;
            }
            _ => {}
        }
    }

    // 等待发送任务结束
    drop(tx);
    let _ = send_task.await;
    println!("[桥接器] 已退出");

    Ok(())
}

/// 轮询任务目录，等待 response.json 出现
/// 同时监听 progress.jsonl 文件的变化来上报进度
async fn poll_task_result(task_id: String, task_dir: PathBuf, tx: mpsc::Sender<String>) {
    let response_file = task_dir.join("response.json");
    let progress_file = task_dir.join("progress.jsonl");

    let start = tokio::time::Instant::now();
    let timeout = Duration::from_secs(TASK_TIMEOUT_SECS);
    let mut last_progress_lines = 0usize;

    println!("[轮询:{}] 等待结果文件: {}", task_id, response_file.display());

    loop {
        // 超时检查
        if start.elapsed() > timeout {
            eprintln!("[轮询:{}] 超时（{}秒），任务失败", task_id, TASK_TIMEOUT_SECS);
            let timeout_msg = json!({
                "jsonrpc": "2.0",
                "method": "tasks/status",
                "params": {
                    "id": task_id,
                    "status": "failed",
                    "error": format!("任务处理超时（{}秒未收到结果文件）", TASK_TIMEOUT_SECS)
                }
            });
            let _ = tx.send(timeout_msg.to_string()).await;
            break;
        }

        // 检查进度文件（追加模式，每行一个JSON）
        if progress_file.exists() {
            if let Ok(content) = fs::read_to_string(&progress_file).await {
                let lines: Vec<&str> = content.lines().collect();
                // 只处理新增的行
                if lines.len() > last_progress_lines {
                    for line in &lines[last_progress_lines..] {
                        if let Ok(progress) = serde_json::from_str::<Value>(line) {
                            let pct = progress["pct"].as_u64().unwrap_or(0);
                            let message =
                                progress["message"].as_str().unwrap_or("处理中...");
                            println!("[轮询:{}] 进度: {}% - {}", task_id, pct, message);

                            let progress_msg = json!({
                                "jsonrpc": "2.0",
                                "method": "tasks/status",
                                "params": {
                                    "id": task_id,
                                    "status": "running",
                                    "progress": pct,
                                    "message": message
                                }
                            });
                            let _ = tx.send(progress_msg.to_string()).await;
                        }
                    }
                    last_progress_lines = lines.len();
                }
            }
        }

        // 检查结果文件
        if response_file.exists() {
            // 短暂等待确保文件写入完成
            sleep(Duration::from_millis(100)).await;

            match fs::read_to_string(&response_file).await {
                Ok(content) => {
                    match serde_json::from_str::<Value>(&content) {
                        Ok(response) => {
                            let status =
                                response["status"].as_str().unwrap_or("completed");
                            let artifact = response
                                .get("artifact")
                                .cloned()
                                .unwrap_or(json!({"text": "（无内容）"}));

                            println!("[轮询:{}] 收到结果，状态: {}", task_id, status);

                            let result_msg = json!({
                                "jsonrpc": "2.0",
                                "method": "tasks/status",
                                "params": {
                                    "id": task_id,
                                    "status": status,
                                    "artifact": artifact
                                }
                            });
                            let _ = tx.send(result_msg.to_string()).await;

                            // 写入完成标记
                            let _ = fs::write(task_dir.join(".done"), "done").await;
                            println!("[轮询:{}] 任务完成 ✓", task_id);
                        }
                        Err(e) => {
                            eprintln!(
                                "[轮询:{}] response.json 解析失败: {}，等待重写...",
                                task_id, e
                            );
                            // 文件可能还在写入中，继续等待
                            sleep(Duration::from_millis(POLL_INTERVAL_MS)).await;
                            continue;
                        }
                    }
                }
                Err(e) => {
                    eprintln!("[轮询:{}] 读取 response.json 失败: {}", task_id, e);
                }
            }
            break;
        }

        sleep(Duration::from_millis(POLL_INTERVAL_MS)).await;
    }
}
