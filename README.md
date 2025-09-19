A microservice (Go + Kafka + PostgreSQL) for receiving order data from Kafka, storing it in Postgres, caching it in memory, and serving an HTTP API + simple web page.

## Features
- Kafka subscription (topic 'orders'), incoming message validation, transactional writing to PostgreSQL.
- Idempotent processing by 'order_uid' (UPSERT).
- LRU cache with TTL and eviction. Cache warmup at startup.
- HTTP API: 'GET /order/{order_uid}' → JSON.
- Web page: 'GET /' — 'order_uid' input form.
- Migrations via 'golang-migrate'.
- Configuration from 'configs/config.yaml' + ENV ('ORDERS_' prefix).

## Quick Start
```bash
docker compose up -d --build
go run ./cmd/producer
# open http://localhost:8081/
```

## Technologies
Go 1.24.5, segmentio/kafka-go, pgx/v5, zap, viper, golang-migrate.
