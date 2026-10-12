package port

// HitCounter tracks request counts per endpoint path.
type HitCounter interface {
	// Increment increases the hit count for a specific path.
	Increment(path string)

	// GetStats returns a snapshot copy of request counts per path since startup.
	GetStats() map[string]int64

	// Reset clears all recorded hit counts.
	Reset()
}
