package httpx

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func newSubrouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Timeout(60 * time.Second))
	return r
}
