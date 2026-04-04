#!/bin/bash

# 数据库管理工具
# 提供数据库的常用操作：备份、恢复、重置等

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

show_help() {
    echo "数据库管理工具"
    echo ""
    echo "用法: $0 <命令> [选项]"
    echo ""
    echo "命令:"
    echo "  backup     备份数据库"
    echo "  restore    恢复数据库"
    echo "  reset      重置数据库"
    echo "  status     查看数据库状态"
    echo "  logs       查看数据库日志"
    echo ""
    echo "选项:"
    echo "  --env <env>    指定环境 (dev|test|prod, 默认: dev)"
    echo "  --file <file>  指定备份文件路径"
    echo ""
    echo "示例:"
    echo "  $0 backup --env dev"
    echo "  $0 restore --env prod --file backup.sql"
}

# 获取数据库容器名
get_db_container() {
    local env=$1
    case $env in
        "dev"|"development")
            echo "personal_disk_db"
            ;;
        "test")
            echo "personal_disk_db_test"
            ;;
        "prod"|"production")
            echo "personal_disk_db_prod"
            ;;
        *)
            echo "personal_disk_db"
            ;;
    esac
}

# 检查容器是否运行
check_container() {
    local container=$1
    if ! docker ps | grep -q $container; then
        print_message $RED "错误: 数据库容器 $container 未运行"
        print_message $YELLOW "请先启动相应环境: ./scripts/start.sh $env"
        exit 1
    fi
}

# 备份数据库
backup_database() {
    local env=$1
    local container=$(get_db_container $env)
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_dir="backup/${env}"
    local backup_file="${backup_dir}/database_${timestamp}.sql"

    print_message $BLUE "备份数据库..."

    check_container $container

    mkdir -p $backup_dir

    docker exec $container mysqldump -u root -p${DB_PASSWORD:-root} personal_disk > $backup_file

    if [ $? -eq 0 ]; then
        print_message $GREEN "数据库备份成功: $backup_file"

        # 压缩备份文件
        gzip $backup_file
        print_message $GREEN "备份文件已压缩: ${backup_file}.gz"
    else
        print_message $RED "数据库备份失败"
        exit 1
    fi
}

# 恢复数据库
restore_database() {
    local env=$1
    local backup_file=$2
    local container=$(get_db_container $env)

    if [ -z "$backup_file" ]; then
        print_message $RED "错误: 请指定备份文件路径"
        exit 1
    fi

    if [ ! -f "$backup_file" ]; then
        print_message $RED "错误: 备份文件不存在: $backup_file"
        exit 1
    fi

    print_message $BLUE "恢复数据库..."
    print_message $YELLOW "警告: 这将覆盖现有数据，请确认操作"

    read -p "确认恢复数据库? (y/N): " confirm
    if [[ $confirm != [yY] ]]; then
        print_message $YELLOW "取消操作"
        exit 0
    fi

    check_container $container

    # 如果是压缩文件，先解压
    if [[ $backup_file == *.gz ]]; then
        gunzip -c $backup_file | docker exec -i $container mysql -u root -p${DB_PASSWORD:-root} personal_disk
    else
        docker exec -i $container mysql -u root -p${DB_PASSWORD:-root} personal_disk < $backup_file
    fi

    if [ $? -eq 0 ]; then
        print_message $GREEN "数据库恢复成功"
    else
        print_message $RED "数据库恢复失败"
        exit 1
    fi
}

# 重置数据库
reset_database() {
    local env=$1
    local container=$(get_db_container $env)

    print_message $YELLOW "警告: 这将删除所有数据并重建数据库"
    read -p "确认重置数据库? (y/N): " confirm
    if [[ $confirm != [yY] ]]; then
        print_message $YELLOW "取消操作"
        exit 0
    fi

    check_container $container

    print_message $BLUE "重置数据库..."

    # 删除并重建数据库
    docker exec $container mysql -u root -p${DB_PASSWORD:-root} -e "DROP DATABASE IF EXISTS personal_disk; CREATE DATABASE personal_disk CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

    if [ $? -eq 0 ]; then
        print_message $GREEN "数据库重置成功"
    else
        print_message $RED "数据库重置失败"
        exit 1
    fi
}

# 查看数据库状态
show_status() {
    local env=$1
    local container=$(get_db_container $env)

    print_message $BLUE "数据库状态:"

    if docker ps | grep -q $container; then
        print_message $GREEN "✓ 容器运行中: $container"

        # 获取数据库信息
        echo ""
        print_message $BLUE "数据库信息:"
        docker exec $container mysql -u root -p${DB_PASSWORD:-root} -e "SHOW DATABASES;" 2>/dev/null

        echo ""
        print_message $BLUE "表信息:"
        docker exec $container mysql -u root -p${DB_PASSWORD:-root} personal_disk -e "SHOW TABLES;" 2>/dev/null

    else
        print_message $RED "✗ 容器未运行: $container"
    fi
}

# 查看数据库日志
show_logs() {
    local env=$1
    local container=$(get_db_container $env)

    if docker ps | grep -q $container; then
        docker logs -f $container
    else
        print_message $RED "容器未运行: $container"
    fi
}

# 主函数
main() {
    local command=""
    local env="dev"
    local backup_file=""

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            backup|restore|reset|status|logs)
                command=$1
                shift
                ;;
            --env)
                env=$2
                shift 2
                ;;
            --file)
                backup_file=$2
                shift 2
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

    if [ -z "$command" ]; then
        print_message $RED "错误: 请指定命令"
        show_help
        exit 1
    fi

    case $command in
        "backup")
            backup_database $env
            ;;
        "restore")
            restore_database $env $backup_file
            ;;
        "reset")
            reset_database $env
            ;;
        "status")
            show_status $env
            ;;
        "logs")
            show_logs $env
            ;;
        *)
            print_message $RED "错误: 未知命令 '$command'"
            show_help
            exit 1
            ;;
    esac
}

main "$@"