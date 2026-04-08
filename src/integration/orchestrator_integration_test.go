//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/MichaelJ43/adaptive-pipe/internal/clients"
	"github.com/MichaelJ43/adaptive-pipe/internal/config"
	"github.com/MichaelJ43/adaptive-pipe/internal/db"
	"github.com/MichaelJ43/adaptive-pipe/internal/dispatcher"
	"github.com/redis/go-redis/v9"
)

// Requires docker compose up: orchestrator worker consumes the queue for this test.
func TestPipelineViaSharedQueue(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1")
	}
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	store := db.NewStore(pool)
	tid, err := store.TenantBySlug(ctx, "demo")
	if err != nil {
		t.Fatal("start compose first (demo tenant from orchestrator seed)", err)
	}
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(opt)
	t.Cleanup(func() { _ = rdb.Close() })
	disp := &dispatcher.Dispatcher{
		RDB:      rdb,
		Store:    store,
		Validate: clients.NewValidateClient(cfg.ValidateURL),
		File:     clients.NewFileClient(cfg.FileURL),
	}

	bn, err := store.NextBuildNumber(ctx, tid, "acme", fmt.Sprintf("integration-%d", time.Now().UnixNano()))
	if err != nil {
		t.Fatal(err)
	}
	repo := fmt.Sprintf("integration-%d", time.Now().UnixNano())
	run, err := store.CreateRun(ctx, tid, "acme", repo, bn, "integsha")
	if err != nil {
		t.Fatal(err)
	}
	if err := disp.Enqueue(ctx, run.ID, "validate"); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		r2, err := store.GetRun(ctx, tid, run.ID)
		if err != nil {
			t.Fatal(err)
		}
		if r2.Status == "succeeded" {
			return
		}
		if r2.Status == "failed" {
			t.Fatalf("failed: %+v", r2.Stages)
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("timeout waiting for pipeline (is orchestrator running?)")
}

func TestValidateServiceHealth(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip()
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.ValidateURL+"/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status %d", res.StatusCode)
	}
}
