package domain

import (
	"testing"
	"time"
)

func TestFocusLossVoids(t *testing.T) {
	for _, tc := range []struct {
		q          int
		background bool
		inactive   time.Duration
		want       bool
	}{
		{9, true, 5 * time.Second, false},
		{10, false, 1500 * time.Millisecond, false},
		{10, false, 1501 * time.Millisecond, true},
		{11, true, 0, true},
		{12, false, 200 * time.Millisecond, false},
		{12, false, 2 * time.Second, true},
	} {
		if got := FocusLossVoids(tc.q, tc.background, tc.inactive); got != tc.want {
			t.Errorf("Q%d background=%v inactive=%v: got %v", tc.q, tc.background, tc.inactive, got)
		}
	}
}
