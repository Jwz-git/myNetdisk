# Personal Disk 项目结构

## 📁 优化后的目录结构

```
personal-disk/
├── 📁 build/               # 构建相关文件
│   └── Dockerfile          # Docker镜像构建文件
├── 📁 config/              # 配置管理
│   ├── config.go           # 配置结构体和加载逻辑
│   ├── init.go            # 配置初始化工具
│   ├── default.yml        # 默认配置文件
│   ├── development.yml    # 开发环境配置
│   ├── test.yml          # 测试环境配置
│   └── production.yml    # 生产环境配置
├── 📁 controller/          # 控制器层
│   ├── file_controller.go # 文件管理控制器
│   └── login_controller.go# 登录认证控制器
├── 📁 deploy/              # 部署相关文件
│   ├── docker-compose.yml     # Docker开发/测试配置
│   ├── docker-compose.prod.yml# Docker生产配置
│   └── env.example        # 环境变量示例
├── 📁 docs/               # 项目文档
│   ├── CHANGELOG.md       # 更新日志
│   ├── CONFIG.md         # 配置管理详细文档
│   ├── CONTRIBUTING.md   # 贡献指南
│   ├── LICENSE           # 开源许可证
│   └── STRUCTURE.md      # 项目结构文档 (本文件)
├── 📁 model/              # 数据模型
│   └── file_model.go     # 文件数据模型
├── 📁 scripts/            # 脚本工具
│   ├── start.sh          # 多环境启动脚本
│   └── db.sh             # 数据库管理工具
├── 📁 static/             # 静态资源
│   ├── css/              # 样式文件
│   │   ├── admin.css     # 管理页面样式
│   │   ├── login.css     # 登录页面样式
│   │   └── visitor.css   # 访客页面样式
│   └── js/               # JavaScript文件
│       ├── admin.js      # 管理页面脚本
│       ├── login.js      # 登录页面脚本
│       └── visitor.js    # 访客页面脚本
├── 📁 templates/          # HTML模板
│   ├── admin.html        # 管理页面
│   ├── index.html        # 首页
│   └── login.html        # 登录页面
├── 📁 third_party/        # 第三方依赖
│   └── go_mysql/         # MySQL驱动
├── 📁 uploads/            # 文件上传目录 (运行时创建)
├── 📄 go.mod              # Go模块定义
├── 📄 go.sum              # Go依赖锁定
├── 📄 main.go             # 应用入口
├── 📄 Makefile            # 构建和管理工具
├── 📄 .gitignore          # Git忽略规则
└── 📄 README.md           # 项目说明文档
```

## 🗂️ 目录分类说明

### 📋 **核心应用代码**
```
├── main.go              # 应用程序入口点
├── controller/          # HTTP请求处理层
├── model/               # 数据访问层
└── config/              # 配置管理层
```

### 🎨 **前端资源**
```
├── static/              # 静态资源文件
│   ├── css/            # 样式表
│   └── js/             # JavaScript脚本
└── templates/          # HTML模板
```

### 🚀 **部署和构建**
```
├── build/              # 构建相关文件
│   └── Dockerfile      # 容器镜像构建
├── deploy/             # 部署配置文件
│   ├── docker-compose.yml      # 开发/测试环境
│   ├── docker-compose.prod.yml # 生产环境
│   └── env.example     # 环境变量模板
└── scripts/            # 自动化脚本
    ├── start.sh        # 启动脚本
    └── db.sh           # 数据库管理
```

### 📚 **文档和工具**
```
├── docs/               # 项目文档
│   ├── CONFIG.md       # 配置说明
│   ├── STRUCTURE.md    # 结构文档
│   ├── CONTRIBUTING.md # 贡献指南
│   ├── CHANGELOG.md    # 更新日志
│   └── LICENSE         # 开源许可
├── Makefile           # 构建自动化
└── README.md          # 项目介绍
```

### 🔧 **Go项目标准文件**
```
├── go.mod              # Go模块定义
├── go.sum              # 依赖版本锁定
├── .gitignore          # Git忽略规则
└── third_party/        # 第三方库
```

## 🎯 **设计原则**

### 1. **关注点分离**
- **应用代码**: 核心业务逻辑
- **配置文件**: 环境相关设置  
- **部署脚本**: 运维自动化
- **文档资料**: 使用和开发指南

### 2. **环境隔离**
- **开发环境**: 本地开发和调试
- **测试环境**: 自动化测试和集成
- **生产环境**: 线上运行配置

### 3. **工具集成**
- **Makefile**: 统一的构建接口
- **Scripts**: 常用操作自动化
- **Docker**: 容器化部署支持

### 4. **文档完整**
- **README**: 项目概览和快速开始
- **CONFIG**: 详细配置说明
- **CONTRIBUTING**: 开发协作规范

## 🚀 **常用操作**

### 快速启动
```bash
# 开发环境
make dev

# 生产环境  
make prod
```

### 数据库管理
```bash
# 备份数据库
./scripts/db.sh backup --env dev

# 查看状态
./scripts/db.sh status --env dev
```

### 代码检查
```bash
# 完整检查
make check

# 格式化代码
make fmt
```

## 📝 **目录命名规范**

- **小写字母**: 所有目录名使用小写
- **下划线分隔**: 多词目录使用下划线 `third_party`
- **复数形式**: 集合性目录使用复数 `scripts`, `docs`
- **功能导向**: 目录名体现功能用途

## 🔄 **扩展建议**

### 可能的新增目录
```
├── api/                # API文档和定义
├── test/               # 测试文件和测试数据
├── tools/              # 开发工具和实用程序
├── internal/           # 内部包（不对外暴露）
├── pkg/                # 可复用的包
└── examples/           # 使用示例
```

### 大型项目结构
```
├── cmd/                # 应用程序入口点
├── internal/           # 私有应用和库代码
├── pkg/                # 外部应用可使用的库代码
├── vendor/             # 应用程序依赖项
├── api/                # API定义文件
├── web/                # Web应用程序相关组件
├── configs/            # 配置文件模板
├── init/               # 系统初始化配置
├── deployments/        # 部署相关的配置和脚本
└── test/               # 额外的外部测试应用和测试数据
```

---

这种目录结构的设计既保持了Go项目的标准规范，又提供了清晰的功能分离，使项目易于维护和扩展。