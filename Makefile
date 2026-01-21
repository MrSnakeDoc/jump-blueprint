.PHONY: start redis-start redis-stop lint lintf test vuln full-stack cov

# Load .env if it exists
-include .env
export

LDFLAGS := -X 'github.com/<your-username>/jump/internal/version.Version=$(VERSION)' \
           -X 'github.com/<your-username>/jump/internal/version.Commit=$(GIT_COMMIT)' \
           -X 'github.com/<your-username>/jump/internal/version.BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)'

BINARY_NAME=jump
# JUMP_DATA_DIR=./data
JUMP_TRUST_PROXY=false


start:
	@echo "Starting ${BINARY_NAME}..."
# 	@echo "Using data directory: ${JUMP_DATA_DIR}"
# 	@JUMP_TRUST_PROXY=${JUMP_TRUST_PROXY} JUMP_DATA_DIR=${JUMP_DATA_DIR} go run -ldflags "$(LDFLAGS)" cmd/${BINARY_NAME}/main.go
	@JUMP_TRUST_PROXY=${JUMP_TRUST_PROXY} go run -ldflags "$(LDFLAGS)" cmd/${BINARY_NAME}/main.go

start-redis:
	@echo "📦 Checking Redis..."
	@if ! docker ps >/dev/null 2>&1; then \
        echo "❌ Docker is not accessible. Try:"; \
        echo "   sudo usermod -aG docker $$USER"; \
        echo "   newgrp docker"; \
        echo "   Or run: sudo docker run -d --name jump-redis -p 6379:6379 redis:7-alpine"; \
        exit 1; \
    fi
	@if ! docker ps | grep -q jump-redis; then \
        if docker ps -a | grep -q jump-redis; then \
            echo "🔄 Starting existing Redis container..."; \
            cd dev/compose && docker compose up -d; \
        else \
            echo "🚀 Creating new Redis container..."; \
            cd dev/compose && docker compose up -d; \
        fi; \
        sleep 1; \
    fi
	@echo "✅ Redis is running"

stop-redis:
	@echo "🛑 Stopping Redis..."
	@cd dev/compose && docker compose down 2>/dev/null || true
	@echo "✅ Redis stopped"

lint:
	@echo "Running linters..."
	@golangci-lint run --config .golangci.yml

lintf:
	@echo "🔍 Running linters..."
	@golangci-lint run --config .golangci.yml --fix

vuln:
	@echo "🔒 Checking for vulnerabilities..."
	@govulncheck ./...

vulnv:
	@echo "🔒 Running vulnerability check in verbose mode..."
	@govulncheck -show verbose ./...


test:
	@echo "🧪 Running tests..."
	@go test -count=1 ./...

full-stack:
	@echo "🚀 Starting full stack tests"
	@make lint
	@make vuln
	@make test

cov:
	@echo "🧪 Running tests with coverage..."
	@go test ./... -covermode=atomic -coverprofile=coverage.out
	@go tool cover -func=coverage.out | tail -n1