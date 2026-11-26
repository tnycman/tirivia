package otp

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeTickets struct{}

func (fakeTickets) Validate(_ context.Context, t string) (Device, error) {
	if t != "good-ticket" {
		return Device{}, errors.New("bad ticket")
	}
	return Device{ID: "dev-1", Platform: "android"}, nil
}

type fakeIntel map[string][2]any // phone -> {lineType, risk}

func (f fakeIntel) Lookup(_ context.Context, p string) (string, int, error) {
	if v, ok := f[p]; ok {
		return v[0].(string), v[1].(int), nil
	}
	return "mobile", 5, nil
}

type fakeSender struct {
	name string
	fail bool
	mu   sync.Mutex
	sent map[string]string // phone -> last code
}

func (f *fakeSender) Name() string { return f.name }
func (f *fakeSender) Send(_ context.Context, phone, code, _ string) error {
	if f.fail {
		return errors.New("provider down")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.sent == nil {
		f.sent = map[string]string{}
	}
	f.sent[phone] = code
	return nil
}

type flag bool

func (f flag) Enabled(context.Context) bool { return bool(f) }

type dir map[string]bool

func (d dir) Exists(_ context.Context, h string) (bool, error) { return d[h], nil }

const al = "+355691234567" // Albanian mobile format

type rig struct {
	svc                   *Service
	sms1, sms2, wa, viber *fakeSender
}

func newRig(t *testing.T, mut func(*Config)) rig {
	t.Helper()
	r := rig{sms1: &fakeSender{name: "sms-a"}, sms2: &fakeSender{name: "sms-b"}, wa: &fakeSender{name: "whatsapp"}, viber: &fakeSender{name: "viber"}}
	cfg := Config{
		Pepper:          []byte(strings.Repeat("p", 32)),
		AllowedPrefixes: []string{"+355", "+383", "+389", "+39", "+30", "+49", "+41", "+44", "+1"},
		Channels:        map[Channel][]Sender{ChannelSMS: {r.sms1, r.sms2}, ChannelWhatsApp: {r.wa}, ChannelViber: {r.viber}},
		Tickets:         fakeTickets{}, Limiter: NewMemLimiter(), Intel: fakeIntel{},
		Signups: flag(true), Directory: dir{}, Store: NewMemStore(),
	}
	if mut != nil {
		mut(&cfg)
	}
	svc, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	r.svc = svc
	return r
}

func TestHappyPathUserChosenChannel(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	res, err := r.svc.Start(ctx, al, "good-ticket", ChannelWhatsApp, "sq")
	if err != nil || res.ChannelUsed != ChannelWhatsApp || res.Provider != "whatsapp" {
		t.Fatalf("start: %+v %v", res, err)
	}
	code := r.wa.sent[al]
	if !regexp.MustCompile(`^[0-9]{6}$`).MatchString(code) {
		t.Fatalf("code %q is not 6 digits", code)
	}
	ch, err := r.svc.Verify(ctx, res.RequestID, code)
	if err != nil || ch.PhoneE164 != al || ch.Device.ID != "dev-1" {
		t.Fatalf("verify: %v", err)
	}
	if _, err := r.svc.Verify(ctx, res.RequestID, code); !errors.Is(err, ErrExpired) {
		t.Fatalf("code reusable: %v", err)
	}
}

func TestFailoverToSecondSMSAndFromWhatsApp(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	r.wa.fail, r.sms1.fail = true, true
	res, err := r.svc.Start(ctx, al, "good-ticket", ChannelWhatsApp, "sq")
	if err != nil || res.ChannelUsed != ChannelSMS || res.Provider != "sms-b" {
		t.Fatalf("expected fallback to second SMS provider: %+v %v", res, err)
	}
	r.sms2.fail = true
	if _, err := r.svc.Start(ctx, "+355697654321", "good-ticket", ChannelSMS, "sq"); !errors.Is(err, ErrSendFailed) {
		t.Fatalf("all down: %v", err)
	}
}

func TestPreSendChecks(t *testing.T) {
	ctx := context.Background()
	intel := fakeIntel{"+355690000001": {"nonFixedVoip", 0}, "+355690000002": {"premium", 0}, "+355690000003": {"mobile", 95}}
	r := newRig(t, func(c *Config) { c.Intel = intel })
	for name, tc := range map[string]struct {
		phone, ticket string
		want          error
	}{
		"no attestation":      {al, "", ErrAttestationRequired},
		"forged ticket":       {al, "forged", ErrAttestationRequired},
		"not E.164":           {"0691234567", "good-ticket", ErrNumberNotAllowed},
		"country not allowed": {"+79991234567", "good-ticket", ErrNumberNotAllowed},
		"VoIP number":         {"+355690000001", "good-ticket", ErrNumberNotAllowed},
		"premium number":      {"+355690000002", "good-ticket", ErrNumberNotAllowed},
		"high pumping risk":   {"+355690000003", "good-ticket", ErrNumberNotAllowed},
	} {
		if _, err := r.svc.Start(ctx, tc.phone, tc.ticket, ChannelSMS, "sq"); !errors.Is(err, tc.want) {
			t.Errorf("%s: got %v want %v", name, err, tc.want)
		}
	}
	if len(r.sms1.sent) != 0 {
		t.Fatalf("a code was sent despite failed checks: %v", r.sms1.sent)
	}
}

func TestRateLimits(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	if _, err := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("resend inside 30 s allowed: %v", err)
	}
	// Device limit: 5 per hour across different numbers.
	r2 := newRig(t, nil)
	var last error
	for i := range 6 {
		_, last = r2.svc.Start(ctx, "+35569100000"+string(rune('0'+i)), "good-ticket", ChannelSMS, "sq")
	}
	if !errors.Is(last, ErrRateLimited) {
		t.Fatalf("6th code from one device in an hour allowed: %v", last)
	}
}

