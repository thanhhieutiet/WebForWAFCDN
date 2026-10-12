package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/http/presenter"
	"labshop/internal/usecase/port"
	tutorialuc "labshop/internal/usecase/tutorial"
)

// TutorialHandler serves HTTP endpoints for tutorial scenarios.
type TutorialHandler struct {
	tutorialUC   *tutorialuc.Usecase
	cacheSeconds int
}

// NewTutorialHandler constructs a new TutorialHandler.
func NewTutorialHandler(tutorialUC *tutorialuc.Usecase, cacheSeconds int) *TutorialHandler {
	if cacheSeconds <= 0 {
		cacheSeconds = 60
	}
	return &TutorialHandler{
		tutorialUC:   tutorialUC,
		cacheSeconds: cacheSeconds,
	}
}

// ListScenarios handles GET /api/tutorial/scenarios.
// Supports optional filtering by query parameters: ?group=cdn|waap|shop and ?priority=P0|P1.
// Complies with Origin Cache Policy: public, max-age=<CATALOG_CACHE_SECONDS>, with ETag and 304 support.
func (h *TutorialHandler) ListScenarios(w http.ResponseWriter, r *http.Request) {
	scenarios, err := h.tutorialUC.ListScenarios(r.Context())
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	// Optional filtering by group or priority
	q := r.URL.Query()
	groupFilter := strings.ToLower(strings.TrimSpace(q.Get("group")))
	priorityFilter := strings.ToUpper(strings.TrimSpace(q.Get("priority")))

	filtered := make([]port.Scenario, 0, len(scenarios))
	for _, s := range scenarios {
		if groupFilter != "" && strings.ToLower(s.Group) != groupFilter {
			continue
		}
		if priorityFilter != "" && strings.ToUpper(s.Priority) != priorityFilter {
			continue
		}
		filtered = append(filtered, s)
	}

	bodyBytes, err := json.Marshal(filtered)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	etag := computeETag(bodyBytes)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", h.cacheSeconds))
	w.Header().Set("ETag", etag)

	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" && (ifNoneMatch == etag || ifNoneMatch == "*") {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bodyBytes)
}

// GetScenario handles GET /api/tutorial/scenarios/{id}.
// Returns a single scenario by ID, supporting ETag and 304 Not Modified.
func (h *TutorialHandler) GetScenario(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	scenario, err := h.tutorialUC.GetScenario(r.Context(), id)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	bodyBytes, err := json.Marshal(scenario)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	etag := computeETag(bodyBytes)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", h.cacheSeconds))
	w.Header().Set("ETag", etag)

	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" && (ifNoneMatch == etag || ifNoneMatch == "*") {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(bodyBytes)
}

func computeETag(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf(`"%s"`, hex.EncodeToString(h[:16]))
}
