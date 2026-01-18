.PHONY: build run test clean dev lint docker

# 变量
BINARY_NAME=agentbox
BUILD_DIR=bin
GO=go
GOFLAGS=-ldflags="-s -w"

# 默认目标
all: build

# 构建
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/agentbox

# 开发模式运行
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# 开发模式 (热重载，需要安装 air)
dev:
	@command -v air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

# 测试
test:
	$(GO) test -v ./...

# 测试覆盖率
test-coverage:
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

# 代码检查
lint:
	@command -v golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# 依赖整理
tidy:
	$(GO) mod tidy

# Docker 构建
docker:
	docker build -t agentbox:latest -f docker/Dockerfile .

# 帮助
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  run            - Build and run"
	@echo "  dev            - Run with hot reload (requires air)"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  tidy           - Run go mod tidy"
	@echo "  docker         - Build Docker image"
