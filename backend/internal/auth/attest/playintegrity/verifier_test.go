package playintegrity

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/tirivia/backend/internal/auth/attest"
)

type fakeDecoder struct {
	v   Verdict
	err error
}

func (f fakeDecoder) Decode(context.Context, string) (Verdict, error) { return f.v, f.err }

func verifier(t *testing.T, d Decoder) *Verifier {
	t.Helper()
	v, err := New(d, expect)
	if err != nil {
		t.Fatal(err)
	}
	v.now = func() time.Time { return now }
	return v
}

func TestVerify(t *testing.T) {
	challenge := make([]byte, attest.ChallengeBytes)
	ev := attest.Evidence{Platform: "android", PlayIntegrityToken: "tok", JKT: "jkt-1"}
	hash := RequestHash(challenge, ev.JKT)

	res, err := verifier(t, fakeDecoder{v: good(hash)}).Verify(context.Background(), challenge, ev)
	if err != nil || !res.PrizeEligible || res.DeviceRef != "android:jkt-1" {
		t.Fatalf("good: %+v %v", res, err)
	}

	weak := good(hash)
	weak.DeviceLabels = []string{"MEETS_BASIC_INTEGRITY"}
	res, err = verifier(t, fakeDecoder{v: weak}).Verify(context.Background(), challenge, ev)
	if err != nil || res.PrizeEligible || !slices.Contains(res.Signals, "pi.device_integrity=missing") {
		t.Fatalf("weak device must play for points with a signal: %+v %v", res, err)
	}

	other := ev
	other.JKT = "jkt-2"
	if _, err := verifier(t, fakeDecoder{v: good(hash)}).Verify(context.Background(), challenge, other); !errors.Is(err, ErrRequest) {
		t.Fatalf("token for another key accepted: %v", err)
	}
	boom := errors.New("decode failed")
	if _, err := verifier(t, fakeDecoder{err: boom}).Verify(context.Background(), challenge, ev); !errors.Is(err, boom) {
		t.Fatalf("decode error swallowed: %v", err)
	}
	if _, err := verifier(t, fakeDecoder{}).Verify(context.Background(), challenge, attest.Evidence{Platform: "android", JKT: "j"}); !errors.Is(err, attest.ErrFailed) {
		t.Fatalf("missing token accepted: %v", err)
	}
}

func TestNewRequiresConfig(t *testing.T) {
	if _, err := New(fakeDecoder{}, Expect{PackageName: "p", Window: time.Minute}); err == nil {
		t.Fatal("missing certificate digests accepted")
	}
}
