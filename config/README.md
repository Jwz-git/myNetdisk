# 配置文件说明

此目录包含 Personal Disk 项目的所有配置文件。

## 配置文件列表

| 文件 | 用途 | 说明 |
|------|------|------|
| `config.go` | 配置结构体定义 | 定义所有配置项的结构和加载逻辑 |
| `init.go` | 配置初始化 | 配置系统的初始化和环境检测 |
| `default.yml` | 默认配置 | 基础配置文件，包含所有配置项的默认值 |
| `development.yml` | 开发环境配置 | 本地开发使用的配置 |
| `test.yml` | 测试环境配置 | 自动化测试使用的配置 |
| `production.yml` | 生产环境配置 | 线上部署使用的配置 |

## 配置加载优先级

1. **环境变量** (最高优先级)
2. **环境配置文件** (如 `development.yml`)
3. **默认配置文件** (`default.yml`)

## 环境切换

通过设置 `APP_ENV` 环境变量来切换配置：

```bash
# 开发环境 (默认)
export APP_ENV=development

# 测试环境
export APP_ENV=test

# 生产环境
export APP_ENV=production
```

## 添加新配置项

1. 在 `config.go` 中添加相应的结构体字段
2. 在各环境配置文件中添加配置值
3. 必要时在 `overrideWithEnv()` 中添加环境变量支持

详细信息请参考 [配置文档](../docs/CONFIG.md)。