package playintegrity

import (
	"crypto/subtle"
	"errors"
	"slices"
	"time"
)

// ErrRequest means the token was not issued for this request; it is rejected outright.
var ErrRequest = errors.New("playintegrity: request details do not match")

// Expect is what a token must match; Window is the attestation challenge TTL, FutureSkew the JWT leeway.
type Expect struct {
	PackageName  string
	CertDigests  []string
	Window       time.Duration
	FutureSkew   time.Duration
	AllowTesting bool
}

// CheckRequest applies Google's first step: package, request hash and timestamp must match the request.
func CheckRequest(v Verdict, e Expect, wantHash string, now time.Time) error {
	ts := time.UnixMilli(v.TimestampMillis)
	switch {
	case !v.HasRequest, v.RequestPackageName != e.PackageName,
		subtle.ConstantTimeCompare([]byte(v.RequestHash), []byte(wantHash)) != 1,
		now.Sub(ts) > e.Window, ts.Sub(now) > e.FutureSkew:
		return ErrRequest
	}
	return nil
}

// PrizeEligible applies Google's sample checks; anything weaker plays for points only (AC-5).
func PrizeEligible(v Verdict, e Expect) bool {
	certOK := len(v.CertDigests) > 0
	for _, d := range v.CertDigests {
		certOK = certOK && slices.Contains(e.CertDigests, d)
	}
	return v.AppRecognition == AppPlayRecognized && v.PackageName == e.PackageName && certOK &&
		v.Licensing == Licensed && slices.Contains(v.DeviceLabels, MeetsDeviceIntegrity) &&
		(e.AllowTesting || !v.Testing)
}
