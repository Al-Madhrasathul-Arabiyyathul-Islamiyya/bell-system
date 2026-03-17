package clock

import (
	"testing"
	"time"
)

func TestNew_DefaultTimezone(t *testing.T) {
	c, err := New("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Location().String() != "Indian/Maldives" {
		t.Errorf("expected Indian/Maldives, got %s", c.Location().String())
	}
}

func TestNew_ExplicitTimezone(t *testing.T) {
	c, err := New("America/New_York")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Location().String() != "America/New_York" {
		t.Errorf("expected America/New_York, got %s", c.Location().String())
	}
}

func TestNew_InvalidTimezone(t *testing.T) {
	_, err := New("Invalid/Zone")
	if err == nil {
		t.Fatal("expected error for invalid timezone")
	}
}

func TestClock_Now_UsesConfiguredLocation(t *testing.T) {
	c, err := New("Indian/Maldives")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	now := c.Now()
	if now.Location().String() != "Indian/Maldives" {
		t.Errorf("expected Indian/Maldives location, got %s", now.Location().String())
	}
}

func TestNewTest_UsesCustomNowFunc(t *testing.T) {
	fixed := time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC)
	c := NewTest(func() time.Time { return fixed })

	got := c.Now()
	if !got.Equal(fixed) {
		t.Errorf("expected %v, got %v", fixed, got)
	}
	if got.Location() != time.UTC {
		t.Errorf("expected UTC, got %s", got.Location().String())
	}
}
