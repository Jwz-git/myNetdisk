#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "${SCRIPT_DIR}/.." && pwd)"

show_help() {
    cat <<'EOF'
Personal Disk 本地启动脚本

用法: ./scripts/start.sh [dev|test|prod]

环境:
  dev   使用 configs/development.yml（默认）
  test  使用 configs/test.yml
  prod  使用 configs/production.yml，必须设置 ADMIN_PASSWORD

环境变量（例如 DB_PATH、UPLOAD_DIRECTORY、SERVER_PORT）会覆盖 YAML 配置。
按 Ctrl+C 停止服务。
EOF
}

environment="${1:-dev}"

case "${environment}" in
    -h|--help)
        show_help
        exit 0
        ;;
    dev|development)
        export APP_ENV=development
        ;;
    test)
        export APP_ENV=test
        ;;
    prod|production)
        export APP_ENV=production
        if [[ -z "${ADMIN_PASSWORD:-}" ]]; then
            echo "错误: 生产环境必须设置 ADMIN_PASSWORD" >&2
            exit 1
        fi
        ;;
    *)
        echo "错误: 未知环境 '${environment}'" >&2
        show_help >&2
        exit 1
        ;;
esac

if ! command -v go >/dev/null 2>&1; then
    echo "错误: 未安装 Go，或 go 不在 PATH 中" >&2
    exit 1
fi

cd "${PROJECT_ROOT}"

echo "启动 Personal Disk（APP_ENV=${APP_ENV}）"
echo "按 Ctrl+C 停止服务"
exec go run ./cmd/netdisk
