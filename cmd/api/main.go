package main

// @title           TalentPass API
// @version         1.0
// @description     İş ilanı ve başvuru yönetimi API'si.
// @schemes         http
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer {token}

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/Ali0NAL/talentpass/docs"

	"github.com/Ali0NAL/talentpass/internal/config"
	"github.com/Ali0NAL/talentpass/internal/db"
	httpx "github.com/Ali0NAL/talentpass/internal/http"
	repo "github.com/Ali0NAL/talentpass/internal/repo"
)

func main() {
	_ = godotenv.Load()
	zerolog.TimeFieldFormat = time.RFC3339

	cfg := config.Load()

	// DB pool
	pool, err := db.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("db connect failed")
	}
	defer pool.Close()

	// sqlc queries
	q := repo.New(pool)

	// Base router (+ /healthz, CORS, rate-limit, swagger vs. zaten NewBaseRouter içinde var)
	r := httpx.NewBaseRouter()

	// /readyz
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	// v1 routes
	r.Route("/v1", func(r chi.Router) {
		// public
		ah := httpx.NewAuthHandler(pool)
		r.Mount("/auth", ah.Router())

		// protected
		r.Group(func(pr chi.Router) {
			pr.Use(httpx.RequireAuth)

			jh := httpx.NewJobsHandler(pool)
			pr.Mount("/jobs", jh.Router())

			ap := httpx.NewApplicationsHandler(pool)
			pr.Mount("/applications", ap.Router())

			oh := httpx.NewOrgsHandler(pool)
			pr.Mount("/orgs", oh.Router())

			// --- EVENTS: /v1/applications/{id}/events ---
			eh := httpx.NewEventHandler(q)
			pr.Route("/applications/{id}/events", func(r chi.Router) {
				r.Post("/", eh.Create)
				r.Get("/", eh.List)
			})
		})
	})

	// HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// run
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("http server starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server listen failed")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("graceful shutdown failed")
	}
	log.Info().Msg("bye 👋")
}
