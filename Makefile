.PHONY: run build docker docker-down migrate migrate-down migrate-create lint tidy

BINARY     := server
MIGRATIONS := migrations
DB_DSN     ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# ── Dev ───────────────────────────────────────────────────────────────────────

run:
	go run ./cmd/api

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) ./cmd/api

# ── Docker ────────────────────────────────────────────────────────────────────

docker:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

# ── Migrations ────────────────────────────────────────────────────────────────

migrate:
	goose -dir $(MIGRATIONS) postgres "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS) postgres "$(DB_DSN)" down

migrate-status:
	goose -dir $(MIGRATIONS) postgres "$(DB_DSN)" status

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir $(MIGRATIONS) create $$name sql

# ── Code quality ──────────────────────────────────────────────────────────────

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
