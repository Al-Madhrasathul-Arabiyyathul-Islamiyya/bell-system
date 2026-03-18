package clock

import (
	"fmt"
	"time"

	_ "time/tzdata" // embed IANA timezone database for portability
)

// Clock provides timezone-aware time for the bell schedule system.
type Clock struct {
	loc     *time.Location
	nowFunc func() time.Time
}

// New creates a Clock for the given IANA timezone name.
// If tzName is empty, defaults to "Indian/Maldives".
func New(tzName string) (*Clock, error) {
	if tzName == "" {
		tzName = "Indian/Maldives"
	}
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone %q: %w", tzName, err)
	}
	return &Clock{loc: loc, nowFunc: time.Now}, nil
}

// NewTest creates a Clock with UTC location and a custom nowFunc, for testing.
func NewTest(nowFunc func() time.Time) *Clock {
	return &Clock{loc: time.UTC, nowFunc: nowFunc}
}

// Now returns the current time in the configured timezone.
func (c *Clock) Now() time.Time {
	return c.nowFunc().In(c.loc)
}

// Location returns the configured timezone.
func (c *Clock) Location() *time.Location {
	return c.loc
}
