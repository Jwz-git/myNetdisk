#!/bin/bash

# Personal Disk 启动脚本
# 支持多环境配置

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 显示帮助信息
show_help() {
    echo "Personal Disk 启动脚本"
    echo ""
    echo "用法: $0 [环境] [选项]"
    echo ""
    echo "环境:"
    echo "  dev          开发环境 (默认)"
    echo "  test         测试环境"
    echo "  prod         生产环境"
    echo ""
    echo "选项:"
    echo "  --build      强制重新构建镜像"
    echo "  --logs       启动后查看日志"
    echo "  --help       显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 dev              # 启动开发环境"
    echo "  $0 prod --build     # 重新构建并启动生产环境"
    echo "  $0 test --logs      # 启动测试环境并查看日志"
}

# 检查 Docker 和 Docker Compose 是否安装
check_requirements() {
    if ! command -v docker &> /dev/null; then
        print_message $RED "错误: Docker 未安装或未在 PATH 中"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_message $RED "错误: Docker Compose 未安装或未在 PATH 中"
        exit 1
    fi
}

# 设置环境变量
setup_environment() {
    local env=$1

    export APP_ENV=$env

    case $env in
        "dev"|"development")
            export APP_ENV=development
            export COMPOSE_FILE="deploy/docker-compose.yml"
            export COMPOSE_PROJECT_NAME="personal-disk-dev"
            print_message $BLUE "配置开发环境..."
            ;;
        "test")
            export APP_ENV=test
            export COMPOSE_FILE="deploy/docker-compose.yml"
            export COMPOSE_PROJECT_NAME="personal-disk-test"
            print_message $YELLOW "配置测试环境..."
            ;;
        "prod"|"production")
            export APP_ENV=production
            export COMPOSE_FILE="deploy/docker-compose.prod.yml"
            export COMPOSE_PROJECT_NAME="personal-disk-prod"
            print_message $GREEN "配置生产环境..."

            # 生产环境安全检查
            if [ -z "$ADMIN_PASSWORD" ]; then
                print_message $RED "错误: 生产环境必须设置 ADMIN_PASSWORD 环境变量"
                exit 1
            fi
            if [ -z "$DB_PASSWORD" ]; then
                print_message $RED "错误: 生产环境必须设置 DB_PASSWORD 环境变量"
                exit 1
            fi
            ;;
        *)
            print_message $RED "错误: 未知环境 '$env'"
            show_help
            exit 1
            ;;
    esac
}

# 创建必要的目录
create_directories() {
    local env=$1

    case $env in
        "development")
            mkdir -p uploads logs
            ;;
        "test")
            mkdir -p uploads_test logs_test
            ;;
        "production")
            mkdir -p uploads logs
            ;;
    esac
}

# 启动服务
start_services() {
    local build_flag=$1
    local logs_flag=$2

    print_message $BLUE "启动服务..."

    if [ "$build_flag" = "true" ]; then
        docker-compose -f $COMPOSE_FILE build --no-cache
    fi

    docker-compose -f $COMPOSE_FILE up -d

    if [ $? -eq 0 ]; then
        print_message $GREEN "服务启动成功!"

        # 显示服务信息
        echo ""
        print_message $BLUE "=== 服务信息 ==="
        echo "环境: $APP_ENV"
        echo "配置文件: $COMPOSE_FILE"
        echo "项目名: $COMPOSE_PROJECT_NAME"

        # 获取应用端口
        local app_port=$(docker-compose -f $COMPOSE_FILE port app 8080 2>/dev/null | cut -d: -f2)
        if [ -n "$app_port" ]; then
            echo ""
            print_message $GREEN "应用访问地址:"
            echo "  主页: http://localhost:$app_port/"
            echo "  登录: http://localhost:$app_port/login"
            echo "  管理: http://localhost:$app_port/admin"
        fi

        echo ""
        print_message $BLUE "常用命令:"
        echo "  查看日志: docker-compose -f $COMPOSE_FILE logs -f"
        echo "  停止服务: docker-compose -f $COMPOSE_FILE down"
        echo "  重启服务: docker-compose -f $COMPOSE_FILE restart"

        if [ "$logs_flag" = "true" ]; then
            echo ""
            print_message $YELLOW "正在显示日志 (Ctrl+C 退出):"
            docker-compose -f $COMPOSE_FILE logs -f
        fi
    else
        print_message $RED "服务启动失败!"
        exit 1
    fi
}

# 主函数
main() {
    local env="dev"
    local build_flag="false"
    local logs_flag="false"

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            dev|development|test|prod|production)
                env=$1
                shift
                ;;
            --build)
                build_flag="true"
                shift
                ;;
            --logs)
                logs_flag="true"
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                print_message $RED "错误: 未知参数 '$1'"
                show_help
                exit 1
                ;;
        esac
    done

    # 检查系统要求
    check_requirements

    # 设置环境
    setup_environment $env

    # 创建目录
    create_directories $env

    # 启动服务
    start_services $build_flag $logs_flag
}

# 脚本入口
main "$@"