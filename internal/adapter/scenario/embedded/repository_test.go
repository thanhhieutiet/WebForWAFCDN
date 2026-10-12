package embedded

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedRepository_ListAndGet(t *testing.T) {
	repo, err := NewRepository()
	require.NoError(t, err)
	require.NotNil(t, repo)

	scenarios, err := repo.ListScenarios()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(scenarios), 23, "must have at least 23 P0 scenarios")

	// Verify required CDN P0 scenarios
	cdnExpected := []string{"C1", "C2", "C2b", "C3", "C4", "C5", "C6", "C7", "C8", "C9", "C10", "C16", "C17", "C18"}
	for _, id := range cdnExpected {
		s, err := repo.GetScenario(id)
		require.NoError(t, err, "expected scenario %s to exist", id)
		assert.Equal(t, id, s.ID)
		assert.Equal(t, "cdn", s.Group)
		assert.Equal(t, "P0", s.Priority)
		assert.NotEmpty(t, s.Title)
		assert.NotEmpty(t, s.Goal)
		assert.NotEmpty(t, s.PortalSteps)
		assert.NotEmpty(t, s.Requests)
		assert.NotEmpty(t, s.Expect)
		assert.NotEmpty(t, s.DocsURL)
	}

	// Verify required WAAP P0 scenarios
	waapExpected := []string{"W1", "W2", "W3", "W4", "W5", "W6", "W7", "W8"}
	for _, id := range waapExpected {
		s, err := repo.GetScenario(id)
		require.NoError(t, err, "expected scenario %s to exist", id)
		assert.Equal(t, id, s.ID)
		assert.Equal(t, "waap", s.Group)
		assert.Equal(t, "P0", s.Priority)
		assert.NotEmpty(t, s.Title)
		assert.NotEmpty(t, s.Goal)
		assert.NotEmpty(t, s.PortalSteps)
		assert.NotEmpty(t, s.Requests)
		assert.NotEmpty(t, s.Expect)
		assert.NotEmpty(t, s.DocsURL)
	}

	// Verify required CDN & WAAP P1 scenarios
	p1Expected := map[string]string{
		"C11": "cdn", "C12": "cdn", "C13": "cdn", "C14": "cdn", "C15": "cdn",
		"W9": "waap", "W10": "waap", "W11": "waap", "W12": "waap", "W13": "waap",
	}
	for id, grp := range p1Expected {
		s, err := repo.GetScenario(id)
		require.NoError(t, err, "expected P1 scenario %s to exist", id)
		assert.Equal(t, id, s.ID)
		assert.Equal(t, grp, s.Group)
		assert.Equal(t, "P1", s.Priority)
		assert.NotEmpty(t, s.Title)
		assert.NotEmpty(t, s.Goal)
		assert.NotEmpty(t, s.PortalSteps)
		assert.NotEmpty(t, s.Requests)
		assert.NotEmpty(t, s.Expect)
		assert.NotEmpty(t, s.DocsURL)
	}

	// Verify required Shop P0 scenario
	shopScenario, err := repo.GetScenario("S1")
	require.NoError(t, err)
	assert.Equal(t, "S1", shopScenario.ID)
	assert.Equal(t, "shop", shopScenario.Group)
	assert.Equal(t, "P0", shopScenario.Priority)
	assert.NotEmpty(t, shopScenario.Requests)
	assert.GreaterOrEqual(t, len(shopScenario.Requests), 8)

	// Verify alias matching for shop
	aliasScenario, err := repo.GetScenario("shop-flow")
	require.NoError(t, err)
	assert.Equal(t, "S1", aliasScenario.ID)

	// Verify case-insensitive retrieval
	sLower, err := repo.GetScenario("c1")
	require.NoError(t, err)
	assert.Equal(t, "C1", sLower.ID)

	// Verify Not Found error
	notFound, err := repo.GetScenario("non_existent_id")
	assert.Error(t, err)
	assert.Nil(t, notFound)
}

func TestMustNewRepository(t *testing.T) {
	assert.NotPanics(t, func() {
		repo := MustNewRepository()
		assert.NotNil(t, repo)
	})
}
