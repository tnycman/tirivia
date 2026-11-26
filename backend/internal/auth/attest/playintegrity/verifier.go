package playintegrity

import (
	"context"
	"errors"
	"time"

	"github.com/tirivia/backend/internal/auth/attest"
)

// Decoder decrypts a standard-request token; only Google's servers can (standard requests page).
type Decoder interface {
	Decode(ctx context.Context, token string) (Verdict, error)
}

// Verifier implements attest.Verifier for Android.
type Verifier struct {
	dec Decoder
	exp Expect
	now func() time.Time
}

// New returns a Verifier.
func New(dec Decoder, exp Expect) (*Verifier, error) {
	if dec == nil || exp.PackageName == "" || len(exp.CertDigests) == 0 || exp.Window <= 0 || exp.FutureSkew < 0 {
		return nil, errors.New("playintegrity: decoder, package, certificate digests and window are required")
	}
	return &Verifier{dec: dec, exp: exp, now: time.Now}, nil
}

// Verify rejects undecodable or mismatched tokens; decoded but weak verdicts play for points only.
func (v *Verifier) Verify(ctx context.Context, challenge []byte, ev attest.Evidence) (attest.Result, error) {
	if ev.PlayIntegrityToken == "" || ev.JKT == "" {
		return attest.Result{}, attest.ErrFailed
	}
	verdict, err := v.dec.Decode(ctx, ev.PlayIntegrityToken)
	if err != nil {
		return attest.Result{}, err
	}
	if err := CheckRequest(verdict, v.exp, RequestHash(challenge, ev.JKT), v.now()); err != nil {
		return attest.Result{}, err
	}
	return attest.Result{DeviceRef: "android:" + ev.JKT, PrizeEligible: PrizeEligible(verdict, v.exp),
		Signals: Signals(verdict)}, nil
}
