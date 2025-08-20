CREATE TABLE IF NOT EXISTS orders (
  order_uid TEXT PRIMARY KEY,
  track_number TEXT NOT NULL,
  entry TEXT NOT NULL,
  locale TEXT NOT NULL,
  internal_signature TEXT,
  customer_id TEXT NOT NULL,
  delivery_service TEXT NOT NULL,
  shardkey TEXT NOT NULL,
  sm_id INTEGER NOT NULL,
  date_created TIMESTAMPTZ NOT NULL,
  oof_shard TEXT NOT NULL,
  raw_json JSONB NOT NULL,                -- для простого восстановления/отладки, но читаем из нормализованных
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS deliveries (
  order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
  name TEXT NOT NULL,
  phone TEXT NOT NULL,
  zip TEXT NOT NULL,
  city TEXT NOT NULL,
  address TEXT NOT NULL,
  region TEXT NOT NULL,
  email TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
  order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
  transaction TEXT NOT NULL,
  request_id TEXT,
  currency TEXT NOT NULL,
  provider TEXT NOT NULL,
  amount INTEGER NOT NULL CHECK (amount >= 0),
  payment_dt BIGINT NOT NULL,
  bank TEXT NOT NULL,
  delivery_cost INTEGER NOT NULL CHECK (delivery_cost >= 0),
  goods_total INTEGER NOT NULL CHECK (goods_total >= 0),
  custom_fee INTEGER NOT NULL CHECK (custom_fee >= 0)
);

CREATE TABLE IF NOT EXISTS items (
  id BIGSERIAL PRIMARY KEY,
  order_uid TEXT NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
  chrt_id BIGINT NOT NULL,
  track_number TEXT NOT NULL,
  price INTEGER NOT NULL CHECK (price >= 0),
  rid TEXT NOT NULL,
  name TEXT NOT NULL,
  sale INTEGER NOT NULL CHECK (sale >= 0),
  size TEXT NOT NULL,
  total_price INTEGER NOT NULL CHECK (total_price >= 0),
  nm_id BIGINT NOT NULL,
  brand TEXT NOT NULL,
  status INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS items_order_uid_idx ON items(order_uid);
