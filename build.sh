#!/bin/bash
# Agent Gateway 编译脚本

set -e

PROJECT_NAME="agent-gateway"
TARGET_DIR="target"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# 默认参数
MODE="release"
CLEAN=false
RUN=false

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --debug)    MODE="debug"; shift ;;
        --release)  MODE="release"; shift ;;
        --clean)    CLEAN=true; shift ;;
        --run)      RUN=true; shift ;;
        --help|-h)
            echo "用法: ./build.sh [选项]"
            echo ""
            echo "选项:"
            echo "  --debug     Debug 模式编译（默认 release）"
            echo "  --release   Release 模式编译（默认）"
            echo "  --clean     编译前清理"
            echo "  --run       编译后运行"
            echo "  -h, --help  显示帮助"
            exit 0
            ;;
        *) error "未知参数: $1" ;;
    esac
done

# 检查 Rust 工具链
if ! command -v cargo &> /dev/null; then
    error "未找到 cargo，请先安装 Rust: https://rustup.rs"
fi

info "Rust 版本: $(rustc --version)"
info "Cargo 版本: $(cargo --version)"
info "编译模式: ${MODE}"

# 清理
if [ "$CLEAN" = true ]; then
    info "清理构建产物..."
    cargo clean
fi

# 编译
if [ "$MODE" = "release" ]; then
    info "开始 Release 编译（启用 LTO，可能较慢）..."
    cargo build --release
    BINARY="${TARGET_DIR}/release/${PROJECT_NAME}"
else
    info "开始 Debug 编译..."
    cargo build
    BINARY="${TARGET_DIR}/debug/${PROJECT_NAME}"
fi

# 编译结果
if [ -f "$BINARY" ]; then
    SIZE=$(du -h "$BINARY" | cut -f1)
    info "编译成功！"
    info "产物路径: ${BINARY}"
    info "文件大小: ${SIZE}"
else
    error "编译产物未找到: ${BINARY}"
fi

# 运行
if [ "$RUN" = true ]; then
    info "启动 ${PROJECT_NAME}..."
    exec "$BINARY"
fi
