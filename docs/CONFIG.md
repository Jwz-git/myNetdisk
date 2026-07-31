# 配置管理文档

## 概述

Personal Disk 项目采用了分层配置管理系统，支持多环境部署和配置解耦。配置系统具有以下特性：

- 🎯 **统一配置管理**: YAML 配置集中在 `configs/`，加载代码位于 `internal/config/`
- 🌍 **多环境支持**: 开发、测试、生产环境独立配置
- 🔒 **环境变量覆盖**: 敏感信息可通过环境变量安全设置
- ✅ **配置验证**: 自动验证配置项的有效性
- 🐳 **Docker 友好**: 完美支持容器化部署

## 目录结构

```
internal/config/
├── config.go          # 配置结构体定义和加载逻辑
└── init.go            # 配置初始化工具

configs/
├── default.yml        # 默认配置文件
├── development.yml    # 开发环境配置
├── test.yml          # 测试环境配置
└── production.yml    # 生产环境配置
```

## 配置文件

### 环境配置优先级

1. **环境变量** (最高优先级)
2. **环境配置文件** (`configs/{environment}.yml`)
3. **默认配置文件** (`configs/default.yml`)
4. **兼容配置文件** (`config.yml` - 向后兼容，不推荐)

### 环境切换

通过 `APP_ENV` 环境变量控制：

```bash
# 开发环境 (默认)
export APP_ENV=development

# 测试环境
export APP_ENV=test

# 生产环境
export APP_ENV=production
```

## 配置项说明

### 数据库配置 (`database`)

| 配置项 | 说明 | 默认值 | 环境变量覆盖 |
|--------|------|--------|--------------|
| `driver` | 数据库类型：`sqlite` 或 `mysql` | `sqlite` | `DB_DRIVER` |
| `path` | SQLite 数据库文件路径 | `data/personal_disk.db` | `DB_PATH` |
| `host` | 数据库主机地址 | `127.0.0.1` | `DB_HOST` |
| `port` | 数据库端口 | `3306` | `DB_PORT` |
| `user` | 数据库用户名 | `root` | `DB_USER` |
| `password` | 数据库密码 | `""` | `DB_PASSWORD` |
| `name` | 数据库名称 | `personal_disk` | `DB_NAME` |
| `charset` | 字符集 | `utf8mb4` | - |
| `parseTime` | 是否解析时间 | `true` | - |
| `location` | 时区 | `Local` | - |

### 服务器配置 (`server`)

| 配置项 | 说明 | 默认值 | 环境变量覆盖 |
|--------|------|--------|--------------|
| `host` | 监听地址 | `0.0.0.0` | `SERVER_HOST` |
| `port` | 监听端口 | `8080` | `SERVER_PORT` |

### 管理员配置 (`admin`)

| 配置项 | 说明 | 默认值 | 环境变量覆盖 |
|--------|------|--------|--------------|
| `username` | 管理员用户名 | `admin` | `ADMIN_USERNAME` |
| `password` | 管理员密码 | 根据环境而定 | `ADMIN_PASSWORD` |

⚠️ **安全提示**: 生产环境必须通过环境变量设置管理员密码！

### 文件上传配置 (`upload`)

| 配置项 | 说明 | 默认值 | 环境变量覆盖 |
|--------|------|--------|--------------|
| `directory` | 上传目录 | `storage` | `UPLOAD_DIRECTORY` |
| `maxSize` | 最大文件大小(字节) | `536870912` (512MB) | `UPLOAD_MAX_SIZE` |
| `allowedTypes` | 允许的文件类型 | 见配置文件 | - |

### 会话配置 (`session`)

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `cookieName` | Cookie名称 | `admin_session` |
| `cookieValue` | Cookie值 | `admin_logged_in` |
| `cookiePath` | Cookie路径 | `/` |
| `maxAge` | 过期时间(秒) | `86400` (24小时) |

### 日志配置 (`logging`)

| 配置项 | 说明 | 默认值 | 环境变量覆盖 |
|--------|------|--------|--------------|
| `level` | 日志级别 | `info` | `LOG_LEVEL` |
| `file` | 日志文件路径 | `""` (输出到控制台) | `LOG_FILE` |
| `maxSize` | 单个日志文件最大大小(MB) | `100` | - |
| `maxBackups` | 保留日志文件数量 | `3` | - |
| `maxAge` | 日志保留天数 | `28` | - |

