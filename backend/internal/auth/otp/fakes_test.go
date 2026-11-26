package otp

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

const al = "+355691234567"

var testLimits = Limits{
	Resend: Limit{1, 30 * time.Second}, PerPhone: Limit{5, 10 * time.Minute}, PerDevice: Limit{5, 10 * time.Minute},
	PerRange: Limit{100, 10 * time.Minute}, RangeDigits: 7, PumpingBlockAt: 90,
}

var strictLimits = Limits{
	Resend: Limit{1, time.Minute}, PerPhone: Limit{2, 10 * time.Minute}, PerDevice: Limit{2, 10 * time.Minute},
	PerRange: Limit{10, 10 * time.Minute}, RangeDigits: 7, PumpingBlockAt: 75,
}

type fakeTickets struct{}

func (fakeTickets) Validate(_ context.Context, t string) (Device, error) {
	if t != "good-ticket" {
		return Device{}, errors.New("bad ticket")
	}
	return Device{ID: "dev-1", Platform: "android"}, nil
}

type fakeIntel map[string][2]any

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
	sent map[string]string
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

type rig struct {
	svc                   *Service
	sms1, sms2, wa, viber *fakeSender
}

func newRig(t *testing.T, mut func(*Config)) rig {
	t.Helper()
	r := rig{sms1: &fakeSender{name: "sms-a"}, sms2: &fakeSender{name: "sms-b"}, wa: &fakeSender{name: "whatsapp"}, viber: &fakeSender{name: "viber"}}
	cfg := Config{
		Pepper: []byte(strings.Repeat("p", 32)), CoreRegions: []string{"AL", "XK", "MK"},
		CoreLimits: testLimits, OtherLimits: strictLimits, Failures: NewMemFailures(),
		Channels: map[Channel][]Sender{ChannelSMS: {r.sms1, r.sms2}, ChannelWhatsApp: {r.wa}, ChannelViber: {r.viber}},
		Tickets:  fakeTickets{}, Limiter: NewMemLimiter(), Intel: fakeIntel{},
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
