package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/MichaelJ43/adaptive-pipe/internal/auth"
	"github.com/MichaelJ43/adaptive-pipe/internal/db"
	"github.com/MichaelJ43/adaptive-pipe/internal/dispatcher"
	"github.com/MichaelJ43/adaptive-pipe/internal/eta"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	Store      *db.Store
	Dispatcher *dispatcher.Dispatcher
	JWTSecret  []byte
}

type loginReq struct {
	TenantSlug string `json:"tenant_slug"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type loginResp struct {
	Token string `json:"token"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	tid, err := h.Store.TenantBySlug(r.Context(), req.TenantSlug)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	u, err := h.Store.UserByUsername(r.Context(), tid, req.Username)
	if err != nil || !db.CheckPassword(u.PasswordHash, req.Password) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	tok, err := auth.SignJWT(h.JWTSecret, tid, u.ID, u.Username, u.Role, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, loginResp{Token: tok})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get("Authorization")
		if !strings.HasPrefix(hdr, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		tok := strings.TrimPrefix(hdr, "Bearer ")
		c, err := auth.ParseJWT(h.JWTSecret, tok)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := WithClaims(r.Context(), c)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, ok := ClaimsFromContext(r.Context())
		if !ok || c.Role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type createRunReq struct {
	GithubOrg   string `json:"github_org"`
	GithubRepo  string `json:"github_repo"`
	CommitSHA   string `json:"commit_sha"`
}

type runResp struct {
	RunID       string            `json:"run_id"`
	BuildNumber int64             `json:"build_number"`
	Status      string            `json:"status"`
	Stages      []stageResp       `json:"stages,omitempty"`
	ETAsSec     map[string]float64 `json:"eta_seconds,omitempty"`
}

type stageResp struct {
	Name   string `json:"name"`
	State  string `json:"state"`
}

func (h *Handler) CreateRun(w http.ResponseWriter, r *http.Request) {
	c, _ := ClaimsFromContext(r.Context())
	var req createRunReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.GithubOrg == "" || req.GithubRepo == "" || req.CommitSHA == "" {
		http.Error(w, "missing org/repo/commit_sha", http.StatusBadRequest)
		return
	}
	bn, err := h.Store.NextBuildNumber(r.Context(), c.TenantID, req.GithubOrg, req.GithubRepo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	run, err := h.Store.CreateRun(r.Context(), c.TenantID, req.GithubOrg, req.GithubRepo, bn, req.CommitSHA)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.Dispatcher.Enqueue(r.Context(), run.ID, "validate"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, runResp{RunID: run.ID.String(), BuildNumber: run.BuildNumber, Status: run.Status})
}

func (h *Handler) ListRuns(w http.ResponseWriter, r *http.Request) {
	c, _ := ClaimsFromContext(r.Context())
	org := chi.URLParam(r, "org")
	repo := chi.URLParam(r, "repo")
	runs, err := h.Store.ListRuns(r.Context(), c.TenantID, org, repo, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]runResp, 0, len(runs))
	for _, run := range runs {
		out = append(out, runResp{RunID: run.ID.String(), BuildNumber: run.BuildNumber, Status: run.Status})
	}
	writeJSON(w, out)
}

func (h *Handler) GetRun(w http.ResponseWriter, r *http.Request) {
	c, _ := ClaimsFromContext(r.Context())
	rid, err := uuid.Parse(chi.URLParam(r, "runID"))
	if err != nil {
		http.Error(w, "bad run id", http.StatusBadRequest)
		return
	}
	run, err := h.Store.GetRun(r.Context(), c.TenantID, rid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	stages := make([]stageResp, 0, len(run.Stages))
	etas := make(map[string]float64)
	for _, s := range run.Stages {
		stages = append(stages, stageResp{Name: s.Name, State: s.State})
		if samples, err := h.Store.RecentStageDurations(r.Context(), c.TenantID, run.GithubOrg, run.GithubRepo, s.Name, 8); err == nil {
			if avg, ok := eta.SimpleMovingAverageMs(samples); ok {
				etas[s.Name] = float64(avg) / 1000.0
			}
		}
	}
	writeJSON(w, runResp{
		RunID:       run.ID.String(),
		BuildNumber: run.BuildNumber,
		Status:      run.Status,
		Stages:      stages,
		ETAsSec:     etas,
	})
}

type settingsResp struct {
	BuildWarmPool   int `json:"build_warm_pool"`
	TestWarmPool    int `json:"test_warm_pool"`
	DeployWarmPool  int `json:"deploy_warm_pool"`
}

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	c, _ := ClaimsFromContext(r.Context())
	p, err := h.Store.GetPlatformSettings(r.Context(), c.TenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, settingsResp{BuildWarmPool: p.BuildWarmPool, TestWarmPool: p.TestWarmPool, DeployWarmPool: p.DeployWarmPool})
}

type patchSettingsReq struct {
	BuildWarmPool   *int `json:"build_warm_pool"`
	TestWarmPool    *int `json:"test_warm_pool"`
	DeployWarmPool  *int `json:"deploy_warm_pool"`
}

func (h *Handler) PatchSettings(w http.ResponseWriter, r *http.Request) {
	c, _ := ClaimsFromContext(r.Context())
	cur, err := h.Store.GetPlatformSettings(r.Context(), c.TenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req patchSettingsReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	b, t, d := cur.BuildWarmPool, cur.TestWarmPool, cur.DeployWarmPool
	if req.BuildWarmPool != nil {
		b = *req.BuildWarmPool
	}
	if req.TestWarmPool != nil {
		t = *req.TestWarmPool
	}
	if req.DeployWarmPool != nil {
		d = *req.DeployWarmPool
	}
	if b < 0 || t < 0 || d < 0 {
		http.Error(w, "warm pools must be non-negative", http.StatusBadRequest)
		return
	}
	if err := h.Store.UpdatePlatformSettings(r.Context(), c.TenantID, b, t, d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, settingsResp{BuildWarmPool: b, TestWarmPool: t, DeployWarmPool: d})
}

type ghPush struct {
	After string `json:"after"`
	Repository *struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

func (h *Handler) GitHubWebhook(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	tid, err := h.Store.TenantBySlug(r.Context(), slug)
	if err != nil {
		http.Error(w, "unknown tenant", http.StatusNotFound)
		return
	}
	delivery := r.Header.Get("X-GitHub-Delivery")
	if delivery != "" {
		rid, ok, err := h.Store.FindWebhookRun(r.Context(), tid, delivery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if ok {
			writeJSON(w, map[string]string{"run_id": rid.String(), "deduplicated": "true"})
			return
		}
	}
	var p ghPush
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if p.Repository == nil || p.After == "" {
		http.Error(w, "not a push event payload", http.StatusBadRequest)
		return
	}
	org := p.Repository.Owner.Login
	repo := p.Repository.Name
	commit := p.After

	bn, err := h.Store.NextBuildNumber(r.Context(), tid, org, repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	run, err := h.Store.CreateRun(r.Context(), tid, org, repo, bn, commit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if delivery != "" {
		_ = h.Store.RecordWebhookDedup(r.Context(), tid, run.ID, delivery)
	}
	if err := h.Dispatcher.Enqueue(r.Context(), run.ID, "validate"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"run_id": run.ID.String(), "build_number": run.BuildNumber, "status": run.Status})
}
