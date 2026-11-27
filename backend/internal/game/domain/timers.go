package domain

import (
	"errors"
	"fmt"
	"time"
)

// Producer-set timer bounds (decision UF-3).
const (
	MinTimer = 6 * time.Second
	MaxTimer = 10 * time.Second
)

// ErrTimerOutOfBounds rejects a timer outside [MinTimer, MaxTimer].
var ErrTimerOutOfBounds = errors.New("domain: question timer must be between 6 s and 10 s")

// ValidateTimer checks a producer-set timer.
func ValidateTimer(d time.Duration) error {
	if d < MinTimer || d > MaxTimer {
		return fmt.Errorf("%w (got %v)", ErrTimerOutOfBounds, d)
	}
	return nil
}

// TimerFor is the default the studio console pre-fills.
func TimerFor(index int) time.Duration {
	switch {
	case index <= 4:
		return 10 * time.Second
	case index <= 8:
		return 8 * time.Second
	default:
		return 7 * time.Second
	}
}
