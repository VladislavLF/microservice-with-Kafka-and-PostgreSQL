package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	kgo "github.com/segmentio/kafka-go"
)

func main() {
	w := &kgo.Writer{
		Addr:         kgo.TCP("localhost:29092"),
		Topic:        "orders",
		RequiredAcks: kgo.RequireOne,
	}
	defer w.Close()

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 10; i++ {
		msg := generateRandomOrder()
		b, _ := json.Marshal(msg)

		if err := w.WriteMessages(context.Background(), kgo.Message{Value: b}); err != nil {
			log.Printf("Failed to send message %d: %v", i+1, err)
			continue
		}

		log.Printf("Sent message %d: %s", i+1, msg["order_uid"])
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("Finished sending messages")
}

func generateRandomOrder() map[string]any {
	orderUID := fmt.Sprintf("order_%d_%d", time.Now().Unix(), rand.Intn(1000))
	trackNumber := fmt.Sprintf("WBIL%d", rand.Intn(100000))

	names := []string{"John Doe", "Jane Smith", "Alex Johnson", "Maria Garcia", "David Brown"}
	emails := []string{"test1@gmail.com", "test2@yahoo.com", "test3@mail.ru", "test4@outlook.com"}
	cities := []string{"Moscow", "New York", "London", "Tokyo", "Berlin"}
	regions := []string{"Moscow Oblast", "NY State", "Greater London", "Kanto", "Brandenburg"}

	currencies := []string{"USD", "EUR", "RUB", "GBP", "JPY"}
	providers := []string{"wbpay", "paypal", "stripe", "yandex", "qiwi"}
	banks := []string{"alpha", "sber", "tinkoff", "gazprom", "raiffeisen"}

	items := []map[string]any{
		{
			"chrt_id":      rand.Int63n(10000000),
			"track_number": trackNumber,
			"price":        rand.Intn(1000) + 100,
			"rid":          fmt.Sprintf("rid_%d", rand.Intn(10000)),
			"name":         "Random Item",
			"sale":         rand.Intn(50),
			"size":         fmt.Sprintf("%d", rand.Intn(5)),
			"total_price":  rand.Intn(500) + 50,
			"nm_id":        rand.Int63n(1000000),
			"brand":        "Test Brand",
			"status":       202,
		},
	}

	return map[string]any{
		"order_uid":    orderUID,
		"track_number": trackNumber,
		"entry":        "WBIL",
		"delivery": map[string]any{
			"name":    names[rand.Intn(len(names))],
			"phone":   fmt.Sprintf("+%d", 1000000000+rand.Intn(9000000000)),
			"zip":     fmt.Sprintf("%d", 100000+rand.Intn(900000)),
			"city":    cities[rand.Intn(len(cities))],
			"address": fmt.Sprintf("Street %d, Building %d", rand.Intn(100)+1, rand.Intn(20)+1),
			"region":  regions[rand.Intn(len(regions))],
			"email":   emails[rand.Intn(len(emails))],
		},
		"payment": map[string]any{
			"transaction":   orderUID,
			"request_id":    "",
			"currency":      currencies[rand.Intn(len(currencies))],
			"provider":      providers[rand.Intn(len(providers))],
			"amount":        rand.Intn(2000) + 500,
			"payment_dt":    time.Now().Unix() - int64(rand.Intn(86400*30)),
			"bank":          banks[rand.Intn(len(banks))],
			"delivery_cost": rand.Intn(500) + 100,
			"goods_total":   rand.Intn(1000) + 200,
			"custom_fee":    rand.Intn(50),
		},
		"items":              items,
		"locale":             "en",
		"internal_signature": "",
		"customer_id":        fmt.Sprintf("customer_%d", rand.Intn(1000)),
		"delivery_service":   "meest",
		"shardkey":           fmt.Sprintf("%d", rand.Intn(10)),
		"sm_id":              rand.Intn(100),
		"date_created":       time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour).Format(time.RFC3339),
		"oof_shard":          fmt.Sprintf("%d", rand.Intn(5)),
	}
}
