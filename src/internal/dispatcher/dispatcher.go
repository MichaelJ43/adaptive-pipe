package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/MichaelJ43/adaptive-pipe/internal/clients"
	"github.com/MichaelJ43/adaptive-pipe/internal/db"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const JobQueueKey = "adaptive-pipe:jobs"

type Job struct {
	RunID uuid.UUID `json:"run_id"`
	Stage string    `json:"stage"`
}

type Dispatcher struct {
	RDB        *redis.Client
	Store      *db.Store
	Validate   *clients.ValidateClient
	File       *clients.FileClient
	Logger     *slog.Logger
}

func (d *Dispatcher) Enqueue(ctx context.Context, runID uuid.UUID, stage string) error {
	j := Job{RunID: runID, Stage: stage}
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return d.RDB.LPush(ctx, JobQueueKey, b).Err()
}

func (d *Dispatcher) Run(ctx context.Context) error {
	if d.Logger == nil {
		d.Logger = slog.Default()
	}
	for {
		res, err := d.RDB.BRPop(ctx, 0, JobQueueKey).Result()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			d.Logger.Error("brpop", "err", err)
			time.Sleep(time.Second)
			continue
		}
		if len(res) != 2 {
			continue
		}
		var job Job
		if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
			d.Logger.Error("job json", "err", err)
			continue
		}
		if err := d.handleJob(ctx, job); err != nil {
			d.Logger.Error("job failed", "run_id", job.RunID, "stage", job.Stage, "err", err)
		}
	}
}

func (d *Dispatcher) handleJob(ctx context.Context, job Job) error {
	run, err := d.Store.GetRunInternal(ctx, job.RunID)
	if err != nil {
		return err
	}
	tenantID := run.TenantID

	if err := d.Store.SetRunStatus(ctx, tenantID, run.ID, "running"); err != nil {
		return err
	}

	switch job.Stage {
	case "validate":
		return d.doValidate(ctx, run)
	case "file":
		return d.doFile(ctx, run)
	case "build":
		return d.doBuild(ctx, run)
	case "test":
		return d.doTest(ctx, run)
	case "deploy":
		return d.doDeploy(ctx, run)
	default:
		return fmt.Errorf("unknown stage %q", job.Stage)
	}
}

func (d *Dispatcher) failRun(ctx context.Context, tenantID, runID uuid.UUID, stage string, cause error) error {
	_ = d.Store.UpdateStageState(ctx, runID, stage, "failed")
	_ = d.Store.SetRunStatus(ctx, tenantID, runID, "failed")
	return cause
}

func (d *Dispatcher) doValidate(ctx context.Context, run *db.Run) error {
	runID := run.ID
	tenantID := run.TenantID
	if err := d.Store.UpdateStageState(ctx, runID, "validate", "running"); err != nil {
		return err
	}
	start := time.Now()
	if err := d.Validate.Validate(ctx, run.GithubOrg, run.GithubRepo, run.CommitSHA); err != nil {
		return d.failRun(ctx, tenantID, runID, "validate", err)
	}
	ms := time.Since(start).Milliseconds()
	_ = d.Store.RecordStageDuration(ctx, tenantID, run.GithubOrg, run.GithubRepo, "validate", ms)
	if err := d.Store.UpdateStageState(ctx, runID, "validate", "succeeded"); err != nil {
		return err
	}
	return d.Enqueue(ctx, runID, "file")
}

func (d *Dispatcher) doFile(ctx context.Context, run *db.Run) error {
	runID := run.ID
	tenantID := run.TenantID
	if err := d.Store.UpdateStageState(ctx, runID, "file", "running"); err != nil {
		return err
	}
	start := time.Now()
	slug, err := d.Store.TenantSlug(ctx, tenantID)
	if err != nil {
		return d.failRun(ctx, tenantID, runID, "file", err)
	}
	ws, err := d.File.InitWorkspace(ctx, slug, run.GithubOrg, run.GithubRepo, runID.String(), run.CommitSHA)
	if err != nil {
		return d.failRun(ctx, tenantID, runID, "file", err)
	}
	if err := d.Store.SetFileNode(ctx, tenantID, runID, ws.NodeID); err != nil {
		return err
	}
	ms := time.Since(start).Milliseconds()
	_ = d.Store.RecordStageDuration(ctx, tenantID, run.GithubOrg, run.GithubRepo, "file", ms)
	if err := d.Store.UpdateStageState(ctx, runID, "file", "succeeded"); err != nil {
		return err
	}
	return d.Enqueue(ctx, runID, "build")
}

func (d *Dispatcher) doBuild(ctx context.Context, run *db.Run) error {
	runID := run.ID
	tenantID := run.TenantID
	if err := d.Store.UpdateStageState(ctx, runID, "build", "running"); err != nil {
		return err
	}
	start := time.Now()
	time.Sleep(80 * time.Millisecond) // stub build / compile
	ms := time.Since(start).Milliseconds()
	_ = d.Store.RecordStageDuration(ctx, tenantID, run.GithubOrg, run.GithubRepo, "build", ms)
	if err := d.Store.UpdateStageState(ctx, runID, "build", "succeeded"); err != nil {
		return err
	}
	return d.Enqueue(ctx, runID, "test")
}

func (d *Dispatcher) doTest(ctx context.Context, run *db.Run) error {
	runID := run.ID
	tenantID := run.TenantID
	if err := d.Store.UpdateStageState(ctx, runID, "test", "running"); err != nil {
		return err
	}
	start := time.Now()
	time.Sleep(50 * time.Millisecond) // stub test node
	ms := time.Since(start).Milliseconds()
	_ = d.Store.RecordStageDuration(ctx, tenantID, run.GithubOrg, run.GithubRepo, "test", ms)
	if err := d.Store.UpdateStageState(ctx, runID, "test", "succeeded"); err != nil {
		return err
	}
	return d.Enqueue(ctx, runID, "deploy")
}

func (d *Dispatcher) doDeploy(ctx context.Context, run *db.Run) error {
	runID := run.ID
	tenantID := run.TenantID
	if err := d.Store.UpdateStageState(ctx, runID, "deploy", "running"); err != nil {
		return err
	}
	start := time.Now()
	time.Sleep(60 * time.Millisecond) // stub Terraform / AWS apply
	ms := time.Since(start).Milliseconds()
	_ = d.Store.RecordStageDuration(ctx, tenantID, run.GithubOrg, run.GithubRepo, "deploy", ms)
	if err := d.Store.UpdateStageState(ctx, runID, "deploy", "succeeded"); err != nil {
		return err
	}
	return d.Store.SetRunStatus(ctx, tenantID, runID, "succeeded")
}
