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
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/Ali0NAL/talentpass/internal/config"
	"github.com/Ali0NAL/talentpass/internal/db"
	httpx "github.com/Ali0NAL/talentpass/internal/http"
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

	// Base router + health
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

	// Swagger UI (hiç argüman yok!)
	// İstersen URL'i sabitleyebilirsin:
	// r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("http://localhost:"+cfg.Port+"/swagger/doc.json")))
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// v1 routes
	r.Route("/v1", func(r chi.Router) {
		ah := httpx.NewAuthHandler(pool)
		r.Mount("/auth", ah.Router())

		r.Group(func(pr chi.Router) {
			pr.Use(httpx.RequireAuth)

			jh := httpx.NewJobsHandler(pool)
			pr.Mount("/jobs", jh.Router())

			ap := httpx.NewApplicationsHandler(pool)
			pr.Mount("/applications", ap.Router())
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("http server starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server listen failed")
		}
	}()

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
