# microservice-with-Kafka-and-PostgreSQL

Микросервис (Go + Kafka + PostgreSQL) для приёма данных заказов из Kafka, сохранения в Postgres, кэширования в памяти и выдачи HTTP API + простой веб-страницы.

## Возможности
- Подписка на Kafka (topic `orders`), валидация входящих сообщений, транзакционная запись в PostgreSQL.
- Идемпотентная обработка по `order_uid` (UPSERT).
- LRU-кэш с TTL и вытеснением. Прогрев кэша на старте.
- HTTP API: `GET /order/{order_uid}` → JSON.
- Web-страница: `GET /` — форма ввода `order_uid`.
- Миграции через `golang-migrate`.
- Конфигурация из `configs/config.yaml` + ENV (префикс `ORDERS_`).

## Быстрый старт
```bash
docker compose up -d --build
go run ./cmd/producer
# открыть http://localhost:8081/
