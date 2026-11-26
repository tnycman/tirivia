package otptest

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/auth/otp"
)

// LimiterWindow is short so RunLimiter can wait for expiry.
const LimiterWindow = 300 * time.Millisecond

// RunLimiter checks l: limit per key per window, keys independent, the window resets.
func RunLimiter(t *testing.T, l otp.Limiter) {
	t.Helper()
	ctx := context.Background()
	key, other := "test:"+hash(t), "test:"+hash(t)
	for i := 1; i <= 4; i++ {
		ok, err := l.Allow(ctx, key, 3, LimiterWindow)
		if err != nil || ok != (i <= 3) {
			t.Fatalf("call %d: allowed=%v err=%v", i, ok, err)
		}
	}
	if ok, err := l.Allow(ctx, other, 3, LimiterWindow); err != nil || !ok {
		t.Fatalf("independent key limited: %v %v", ok, err)
	}
	time.Sleep(LimiterWindow + 200*time.Millisecond)
	if ok, err := l.Allow(ctx, key, 3, LimiterWindow); err != nil || !ok {
		t.Fatalf("window did not reset: %v %v", ok, err)
	}
}

// RunLimiterRace checks that concurrent calls never exceed the limit.
func RunLimiterRace(t *testing.T, l otp.Limiter) {
	t.Helper()
	const calls, limit = 50, 10
	key := "test:" + hash(t)
	var allowed atomic.Int32
	var wg sync.WaitGroup
	for range calls {
		wg.Go(func() {
			if ok, err := l.Allow(context.Background(), key, limit, time.Minute); err == nil && ok {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()
	if allowed.Load() != limit {
		t.Fatalf("%d of %d concurrent calls allowed; want %d", allowed.Load(), calls, limit)
	}
}
