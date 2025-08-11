package config

import "os"

type Config struct {
	HTTPAddr     string
	PostgresDSN  string
	KafkaBrokers []string
	KafkaTopic   string
}

func Load() *Config {
	return &Config{
		HTTPAddr:     getEnv("HTTP_ADDR", ":8081"),
		PostgresDSN:  getEnv("PG_DSN", "postgres://administrator:password@localhost:5432/L0"),
		KafkaBrokers: getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "orders"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	return fallback
}
