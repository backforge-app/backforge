SHELL := /bin/bash
.DEFAULT_GOAL := help

APP_NAME := waybill
COMPOSE := docker compose
COMPOSE_FILE := deploy/docker-compose.yml
DOCKERFILE := deploy/Dockerfile
ENV_FILE := .env
GO := go
SERVER_PKG := ./cmd/server
GOLANGCI_LINT := golangci-lint

.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Docker:"
	@echo "  make up             Start all services"
	@echo "  make up-build       Build images and start all services"
	@echo "  make up-detach      Start all services in detached mode"
	@echo "  make down           Stop all services"
	@echo "  make down-volumes   Stop all services and remove volumes"
	@echo "  make logs           Follow logs"
	@echo "  make ps             Show containers"
	@echo ""
	@echo "Migrations:"
	@echo "  make migrate-up     Run goose migrations (via migrator service)"
	@echo "  make migrate-down   Rollback 1 migration (via migrator service)"
	@echo ""
	@echo "Go:"
	@echo "  make tidy           go mod tidy"
	@echo "  make fmt            gofmt"
	@echo "  make test           Run all tests"
	@echo "  make test-integration Run integration tests"
	@echo "  make build          Build server locally into ./bin/"
	@echo ""
	@echo "Lint/format:"
	@echo "  make lint           Run golangci-lint"
	@echo "  make lint-fix       Run golangci-lint with --fix"

# Docker
.PHONY: up
up:
	$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up

.PHONY: up-build
up-build:
	docker build -f $(DOCKERFILE) -t $(APP_NAME) .
	$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up --build

.PHONY: up-detach
up-detach:
	$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) up -d

.PHONY: down
down:
	$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down

.PHONY: down-volumes
down-volumes:
	$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) down -v

.PHONY: logs
logs:
	$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) logs -f --tail=200

.PHONY: ps
ps:
	$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) ps

# Migrations
.PHONY: migrate-up
migrate-up:
	$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) run --rm migrator up

.PHONY: migrate-down
migrate-down:
	$(COMPOSE) --env-file $(ENV_FILE) -f $(COMPOSE_FILE) run --rm migrator down

# Go
.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: fmt
fmt:
	gofmt -w .

.PHONY: test
test:
	$(GO) test ./...

.PHONY: test-integration
test-integration:
	$(GO) test -v -tags=integration ./tests/integration...

.PHONY: build
build:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o bin/$(APP_NAME) $(SERVER_PKG)

# Lint
.PHONY: lint
lint:
	$(GOLANGCI_LINT) run ./...

.PHONY: lint-fix
lint-fix:
	$(GOLANGCI_LINT) run --fix ./...