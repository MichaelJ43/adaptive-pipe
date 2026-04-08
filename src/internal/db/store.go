package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}

type User struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Username     string
	PasswordHash string
	Role         string
}

type Run struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	GithubOrg    string
	GithubRepo   string
	BuildNumber  int64
	CommitSHA    string
	Status       string
	FileNodeID   *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Stages       []Stage
}

type Stage struct {
	ID         uuid.UUID
	RunID      uuid.UUID
	Name       string
	State      string
	StartedAt  *time.Time
	FinishedAt *time.Time
	LogRef     *string
}

type PlatformSettings struct {
	TenantID        uuid.UUID
	BuildWarmPool   int
	TestWarmPool    int
	DeployWarmPool  int
	UpdatedAt       time.Time
}

func (s *Store) EnsureSeed(ctx context.Context, demoPassword string) error {
	var n int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	tid := uuid.New()
	slug := "demo"
	name := "Demo SaaS tenant"
	if _, err := s.pool.Exec(ctx, `INSERT INTO tenants (id, name, slug) VALUES ($1,$2,$3)`, tid, name, slug); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `INSERT INTO platform_settings (tenant_id) VALUES ($1)`, tid); err != nil {
		return err
	}
	hash, err := HashPassword(demoPassword)
	if err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, `INSERT INTO users (tenant_id, username, password_hash, role) VALUES ($1,$2,$3,'admin')`,
		tid, "admin", hash); err != nil {
		return err
	}
	return nil
}

func (s *Store) TenantBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `SELECT id FROM tenants WHERE slug=$1`, slug).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *Store) UserByUsername(ctx context.Context, tenantID uuid.UUID, username string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`SELECT id, tenant_id, username, password_hash, role FROM users WHERE tenant_id=$1 AND username=$2`,
		tenantID, username,
	).Scan(&u.ID, &u.TenantID, &u.Username, &u.PasswordHash, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) NextBuildNumber(ctx context.Context, tenantID uuid.UUID, org, repo string) (int64, error) {
	var max *int64
	err := s.pool.QueryRow(ctx,
		`SELECT MAX(build_number) FROM runs WHERE tenant_id=$1 AND github_org=$2 AND github_repo=$3`,
		tenantID, org, repo).Scan(&max)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	if max == nil {
		return 1, nil
	}
	return *max + 1, nil
}

func (s *Store) FindWebhookRun(ctx context.Context, tenantID uuid.UUID, deliveryID string) (uuid.UUID, bool, error) {
	var rid uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT run_id FROM webhook_idempotency WHERE tenant_id=$1 AND delivery_id=$2`,
		tenantID, deliveryID).Scan(&rid)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, err
	}
	return rid, true, nil
}

func (s *Store) RecordWebhookDedup(ctx context.Context, tenantID, runID uuid.UUID, deliveryID string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO webhook_idempotency (tenant_id, delivery_id, run_id) VALUES ($1,$2,$3)`,
		tenantID, deliveryID, runID)
	return err
}

