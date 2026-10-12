package port

import "time"

// Clock provides current timestamps, enabling deterministic tests.
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using time.Now().
type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}
