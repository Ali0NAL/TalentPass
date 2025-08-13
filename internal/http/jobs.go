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
	q *repo.Queries
}

func NewJobsHandler(pool *pgxpool.Pool) *JobsHandler {
	return &JobsHandler{q: repo.New(pool)}
}

func (h *JobsHandler) Router() http.Handler {
	r := newSubrouter()

	r.Post("/", h.createJob)
	r.Get("/", h.listJobs)
	r.Get("/{id}", h.getJob)
	r.Put("/{id}", h.updateJob)
	r.Delete("/{id}", h.deleteJob)

	return r
}

type CreateJobReq struct {
	OrgID    *int64   `json:"org_id,omitempty"`
	Title    string   `json:"title"`
	Company  string   `json:"company"`
	URL      *string  `json:"url,omitempty"`
	Location *string  `json:"location,omitempty"`
	Tags     []string `json:"tags"`
}

// @Summary Create job
// @Tags jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateJobReq true "job payload"
// @Success 201 {object} repo.Job
// @Failure 400 {object} map[string]string
// @Router /v1/jobs [post]
func (h *JobsHandler) createJob(w http.ResponseWriter, r *http.Request) {
	var req CreateJobReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Title == "" || req.Company == "" {
		writeError(w, http.StatusBadRequest, "title and company required")
		return
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	job, err := h.q.CreateJob(ctx, repo.CreateJobParams{
		OrgID:    req.OrgID,
		Title:    req.Title,
		Company:  req.Company,
		Url:      req.URL,
		Location: req.Location,
		Tags:     req.Tags,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

// @Summary List jobs
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param company query string false "filter by company (ILIKE)"
// @Param title   query string false "filter by title (ILIKE)"
// @Param limit   query int    false "limit (default 20)"
// @Param offset  query int    false "offset (default 0)"
// @Success 200 {object} map[string]any
// @Router /v1/jobs [get]
func (h *JobsHandler) listJobs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	company := q.Get("company")
	title := q.Get("title")

	limit := int32(20)
	offset := int32(0)
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = int32(n)
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}

	var companyPtr *string
	if company != "" {
		companyPtr = &company
	}
	var titlePtr *string
	if title != "" {
		titlePtr = &title
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	items, err := h.q.ListJobs(ctx, repo.ListJobsParams{
		Company: companyPtr,
		Title:   titlePtr,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

// @Summary Get job by ID
// @Tags jobs
// @Security BearerAuth
// @Produce json
// @Param id path int true "Job ID"
// @Success 200 {object} repo.Job
// @Failure 404 {object} map[string]string
// @Router /v1/jobs/{id} [get]
func (h *JobsHandler) getJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	job, err := h.q.GetJobByID(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	writeJSON(w, http.StatusOK, job)
}

type UpdateJobReq struct {
	Title    *string   `json:"title,omitempty"`
	Company  *string   `json:"company,omitempty"`
	URL      *string   `json:"url,omitempty"`
	Location *string   `json:"location,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
}

// @Summary Update job
// @Tags jobs
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Job ID"
// @Param body body UpdateJobReq true "job payload"
// @Success 200 {object} repo.Job
// @Router /v1/jobs/{id} [put]
func (h *JobsHandler) updateJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdateJobReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	params := repo.UpdateJobParams{
		ID: id,
	}
	if req.Title != nil {
		params.Title = req.Title
	}
	if req.Company != nil {
		params.Company = req.Company
	}
	if req.URL != nil {
		params.Url = req.URL
	}
	if req.Location != nil {
		params.Location = req.Location
	}
	if req.Tags != nil {
		params.Tags = *req.Tags
	}

	job, err := h.q.UpdateJob(ctx, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// @Summary Delete job
// @Tags jobs
// @Security BearerAuth
// @Param id path int true "Job ID"
// @Success 204 "No Content"
// @Router /v1/jobs/{id} [delete]
func (h *JobsHandler) deleteJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.q.DeleteJob(ctx, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
