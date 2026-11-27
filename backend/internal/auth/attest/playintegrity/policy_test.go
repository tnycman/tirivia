package playintegrity

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

var (
	now    = time.UnixMilli(1_800_000_000_000)
	expect = Expect{PackageName: "al.tirivia.app", CertDigests: []string{"cert-a"}, Window: 5 * time.Minute, FutureSkew: 30 * time.Second}
)

func good(hash string) Verdict {
	return Verdict{HasRequest: true, RequestPackageName: expect.PackageName, RequestHash: hash, TimestampMillis: now.UnixMilli(),
		AppRecognition: AppPlayRecognized, PackageName: expect.PackageName, CertDigests: []string{"cert-a"},
		Licensing: Licensed, DeviceLabels: []string{MeetsDeviceIntegrity}}
}

// Expected value computed independently with Node crypto: sha256(32×0x01 ‖ "jkt-example"), base64url.
func TestRequestHashVector(t *testing.T) {
	got := RequestHash(bytes.Repeat([]byte{0x01}, 32), "jkt-example")
	if got != "-maAfZhsG0rRTCGoV5bozn4vcnKO4dQjvybi10jjoMQ" || len(got) != 43 {
		t.Fatalf("vector changed: %s", got)
	}
}

func TestCheckRequest(t *testing.T) {
	cases := map[string]func(*Verdict){
		"no request details": func(v *Verdict) { v.HasRequest = false },
		"other package":      func(v *Verdict) { v.RequestPackageName = "com.evil" },
		"other request":      func(v *Verdict) { v.RequestHash = "other" },
		"stale":              func(v *Verdict) { v.TimestampMillis = now.Add(-6 * time.Minute).UnixMilli() },
		"from the future":    func(v *Verdict) { v.TimestampMillis = now.Add(time.Minute).UnixMilli() },
	}
	if err := CheckRequest(good("h"), expect, "h", now); err != nil {
		t.Fatalf("good verdict rejected: %v", err)
	}
	for name, mutate := range cases {
		v := good("h")
		mutate(&v)
		if err := CheckRequest(v, expect, "h", now); !errors.Is(err, ErrRequest) {
			t.Errorf("%s: got %v", name, err)
		}
	}
}

func TestPrizeEligible(t *testing.T) {
	weak := map[string]func(*Verdict){
		"unrecognized app":   func(v *Verdict) { v.AppRecognition = "UNRECOGNIZED_VERSION" },
		"unevaluated app":    func(v *Verdict) { v.AppRecognition = "UNEVALUATED" },
		"other package":      func(v *Verdict) { v.PackageName = "com.evil" },
		"unknown cert":       func(v *Verdict) { v.CertDigests = []string{"cert-a", "cert-x"} },
		"no cert":            func(v *Verdict) { v.CertDigests = nil },
		"unlicensed":         func(v *Verdict) { v.Licensing = "UNLICENSED" },
		"basic only":         func(v *Verdict) { v.DeviceLabels = []string{"MEETS_BASIC_INTEGRITY"} },
		"no device labels":   func(v *Verdict) { v.DeviceLabels = nil },
		"testing response":   func(v *Verdict) { v.Testing = true },
		"replayed (cleared)": func(v *Verdict) { v.DeviceLabels, v.AppRecognition, v.Licensing = nil, "UNEVALUATED", "UNEVALUATED" },
	}
	if !PrizeEligible(good("h"), expect) {
		t.Fatal("good verdict must be prize eligible")
	}
	for name, mutate := range weak {
		v := good("h")
		mutate(&v)
		if PrizeEligible(v, expect) {
			t.Errorf("%s: must be points only", name)
		}
	}
	v := good("h")
	v.Testing = true
	withTesting := expect
	withTesting.AllowTesting = true
	if !PrizeEligible(v, withTesting) {
		t.Error("testing responses allowed when configured")
	}
}