func (s *Store) CreateRun(ctx context.Context, tenantID uuid.UUID, org, repo string, buildNumber int64, commit string) (*Run, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	runID := uuid.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO runs (id, tenant_id, github_org, github_repo, build_number, commit_sha, status)
		VALUES ($1,$2,$3,$4,$5,$6,'queued')`,
		runID, tenantID, org, repo, buildNumber, commit); err != nil {
		return nil, err
	}
	stageNames := []string{"validate", "file", "build", "test", "deploy"}
	for _, n := range stageNames {
		if _, err := tx.Exec(ctx, `INSERT INTO stages (run_id, name, state) VALUES ($1,$2,'pending')`, runID, n); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetRun(ctx, tenantID, runID)
}

func (s *Store) TenantSlug(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var slug string
	err := s.pool.QueryRow(ctx, `SELECT slug FROM tenants WHERE id=$1`, tenantID).Scan(&slug)
	return slug, err
}

// GetRunInternal loads a run by id without tenant check (dispatcher only).
func (s *Store) GetRunInternal(ctx context.Context, runID uuid.UUID) (*Run, error) {
	var r Run
	err := s.pool.QueryRow(ctx, `
		SELECT id, tenant_id, github_org, github_repo, build_number, commit_sha, status, file_node_id, created_at, updated_at
		FROM runs WHERE id=$1`, runID).
		Scan(&r.ID, &r.TenantID, &r.GithubOrg, &r.GithubRepo, &r.BuildNumber, &r.CommitSHA, &r.Status, &r.FileNodeID, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT id, run_id, name, state, started_at, finished_at, log_ref FROM stages WHERE run_id=$1 ORDER BY
		CASE name WHEN 'validate' THEN 1 WHEN 'file' THEN 2 WHEN 'build' THEN 3 WHEN 'test' THEN 4 WHEN 'deploy' THEN 5 ELSE 99 END`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var st Stage
		if err := rows.Scan(&st.ID, &st.RunID, &st.Name, &st.State, &st.StartedAt, &st.FinishedAt, &st.LogRef); err != nil {
			return nil, err
		}
		r.Stages = append(r.Stages, st)
	}
	return &r, nil
}

func (s *Store) GetRun(ctx context.Context, tenantID, runID uuid.UUID) (*Run, error) {
	var r Run
	err := s.pool.QueryRow(ctx, `
		SELECT id, tenant_id, github_org, github_repo, build_number, commit_sha, status, file_node_id, created_at, updated_at
		FROM runs WHERE id=$1 AND tenant_id=$2`, runID, tenantID).
		Scan(&r.ID, &r.TenantID, &r.GithubOrg, &r.GithubRepo, &r.BuildNumber, &r.CommitSHA, &r.Status, &r.FileNodeID, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `SELECT id, run_id, name, state, started_at, finished_at, log_ref FROM stages WHERE run_id=$1 ORDER BY
		CASE name WHEN 'validate' THEN 1 WHEN 'file' THEN 2 WHEN 'build' THEN 3 WHEN 'test' THEN 4 WHEN 'deploy' THEN 5 ELSE 99 END`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var st Stage
		if err := rows.Scan(&st.ID, &st.RunID, &st.Name, &st.State, &st.StartedAt, &st.FinishedAt, &st.LogRef); err != nil {
			return nil, err
		}
		r.Stages = append(r.Stages, st)
	}
	return &r, nil
}

func (s *Store) ListRuns(ctx context.Context, tenantID uuid.UUID, org, repo string, limit int) ([]Run, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, tenant_id, github_org, github_repo, build_number, commit_sha, status, file_node_id, created_at, updated_at
		FROM runs WHERE tenant_id=$1 AND github_org=$2 AND github_repo=$3
		ORDER BY build_number DESC LIMIT $4`, tenantID, org, repo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(&r.ID, &r.TenantID, &r.GithubOrg, &r.GithubRepo, &r.BuildNumber, &r.CommitSHA, &r.Status, &r.FileNodeID, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *Store) SetRunStatus(ctx context.Context, tenantID, runID uuid.UUID, status string) error {
	ct, err := s.pool.Exec(ctx, `UPDATE runs SET status=$3, updated_at=now() WHERE id=$1 AND tenant_id=$2`, runID, tenantID, status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) SetFileNode(ctx context.Context, tenantID, runID uuid.UUID, nodeID string) error {
	ct, err := s.pool.Exec(ctx, `UPDATE runs SET file_node_id=$3, updated_at=now() WHERE id=$1 AND tenant_id=$2`, runID, tenantID, nodeID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) UpdateStageState(ctx context.Context, runID uuid.UUID, stageName, state string) error {
	now := time.Now().UTC()
	switch state {
	case "running":
		_, err := s.pool.Exec(ctx, `UPDATE stages SET state=$3, started_at=$4 WHERE run_id=$1 AND name=$2`, runID, stageName, state, now)
		return err
	case "succeeded", "failed", "skipped":
		_, err := s.pool.Exec(ctx, `UPDATE stages SET state=$3, finished_at=$4 WHERE run_id=$1 AND name=$2`, runID, stageName, state, now)
		return err
	default:
		_, err := s.pool.Exec(ctx, `UPDATE stages SET state=$3 WHERE run_id=$1 AND name=$2`, runID, stageName, state)
		return err
	}
}

func (s *Store) RecordStageDuration(ctx context.Context, tenantID uuid.UUID, org, repo, stage string, ms int64) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO build_stage_durations (tenant_id, github_org, github_repo, stage_name, duration_ms)
		VALUES ($1,$2,$3,$4,$5)`, tenantID, org, repo, stage, ms)
	return err
}

