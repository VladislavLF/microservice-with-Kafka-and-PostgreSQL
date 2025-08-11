package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"L0/internal/cache"
	"L0/internal/config"
	"L0/internal/db"
	"L0/internal/handler"
	"L0/internal/kafka"
)

func main() {
	cfg := config.Load()

	pg, err := db.NewPostgres(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pg.Close()

	orderCache := cache.New()
	if err := orderCache.LoadFromDB(context.Background(), pg); err != nil {
		log.Fatalf("Failed to load cache: %v", err)
	}

	webRoot, err := filepath.Abs("../../web")
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	requiredFiles := []string{
		filepath.Join(webRoot, "templates", "index.html"),
		filepath.Join(webRoot, "static", "css", "style.css"),
		filepath.Join(webRoot, "static", "js", "app.js"),
	}

	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Fatalf("Required file does not exist: %s", file)
		}
	}

	kafkaConsumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, pg, orderCache)
	go kafkaConsumer.Start()
	defer kafkaConsumer.Stop()

	httpHandler := handler.New(orderCache, webRoot)
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: httpHandler.SetupRoutes(),
	}

	go func() {
		log.Printf("HTTP server started on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped")
}
