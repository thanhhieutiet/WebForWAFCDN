package memory

import (
	"sync"

	"labshop/internal/usecase/port"
)

// MemoryHitCounter provides a thread-safe in-memory implementation of port.HitCounter.
type MemoryHitCounter struct {
	mu   sync.RWMutex
	hits map[string]int64
}

// New creates a new initialized MemoryHitCounter.
func New() *MemoryHitCounter {
	return &MemoryHitCounter{
		hits: make(map[string]int64),
	}
}

// Ensure MemoryHitCounter satisfies port.HitCounter interface.
var _ port.HitCounter = (*MemoryHitCounter)(nil)

// Increment atomically increments the request count for the specified path.
func (c *MemoryHitCounter) Increment(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hits[path]++
}

// GetStats returns a point-in-time snapshot of the path counts.
func (c *MemoryHitCounter) GetStats() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	snapshot := make(map[string]int64, len(c.hits))
	for k, v := range c.hits {
		snapshot[k] = v
	}
	return snapshot
}

// Reset clears all counters back to empty state.
func (c *MemoryHitCounter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hits = make(map[string]int64)
}
