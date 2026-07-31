# Personal Disk

基于 Go 的轻量级个人网盘，支持文件与文件夹上传、浏览、下载、重命名和删除。

默认使用 SQLite，首次运行无需准备数据库；需要独立数据库服务时可以切换到 MySQL。

## 主要功能

- 上传单个或多个文件；文件夹会先在浏览器内压缩成 ZIP，再作为单个文件上传
- 下载文件；下载文件夹时实时生成 ZIP
- 管理员登录和带有效期的签名会话
- 登录后执行上传、重命名和删除操作
- 校验上传大小、文件类型和相对路径
- 可配置 CORS 和基于客户端 IP 的请求限流
- 支持 SQLite 与 MySQL
- 支持本地运行和 Docker 部署

> 当前文件列表和下载接口无需登录。如果网盘内容不能公开访问，应在部署前为这两个接口增加认证，或通过反向代理限制访问。

## 快速开始

### 本地运行

要求 Go 1.22 或更高版本。

```bash
git clone https://github.com/Jwz-git/myNetdisk.git
cd myNetdisk
./scripts/start.sh dev
```

启动后访问：

- 首页：<http://localhost:8080/>
- 登录页：<http://localhost:8080/login>
- 管理页：<http://localhost:8080/admin>

开发环境默认管理员为 `admin` / `123456`，仅供本地体验。生产部署必须设置自己的密码。

首次启动会自动创建 SQLite 数据库 `data/personal_disk.db`；`storage/` 会在第一次上传时创建。

### Docker 运行

开发配置同样默认使用 SQLite，只启动一个应用容器：

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

查看日志或停止服务：

```bash
docker compose -f deploy/docker-compose.yml logs -f app
docker compose -f deploy/docker-compose.yml down
```

SQLite 数据和上传内容保存在 Docker 命名卷中。上述 `down` 命令不会删除数据；执行 `docker compose -f deploy/docker-compose.yml down -v` 会删除这些卷和其中的数据。

## 配置

应用通过 `APP_ENV` 选择配置文件，环境变量的值优先于 YAML：

| `APP_ENV` | 配置文件 | 数据库 |
|---|---|---|
| 未设置或 `development` | `configs/development.yml` | SQLite |
| `test` | `configs/test.yml` | SQLite |
| `production` | `configs/production.yml` | SQLite |

常用环境变量：

| 变量 | 用途 | 默认值 |
|---|---|---|
| `DB_DRIVER` | 数据库驱动：`sqlite` 或 `mysql` | `sqlite` |
| `DB_PATH` | SQLite 文件路径 | `data/personal_disk.db` |
| `DB_HOST` | MySQL 地址 | `127.0.0.1` |
| `DB_PORT` | MySQL 端口 | `3306` |
| `DB_USER` | MySQL 用户名 | `root` |
| `DB_PASSWORD` | MySQL 密码 | 空 |
| `DB_NAME` | MySQL 数据库名 | `personal_disk` |
| `SERVER_HOST` | 服务监听地址 | 由环境配置决定 |
| `SERVER_PORT` | 服务监听端口 | `8080` |
| `ADMIN_USERNAME` | 管理员用户名 | `admin` |
| `ADMIN_PASSWORD` | 管理员密码 | 由环境配置决定 |
| `UPLOAD_DIRECTORY` | 文件保存目录 | `storage` |
| `UPLOAD_MAX_SIZE` | 单文件上限，单位为字节 | 由环境配置决定 |

本地使用示例文件：

```bash
cp env.example .env
set -a
source .env
set +a
go run ./cmd/netdisk
```

应用本身不会读取 `.env`，因此本地运行前需要将变量导入当前 Shell。Docker Compose 会自动读取项目根目录的 `.env`。

启动脚本在前台运行，按 `Ctrl+C` 停止服务。它只依赖 Go，不依赖 Docker；也可直接执行 `go run ./cmd/netdisk`。

完整配置说明见 [docs/CONFIG.md](docs/CONFIG.md)。

## 使用 MySQL

MySQL 是显式选项，不会随默认 Docker Compose 一起启动。先创建数据库：

```bash
mysql -u root -p -e \
  'CREATE DATABASE personal_disk CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;'
```

然后提供连接参数：

```bash
DB_DRIVER=mysql \
DB_HOST=127.0.0.1 \
DB_PORT=3306 \
DB_USER=root \
DB_PASSWORD='your-password' \
DB_NAME=personal_disk \
go run ./cmd/netdisk
```

应用会自动创建或补齐 `file_info` 表，但不会在 SQLite 与 MySQL 之间迁移已有数据。

容器连接外部 MySQL 时，`DB_HOST` 必须是容器内可访问的地址，通常不能填写宿主机的 `127.0.0.1`。

## 生产部署

生产 Compose 要求显式设置管理员密码：

```bash
ADMIN_PASSWORD='replace-with-a-strong-password' \
docker compose -f deploy/docker-compose.prod.yml up -d --build
```

生产环境建议同时做到：

- 使用高强度 `ADMIN_PASSWORD`
- 通过 HTTPS 反向代理提供服务
- 限制公开访问范围；当前列表和下载接口默认公开
- 同时备份数据库和上传目录
- 多实例部署时使用 MySQL，并为上传目录提供共享存储

## API

| 方法 | 路径 | 用途 | 需要登录 |
|---|---|---|---|
| `POST` | `/api/login` | 登录 | 否 |
| `POST` | `/api/upload` | 上传文件或文件夹 | 是 |
| `GET` | `/api/files` | 获取文件列表 | 否 |
| `GET` | `/api/download/{id}` | 下载文件或文件夹 ZIP | 否 |
| `DELETE` | `/api/delete/{id}` | 删除文件或文件夹 | 是 |
| `POST` / `PUT` | `/api/rename/{id}` | 重命名 | 是 |
| `GET` | `/logout` | 退出登录 | 否 |

重命名请求体：

```json
{"file_name":"new-name.txt"}
```

## 开发与验证

```bash
go test ./...
go vet ./...
CGO_ENABLED=0 go build ./cmd/netdisk
```

常用 Make 目标：

```bash
make run      # 本地运行
make test     # 执行测试
make check    # 格式化、静态检查和测试
make build    # 构建 Docker 镜像
make clean    # 清理测试与构建临时文件
```

`make clean` 不会删除 `data/`、`storage/` 或 Docker 数据卷。

## 项目结构

```text
cmd/netdisk/          应用入口与路由组装
internal/config/      配置加载和校验
internal/controller/  页面与 API 处理器
internal/middleware/  CORS 和请求限流
internal/model/       数据库初始化与文件元数据访问
configs/              YAML 环境配置
web/static/           CSS 和 JavaScript
web/templates/        HTML 模板
build/Dockerfile      应用镜像的构建规则
deploy/               Docker Compose 运行编排
docs/                 配置、快速开始和结构说明
```

`build/` 与 `deploy/` 都包含 Docker 相关内容，但职责不同：前者定义镜像如何生成，后者定义镜像如何运行。完整目录说明见 [docs/STRUCTURE.md](docs/STRUCTURE.md)。

## 数据边界

- 数据库只保存文件元数据，文件内容保存在上传目录。
- 恢复服务需要同时恢复数据库与上传内容。
- SQLite 适合单实例运行；它不是多容器并发共享数据库的替代品。
- 文件夹压缩发生在浏览器内存中，选择超大文件夹时需要足够的客户端可用内存。
