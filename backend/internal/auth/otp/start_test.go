package otp

import (
	"context"
	"errors"
	"regexp"
	"testing"
)

func TestHappyPathUserChosenChannel(t *testing.T) {
	r := newRig(t, nil)
	res, err := r.svc.Start(context.Background(), al, "good-ticket", ChannelWhatsApp, "sq")
	if err != nil || res.ChannelUsed != ChannelWhatsApp || res.Provider != "whatsapp" {
		t.Fatalf("start: %+v %v", res, err)
	}
	if !regexp.MustCompile(`^[0-9]{6}$`).MatchString(r.wa.sent[al]) {
		t.Fatalf("code %q is not 6 digits", r.wa.sent[al])
	}
}

func TestFailover(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	r.wa.fail, r.sms1.fail = true, true
	res, err := r.svc.Start(ctx, al, "good-ticket", ChannelWhatsApp, "sq")
	if err != nil || res.ChannelUsed != ChannelSMS || res.Provider != "sms-b" {
		t.Fatalf("expected second SMS provider: %+v %v", res, err)
	}
	r.sms2.fail = true
	if _, err := r.svc.Start(ctx, "+355697654321", "good-ticket", ChannelSMS, "sq"); !errors.Is(err, ErrSendFailed) {
		t.Fatalf("all down: %v", err)
	}
}

func TestPreSendChecks(t *testing.T) {
	intel := fakeIntel{"+355690000001": {"nonFixedVoip", 0}, "+355690000002": {"premium", 0}, "+355690000003": {"mobile", 90}}
	r := newRig(t, func(c *Config) { c.Intel = intel })
	for name, tc := range map[string]struct {
		phone, ticket string
		want          error
	}{
		"no attestation":    {al, "", ErrAttestationRequired},
		"forged ticket":     {al, "forged", ErrAttestationRequired},
		"not E.164":         {"0691234567", "good-ticket", ErrNumberNotAllowed},
		"invalid number":    {"+355123", "good-ticket", ErrNumberNotAllowed},
		"landline":          {"+35542123456", "good-ticket", ErrNumberNotAllowed},
		"non-canonical":     {"+355 69 123 4567", "good-ticket", ErrNumberNotAllowed},
		"VoIP":              {"+355690000001", "good-ticket", ErrNumberNotAllowed},
		"premium":           {"+355690000002", "good-ticket", ErrNumberNotAllowed},
		"high pumping risk": {"+355690000003", "good-ticket", ErrNumberNotAllowed},
	} {
		if _, err := r.svc.Start(context.Background(), tc.phone, tc.ticket, ChannelSMS, "sq"); !errors.Is(err, tc.want) {
			t.Errorf("%s: got %v want %v", name, err, tc.want)
		}
	}
	if len(r.sms1.sent) != 0 {
		t.Fatalf("code sent despite failed checks: %v", r.sms1.sent)
	}
}

func TestRateLimits(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	if _, err := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("resend inside window allowed: %v", err)
	}
	r2 := newRig(t, nil)
	var last error
	for i := range testLimits.PerDevice.Max + 1 {
		_, last = r2.svc.Start(ctx, "+35569100000"+string(rune('0'+i)), "good-ticket", ChannelSMS, "sq")
	}
	if !errors.Is(last, ErrRateLimited) {
		t.Fatalf("device limit not enforced: %v", last)
	}
}

func TestSignupsPaused(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, func(c *Config) { c.Signups = flag(false) })
	if _, err := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq"); !errors.Is(err, ErrSignupsPaused) {
		t.Fatalf("new number while paused: %v", err)
	}
	r.svc.cfg.Directory = dir{r.svc.PhoneHash(al): true}
	if _, err := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq"); err != nil {
		t.Fatalf("existing user blocked: %v", err)
	}
}

func TestConfigRequired(t *testing.T) {
	if _, err := New(Config{Pepper: make([]byte, 32), Channels: map[Channel][]Sender{ChannelSMS: {&fakeSender{}}}}); err == nil {
		t.Fatal("missing limits accepted")
	}
}

func TestAllCountriesWithTieredLimits(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	other := "+12025550123"
	for i := range strictLimits.PerPhone.Max {
		r.svc.cfg.Limiter.(*MemLimiter).m = map[string]window{}
		if _, err := r.svc.Start(ctx, other, "good-ticket", ChannelSMS, "en"); err != nil {
			t.Fatalf("non-core country send %d refused: %v", i, err)
		}
	}
	r2 := newRig(t, func(c *Config) { c.Intel = fakeIntel{other: {"mobile", strictLimits.PumpingBlockAt}} })
	if _, err := r2.svc.Start(ctx, other, "good-ticket", ChannelSMS, "en"); !errors.Is(err, ErrNumberNotAllowed) {
		t.Fatalf("stricter pumping threshold not applied outside core regions: %v", err)
	}
	r3 := newRig(t, func(c *Config) { c.Intel = fakeIntel{al: {"mobile", strictLimits.PumpingBlockAt}} })
	if _, err := r3.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq"); err != nil {
		t.Fatalf("core region should use the core threshold: %v", err)
	}
}
