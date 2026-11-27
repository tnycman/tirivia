package otp

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestVerifySingleUse(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	res, _ := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq")
	ch, err := r.svc.Verify(ctx, res.RequestID, r.sms1.sent[al], "")
	if err != nil || ch.PhoneE164 != al || ch.Device.ID != "dev-1" {
		t.Fatalf("verify: %v", err)
	}
	if _, err := r.svc.Verify(ctx, res.RequestID, r.sms1.sent[al], ""); !errors.Is(err, ErrExpired) {
		t.Fatalf("code reusable: %v", err)
	}
}

func TestAttemptsCapped(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	res, _ := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq")
	right, wrong := r.sms1.sent[al], "000000"
	if right == wrong {
		wrong = "111111"
	}
	for i := 1; i <= MaxAttempts; i++ {
		if _, err := r.svc.Verify(ctx, res.RequestID, wrong, ""); !errors.Is(err, ErrWrongCode) {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	if _, err := r.svc.Verify(ctx, res.RequestID, right, ""); err == nil {
		t.Fatal("correct code accepted after max attempts")
	}
}

func TestExpiry(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	res, _ := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq")
	r.svc.now = func() time.Time { return time.Now().Add(CodeTTL + time.Second) }
	if _, err := r.svc.Verify(ctx, res.RequestID, r.sms1.sent[al], ""); !errors.Is(err, ErrExpired) {
		t.Fatalf("expired code accepted: %v", err)
	}
}

func TestParallelGuessingCapped(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	res, _ := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq")
	var wg sync.WaitGroup
	var mu sync.Mutex
	compared := 0
	for range 20 {
		wg.Go(func() {
			if _, err := r.svc.Verify(ctx, res.RequestID, "x", ""); errors.Is(err, ErrWrongCode) {
				mu.Lock()
				compared++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if compared > MaxAttempts {
		t.Fatalf("%d guesses compared; max %d", compared, MaxAttempts)
	}
}

func TestCodesUniform(t *testing.T) {
	seen := map[string]bool{}
	for range 2000 {
		c, err := newCode()
		if err != nil || len(c) != CodeDigits {
			t.Fatalf("bad code %q %v", c, err)
		}
		seen[c] = true
	}
	if len(seen) < 1990 {
		t.Fatalf("only %d distinct codes in 2000 draws", len(seen))
	}
}

func TestConsecutiveFailuresSurviveNewCodes(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	f := r.svc.cfg.Failures.(*MemFailures)
	f.m[r.svc.PhoneHash(al)] = MaxConsecutiveFailures - 1
	res, _ := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq")
	if _, err := r.svc.Verify(ctx, res.RequestID, "nope", ""); !errors.Is(err, ErrWrongCode) {
		t.Fatal(err)
	}
	r.svc.cfg.Limiter = NewMemLimiter()
	if _, err := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("new code allowed after %d consecutive failures: %v", MaxConsecutiveFailures, err)
	}
}

func TestSuccessResetsFailures(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	res, _ := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq")
	_, _ = r.svc.Verify(ctx, res.RequestID, "nope", "")
	if _, err := r.svc.Verify(ctx, res.RequestID, r.sms1.sent[al], ""); err != nil {
		t.Fatal(err)
	}
	if n, _ := r.svc.cfg.Failures.Count(ctx, r.svc.PhoneHash(al)); n != 0 {
		t.Fatalf("failures not reset on success: %d", n)
	}
}
