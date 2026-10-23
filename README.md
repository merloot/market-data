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
- 🐳 **Docker**: multi-stage сборка, docker-compose с Postgres + Redis
- 📈 **Observability**: structured logs (`slog`)

## Архитектура

┌─────────────┐ ┌──────────────┐ ┌─────────────────┐
│ cmd/api │─────▶│ marketservice│─────▶│ provider/* │
│ HTTP + SSE │ │ (fan-out) │ │ coingecko │
│ WS │ └──────┬───────┘ │ coinmarketcap │
└─────────────┘ │ └─────────────────┘
│
┌───────▼───────┐
│ storage │
│ postgres │
└───────┬───────┘
│
┌─────────────┐ ┌──────▼───────┐ ┌─────────────────┐
│ cmd/worker │◀─────│ realtime │◀─────│ cmd/scheduler │
│ Asynq/River │ │ memory/redis │ │ cron (2m) │
└─────────────┘ └──────────────┘ └─────────────────┘


**Слои:**

| Слой | Пакет | Ответственность |
|------|-------|-----------------|
| Домен | `internal/domain/market` | Типы, ошибки, Timeframe |
| Use case | `internal/marketservice` | Реестр провайдеров, fan-out |
| Use case | `internal/tasks/oracle` | Сбор истории + actual |
| Инфраструктура | `internal/provider/*` | HTTP-клиенты API |
| Инфраструктура | `internal/storage/postgres` | Репозитории |
| Инфраструктура | `internal/realtime/*` | События (memory/redis) |
| Планирование | `internal/scheduling` | Task, Scheduler, Registry |
| Выполнение | `internal/worker` | Handler, Registry |
| HTTP | `internal/httpapi` | Хендлеры, middleware |

## Требования

- **Go** 1.23+
- **Docker** + **Docker Compose**
- **Postgres** 16+ (через Docker)
- **Redis** 7+ (через Docker)
- **goose**: `go install github.com/pressly/goose/v3/cmd/goose@latest`
- **jq**: `sudo apt install jq`
- **websocat** (опционально): `cargo install websocat`
- **redis-cli** (опционально): `sudo apt install redis-tools`

## Быстрый старт

```bash
git clone <repo-url>
cd market-data

cp .env.example .env
# Отредактировать: API-ключи провайдеров.

make docker-up
make migrate-up
make seed
make run-api

# В других терминалах:
make run-worker
make run-scheduler

# Проверка:
make curl-currencies
make curl-sse
```

**Проверка:**

- `http://localhost:8080/healthz` → `ok`
- `http://localhost:8080/v1/market-data/currencies` → список валют
- `http://localhost:8080/v1/market-data?timeframe=24h` → данные с diff

## Переменные окружения

Скопируй `.env.example` → `.env`.

### Database

| Переменная | Default | Описание |
|-----------|---------|----------|
| `DATABASE_URL` | — | **Обязательно**. `postgres://user:pass@host:port/db?sslmode=disable` |
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

### Realtime

| Переменная | Default | Описание |
|-----------|---------|----------|
| `REALTIME_BACKEND` | `memory` | `memory` или `redis` |

### Queue

| Переменная | Default | Описание |
|-----------|---------|----------|
| `QUEUE_BACKEND` | `redis` | `redis` (Asynq) или `postgres` (River) |

### App

| Переменная | Default | Описание |
|-----------|---------|----------|
| `ENV` | `dev` | `dev` / `prod` — влияет на уровень логов |

## API

Базовый URL: `http://localhost:8080`

### `GET /healthz`

Проверка живости.

```bash
curl http://localhost:8080/healthz
# ok
```

### `GET /v1/market-data/currencies`

Список поддерживаемых валют.

```bash
curl http://localhost:8080/v1/market-data/currencies | jq .
```

**Ответ:**

```json
{
  "data": [
    { "coinId": "bitcoin", "coinName": "BTC" },
    { "coinId": "ethereum", "coinName": "ETH" }
  ]
}
```

### `GET /v1/market-data?timeframe=24h`

Данные валют с diff за период.

**Query:**

| Параметр | Default | Описание |
|----------|---------|----------|
| `timeframe` | `24h` | `1h`, `24h`, `7d`, `30d`, `90d`, `365d`, `custom` |
| `from` | — | RFC3339 (для `custom`) |

**Ответ:**

```json
{
  "data": [
    {
      "coinName": "BTC",
      "logoUrl": "https://...",
      "price": 50000,
      "marketCap": 1e12,
      "circulatingSupply": 19000000,
      "totalSupply": 21000000,
      "priceDiff": 1000,
      "marketCapDiff": 1e10,
      "lastUpdated": "2026-10-23T06:40:00Z"
    }
  ]
}
```

### `GET /v1/market-data/price-history-chart?coinName=bitcoin&timeframe=24h`

График цены с bucket-агрегацией.

**Query:**

| Параметр | Default | Описание |
|----------|---------|----------|
| `coinName` | — | **Обязательно** |
| `timeframe` | `24h` | `1h`, `24h`, `7d`, `30d`, `90d`, `365d`, `custom` |
| `from` | — | RFC3339 (для `custom`) |

**Ответ:**

```json
{
  "prices": [
    [1729670400000, 50000],
    [1729670700000, 50100]
  ]
}
```

## Real-time

### SSE

```bash
curl -N http://localhost:8080/v1/stream/sse
```

**Формат:**

```
event: connected
data: {"status":"ok"}

event: price.updated
data: {"CoinName":"BTC","Price":50000,...}

event: marketcap.updated
data: {"CoinName":"ETH","MarketCap":400000000000,...}
```

**JavaScript:**

```javascript
const es = new EventSource("http://localhost:8080/v1/stream/sse");
es.addEventListener("price.updated", (e) => console.log(JSON.parse(e.data)));
```

### WebSocket

```bash
make ws-connect
# или
websocat ws://localhost:8080/v1/stream/ws
```

**Формат:**

```json
{"event":"connected","data":{"status":"ok"}}
{"event":"price.updated","data":{"CoinName":"BTC","Price":50000}}
```

**JavaScript:**

```javascript
const ws = new WebSocket("ws://localhost:8080/v1/stream/ws");
ws.onmessage = (e) => console.log(JSON.parse(e.data));
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

### Задачи

| Type | Interval | Что делает |
|------|----------|-----------|
| `market-data-oracle` | `@every 2m` | Сбор истории + actual → БД → события |

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
open coverage.html

make test-integration
```

**Требования для интеграционных:** Docker должен быть запущен.

**Если `permission denied`:**

```bash
sudo usermod -aG docker $USER
newgrp docker
```

## Миграции

```bash
make migrate-up
make migrate-down
make migrate-status
make migrate-create NAME=add_index
```

**Требуется `DATABASE_URL`** в `.env` или окружении.

## Docker

### Инфраструктура

```bash
make docker-up
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

**Multi-stage Dockerfile** — статический бинарник, non-root пользователь, минимальный размер.

## Структура проекта

```
market-data/
├── cmd/
│   ├── api/          # HTTP-сервер
│   ├── worker/       # выполнение задач
│   ├── scheduler/    # cron
│   ├── migrate/      # миграции
│   └── seed/         # сиды
├── internal/
│   ├── domain/
│   │   └── market/             # типы, ошибки, Timeframe
│   ├── marketservice/          # реестр провайдеров, fan-out
│   ├── provider/
│   │   ├── coingecko/
│   │   └── coinmarketcap/
│   ├── storage/
│   │   └── postgres/
│   │       ├── migrations/
│   │       ├── currency_repo.go
│   │       └── market_data_history_repo.go
│   ├── realtime/
│   │   ├── realtime.go
│   │   ├── memory/
│   │   └── redis/
│   ├── tasks/
│   │   └── oracle/
│   ├── scheduling/
│   ├── worker/
│   └── httpapi/
│       ├── v1/
│       │   ├── marketdata/
│       │   └── stream/
│       ├── debug/
│       └── router.go
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## Лицензия

MIT