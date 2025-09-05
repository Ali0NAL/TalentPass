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

type OrgsHandler struct {
	q *repo.Queries
}

func NewOrgsHandler(pool *pgxpool.Pool) *OrgsHandler {
	return &OrgsHandler{q: repo.New(pool)}
}

func (h *OrgsHandler) Router() http.Handler {
	r := newSubrouter()
	r.Post("/", h.createOrg)             // POST /v1/orgs
	r.Get("/", h.listMyOrgs)             // GET  /v1/orgs
	r.Post("/{id}/members", h.addMember) // POST /v1/orgs/{id}/members
	return r
}

type createOrgReq struct {
	Name string `json:"name"`
}

func (h *OrgsHandler) createOrg(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req createOrgReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	org, err := h.q.CreateOrganization(ctx, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// kurucuyu owner yap
	if err := h.q.AddOrgMember(ctx, repo.AddOrgMemberParams{
		OrgID: org.ID, UserID: uid, Role: "owner",
	}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, org)
}

func (h *OrgsHandler) listMyOrgs(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	orgs, err := h.q.ListMyOrganizations(ctx, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, orgs)
}

type addMemberReq struct {
	Email string `json:"email"`
	Role  string `json:"role"` // owner|admin|member
}

func (h *OrgsHandler) addMember(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	orgIDStr := chi.URLParam(r, "id")
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil || orgID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	var req addMemberReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}
	switch req.Role {
	case "owner", "admin", "member":
	default:
		writeError(w, http.StatusBadRequest, "invalid role")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// çağıran yetkili mi? (owner/admin)
	role, err := h.q.GetOrgMemberRole(ctx, repo.GetOrgMemberRoleParams{
		OrgID: orgID, UserID: uid,
	})
	if err != nil || (role != "owner" && role != "admin") {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	// eklenecek kullanıcıyı bul
	u, err := h.q.GetUserByEmail(ctx, req.Email)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	// ekle/güncelle
	if err := h.q.AddOrgMember(ctx, repo.AddOrgMemberParams{
		OrgID: orgID, UserID: u.ID, Role: req.Role,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"org_id": orgID,
		"user":   map[string]any{"id": u.ID, "email": u.Email},
		"role":   req.Role,
	})
}
