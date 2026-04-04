# Personal Disk Makefile
# 简化常用开发和部署操作

.PHONY: help build run test clean dev prod stop logs fmt vet tidy

# 默认目标
help:
	@echo "Personal Disk 项目命令"
	@echo ""
	@echo "开发相关:"
	@echo "  dev       启动开发环境"
	@echo "  test      运行测试"
	@echo "  fmt       格式化代码"
	@echo "  vet       代码检查"
	@echo "  tidy      整理依赖"
	@echo ""
	@echo "部署相关:"
	@echo "  build     构建Docker镜像"
	@echo "  run       启动应用 (本地)"
	@echo "  prod      启动生产环境"
	@echo "  stop      停止服务"
	@echo "  logs      查看日志"
	@echo ""
	@echo "维护相关:"
	@echo "  clean     清理临时文件"
	@echo "  backup    备份数据"

# 开发环境
dev:
	@echo "启动开发环境..."
	./scripts/start.sh dev

# 本地运行
run:
	@echo "启动应用..."
	go run main.go

# 测试
test:
	@echo "运行测试..."
	go test -v ./...

# 代码格式化
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
vet:
	@echo "检查代码..."
	go vet ./...

# 整理依赖
tidy:
	@echo "整理依赖..."
	go mod tidy

# 构建Docker镜像
build:
	@echo "构建Docker镜像..."
	docker build -f build/Dockerfile -t personal-disk .

# 生产环境
prod:
	@echo "启动生产环境..."
	@if [ -z "$$ADMIN_PASSWORD" ]; then \
		echo "错误: 请设置 ADMIN_PASSWORD 环境变量"; \
		exit 1; \
	fi
	@if [ -z "$$DB_PASSWORD" ]; then \
		echo "错误: 请设置 DB_PASSWORD 环境变量"; \
		exit 1; \
	fi
	./scripts/start.sh prod

# 停止服务
stop:
	@echo "停止服务..."
	@if [ -f "deploy/docker-compose.yml" ]; then \
		docker-compose -f deploy/docker-compose.yml down; \
	fi
	@if [ -f "deploy/docker-compose.prod.yml" ]; then \
		docker-compose -f deploy/docker-compose.prod.yml down; \
	fi

# 查看日志
logs:
	@if docker-compose -f deploy/docker-compose.yml ps | grep -q "personal_disk"; then \
		docker-compose -f deploy/docker-compose.yml logs -f; \
	elif docker-compose -f deploy/docker-compose.prod.yml ps | grep -q "personal_disk"; then \
		docker-compose -f deploy/docker-compose.prod.yml logs -f; \
	else \
		echo "没有运行中的服务"; \
	fi

# 清理
clean:
	@echo "清理临时文件..."
	go clean
	docker system prune -f
	rm -rf uploads_test logs_test

# 备份数据
backup:
	@echo "备份数据..."
	@mkdir -p backup/$(shell date +%Y%m%d)
	@if docker ps | grep -q personal_disk_db; then \
		docker exec personal_disk_db mysqldump -u root -p personal_disk > backup/$(shell date +%Y%m%d)/database.sql; \
		echo "数据库备份完成: backup/$(shell date +%Y%m%d)/database.sql"; \
	fi
	@if [ -d "uploads" ]; then \
		tar -czf backup/$(shell date +%Y%m%d)/uploads.tar.gz uploads/; \
		echo "文件备份完成: backup/$(shell date +%Y%m%d)/uploads.tar.gz"; \
	fi

# 安装开发依赖
install:
	@echo "安装开发依赖..."
	go mod download

# 安全检查
security:
	@echo "运行安全检查..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "提示: 安装 gosec 进行安全检查: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# 完整检查
check: fmt vet test
	@echo "代码检查完成"

# 快速部署 (开发环境)
quick-dev: tidy fmt vet dev

# 完整部署 (生产环境)
deploy-prod: tidy fmt vet test build prod

# 版本信息
version:
	@echo "Personal Disk 版本信息"
	@echo "Go 版本: $(shell go version)"
	@echo "Git 提交: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "构建时间: $(shell date)"