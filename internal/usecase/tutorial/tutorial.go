package tutorial

import (
	"context"
	"strings"

	"labshop/internal/domain/errs"
	"labshop/internal/usecase/port"
)

// Usecase provides business logic for tutorial scenarios.
type Usecase struct {
	repo port.ScenarioRepository
}

// New constructs a new tutorial Usecase.
func New(repo port.ScenarioRepository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

// ListScenarios returns all available tutorial scenarios.
func (u *Usecase) ListScenarios(ctx context.Context) ([]port.Scenario, error) {
	return u.repo.ListScenarios()
}

// GetScenario returns a single scenario by its ID.
func (u *Usecase) GetScenario(ctx context.Context, id string) (*port.Scenario, error) {
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return nil, errs.NewValidation("scenario id is required", errs.FieldViolation{Field: "id", Issue: "cannot be empty"})
	}
	return u.repo.GetScenario(trimmed)
}
