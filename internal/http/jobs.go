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

type JobsHandler struct {
	queries *repo.Queries
	pool    *pgxpool.Pool
}

func NewJobsHandler(pool *pgxpool.Pool) *JobsHandler {
	return &JobsHandler{
		queries: repo.New(pool),
		pool:    pool,
	}
}

func (h *JobsHandler) Router() http.Handler {
	r := chi.NewRouter()

	// /v1/jobs
	r.Get("/", h.listJobs)
	r.Post("/", h.createJob)

	return r
}

type createJobReq struct {
	OrgID    *int64   `json:"org_id"` // opsiyonel; multi-tenant için
	Title    string   `json:"title"`
	Company  string   `json:"company"`
	URL      *string  `json:"url"`
	Location *string  `json:"location"`
	Tags     []string `json:"tags"`
}

func (h *JobsHandler) createJob(w http.ResponseWriter, r *http.Request) {
	var req createJobReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Title == "" || req.Company == "" {
		writeError(w, http.StatusBadRequest, "title and company are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	orgID := (*int64)(nil)
	if req.OrgID != nil {
		orgID = req.OrgID
	}

	url := (*string)(nil)
	if req.URL != nil {
		url = req.URL
	}

	loc := (*string)(nil)
	if req.Location != nil {
		loc = req.Location
	}

	args := repo.CreateJobParams{
		OrgID:    orgID,
		Title:    req.Title,
		Company:  req.Company,
		Url:      url,
		Location: loc,
		Tags:     req.Tags,
	}

	job, err := h.queries.CreateJob(ctx, args)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

func (h *JobsHandler) listJobs(w http.ResponseWriter, r *http.Request) {
	// Query string: ?org_id=&q=&limit=&offset=
	q := r.URL.Query()

	var orgID *int64
	if s := q.Get("org_id"); s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			orgID = &v
		}
	}

	// q'yu hem company hem title için kullan
	var needle *string
	if s := q.Get("q"); s != "" {
		needle = &s
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

	company := (*string)(nil)
	title := (*string)(nil)
	if needle != nil {
		company = needle
		title = needle
	}

	rows, err := h.queries.ListJobs(ctx, repo.ListJobsParams{
		OrgID:   orgID,
		Company: company,
		Title:   title,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": rows,
		// Basit sayfalama bilgisi
		"limit":  limit,
		"offset": offset,
	})
}
