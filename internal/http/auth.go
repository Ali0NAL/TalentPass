package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ali0NAL/talentpass/internal/auth"
	"github.com/Ali0NAL/talentpass/internal/repo"
)

type AuthHandler struct {
	q    *repo.Queries
	pool *pgxpool.Pool
}

func NewAuthHandler(pool *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{q: repo.New(pool), pool: pool}
}

func (h *AuthHandler) Router() http.Handler {
	r := newSubrouter()
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	return r
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "email or password too short")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash error")
		return
	}
	u, err := h.q.CreateUser(ctx, repo.CreateUserParams{
		Email:        req.Email,
		PasswordHash: hash,
	})
	if err != nil {
		// muhtemel UNIQUE ihlali
		writeError(w, http.StatusConflict, "email already exists")
		return
	}
	token, exp, err := auth.GenerateAccessToken(u.ID, u.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"user":         map[string]any{"id": u.ID, "email": u.Email},
		"access_token": token,
		"expires_at":   exp.UTC(),
	})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	u, err := h.q.GetUserByEmail(ctx, req.Email)
	if err != nil || !auth.CheckPassword(u.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	token, exp, err := auth.GenerateAccessToken(u.ID, u.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user":         map[string]any{"id": u.ID, "email": u.Email},
		"access_token": token,
		"expires_at":   exp.UTC(),
	})
}
