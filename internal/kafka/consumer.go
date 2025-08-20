package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"order-svc/internal/model"
	"order-svc/internal/service"

	kgo "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	reader *kgo.Reader
	svc    *service.Service
	log    *zap.Logger
}

type Config struct {
	Brokers        []string
	Topic          string
	GroupID        string
	MinBytes       int
	MaxBytes       int
	CommitInterval time.Duration
}

func New(cfg Config, svc *service.Service, log *zap.Logger) *Consumer {
	r := kgo.NewReader(kgo.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		CommitInterval: cfg.CommitInterval,
	})
	return &Consumer{reader: r, svc: svc, log: log}
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.reader.Close()
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			c.log.Error("fetch message", zap.Error(err))
			continue
		}

		var o model.Order
		if err := json.Unmarshal(m.Value, &o); err != nil {
			c.log.Warn("invalid message json, skip", zap.Error(err))
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		if validationErrors := o.Validate(); len(validationErrors) > 0 {
			c.log.Warn("invalid order data, skip",
				zap.String("order_uid", o.OrderUID),
				zap.Any("validation_errors", validationErrors),
			)
			_ = c.reader.CommitMessages(ctx, m)
			continue
		}

		if o.DateCreated.IsZero() {
			c.log.Warn("order has zero date, using current time",
				zap.String("order_uid", o.OrderUID),
			)
			o.DateCreated = time.Now()
		}

		if err := c.svc.Put(ctx, &o); err != nil {
			if errors.Is(err, errors.New("validation error")) {
				c.log.Warn("service validation failed, skip",
					zap.String("order_uid", o.OrderUID),
					zap.Error(err),
				)
				_ = c.reader.CommitMessages(ctx, m)
				continue
			}

			c.log.Error("store order failed", zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		c.log.Info("order processed successfully",
			zap.String("order_uid", o.OrderUID),
		)

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.log.Error("commit failed", zap.Error(err))
			continue
		}
	}
}

func (c *Consumer) ValidateMessage(ctx context.Context, message []byte) ([]model.ValidationError, error) {
	var o model.Order
	if err := json.Unmarshal(message, &o); err != nil {
		return nil, err
	}

	return o.Validate(), nil
}

func (c *Consumer) ProcessMessageWithValidation(ctx context.Context, m kgo.Message) error {
	var o model.Order
	if err := json.Unmarshal(m.Value, &o); err != nil {
		return err
	}

	if validationErrors := o.Validate(); len(validationErrors) > 0 {
		c.log.Warn("order validation failed",
			zap.String("order_uid", o.OrderUID),
			zap.Any("errors", validationErrors),
		)
		return errors.New("validation failed")
	}

	if o.Payment.Amount <= 0 {
		c.log.Warn("order has invalid payment amount",
			zap.String("order_uid", o.OrderUID),
			zap.Int("amount", o.Payment.Amount),
		)
		return errors.New("invalid payment amount")
	}

	if len(o.Items) == 0 {
		c.log.Warn("order has no items",
			zap.String("order_uid", o.OrderUID),
		)
		return errors.New("no items in order")
	}

	if err := c.svc.Put(ctx, &o); err != nil {
		return err
	}

	return c.reader.CommitMessages(ctx, m)
}
