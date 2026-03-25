# 个人网盘

一个基于 Go + MySQL 的简单个人网盘。

## 快速开始（推荐）

```bash
git clone https://github.com/Jwz-git/myNetdisk.git
cd myNetdisk
docker-compose up -d
```

访问地址（局域网访问请替换 `localhost` 为服务器 IP 地址）：
- 访客页面：http://localhost:8080/
- 登录页面：http://localhost:8080/login
- 管理页面：http://localhost:8080/admin（需要登录）

默认管理员凭据（在controller/login_controller.go 中配置）：
- 用户名：admin
- 密码：123456

常用命令：

```bash
docker-compose ps
docker-compose logs -f
docker-compose down
docker-compose down -v
```

## 源码运行

推荐使用脚本一键启动：

```bash
bash run_source.sh
```

可选参数示例：

```bash
DB_PASSWORD=your_password INIT_DB=1 bash run_source.sh
```

环境要求：
- Go 1.18+
- MySQL 5.7+ / 8.0+

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `DB_HOST` | `127.0.0.1` | MySQL 主机 |
| `DB_PORT` | `3306` | MySQL 端口 |
| `DB_USER` | `root` | MySQL 用户 |
| `DB_PASSWORD` | `` | MySQL 密码 |
| `DB_NAME` | `personal_disk` | 数据库名 |

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/upload` | 上传文件 |
| GET | `/api/files` | 获取文件列表 |
| GET | `/api/download/{id}` | 下载文件 |
| POST | `/api/delete/{id}` | 删除文件 |
| POST | `/api/rename/{id}` | 重命名文件 |
| POST | `/api/login` | 管理员登录 |
| GET | `/api/logout` | 管理员退出登录 |

## 说明

- 应用启动时会自动初始化表结构。
- 上传文件保存在 `uploads/`，MySQL 数据保存在 `mysql-data` 卷。
- M1/M2 Mac 已在 Compose 中设置 `platform: linux/amd64` 兼容 MySQL 镜像。
