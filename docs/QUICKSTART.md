# 快速开始指南

这是 Personal Disk 项目的快速开始指南，帮助您在几分钟内启动和运行系统。

## 🚀 一键启动

### 方式一：使用 Make 命令（推荐）

```bash
# 启动开发环境
make dev

# 启动生产环境
export ADMIN_PASSWORD="your_secure_password"
make prod
```

### 方式二：使用启动脚本

```bash
# 启动开发环境
./scripts/start.sh dev

# 启动生产环境
ADMIN_PASSWORD="your_secure_password" ./scripts/start.sh prod
```

## 📋 系统要求

- **本地开发**: Go 1.22+；默认使用 SQLite，选择 MySQL 时需 MySQL 8.0+
- **容器部署（可选）**: Docker 20.0+ 和 Docker Compose

## 本地零配置启动

```bash
go run ./cmd/netdisk
```

默认数据写入 `data/personal_disk.db`。改用 MySQL 时设置 `DB_DRIVER=mysql` 及 `DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASSWORD`、`DB_NAME`。

## 🎯 访问地址

启动后访问以下地址：

- **主页**: http://localhost:8080/
- **登录**: http://localhost:8080/login  
- **管理**: http://localhost:8080/admin

## 🔑 默认账号

| 环境 | 用户名 | 密码 |
|------|--------|------|
| 开发 | `admin` | `123456` |
| 测试 | `test_admin` | `test123456` |
| 生产 | `admin` | **通过环境变量设置** |

## 🛠️ 常用操作

### 查看服务状态
```bash
docker ps
```

### 查看日志
```bash
make logs
# 或
docker logs personal_disk_app -f
```

### 停止服务
```bash
make stop
```

### 重新构建
```bash
make build
```

## 🐛 故障排除

### 端口被占用
```bash
# 查找占用端口的进程
lsof -i :8080

# 停止占用进程
kill $(lsof -ti :8080)
```

### 数据库连接失败

- SQLite：检查 `data/` 目录是否可写。
- MySQL：检查 `DB_HOST`、端口、账号密码和容器网络连通性。

### 清理 Docker 资源
```bash
# 清理所有资源
make clean

# 手动清理
docker system prune -f
```

## 📚 更多帮助

- [完整文档](README.md)
- [配置说明](docs/CONFIG.md)
- [项目结构](docs/STRUCTURE.md)

## 🤝 获取支持

遇到问题？

1. 查看 [故障排除文档](README.md#故障排除)
2. 提交 [Issue](../../issues)
3. 查看项目 [Wiki](../../wiki)

---

🎉 恭喜！您已成功启动 Personal Disk 系统！