func (s *Store) RecentStageDurations(ctx context.Context, tenantID uuid.UUID, org, repo, stage string, limit int) ([]int64, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.pool.Query(ctx, `
		SELECT duration_ms FROM build_stage_durations
		WHERE tenant_id=$1 AND github_org=$2 AND github_repo=$3 AND stage_name=$4
		ORDER BY recorded_at DESC LIMIT $5`, tenantID, org, repo, stage, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ds []int64
	for rows.Next() {
		var ms int64
		if err := rows.Scan(&ms); err != nil {
			return nil, err
		}
		ds = append(ds, ms)
	}
	return ds, nil
}

func (s *Store) GetPlatformSettings(ctx context.Context, tenantID uuid.UUID) (*PlatformSettings, error) {
	var p PlatformSettings
	err := s.pool.QueryRow(ctx, `
		SELECT tenant_id, build_warm_pool, test_warm_pool, deploy_warm_pool, updated_at
		FROM platform_settings WHERE tenant_id=$1`, tenantID).
		Scan(&p.TenantID, &p.BuildWarmPool, &p.TestWarmPool, &p.DeployWarmPool, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) UpdatePlatformSettings(ctx context.Context, tenantID uuid.UUID, build, test, deploy int) error {
	ct, err := s.pool.Exec(ctx, `
		UPDATE platform_settings SET build_warm_pool=$2, test_warm_pool=$3, deploy_warm_pool=$4, updated_at=now()
		WHERE tenant_id=$1`, tenantID, build, test, deploy)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("no platform_settings for tenant")
	}
	return nil
}

func (s *Store) ListRunNumbersForRepo(ctx context.Context, tenantID uuid.UUID, org, repo string) ([]int64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT build_number FROM runs WHERE tenant_id=$1 AND github_org=$2 AND github_repo=$3
		ORDER BY build_number DESC`, tenantID, org, repo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var nums []int64
	for rows.Next() {
		var n int64
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		nums = append(nums, n)
	}
	return nums, nil
}

func (s *Store) RunIDByBuildNumber(ctx context.Context, tenantID uuid.UUID, org, repo string, buildNumber int64) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT id FROM runs WHERE tenant_id=$1 AND github_org=$2 AND github_repo=$3 AND build_number=$4`,
		tenantID, org, repo, buildNumber).Scan(&id)
	return id, err
}

func (s *Store) DeleteRun(ctx context.Context, tenantID, runID uuid.UUID) error {
	ct, err := s.pool.Exec(ctx, `DELETE FROM runs WHERE id=$1 AND tenant_id=$2`, runID, tenantID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) IsRunActive(ctx context.Context, tenantID, runID uuid.UUID) (bool, error) {
	var st string
	err := s.pool.QueryRow(ctx, `SELECT status FROM runs WHERE id=$1 AND tenant_id=$2`, runID, tenantID).Scan(&st)
	if err != nil {
		return false, err
	}
	return st == "queued" || st == "running", nil
}
