# 项目结构

项目按“可执行入口、内部代码、运行配置、Web 资源、部署资料”划分。运行时生成的 `data/`、`storage/` 和 `logs/` 不提交到 Git。

```text
personal-disk/
├── cmd/
│   └── netdisk/
│       └── main.go                 应用入口、路由注册和服务启动
├── internal/
│   ├── config/                     配置结构、加载、覆盖和校验
│   ├── controller/                 页面与 API 处理器
│   ├── middleware/                 CORS 和请求限流
│   └── model/                      数据库初始化与文件元数据访问
├── configs/
│   ├── default.yml                 兜底配置
│   ├── development.yml             开发环境
│   ├── test.yml                    测试环境
│   └── production.yml              生产环境
├── web/
│   ├── static/                     CSS 和 JavaScript
│   └── templates/                  HTML 页面
├── build/
│   └── Dockerfile
├── deploy/
│   ├── docker-compose.yml          开发部署
│   ├── docker-compose.prod.yml     生产部署
│   └── env.example
├── docs/                            配置、快速开始和变更记录
├── scripts/
│   └── start.sh                    Docker 多环境启动脚本
├── Makefile
├── env.example
├── go.mod
├── go.sum
└── README.md
```

## 边界说明

- `cmd/netdisk` 只负责组装依赖、注册路由和启动服务。
- `internal` 中的包只能由本模块内部导入，避免形成无意的公共 API。
- `configs` 只保存数据文件，不混放 Go 源码。
- `web` 统一保存需要随应用发布的浏览器资源。
- `build` 保存镜像构建定义，`deploy` 保存具体运行编排。
- 测试文件与被测包放在同一目录，便于发现和维护。

## 常用命令

```bash
go run ./cmd/netdisk
go test ./...
go vet ./...
CGO_ENABLED=0 go build ./cmd/netdisk
```

命令应在项目根目录执行，因为配置和 Web 资源路径以项目工作目录为基准。
