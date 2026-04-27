.PHONY: test build run migrate docker-up docker-down push tidy lint web-dev web-build all

GO          ?= go
MIGRATE     ?= migrate
POSTGRES_URL ?= postgres://chess:chess@localhost:5432/chess?sslmode=disable
MIGRATIONS  := ./migrations

all: test build

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./... -count=1

test-race:
	CGO_ENABLED=1 $(GO) test ./... -race -count=1

build:
	$(GO) build -o bin/api ./cmd/api

run:
	$(GO) run ./cmd/api

migrate:
	$(MIGRATE) -path $(MIGRATIONS) -database "$(POSTGRES_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS) -database "$(POSTGRES_URL)" down 1

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

logs:
	docker compose logs -f api

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

lint:
	$(GO) vet ./...

push:
	git add -A
	git commit -m "$(msg)" || true
	git push origin main
