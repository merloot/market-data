# ============================================================
# Makefile — market-data
# ============================================================

.PHONY: help
help: ## Показать все команды
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'

# ============================================================
# Разработка
# ============================================================

.PHONY: run-api run-worker run-scheduler
run-api: ## Запустить API
	go run ./cmd/api

run-worker: ## Запустить worker
	go run ./cmd/worker

run-scheduler: ## Запустить scheduler
	go run ./cmd/scheduler

# ============================================================
# Сборка
# ============================================================

.PHONY: build build-api build-worker build-scheduler build-migrate build-seed
BIN_DIR := bin

build: build-api build-worker build-scheduler build-migrate build-seed ## Собрать все бинарники

build-api: ## Собрать api
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/api ./cmd/api

build-worker: ## Собрать worker
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/worker ./cmd/worker

build-scheduler: ## Собрать scheduler
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/scheduler ./cmd/scheduler

build-migrate: ## Собрать migrate
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/migrate ./cmd/migrate

build-seed: ## Собрать seed
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/seed ./cmd/seed

# ============================================================
# Тесты
# ============================================================

.PHONY: test test-race test-cover test-integration
test: ## Запустить все тесты
	go test ./... -count=1

test-race: ## Тесты с race detector
	go test ./... -race -count=1

test-cover: ## Тесты с покрытием (HTML)
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html
	@echo "Открой coverage.html в браузере"

test-integration: ## Только интеграционные тесты (testcontainers)
	go test ./internal/storage/... ./internal/realtime/... -count=1 -v

# ============================================================
# Миграции (goose)
# ============================================================

.PHONY: migrate-up migrate-down migrate-status migrate-version migrate-create
MIGRATIONS_DIR := internal/storage/postgres/migrations

migrate-up: ## Применить все миграции
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down: ## Откатить последнюю миграцию
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status: ## Статус миграций
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-version: ## Текущая версия
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" version

migrate-create: ## Создать миграцию: make migrate-create NAME=add_foo
	@if [ -z "$(NAME)" ]; then echo "usage: make migrate-create NAME=add_foo"; exit 1; fi
	goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

# ============================================================
# Сиды
# ============================================================

.PHONY: seed
seed: ## Засеять БД (идемпотентно)
	go run ./cmd/seed

# ============================================================
# Redis (проверка событий)
# ============================================================

.PHONY: redis-price redis-marketcap
redis-price: ## Подписаться на price.updated
	redis-cli SUBSCRIBE price.updated

redis-marketcap: ## Подписаться на marketcap.updated
	redis-cli SUBSCRIBE marketcap.updated

# ============================================================
# HTTP (проверка API)
# ============================================================

.PHONY: curl-health curl-currencies curl-info curl-chart curl-sse
API_URL := http://localhost:8080

curl-health: ## Проверить /healthz
	curl -s $(API_URL)/healthz

curl-currencies: ## Список валют
	curl -s $(API_URL)/v1/market-data/currencies | jq .

curl-info: ## Информация по валютам
	curl -s "$(API_URL)/v1/market-data?timeframe=24h" | jq .

curl-chart: ## График за 24h (BTC)
	curl -s "$(API_URL)/v1/market-data/price-history-chart?coinName=bitcoin&timeframe=24h" | jq .

curl-sse: ## SSE-поток (Ctrl+C для выхода)
	curl -N $(API_URL)/v1/stream/sse

# ============================================================
# Утилиты
# ============================================================

.PHONY: tidy lint vet fmt clean
tidy: ## go mod tidy + verify
	go mod tidy
	go mod verify

lint: ## go vet + staticcheck
	go vet ./...
	@which staticcheck > /dev/null && staticcheck ./... || echo "staticcheck not installed"

vet: ## go vet
	go vet ./...

fmt: ## gofmt + goimports
	gofmt -w .
	@which goimports > /dev/null && goimports -w . || echo "goimports not installed"

clean: ## Удалить артефакты
	rm -rf $(BIN_DIR) coverage.out coverage.html

# ============================================================
# Docker (если нужно)
# ============================================================

.PHONY: docker-up docker-down docker-logs
docker-up: ## Поднять Postgres + Redis
	docker compose up -d

docker-down: ## Остановить
	docker compose down

docker-logs: ## Логи контейнеров
	docker compose logs -f

# ============================================================
# Полный цикл
# ============================================================

.PHONY: dev
dev: ## Полный цикл: миграции + сиды + запуск API
	$(MAKE) migrate-up
	$(MAKE) seed
	$(MAKE) run-api

.PHONY: check
check: tidy vet test-race ## Проверка перед коммитом
	@echo "✅ Всё чисто"


	# ============================================================
# Docker — инфраструктура
# ============================================================

.PHONY: docker-up docker-down docker-logs docker-ps docker-restart
docker-up: ## Поднять Postgres + Redis
	docker compose up -d
	@echo "Ждём healthcheck..."
	@sleep 3
	@docker compose ps

docker-down: ## Остановить инфраструктуру
	docker compose down

docker-down-v: ## Остановить + удалить volumes (ОСТОРОЖНО!)
	docker compose down -v

docker-logs: ## Логи контейнеров
	docker compose logs -f

docker-ps: ## Статус контейнеров
	docker compose ps

docker-restart: ## Перезапустить
	docker compose restart

# ============================================================
# Docker — приложение
# ============================================================

.PHONY: docker-build docker-run-api docker-run-worker docker-run-scheduler
docker-build: ## Собрать образ приложения
	docker build -t market-data:latest .

docker-run-api: docker-build ## Запустить API в Docker
	docker run --rm --network host \
		-e DATABASE_URL="$(DATABASE_URL)" \
		-e REDIS_ADDR="$(REDIS_ADDR)" \
		market-data:latest /usr/local/bin/api

docker-run-worker: docker-build ## Запустить worker в Docker
	docker run --rm --network host \
		-e DATABASE_URL="$(DATABASE_URL)" \
		-e REDIS_ADDR="$(REDIS_ADDR)" \
		market-data:latest /usr/local/bin/worker

docker-run-scheduler: docker-build ## Запустить scheduler в Docker
	docker run --rm --network host \
		-e DATABASE_URL="$(DATABASE_URL)" \
		-e REDIS_ADDR="$(REDIS_ADDR)" \
		market-data:latest /usr/local/bin/scheduler

# ============================================================
# Docker — миграции / сиды
# ============================================================

.PHONY: docker-migrate docker-seed
docker-migrate: ## Миграции в Docker
	goose -dir internal/storage/postgres/migrations postgres \
		"postgres://user:pass@localhost:5432/market_data_go?sslmode=disable" up

docker-seed: ## Сиды в Docker
	docker run --rm --network host \
		-e DATABASE_URL="$(DATABASE_URL)" \
		market-data:latest /usr/local/bin/seed