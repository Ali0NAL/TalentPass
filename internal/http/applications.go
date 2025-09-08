package httpx

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ali0NAL/talentpass/internal/repo"
)

// ApplicationsHandler: applications (job başvuruları) endpoint'leri
type ApplicationsHandler struct {
	q    *repo.Queries
	pool *pgxpool.Pool
}

func NewApplicationsHandler(pool *pgxpool.Pool) *ApplicationsHandler {
	return &ApplicationsHandler{q: repo.New(pool), pool: pool}
}

func (h *ApplicationsHandler) Router() http.Handler {
	r := chi.NewRouter()
	r.Post("/", h.create)                   // POST   /v1/applications
	r.Get("/", h.list)                      // GET    /v1/applications
	r.Patch("/{id}:status", h.updateStatus) // PATCH  /v1/applications/{id}:status
	return r
}

type CreateAppReq struct {
	JobID        int64   `json:"job_id"`
	Status       *string `json:"status"`         // optional: applied/interview/offer/denied
	Notes        *string `json:"notes"`          // optional
	NextActionAt *string `json:"next_action_at"` // RFC3339 (optional)
}

// @Summary      Create application
// @Description  Bir ilana başvuru oluşturur
// @Tags         applications
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  CreateAppReq  true  "application payload"
// @Success      201   {object}  repo.Application
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Router       /v1/applications [post]
func (h *ApplicationsHandler) create(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateAppReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.JobID == 0 {
		writeError(w, http.StatusBadRequest, "job_id required")
		return
	}

	// next_action_at parse (opsiyonel)
	var nextAt *time.Time
	if req.NextActionAt != nil && *req.NextActionAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.NextActionAt); err == nil {
			nextAt = &t
		} else {
			writeError(w, http.StatusBadRequest, "invalid next_action_at (RFC3339)")
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	app, err := h.q.CreateApplication(ctx, repo.CreateApplicationParams{
		JobID:        req.JobID,
		UserID:       uid,
		Status:       req.Status,
		Notes:        req.Notes,
		NextActionAt: nextAt,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, app)
}

// @Summary      List my applications
// @Description  Kullanıcının kendi başvurularını listeler
// @Tags         applications
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        status  query   string  false  "applied|interview|offer|denied"
// @Param        limit   query   int     false  "limit (1-100)"
// @Param        offset  query   int     false  "offset"
// @Success      200     {object}  map[string]any
// @Failure      401     {object}  map[string]string
// @Router       /v1/applications [get]
func (h *ApplicationsHandler) list(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	var status *string
	if s := q.Get("status"); s != "" {
		status = &s
	}
	limit := int32(20)
	if s := q.Get("limit"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 32); err == nil && v > 0 && v <= 100 {
			limit = int32(v)
		}
	}
	offset := int32(0)
	if s := q.Get("offset"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 32); err == nil && v >= 0 {
			offset = int32(v)
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rows, err := h.q.ListApplicationsByUser(ctx, repo.ListApplicationsByUserParams{
		UserID: uid,
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items":  rows,
		"limit":  limit,
		"offset": offset,
	})
}

type UpdateStatusReq struct {
	Status string `json:"status"` // applied/interview/offer/denied
}

// @Summary      Update application status
// @Description  Başvuru durumunu günceller
// @Tags         applications
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path   int64            true  "application id"
// @Param        body  body   UpdateStatusReq  true  "status payload"
// @Success      200   {object}  repo.Application
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Router       /v1/applications/{id}:status [patch]
func (h *ApplicationsHandler) updateStatus(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	appID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || appID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdateStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	switch req.Status {
	case "applied", "interview", "offer", "denied":
	default:
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// 1) status güncelle
	app, err := h.q.UpdateApplicationStatus(ctx, repo.UpdateApplicationStatusParams{
		ID:     appID,
		UserID: uid, // sahiplik kontrolü
		Status: req.Status,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}

	// 2) audit event
	payload, _ := json.Marshal(map[string]any{
		"application_id": app.ID,
		"user_id":        uid,
		"new_status":     app.Status,
		"updated_at":     app.UpdatedAt,
	})
	_, _ = h.q.CreateEvent(ctx, repo.CreateEventParams{
		UserID:        uid,
		ApplicationID: appID,
		Type:          "application.status.changed",
		PayloadJson:   payload,
	})

	writeJSON(w, http.StatusOK, app)
}
