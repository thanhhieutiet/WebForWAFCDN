package memory

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryHitCounter_Basic(t *testing.T) {
	c := New()

	assert.Empty(t, c.GetStats())

	c.Increment("/api/health")
	c.Increment("/api/health")
	c.Increment("/api/time")

	stats := c.GetStats()
	assert.Equal(t, int64(2), stats["/api/health"])
	assert.Equal(t, int64(1), stats["/api/time"])
	assert.Equal(t, int64(0), stats["/nonexistent"])
}

func TestMemoryHitCounter_Concurrent(t *testing.T) {
	c := New()
	const numGoroutines = 50
	const incPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incPerGoroutine; j++ {
				c.Increment("/api/endpoint-a")
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < incPerGoroutine; j++ {
				c.Increment("/api/endpoint-b")
			}
		}()
	}

	wg.Wait()

	stats := c.GetStats()
	assert.Equal(t, int64(numGoroutines*incPerGoroutine), stats["/api/endpoint-a"])
	assert.Equal(t, int64(numGoroutines*incPerGoroutine), stats["/api/endpoint-b"])
}

func TestMemoryHitCounter_Reset(t *testing.T) {
	c := New()
	c.Increment("/api/health")
	c.Increment("/api/time")

	assert.Len(t, c.GetStats(), 2)

	c.Reset()

	stats := c.GetStats()
	assert.Empty(t, stats)
	assert.Equal(t, int64(0), stats["/api/health"])
}

func TestMemoryHitCounter_Isolation(t *testing.T) {
	c := New()
	c.Increment("/api/test")

	stats := c.GetStats()
	assert.Equal(t, int64(1), stats["/api/test"])

	// Mutate returned map
	stats["/api/test"] = 999
	stats["/api/injected"] = 100

	// Internal state should be unchanged
	internalStats := c.GetStats()
	assert.Equal(t, int64(1), internalStats["/api/test"])
	assert.NotContains(t, internalStats, "/api/injected")
}
