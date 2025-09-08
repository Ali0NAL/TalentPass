package httpx

import (
	"encoding/json"
	"net/http"
	"strconv"

	repo "github.com/Ali0NAL/talentpass/internal/repo"
	"github.com/go-chi/chi/v5"
)

type CreateEventRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload_json"`
}

type EventHandler struct {
	Q *repo.Queries
}

func NewEventHandler(q *repo.Queries) *EventHandler { return &EventHandler{Q: q} }

// POST /v1/applications/{id}/events
func (h *EventHandler) Create(w http.ResponseWriter, r *http.Request) {
	appID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || appID <= 0 {
		http.Error(w, "invalid application id", http.StatusBadRequest)
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}
	if len(req.Payload) == 0 {
		req.Payload = json.RawMessage(`{}`)
	}

	// TODO: auth'tan gerçek userID al
	userID := int64(1)

	ev, err := h.Q.CreateEvent(r.Context(), repo.CreateEventParams{
		UserID:        userID,
		ApplicationID: appID,
		Type:          req.Type,
		PayloadJson:   req.Payload, // sqlc => column payload_json -> field PayloadJson
	})
	if err != nil {
		http.Error(w, "create event failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ev)
}

// GET /v1/applications/{id}/events?limit=&offset=
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	appID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || appID <= 0 {
		http.Error(w, "invalid application id", http.StatusBadRequest)
		return
	}
	limit := int32(20)
	offset := int32(0)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}

	rows, err := h.Q.ListEventsByApplication(r.Context(), repo.ListEventsByApplicationParams{
		ApplicationID: appID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		http.Error(w, "list events failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items":  rows,
		"limit":  limit,
		"offset": offset,
	})
}
