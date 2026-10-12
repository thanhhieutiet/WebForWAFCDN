package tutorial

import (
	"context"
	"errors"
	"testing"

	"labshop/internal/domain/errs"
	"labshop/internal/usecase/port"
)

type mockScenarioRepo struct {
	scenarios []port.Scenario
	err       error
}

func (m *mockScenarioRepo) ListScenarios() ([]port.Scenario, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.scenarios, nil
}

func (m *mockScenarioRepo) GetScenario(id string) (*port.Scenario, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, s := range m.scenarios {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, errs.NewNotFound("scenario", "not found")
}

func TestTutorialUsecase_ListScenarios(t *testing.T) {
	fakeData := []port.Scenario{
		{ID: "C1", Title: "Onboard site", Group: "cdn", Priority: "P0"},
		{ID: "W1", Title: "Onboard WAAP", Group: "waap", Priority: "P0"},
	}

	repo := &mockScenarioRepo{scenarios: fakeData}
	uc := New(repo)

	ctx := context.Background()
	res, err := uc.ListScenarios(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 scenarios, got %d", len(res))
	}
	if res[0].ID != "C1" {
		t.Fatalf("expected C1, got %s", res[0].ID)
	}

	// Error case
	repoErr := &mockScenarioRepo{err: errors.New("db error")}
	ucErr := New(repoErr)
	_, err = ucErr.ListScenarios(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTutorialUsecase_GetScenario(t *testing.T) {
	fakeData := []port.Scenario{
		{ID: "C1", Title: "Onboard site", Group: "cdn", Priority: "P0"},
	}

	repo := &mockScenarioRepo{scenarios: fakeData}
	uc := New(repo)

	ctx := context.Background()
	// Validation error for empty ID
	_, err := uc.GetScenario(ctx, "   ")
	if err == nil {
		t.Fatal("expected error for empty ID, got nil")
	}
	de, ok := errs.AsDomainError(err)
	if !ok || de.Kind != errs.KindValidation {
		t.Fatalf("expected validation error, got: %v", err)
	}

	// Found
	s, err := uc.GetScenario(ctx, "C1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != "C1" {
		t.Fatalf("expected C1, got %s", s.ID)
	}

	// Not Found
	_, err = uc.GetScenario(ctx, "unknown")
	if err == nil {
		t.Fatal("expected error for unknown ID, got nil")
	}
	de, ok = errs.AsDomainError(err)
	if !ok || de.Kind != errs.KindNotFound {
		t.Fatalf("expected not found error, got: %v", err)
	}
}
