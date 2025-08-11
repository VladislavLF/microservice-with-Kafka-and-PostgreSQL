package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"L0/internal/config"
	"L0/internal/model"

	"github.com/IBM/sarama"
)

func main() {
	cfg := config.Load()

	brokers := cfg.KafkaBrokers
	topic := cfg.KafkaTopic

	sendInterval := 3 * time.Second

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	log.Printf("Connected to Kafka at %v", brokers)
	log.Printf("Sending messages to topic '%s' every %v", topic, sendInterval)

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)

	counter := 1
	for {
		select {
		case <-sigchan:
			log.Println("Interrupted, shutting down...")
			return
		default:
			order := generateOrder(counter)
			message, err := json.Marshal(order)
			if err != nil {
				log.Printf("Error marshaling order: %v", err)
				continue
			}

			msg := &sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.StringEncoder(message),
			}

			partition, offset, err := producer.SendMessage(msg)
			if err != nil {
				log.Printf("Failed to send message: %v", err)
			} else {
				log.Printf("Sent order %s [partition %d, offset %d]", order.OrderUID, partition, offset)
			}

			counter++
			time.Sleep(sendInterval)
		}
	}
}

func generateOrder(id int) model.Order {
	now := time.Now()
	orderUID := fmt.Sprintf("order%d-%d", id, now.Unix())

	var items []model.Item
	itemCount := rand.Intn(3) + 1
	for i := 0; i < itemCount; i++ {
		items = append(items, model.Item{
			ChrtID:      rand.Intn(10000000),
			TrackNumber: fmt.Sprintf("WBILMTESTTRACK%d", id),
			Price:       rand.Intn(10000) + 100,
			Rid:         fmt.Sprintf("ab4219087a764ae0btest%d", id),
			Name:        []string{"Mascaras", "Lipstick", "Foundation", "Eyeshadow"}[rand.Intn(4)],
			Sale:        []int{0, 10, 20, 30}[rand.Intn(4)],
			Size:        "0",
			TotalPrice:  rand.Intn(500) + 50,
			NmID:        rand.Intn(1000000),
			Brand:       []string{"Vivienne Sabo", "Maybelline", "L'Oreal", "NYX"}[rand.Intn(4)],
			Status:      202,
		})
	}

	return model.Order{
		OrderUID:    orderUID,
		TrackNumber: fmt.Sprintf("WBILMTESTTRACK%d", id),
		Entry:       "WBIL",
		Delivery: model.Delivery{
			Name:    []string{"John Doe", "Jane Smith", "Alex Johnson"}[rand.Intn(3)],
			Phone:   "+972" + fmt.Sprintf("%09d", rand.Intn(1000000000)),
			Zip:     fmt.Sprintf("%07d", rand.Intn(10000000)),
			City:    []string{"Moscow", "New York", "London", "Berlin"}[rand.Intn(4)],
			Address: fmt.Sprintf("%d Main St", rand.Intn(100)+1),
			Region:  []string{"Moscow", "NY", "England", "Berlin"}[rand.Intn(4)],
			Email:   fmt.Sprintf("user%d@gmail.com", id),
		},
		Payment: model.Payment{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     []string{"USD", "EUR", "RUB"}[rand.Intn(3)],
			Provider:     "wbpay",
			Amount:       rand.Intn(10000) + 1000,
			PaymentDt:    now.Unix(),
			Bank:         []string{"alpha", "sber", "tinkoff"}[rand.Intn(3)],
			DeliveryCost: 1500,
			GoodsTotal:   rand.Intn(500) + 100,
			CustomFee:    0,
		},
		Items:             items,
		Locale:            []string{"en", "ru"}[rand.Intn(2)],
		InternalSignature: "",
		CustomerID:        fmt.Sprintf("customer%d", id),
		DeliveryService:   "meest",
		Shardkey:          fmt.Sprintf("%d", rand.Intn(10)+1),
		SmID:              rand.Intn(100),
		DateCreated:       now,
		OofShard:          "1",
	}
}