### 安全配置 (`security`)

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `enableCORS` | 是否启用跨域 | `true` (开发), `false` (生产) |
| `allowedOrigins` | 允许的跨域来源 | 见配置文件 |
| `enableRateLimit` | 是否启用速率限制 | `false` (开发), `true` (生产) |

## 使用方法

### 1. 在代码中使用配置

```go
package main

import (
    "personal-disk/internal/config"
)

func main() {
    // 初始化配置
    cfg := config.InitConfig()
    
    // 使用配置
    fmt.Println("数据库连接:", cfg.GetDSN())
    fmt.Println("服务器地址:", cfg.GetServerAddr())
    
    // 或者获取全局配置
    cfg = config.MustGetConfig()
}
```

### 2. 本地开发

```bash
# 1. 复制环境变量示例文件
cp env.example .env

# 2. 编辑 .env 文件设置本地配置
vim .env

# 3. 启动应用 (会自动加载开发环境配置)
go run ./cmd/netdisk

# 或使用启动脚本
./scripts/start.sh dev
```

### 3. 启动脚本

```bash
# 本地开发环境
./scripts/start.sh dev

# 本地测试环境
./scripts/start.sh test

# 本地生产环境（需要设置环境变量）
export ADMIN_PASSWORD="your_secure_password"
./scripts/start.sh prod
```

### 4. 手动 Docker Compose

```bash
# 开发环境
docker compose -f deploy/docker-compose.yml up -d --build

# 生产环境
ADMIN_PASSWORD="your_secure_password" \
docker compose -f deploy/docker-compose.prod.yml up -d --build
```

## 环境变量设置

### 开发环境

创建 `.env` 文件：

```bash
APP_ENV=development
DB_DRIVER=sqlite
DB_PATH=data/personal_disk.db
ADMIN_PASSWORD=123456
```

改用 MySQL：

```bash
DB_DRIVER=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=dev_password
DB_NAME=personal_disk
```

### 生产环境

```bash
# 必须设置的环境变量
export APP_ENV=production
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=your_secure_password_here

# 可选的环境变量
export SERVER_PORT=8080
export UPLOAD_MAX_SIZE=536870912
export LOG_LEVEL=warn
```

## 配置验证

配置系统会自动验证以下项目：

- ✅ 必填字段不能为空
- ✅ MySQL 模式下端口号格式正确 (1-65535)
- ✅ SQLite 模式下数据库路径非空
- ✅ 管理员账号和密码不能为空

## 最佳实践

### 🔒 安全实践

1. **生产环境密码**: 永远不要在配置文件中硬编码管理员密码或 MySQL 密码
2. **环境变量**: 敏感信息使用环境变量传递
3. **权限控制**: 确保配置文件和日志目录有适当的权限

### 🔧 配置管理

1. **环境分离**: 不同环境使用独立的配置文件
2. **默认值**: 为所有配置项提供合理的默认值
3. **文档更新**: 添加新配置项时同步更新文档

### 🚀 部署实践

1. **配置验证**: 部署前验证配置文件语法
2. **环境检查**: 确认环境变量正确设置
3. **日志监控**: 配置适当的日志级别和保留策略

## 故障排除

### 配置加载失败

```
错误: 配置文件不存在: configs/production.yml
```

**解决**: 确保对应环境的配置文件存在，或设置正确的 `APP_ENV`

### 环境变量未生效

```
错误: 生产环境必须设置 ADMIN_PASSWORD 环境变量
```

**解决**: 
```bash
export ADMIN_PASSWORD=your_password
# 或者在 docker-compose 中设置环境变量
```

### 端口冲突

```
错误: 服务器端口格式不正确: abc
```

**解决**: 确保端口号是有效的数字 (1-65535)

### 数据库连接失败

1. 检查数据库配置是否正确
2. 确认数据库服务已启动
3. 验证网络连接和防火墙设置

## 配置迁移

如果从旧版本升级，请按以下步骤操作：

1. **备份现有配置**: `cp config.yml config.yml.backup`
2. **更新代码**: 拉取最新代码
3. **安装依赖**: `go mod tidy`
4. **迁移配置**: 将旧配置项迁移到新的配置结构
5. **测试验证**: 在开发环境验证配置正确性

---

📚 更多信息请参考项目源码中的配置文件示例和注释。
