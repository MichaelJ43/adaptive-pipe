package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(h *Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/login", h.Login)
		r.Post("/tenants/{slug}/webhooks/github", h.GitHubWebhook)

		r.Group(func(r chi.Router) {
			r.Use(h.AuthMiddleware)
			r.Post("/runs", h.CreateRun)
			r.Get("/repos/{org}/{repo}/runs", h.ListRuns)
			r.Get("/runs/{runID}", h.GetRun)
			r.Get("/platform/settings", h.GetSettings)
			r.Group(func(r chi.Router) {
				r.Use(h.AdminMiddleware)
				r.Patch("/platform/settings", h.PatchSettings)
			})
		})
	})
	return r
}
