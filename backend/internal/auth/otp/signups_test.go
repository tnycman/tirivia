package otp

import (
	"context"
	"errors"
	"testing"
)

// ID-8: while sign-ups are paused, each probe counts against the device limit.
func TestPausedProbesRateLimited(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, func(c *Config) { c.Signups = flag(false) })
	var last error
	for i := range testLimits.PerDevice.Max + 1 {
		_, last = r.svc.Start(ctx, "+35569100000"+string(rune('0'+i)), "good-ticket", ChannelSMS, "sq")
	}
	if !errors.Is(last, ErrRateLimited) {
		t.Fatalf("paused probes skip the device limit: %v", last)
	}
}
