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
	r.Post("/refresh", h.refresh)
	return r
}

type RegisterReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// @Summary      Register
// @Description  Yeni kullanıcı oluşturur ve access token döner
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  RegisterReq  true  "register payload"
// @Success      201   {object}  map[string]any
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Router       /v1/auth/register [post]
func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	var req RegisterReq
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

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// @Summary      Login
// @Description  E-posta/şifre ile giriş yapar ve access token döner
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  LoginReq  true  "login payload"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Router       /v1/auth/login [post]
func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req LoginReq
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

	// Access token üret
	accessToken, accessExp, err := auth.GenerateAccessToken(u.ID, u.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}

	// Refresh token üret + DB'ye kaydet
	refreshPlain, refreshHash, refreshExp, err := auth.NewRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}
	if _, err := h.q.CreateRefreshToken(ctx, repo.CreateRefreshTokenParams{
		UserID:    u.ID,
		TokenHash: refreshHash,
		ExpiresAt: refreshExp,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	// Cevap dön
	writeJSON(w, http.StatusOK, map[string]any{
		"user":          map[string]any{"id": u.ID, "email": u.Email},
		"access_token":  accessToken,
		"access_exp":    accessExp.UTC(),
		"refresh_token": refreshPlain,
		"refresh_exp":   refreshExp.UTC(),
	})
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  RefreshReq  true  "refresh payload"
// @Success      200   {object}  map[string]any
// @Failure      400   {object}  map[string]map[string]string
// @Failure      401   {object}  map[string]map[string]string
// @Router       /v1/auth/refresh [post]
func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Gelen plain refresh token'ı hash'le
	hash := auth.HashRefreshToken(req.RefreshToken)

	// DB'den refresh token kaydını çek
	rt, err := h.q.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		// bulunamadı
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	// revoke ya da süresi geçmiş mi?
	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		writeError(w, http.StatusUnauthorized, "token expired or revoked")
		return
	}

	// Kullanıcıyı getir (GetUserByID sorgun yoksa eklemelisin)
	user, err := h.q.GetUserByID(ctx, rt.UserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	// Yeni access token üret
	access, err := auth.NewAccessToken(user.ID, user.Email, time.Hour) // 1 saatlik access
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}

	// Rotation: eski refresh'i revoke et
	if err := h.q.RevokeRefreshToken(ctx, rt.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	// Yeni refresh üret ve kaydet
	plain, newHash, exp, err := auth.NewRefreshToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}
	if _, err := h.q.CreateRefreshToken(ctx, repo.CreateRefreshTokenParams{
		UserID:    rt.UserID,
		TokenHash: newHash,
		ExpiresAt: exp,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	// Cevap
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  access,
		"refresh_token": plain,
		"expires_in":    3600,
	})
}
