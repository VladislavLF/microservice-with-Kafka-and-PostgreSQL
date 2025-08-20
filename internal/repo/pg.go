package repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"order-svc/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrValidation = errors.New("validation error")

type PG struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *PG { return &PG{db: db} }

func (p *PG) UpsertOrder(ctx context.Context, o *model.Order) error {
	if validationErrors := o.Validate(); len(validationErrors) > 0 {
		return fmt.Errorf("%w: %v", ErrValidation, validationErrors)
	}

	raw, err := json.Marshal(o)
	if err != nil {
		return fmt.Errorf("marshal order: %w", err)
	}

	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
		INSERT INTO orders(order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, raw_json)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (order_uid) DO UPDATE SET
		  track_number=EXCLUDED.track_number, entry=EXCLUDED.entry, locale=EXCLUDED.locale,
		  internal_signature=EXCLUDED.internal_signature, customer_id=EXCLUDED.customer_id,
		  delivery_service=EXCLUDED.delivery_service, shardkey=EXCLUDED.shardkey, sm_id=EXCLUDED.sm_id,
		  date_created=EXCLUDED.date_created, oof_shard=EXCLUDED.oof_shard, raw_json=EXCLUDED.raw_json
	`, o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature, o.CustomerID, o.DeliveryService, o.ShardKey, o.SmID, o.DateCreated, o.OofShard, raw)
	if err != nil {
		return fmt.Errorf("upsert order: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO deliveries(order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO UPDATE SET
		  name=EXCLUDED.name, phone=EXCLUDED.phone, zip=EXCLUDED.zip, city=EXCLUDED.city,
		  address=EXCLUDED.address, region=EXCLUDED.region, email=EXCLUDED.email
	`, o.OrderUID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.ZIP, o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email)
	if err != nil {
		return fmt.Errorf("upsert delivery: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payments(order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (order_uid) DO UPDATE SET
		  transaction=EXCLUDED.transaction, request_id=EXCLUDED.request_id, currency=EXCLUDED.currency, provider=EXCLUDED.provider,
		  amount=EXCLUDED.amount, payment_dt=EXCLUDED.payment_dt, bank=EXCLUDED.bank, delivery_cost=EXCLUDED.delivery_cost,
		  goods_total=EXCLUDED.goods_total, custom_fee=EXCLUDED.custom_fee
	`, o.OrderUID, o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency, o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDT, o.Payment.Bank, o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("upsert payment: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM items WHERE order_uid=$1`, o.OrderUID)
	if err != nil {
		return fmt.Errorf("clear items: %w", err)
	}
	for _, it := range o.Items {
		if itemErrors := it.Validate(); len(itemErrors) > 0 {
			return fmt.Errorf("%w: item validation failed: %v", ErrValidation, itemErrors)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO items(order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		`, o.OrderUID, it.ChrtID, it.TrackNumber, it.Price, it.RID, it.Name, it.Sale, it.Size, it.TotalPrice, it.NMID, it.Brand, it.Status)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (p *PG) ValidateOrder(ctx context.Context, o *model.Order) error {
	if validationErrors := o.Validate(); len(validationErrors) > 0 {
		return fmt.Errorf("%w: %v", ErrValidation, validationErrors)
	}
	return nil
}

func (p *PG) UpsertOrderIfValid(ctx context.Context, o *model.Order) error {
	if err := p.ValidateOrder(ctx, o); err != nil {
		return err
	}
	return p.UpsertOrder(ctx, o)
}

type OrderFull struct {
	Order model.Order
}

func (p *PG) GetOrder(ctx context.Context, uid string) (*model.Order, error) {
	row := p.db.QueryRow(ctx, `
		SELECT raw_json
		FROM orders
		WHERE order_uid=$1
	`, uid)
	var raw []byte
	if err := row.Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		return nil, fmt.Errorf("scan order: %w", err)
	}
	var o model.Order
	if err := json.Unmarshal(raw, &o); err != nil {
		return nil, fmt.Errorf("unmarshal raw_json: %w", err)
	}

	if validationErrors := o.Validate(); len(validationErrors) > 0 {
		return nil, fmt.Errorf("corrupted data in DB: %w: %v", ErrValidation, validationErrors)
	}

	return &o, nil
}

func (p *PG) LoadRecent(ctx context.Context, limit int) ([]*model.Order, error) {
	rows, err := p.db.Query(ctx, `
		SELECT raw_json
		FROM orders
		ORDER BY date_created DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("load recent: %w", err)
	}
	defer rows.Close()

	var res []*model.Order
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		var o model.Order
		if err := json.Unmarshal(raw, &o); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}

		if validationErrors := o.Validate(); len(validationErrors) > 0 {
			fmt.Printf("Warning: skipping invalid order %s: %v\n", o.OrderUID, validationErrors)
			continue
		}

		res = append(res, &o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return res, nil
}

func (p *PG) LoadRecentValid(ctx context.Context, limit int) ([]*model.Order, error) {
	orders, err := p.LoadRecent(ctx, limit)
	if err != nil {
		return nil, err
	}

	var validOrders []*model.Order
	for _, order := range orders {
		if order.IsValid() {
			validOrders = append(validOrders, order)
		}
	}

	return validOrders, nil
}
