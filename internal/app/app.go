package app

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"order-svc/internal/cache"
	"order-svc/internal/config"
	h "order-svc/internal/http"
	kc "order-svc/internal/kafka"
	"order-svc/internal/log"
	"order-svc/internal/model"
	"order-svc/internal/repo"
	"order-svc/internal/service"
)

type App struct {
	log  *zap.Logger
	cfg  *config.Config
	db   *pgxpool.Pool
	svc  *service.Service
	kc   *kc.Consumer
	http *http.Server
}

func New() (*App, error) {
	logger, err := log.New()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pcfg, err := pgxpool.ParseConfig(cfg.DB.DSN)
	if err != nil {
		return nil, err
	}
	pcfg.MaxConns = int32(cfg.DB.MaxOpenConns)
	db, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	if err := repo.RunMigrations(cfg.DB.DSN, true); err != nil {
		logger.Error("migrations up error", zap.Error(err))
		return nil, err
	}

	r := repo.New(db)
	lru := cache.NewLRU[string, *model.Order](cfg.Cache.Capacity, cfg.Cache.TTL)
	svc := service.New(r, lru)

	_ = svc.Warmup(ctx, cfg.Cache.Capacity)

	mux := http.NewServeMux()
	hd := h.NewHandler(svc, logger)
	mux.HandleFunc("GET /order/", hd.GetOrder)
	if cfg.UI.Enable {
		fs := http.FileServer(http.Dir(cfg.UI.StaticDir))
		mux.Handle("/", fs)
	}

	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	k := kc.New(kc.Config{
		Brokers:        cfg.Kafka.Brokers,
		Topic:          cfg.Kafka.Topic,
		GroupID:        cfg.Kafka.GroupID,
		MinBytes:       cfg.Kafka.MinBytes,
		MaxBytes:       cfg.Kafka.MaxBytes,
		CommitInterval: cfg.Kafka.CommitInterval,
	}, svc, logger)

	return &App{log: logger, cfg: cfg, db: db, svc: svc, kc: k, http: srv}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		a.log.Info("http listen", zap.String("addr", a.cfg.Server.Addr))
		if err := a.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Fatal("http server", zap.Error(err))
		}
	}()

	go func() {
		a.log.Info("kafka consumer started")
		if err := a.kc.Run(ctx); err != nil {
			a.log.Error("kafka run", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = a.http.Shutdown(shutdownCtx)
	a.db.Close()
	a.log.Sync()
	return nil
}
