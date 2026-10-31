# Market Data

Сервис сбора, хранения и отдачи рыночных данных криптовалют с поддержкой нескольких провайдеров (CoinGecko, CoinMarketCap), фоновым сбором истории через Redis-очереди и realtime-стримингом событий (SSE + WebSocket).

## Содержание

- [Возможности](#возможности)
- [Архитектура](#архитектура)
- [Требования](#требования)
- [Быстрый старт](#быстрый-старт)
- [Переменные окружения](#переменные-окружения)
- [API](#api)
- [Real-time](#real-time)
- [Фоновые задачи](#фоновые-задачи)
- [Observability (Метрики и Логи)](#observability-метрики-и-логи)
- [Разработка](#разработка)
- [Тестирование](#тестирование)
- [Миграции](#миграции)
- [Docker](#docker)
- [Структура проекта](#структура-проекта)

## Возможности

- 📊 **Мульти-провайдеры**: CoinGecko и CoinMarketCap с fan-out и частичным результатом
- 🕒 **История котировок**: сбор точек в БД с bucket-агрегацией для графиков
- ⚡ **Real-time**: SSE и WebSocket для `price.updated` / `marketcap.updated`
- ⏰ **Cron-задачи**: Asynq (Redis) или River (Postgres)
- 🔁 **Идемпотентность**: `ON CONFLICT DO NOTHING` на `(currency, minute)`
- 🧪 **Тесты**: unit + integration (testcontainers)
- 🐳 **Docker**: multi-stage сборка, docker-compose с Postgres + Redis + Prometheus + Grafana
- 📈 **Observability**: Prometheus-метрики (HTTP, DB, Провайдеры), дашборды Grafana, structured logs (`slog`)

## Архитектура

```text
                  ┌─────────────────────────────────────────┐
                  │              cmd/scheduler              │
                  │   (1 реплика, cron @every 2m)           │
                  │                                         │
                  │   1. Ставит задачу "oracle:run"         │
                  └──────────────────┬──────────────────────┘
                                     │
                                     ▼
                         ┌───────────────────────┐
                         │      Queue (Redis)    │
                         │  (Asynq / River)      │
                         └───────────┬───────────┘
                                     │
                                     │ 2. Worker забирает задачу
                                     ▼
┌─────────────────────────────────────────────────────────────┐
│                         cmd/worker                          │
│                                                             │
│  3. Выполняет oracle.Service.Execute()                      │
│     ├─▶ marketservice (fan-out по провайдерам)              │
│     │     ├─▶ coingecko (метрики через декоратор)           │
│     │     └─▶ coinmarketcap (метрики через декоратор)       │
│     ├─▶ storage/postgres (сохранение истории)               │
│     └─▶ realtime (публикация событий)                       │
│                                                             │
│  4. Отдает метрики на :9091/metrics                         │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           │ 5. События (Pub/Sub)
                           ▼
                 ┌───────────────────┐
                 │  Redis (Pub/Sub)  │
                 └─────────┬─────────┘
                           │
                           │ 6. Подписка на события
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                          cmd/api                            │
│                                                             │
│  7. Читает данные из storage/postgres для REST API          │
│  8. Отдает SSE/WS стримы из realtime (memory/redis)         │
│  9. Отдает метрики на :8080/metrics                         │
└─────────────────────────────────────────────────────────────┘
```

**Слои:**

| Слой | Пакет | Ответственность |
|------|-------|-----------------|
| Домен | `internal/domain/market` | Типы, ошибки, Timeframe |
| Use case | `internal/marketservice` | Реестр провайдеров, fan-out |
| Use case | `internal/tasks/oracle` | Сбор истории + actual |
| Инфраструктура | `internal/provider/*` | HTTP-клиенты API, декоратор метрик |
| Инфраструктура | `internal/storage/postgres` | Репозитории, метрики пула (`pgxpool`) |
| Инфраструктура | `internal/realtime/*` | События (memory/redis) |
| Планирование | `internal/scheduling` | Task, Scheduler, Registry |
| Выполнение | `internal/worker` | Handler, Registry |
| HTTP | `internal/httpapi` | Хендлеры, middleware (логи, recovery, метрики) |

## Требования

- **Go** 1.23+
- **Docker** + **Docker Compose**
- **Postgres** 16+ (или локальный)
- **Redis** 7+ (или локальный)
- **goose**: `go install github.com/pressly/goose/v3/cmd/goose@latest`
- **jq**: `sudo apt install jq`
- **websocat** (опционально): `cargo install websocat`
- **redis-cli** (опционально): `sudo apt install redis-tools`

## Быстрый старт

```bash
git clone <repo-url>
cd market-data

cp .env.example .env
# Отредактировать: DATABASE_URL, API-ключи провайдеров.

# Запуск инфраструктуры (Postgres, Redis, Prometheus, Grafana, Exporter)
make docker-up

# Миграции и сиды
make migrate-up
make seed

# Запуск API
make run-api

# В других терминалах:
make run-worker
make run-scheduler
```

**Проверка:**

- `http://localhost:8080/healthz` → `ok`
- `http://localhost:8080/v1/market-data/currencies` → список валют
- `http://localhost:8080/metrics` → **метрики Prometheus**
- `http://localhost:9090` → **Prometheus UI**
- `http://localhost:3000` → **Grafana (admin/admin)**

## Переменные окружения

Скопируй `.env.example` → `.env`.

### Database

| Переменная | Default | Описание |
|-----------|---------|----------|
| `DB_URL` | — | **Обязательно**. `postgres://user:pass@host:port/db?sslmode=disable` |
| `DB_MAX_CONNS` | `10` | Максимум соединений |
| `DB_MIN_CONNS` | `2` | Минимум соединений |
| `DB_MAX_CONN_LIFETIME` | `1h` | Время жизни соединения |
| `DB_MAX_CONN_IDLE_TIME` | `30m` | Время простоя |
| `DB_CONNECT_TIMEOUT` | `5s` | Таймаут подключения |

### Redis

| Переменная | Default | Описание |
|-----------|---------|----------|
| `REDIS_ADDR` | `localhost:6379` | **host:port**, не URL |
| `REDIS_PASSWORD` | — | Пароль |
| `REDIS_DB` | `0` | Номер БД |

### HTTP

| Переменная | Default | Описание |
|-----------|---------|----------|
| `HTTP_ADDR` | `:8080` | Адрес |
| `HTTP_READ_TIMEOUT` | `10s` | |
| `HTTP_WRITE_TIMEOUT` | `30s` | |
| `HTTP_SHUTDOWN_TIMEOUT` | `15s` | Graceful shutdown |

### Providers

| Переменная | Default | Описание |
|-----------|---------|----------|
| `COINGECKO_API_KEY` | — | API-ключ CoinGecko (опционально) |
| `COINMARKETCAP_API_KEY` | — | **Обязательно** для CMC |

### Realtime & Queue

| Переменная | Default | Описание |
|-----------|---------|----------|
| `REALTIME_BACKEND` | `memory` | `memory` или `redis` |
| `QUEUE_BACKEND` | `redis` | `redis` (Asynq) или `postgres` (River) |

### App

| Переменная | Default | Описание |
|-----------|---------|----------|
| `ENV` | `dev` | `dev` / `prod` — влияет на уровень логов |

## API

Базовый URL: `http://localhost:8080`

### `GET /healthz`

Проверка живости.

### `GET /v1/market-data/currencies`

Список поддерживаемых валют.

### `GET /v1/market-data?timeframe=24h`

Данные валют с diff за период.

### `GET /v1/market-data/price-history-chart?coinName=bitcoin&timeframe=24h`

График цены с bucket-агрегацией.

## Real-time

### SSE

```bash
curl -N http://localhost:8080/v1/stream/sse
```

### WebSocket

```bash
make ws-connect
# или
websocat ws://localhost:8080/v1/stream/ws
```

## Фоновые задачи

Сбор данных — через **scheduler** (cron) + **worker** (выполнение).

### Asynq (Redis)

```bash
QUEUE_BACKEND=redis
make run-scheduler
make run-worker
```

### River (Postgres)

```bash
QUEUE_BACKEND=postgres
make run-scheduler
make run-worker
```

## Observability (Метрики и Логи)

### 📝 Логирование

Сервис использует встроенный `log/slog` (структурированное логирование в JSON).

- **Dev (`ENV=dev`)**: уровень `DEBUG` (видны SQL-запросы, работа воркеров).
- **Prod (`ENV=prod`)**: уровень `INFO` (только важные события).

**Пример лога:**
```json
{"time":"2026-10-31T10:00:00Z","level":"INFO","msg":"request","method":"GET","path":"/v1/market-data","status":200,"duration_ms":12}
```

### 📊 Метрики (Prometheus)

Метрики доступны на эндпоинте **`/metrics`** (порт 8080 для API и 9091 для Worker).

**Собираемые метрики:**
- **HTTP**: `marketdata_http_requests_total`, `marketdata_http_request_duration_seconds` (RPS, latency, статусы).
- **Провайдеры (Декоратор)**: `marketdata_provider_requests_total`, `marketdata_provider_request_duration_seconds` (запросы к CoinGecko/CMC, ошибки).
- **База данных (Go-приложение)**: `pgx_pool_acquired_connections`, `pgx_pool_idle_connections` (метрики пула `pgxpool`).
- **База данных (Postgres Exporter)**: `pg_stat_database_tup_inserted`, `pg_stat_activity_count` (нагрузка на БД).
- **Oracle**: `marketdata_oracle_runs_total`, `marketdata_oracle_saved_total` (статистика сбора данных).

**Проверка метрик вручную:**
```bash
# Метрики API
curl -s localhost:8080/metrics | grep marketdata_

# Метрики Worker
curl -s localhost:9091/metrics | grep marketdata_provider

# Метрики БД
curl -s localhost:9187/metrics | grep pg_stat_database
```

### 📈 Визуализация (Grafana)

Для визуализации используется **Grafana** с предустановленными дашбордами (`deploy/grafana/dashboards/`).

1. Открой **http://localhost:3000** (admin/admin).
2. Перейди в **Dashboards → Market Data**.
3. Доступны панели:
   - **HTTP RPS & p95 Latency**
   - **Provider Requests & Errors** (по провайдерам)
   - **DB Pool** (занятые/свободные соединения)
   - **PostgreSQL** (вставки, транзакции, сессии, кэш)

> 💡 **Совет:** если панели пустые, проверь, что в выпадающих списках (Datasource, Database) выбраны `Prometheus` и `market_data_go`.

## Разработка

```bash
make run-api
make run-worker
make run-scheduler

make build
make tidy
make vet
make fmt
make lint
make clean
make help
```

## Тестирование

```bash
make test
make test-race
make test-cover

make test-integration
```

## Миграции

```bash
make migrate-up
make migrate-down
make migrate-status
make migrate-create NAME=add_index
```

## Docker

### Инфраструктура

```bash
make docker-up      # Postgres, Redis, Prometheus, Grafana, Exporter
make docker-down
make docker-down-v
make docker-logs
```

### Приложение

```bash
make docker-build
make docker-run-api
make docker-run-worker
make docker-run-scheduler
```

## Структура проекта

```text
market-data/
├── cmd/
│   ├── api/          # HTTP-сервер, метрики
│   ├── worker/       # выполнение задач, метрики
│   ├── scheduler/    # cron
│   ├── migrate/      # миграции
│   └── seed/         # сиды
├── internal/
│   ├── domain/
│   │   └── market/             # типы, ошибки, Timeframe
│   ├── marketservice/          # реестр провайдеров, fan-out
│   ├── provider/
│   │   ├── coingecko/
│   │   ├── coinmarketcap/
│   │   └── metrics/            # Декоратор для метрик провайдеров
│   ├── storage/
│   │   ├── postgres/
│   │   │   ├── migrations/
│   │   │   ├── currency_repo.go
│   │   │   ├── market_data_history_repo.go
│   │   │   └── options.go      # WithMetrics()
│   │   └── ...
│   ├── realtime/
│   │   ├── realtime.go
│   │   ├── memory/
│   │   └── redis/
│   ├── tasks/
│   │   └── oracle/
│   ├── scheduling/
│   ├── worker/
│   ├── observability/          # Метрики (Prometheus)
│   └── httpapi/
│       ├── v1/
│       │   ├── marketdata/
│       │   └── stream/
│       ├── debug/
│       └── router.go
├── deploy/
│   ├── prometheus.yml          # Конфиг Prometheus
│   └── grafana/
│       ├── dashboards/         # JSON дашборды
│       └── provisioning/       # Автоподключение
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## Лицензия

MIT