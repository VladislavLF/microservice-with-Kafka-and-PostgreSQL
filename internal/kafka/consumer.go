package kafka

import (
	"context"
	"encoding/json"
	"log"

	"L0/internal/model"

	"github.com/IBM/sarama"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	handler  sarama.ConsumerGroupHandler
}

type OrderHandler interface {
	SaveOrder(ctx context.Context, order *model.Order) error
}

type Cache interface {
	Add(order model.Order)
}

func NewConsumer(brokers []string, topic string, db OrderHandler, cache Cache) *Consumer {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	log.Printf("Connecting to Kafka brokers: %v", brokers)

	consumer, err := sarama.NewConsumerGroup(brokers, "L0", config)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	handler := &kafkaHandler{db: db, cache: cache}
	return &Consumer{consumer: consumer, handler: handler}
}

func (c *Consumer) Start() {
	ctx := context.Background()
	for {
		if err := c.consumer.Consume(ctx, []string{"orders"}, c.handler); err != nil {
			log.Printf("Kafka consumer error: %v", err)
		}
	}
}

func (c *Consumer) Stop() {
	c.consumer.Close()
}

type kafkaHandler struct {
	db    OrderHandler
	cache Cache
}

func (h *kafkaHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *kafkaHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h *kafkaHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var order model.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("Failed to parse Kafka message: %v", err)
			continue
		}

		if err := h.db.SaveOrder(context.Background(), &order); err != nil {
			log.Printf("Failed to save order: %v", err)
			continue
		}

		h.cache.Add(order)

		session.MarkMessage(msg, "")
	}
	return nil
}
