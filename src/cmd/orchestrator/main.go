package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MichaelJ43/adaptive-pipe/internal/clients"
	"github.com/MichaelJ43/adaptive-pipe/internal/config"
	"github.com/MichaelJ43/adaptive-pipe/internal/db"
	"github.com/MichaelJ43/adaptive-pipe/internal/dispatcher"
	"github.com/MichaelJ43/adaptive-pipe/internal/httpapi"
	"github.com/redis/go-redis/v9"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		slog.Error("migrate", "err", err)
		os.Exit(1)
	}

	store := db.NewStore(pool)
	seedPW := cfg.SeedAdminPW
	if seedPW == "" {
		seedPW = "admin123"
	}
	if err := store.EnsureSeed(ctx, seedPW); err != nil {
		slog.Error("seed", "err", err)
		os.Exit(1)
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Error("redis url", "err", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	defer func() { _ = rdb.Close() }()
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("redis ping", "err", err)
		os.Exit(1)
	}

	disp := &dispatcher.Dispatcher{
		RDB:      rdb,
		Store:    store,
		Validate: clients.NewValidateClient(cfg.ValidateURL),
		File:     clients.NewFileClient(cfg.FileURL),
		Logger:   slog.Default(),
	}
	go func() {
		if err := disp.Run(ctx); err != nil && ctx.Err() == nil {
			slog.Error("dispatcher", "err", err)
			os.Exit(1)
		}
	}()

	h := &httpapi.Handler{
		Store:      store,
		Dispatcher: disp,
		JWTSecret:  []byte(cfg.JWTSecret),
	}
	srv := &http.Server{Addr: cfg.ListenAddr, Handler: httpapi.NewRouter(h)}
	go func() {
		slog.Info("listening", "addr", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
