package embedded

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"labshop/internal/domain/errs"
	"labshop/internal/usecase/port"
)

//go:embed scenarios.json
var scenariosData []byte

// Repository implements port.ScenarioRepository backed by embedded JSON data.
type Repository struct {
	scenarios []port.Scenario
	byID      map[string]port.Scenario
	once      sync.Once
	initErr   error
}

// NewRepository creates and initializes a new embedded ScenarioRepository.
func NewRepository() (*Repository, error) {
	r := &Repository{
		byID: make(map[string]port.Scenario),
	}
	if err := r.load(); err != nil {
		return nil, err
	}
	return r, nil
}

// MustNewRepository creates a new Repository or panics if loading fails.
func MustNewRepository() *Repository {
	r, err := NewRepository()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize embedded scenarios repository: %v", err))
	}
	return r
}

func (r *Repository) load() error {
	r.once.Do(func() {
		var list []port.Scenario
		if err := json.Unmarshal(scenariosData, &list); err != nil {
			r.initErr = fmt.Errorf("failed to unmarshal embedded scenarios: %w", err)
			return
		}
		r.scenarios = list
		for _, s := range list {
			r.byID[s.ID] = s
			r.byID[strings.ToLower(s.ID)] = s
			r.byID[strings.ToUpper(s.ID)] = s
			if strings.EqualFold(s.ID, "S1") {
				r.byID["shop-flow"] = s
				r.byID["shop"] = s
			}
		}
	})
	return r.initErr
}

// ListScenarios returns all embedded tutorial scenarios.
func (r *Repository) ListScenarios() ([]port.Scenario, error) {
	if r.initErr != nil {
		return nil, r.initErr
	}
	out := make([]port.Scenario, len(r.scenarios))
	copy(out, r.scenarios)
	return out, nil
}

// GetScenario returns a single scenario by its ID (supports case-insensitive matching and aliases).
func (r *Repository) GetScenario(id string) (*port.Scenario, error) {
	if r.initErr != nil {
		return nil, r.initErr
	}
	trimmed := strings.TrimSpace(id)
	if s, ok := r.byID[trimmed]; ok {
		cp := s
		return &cp, nil
	}
	if s, ok := r.byID[strings.ToLower(trimmed)]; ok {
		cp := s
		return &cp, nil
	}
	return nil, errs.NewNotFound("scenario", fmt.Sprintf("scenario '%s' not found", id))
}