func TestSignupsPausedStillLetsExistingUsersIn(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, func(c *Config) { c.Signups = flag(false) })
	if _, err := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq"); !errors.Is(err, ErrSignupsPaused) {
		t.Fatalf("new number while paused: %v", err)
	}
	existing := "+355692222222"
	r2 := newRig(t, func(c *Config) { c.Signups = flag(false) })
	r2.svc.cfg.Directory = dir{r2.svc.PhoneHash(existing): true}
	if _, err := r2.svc.Start(ctx, existing, "good-ticket", ChannelSMS, "sq"); err != nil {
		t.Fatalf("existing user blocked by sign-up pause: %v", err)
	}
}

func TestAttemptsAndExpiry(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	res, _ := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq")
	right := r.sms1.sent[al]
	wrong := "000000"
	if right == wrong {
		wrong = "111111"
	}
	for i := 1; i <= MaxAttempts; i++ {
		if _, err := r.svc.Verify(ctx, res.RequestID, wrong); !errors.Is(err, ErrWrongCode) {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	if _, err := r.svc.Verify(ctx, res.RequestID, right); err == nil {
		t.Fatal("correct code accepted after 5 failed attempts")
	}

	r2 := newRig(t, nil)
	res2, _ := r2.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq")
	r2.svc.now = func() time.Time { return time.Now().Add(CodeTTL + time.Second) }
	if _, err := r2.svc.Verify(ctx, res2.RequestID, r2.sms1.sent[al]); !errors.Is(err, ErrExpired) {
		t.Fatalf("expired code accepted: %v", err)
	}
}

// 20 parallel guesses must never exceed MaxAttempts comparisons.
func TestParallelGuessingIsCapped(t *testing.T) {
	ctx := context.Background()
	r := newRig(t, nil)
	res, _ := r.svc.Start(ctx, al, "good-ticket", ChannelSMS, "sq")
	var wg sync.WaitGroup
	var mu sync.Mutex
	wrongCount := 0
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := r.svc.Verify(ctx, res.RequestID, "999999x")
			if errors.Is(err, ErrWrongCode) {
				mu.Lock()
				wrongCount++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if wrongCount > MaxAttempts {
		t.Fatalf("%d guesses were compared; max is %d", wrongCount, MaxAttempts)
	}
}

func TestCodesLookRandom(t *testing.T) {
	seen := map[string]bool{}
	for range 2000 {
		c, err := newCode()
		if err != nil || len(c) != CodeDigits {
			t.Fatalf("bad code %q %v", c, err)
		}
		seen[c] = true
	}
	if len(seen) < 1990 { // birthday bound for 2000 draws from 10^6 is ~2 collisions
		t.Fatalf("only %d distinct codes in 2000 draws", len(seen))
	}
}
