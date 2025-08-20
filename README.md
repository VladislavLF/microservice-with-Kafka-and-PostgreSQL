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
```

## Переменные окружения

- ORDERS_DB_DSN, ORDERS_SERVER_ADDR, ORDERS_KAFKA_BROKERS и т.п. — перекрывают config.yaml.

## Надёжность

- Транзакции в БД, коммит offset’а в Kafka только после успешной записи.

- Ошибочные/невалидные сообщения логируются и подтверждаются (чтобы не блокировать поток).

- Ошибки не сравниваются через ==, вместо этого — errors.Is (см. код).

- Мелкозернистые блокировки в кэше (RWMutex + короткие критические секции).

## Производительность

- Кэш ускоряет повторные запросы (LRU + TTL, без утечки памяти).

- Пулы подключений к БД настраиваются.

## Технологии

Go 1.22, segmentio/kafka-go, pgx/v5, zap, viper, golang-migrate.

## Тестовые данные

Пример в make seed соответствует model.json из задания.
