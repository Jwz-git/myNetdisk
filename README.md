# Personal Disk 🗂️

> 一个简洁、高效的个人网盘系统，支持文件上传、下载、管理等功能

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Supported-2496ED?style=flat-square&logo=docker&logoColor=white)
![MySQL](https://img.shields.io/badge/MySQL-8.0+-4479A1?style=flat-square&logo=mysql&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

## ✨ 特性

- 🚀 **高性能**: 基于 Go 语言，支持高并发文件操作
- 🔧 **易部署**: 支持 Docker 一键部署，多环境配置
- 🔐 **安全可靠**: 管理员认证，文件类型限制，安全上传
- 📱 **响应式**: 现代化 Web 界面，支持移动端访问
- ⚡ **轻量级**: 最小化依赖，快速启动
- 🌍 **多环境**: 开发、测试、生产环境独立配置

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Browser   │───▶│  Personal Disk  │───▶│   MySQL DB      │
│                 │    │   (Go Server)   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                        ┌─────────────────┐
                        │  File Storage   │
                        │   (uploads/)    │
                        └─────────────────┘
```

## 🚀 快速开始

### 方式一：Docker 部署（推荐）

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd personal-disk
   ```

2. **配置环境变量**
   ```bash
   cp env.example .env
   # 编辑 .env 文件，设置数据库密码等
   ```

3. **启动服务**
   ```bash
   # 开发环境
   ./scripts/start.sh dev
   
   # 或使用Makefile
   make dev
   
   # 生产环境
   export ADMIN_PASSWORD="your_secure_password"
   export DB_PASSWORD="your_db_password"
   ./scripts/start.sh prod
   # 或
   make prod
   ```

### 方式二：本地开发

1. **环境要求**
   - Go 1.20+
   - MySQL 8.0+

2. **安装依赖**
   ```bash
   go mod tidy
   ```

3. **配置数据库**
   ```bash
   # 创建数据库
   mysql -u root -p -e "CREATE DATABASE personal_disk CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
   ```

4. **启动应用**
   ```bash
   go run main.go
   ```

## 📖 使用说明

### 访问地址

- **主页**: http://localhost:8080/
- **管理后台**: http://localhost:8080/admin
- **登录页面**: http://localhost:8080/login

### 默认账号

| 环境 | 用户名 | 密码 | 说明 |
|------|--------|------|------|
| 开发环境 | `admin` | `dev123456` | 开发测试用 |
| 测试环境 | `test_admin` | `test123456` | 测试用 |
| 生产环境 | `admin` | **必须通过环境变量设置** | 安全要求 |

### 主要功能

#### 🔐 用户认证
- 管理员登录/登出
- 会话管理
- 访问权限控制

#### 📁 文件管理
- **文件上传**: 支持多文件同时上传
- **文件列表**: 查看所有已上传文件
- **文件下载**: 安全的文件下载
- **文件删除**: 删除不需要的文件
- **文件重命名**: 在线重命名文件

#### 🛡️ 安全特性
- 文件类型限制
- 文件大小限制
- 上传路径安全检查
- 管理员权限验证

## ⚙️ 配置说明

### 环境配置

通过 `APP_ENV` 环境变量切换环境：

```bash
# 开发环境 (默认)
export APP_ENV=development  # 使用 config/development.yml

# 测试环境
export APP_ENV=test         # 使用 config/test.yml

# 生产环境
export APP_ENV=production   # 使用 config/production.yml
```

配置文件优先级：
1. **环境特定配置** - `config/{environment}.yml`
2. **默认配置** - `config/default.yml`
3. **兼容配置** - `config.yml` (向后兼容，不推荐)

### 重要配置项

| 配置项 | 环境变量 | 说明 | 默认值 |
|--------|----------|------|--------|
| 数据库主机 | `DB_HOST` | MySQL 服务器地址 | `127.0.0.1` |
| 数据库端口 | `DB_PORT` | MySQL 端口 | `3306` |
| 数据库用户 | `DB_USER` | 数据库用户名 | `root` |
| 数据库密码 | `DB_PASSWORD` | 数据库密码 | 空 |
| 服务器端口 | `SERVER_PORT` | 应用监听端口 | `8080` |
| 管理员用户名 | `ADMIN_USERNAME` | 管理员用户名 | `admin` |
| 管理员密码 | `ADMIN_PASSWORD` | 管理员密码 | **必须设置** |
| 上传目录 | `UPLOAD_DIRECTORY` | 文件上传路径 | `uploads` |
| 最大文件大小 | `UPLOAD_MAX_SIZE` | 单文件最大大小(字节) | `536870912` (512MB) |

📋 详细配置说明请参考：[配置文档](docs/CONFIG.md)

## 🐳 Docker 部署

### 开发环境

```bash
# 快速启动
make dev
# 或
./scripts/start.sh dev

# 手动启动
docker-compose -f deploy/docker-compose.yml up -d
```

### 生产环境

```bash
# 设置必要的环境变量
export ADMIN_PASSWORD="your_secure_password_here"
export DB_PASSWORD="your_database_password_here"

# 启动生产环境
make prod
# 或
./scripts/start.sh prod

# 或使用生产配置文件
docker-compose -f deploy/docker-compose.prod.yml up -d
```

### Docker 容器管理

```bash
# 查看日志
make logs

# 停止服务
make stop

# 重启服务
docker-compose -f deploy/docker-compose.yml restart

# 重新构建
make build
```

## 🔧 开发指南

### 项目结构

详细的项目结构说明请参考：[项目结构文档](docs/STRUCTURE.md)

### 本地开发

1. **启动开发环境**
   ```bash
   ./scripts/start.sh dev --logs
   ```

2. **代码热重载**（可选）
   ```bash
   # 安装 air
   go install github.com/cosmtrek/air@latest
   
   # 启动热重载
   air
   ```

3. **数据库操作**
   ```bash
   # 连接开发数据库
   docker exec -it personal_disk_db mysql -u root -p personal_disk_dev
   ```

### API 接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| `POST` | `/api/login` | 用户登录 | 否 |
| `POST` | `/api/upload` | 上传文件 | 是 |
| `GET` | `/api/files` | 获取文件列表 | 是 |
| `GET` | `/api/download/{id}` | 下载文件 | 是 |
| `DELETE` | `/api/delete/{id}` | 删除文件 | 是 |
| `PUT` | `/api/rename/{id}` | 重命名文件 | 是 |

## 📊 性能特性

- ⚡ **并发上传**: 支持多文件并发上传
- 🗄️ **流式下载**: 大文件流式传输，节省内存
- 📝 **连接池**: 数据库连接池优化
- 🔄 **会话管理**: 高效的用户会话处理

## 🛡️ 安全考虑

- ✅ **文件类型验证**: 严格的文件类型白名单
- ✅ **路径安全**: 防止目录遍历攻击
- ✅ **大小限制**: 防止大文件攻击
- ✅ **认证保护**: 管理功能需要认证
- ✅ **CORS 控制**: 可配置的跨域策略
- ✅ **错误处理**: 安全的错误信息处理

## 🚨 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 查找占用端口的进程
   lsof -i :8080
   # 或更换端口
   export SERVER_PORT=8081
   ```

2. **数据库连接失败**
   - 检查 MySQL 服务是否启动
   - 验证数据库连接配置
   - 确认网络连通性

3. **文件上传失败**
   - 检查上传目录权限
   - 验证文件大小限制
   - 确认文件类型是否允许

4. **Docker 启动失败**
   ```bash
   # 查看详细日志
   docker-compose logs
   
   # 重新构建镜像
   docker-compose build --no-cache
   ```

### 日志查看

```bash
# Docker 环境
docker-compose logs -f app

# 本地开发
tail -f app.log
```

## 📈 监控和维护

### 健康检查

```bash
# 检查应用状态
curl http://localhost:8080/

# 检查数据库连接
docker exec personal_disk_db mysqladmin ping
```

### 备份建议

```bash
# 数据库备份
docker exec personal_disk_db mysqldump -u root -p personal_disk > backup.sql

# 文件备份
tar -czf uploads_backup.tar.gz uploads/
```

---

<div align="center">
  <p>如果这个项目对您有帮助，请考虑给它一个 ⭐</p>
  <p>有问题或建议？欢迎提交 <a href="../../issues">Issue</a></p>
</div>