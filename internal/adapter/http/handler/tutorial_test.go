package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"labshop/internal/adapter/scenario/embedded"
	"labshop/internal/usecase/port"
	tutorialuc "labshop/internal/usecase/tutorial"
)

func setupTutorialHandlerTest(t *testing.T) (*TutorialHandler, *chi.Mux) {
	t.Helper()
	repo := embedded.MustNewRepository()
	uc := tutorialuc.New(repo)
	h := NewTutorialHandler(uc, 60)

	r := chi.NewRouter()
	r.Get("/api/tutorial/scenarios", h.ListScenarios)
	r.Get("/api/tutorial/scenarios/{id}", h.GetScenario)

	return h, r
}

func TestTutorialHandler_ListScenarios(t *testing.T) {
	_, r := setupTutorialHandlerTest(t)

	// Normal list
	req := httptest.NewRequest(http.MethodGet, "/api/tutorial/scenarios", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, "public, max-age=60", rec.Header().Get("Cache-Control"))
	etag := rec.Header().Get("ETag")
	assert.NotEmpty(t, etag)

	var list []port.Scenario
	err := json.Unmarshal(rec.Body.Bytes(), &list)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 23)

	// ETag 304 Not Modified check
	req304 := httptest.NewRequest(http.MethodGet, "/api/tutorial/scenarios", nil)
	req304.Header.Set("If-None-Match", etag)
	rec304 := httptest.NewRecorder()
	r.ServeHTTP(rec304, req304)

	assert.Equal(t, http.StatusNotModified, rec304.Code)
	assert.Empty(t, rec304.Body.String())
}

func TestTutorialHandler_ListScenarios_Filter(t *testing.T) {
	_, r := setupTutorialHandlerTest(t)

	// Filter by group=cdn
	req := httptest.NewRequest(http.MethodGet, "/api/tutorial/scenarios?group=cdn", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var cdnList []port.Scenario
	err := json.Unmarshal(rec.Body.Bytes(), &cdnList)
	require.NoError(t, err)
	assert.NotEmpty(t, cdnList)
	for _, s := range cdnList {
		assert.Equal(t, "cdn", s.Group)
	}

	// Filter by group=waap
	reqW := httptest.NewRequest(http.MethodGet, "/api/tutorial/scenarios?group=waap", nil)
	recW := httptest.NewRecorder()
	r.ServeHTTP(recW, reqW)

	assert.Equal(t, http.StatusOK, recW.Code)
	var waapList []port.Scenario
	err = json.Unmarshal(recW.Body.Bytes(), &waapList)
	require.NoError(t, err)
	assert.NotEmpty(t, waapList)
	for _, s := range waapList {
		assert.Equal(t, "waap", s.Group)
	}
}

func TestTutorialHandler_GetScenario(t *testing.T) {
	_, r := setupTutorialHandlerTest(t)

	// Get C1
	req := httptest.NewRequest(http.MethodGet, "/api/tutorial/scenarios/C1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "public, max-age=60", rec.Header().Get("Cache-Control"))
	etag := rec.Header().Get("ETag")
	assert.NotEmpty(t, etag)

	var s port.Scenario
	err := json.Unmarshal(rec.Body.Bytes(), &s)
	require.NoError(t, err)
	assert.Equal(t, "C1", s.ID)
	assert.Equal(t, "cdn", s.Group)

	// 304 Not Modified
	req304 := httptest.NewRequest(http.MethodGet, "/api/tutorial/scenarios/C1", nil)
	req304.Header.Set("If-None-Match", etag)
	rec304 := httptest.NewRecorder()
	r.ServeHTTP(rec304, req304)
	assert.Equal(t, http.StatusNotModified, rec304.Code)

	// Get non-existent
	reqNotFound := httptest.NewRequest(http.MethodGet, "/api/tutorial/scenarios/unknown_id", nil)
	recNotFound := httptest.NewRecorder()
	r.ServeHTTP(recNotFound, reqNotFound)
	assert.Equal(t, http.StatusNotFound, recNotFound.Code)
}
