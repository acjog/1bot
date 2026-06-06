//! examples/mock_nanobot.rs
//!
//! 模拟一个 Remote Bot (Nanobot) 通过 WebSocket 连接 Gateway
//! 使用 A2A JSON-RPC 2.0 协议通信
//!
//! 运行方式:
//! cargo run --example mock_nanobot -- ws://127.0.0.1:8080/agent/register nanobot-text-001

use std::env;
use std::time::Duration;

use futures_util::{SinkExt, StreamExt};
use serde_json::json;
use tokio::time::sleep;
use tokio_tungstenite::{connect_async, tungstenite::Message};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args: Vec<String> = env::args().collect();
    let url = args
        .get(1)
        .cloned()
        .unwrap_or_else(|| "ws://127.0.0.1:8080/agent/register".into());
    let name = args.get(2).cloned().unwrap_or_else(|| "nanobot-mock-001".into());

    println!("Connecting to {} as {}", url, name);

    let (ws_stream, _) = connect_async(&url).await?;
    let (mut write, mut read) = ws_stream.split();

    // 1. 发送 Register（A2A JSON-RPC: x-agent/register）
    let register = json!({
        "jsonrpc": "2.0",
        "method": "x-agent/register",
        "params": {
            "card": {
                "name": name,
                "version": "1.0.0",
                "description": "Mock Nanobot for testing",
                "skills": [
                    {"id": "summarize", "name": "Text Summarization", "input_modes": ["text"], "output_modes": ["text"]},
                    {"id": "translate", "name": "Translation", "input_modes": ["text"], "output_modes": ["text"]}
                ],
                "capabilities": {"streaming": true, "push_notifications": false},
                "max_concurrency": 3
            },
            "namespace": "default"
        },
        "id": "register-1"
    });
    write.send(Message::Text(register.to_string())).await?;
    println!("Registered. Waiting for tasks...");

    // 2. 处理下行消息（A2A JSON-RPC 2.0）
    while let Some(msg) = read.next().await {
        let msg = msg?;
        match msg {
            Message::Text(text) => {
                println!("Received: {}", text);
                let v: serde_json::Value = serde_json::from_str(&text)?;

                let method = v["method"].as_str().unwrap_or("");

                match method {
                    // 响应 Gateway 心跳探活（x-agent/ping）
                    "x-agent/ping" => {
                        let id = v["id"].clone();
                        let pong = json!({
                            "jsonrpc": "2.0",
                            "result": {"status": "ok"},
                            "id": id
                        });
                        write.send(Message::Text(pong.to_string())).await?;
                        println!("Replied pong to gateway");
                    }

                    // 处理任务分派（tasks/send）
                    "tasks/send" => {
                        let params = &v["params"];
                        let task_id = params["id"].as_str().unwrap_or("unknown").to_string();
                        let instruction = params["message"]["parts"][0]["text"]
                            .as_str()
                            .unwrap_or("")
                            .to_string();

                        println!("Task assigned: {} - {}", task_id, instruction);

                        // 模拟异步执行：每10%报一次进度，每步延时6秒
                        for pct in (10..=100).step_by(10) {
                            let progress = json!({
                                "jsonrpc": "2.0",
                                "method": "tasks/status",
                                "params": {
                                    "id": task_id,
                                    "status": "running",
                                    "progress": pct,
                                    "message": format!("processing {}%...", pct)
                                }
                            });
                            write.send(Message::Text(progress.to_string())).await?;
                            println!("Task {} progress: {}%", task_id, pct);
                            sleep(Duration::from_secs(6)).await;
                        }

                        // 发送完成结果
                        let result = json!({
                            "jsonrpc": "2.0",
                            "method": "tasks/status",
                            "params": {
                                "id": task_id,
                                "status": "completed",
                                "artifact": {
                                    "text": format!("[mock summary] {}", instruction.chars().take(50).collect::<String>())
                                }
                            }
                        });
                        write.send(Message::Text(result.to_string())).await?;
                        println!("Completed task {}", task_id);
                    }

                    // 取消任务（tasks/cancel）
                    "tasks/cancel" => {
                        let task_id = v["params"]["id"].as_str().unwrap_or("unknown");
                        println!("Task cancelled: {}", task_id);
                    }

                    _ => {
                        println!("Unknown method: {}", method);
                    }
                }
            }
            Message::Ping(p) => {
                write.send(Message::Pong(p)).await?;
            }
            Message::Close(_) => break,
            _ => {}
        }
    }

    Ok(())
}
