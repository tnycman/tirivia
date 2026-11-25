package cuesign

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"math"
	"strings"
	"testing"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
)

const uuidLike = "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b" // 36 chars, UUIDv7 text form

func newKeys(t *testing.T) (*Signer, *Verifier) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewSigner(7, priv)
	if err != nil {
		t.Fatal(err)
	}
	return s, NewVerifier(map[uint32]ed25519.PublicKey{7: pub})
}

// repeatToBytes returns s repeated until it is exactly n bytes (s must divide n).
func repeatToBytes(s string, n int) string {
	return strings.Repeat(s, n/len(s))
}

// worstCaseQuestionOpen builds the largest question_open cue the question bank allows:
// every character is a 2-byte UTF-8 letter (ë), ids are full UUID text, counters are huge.
func worstCaseQuestionOpen() *cuev1.CueBody {
	opts := make([]*cuev1.Option, OptionsPerQuestion)
	for i := range opts {
		opts[i] = &cuev1.Option{OptionId: "a", Text: repeatToBytes("ç", MaxOptionTextBytes)}
	}
	return &cuev1.CueBody{
		ShowId:          uuidLike,
		Seq:             math.MaxUint32,
		EmittedAtUnixMs: math.MaxInt64,
		Kind: &cuev1.CueBody_QuestionOpen{QuestionOpen: &cuev1.QuestionOpen{
			QuestionId:     uuidLike,
			Index:          12,
			DurationMs:     10_000,
			Text:           repeatToBytes("ë", MaxQuestionTextBytes),
			Options:        opts,
			IsFinal:        true,
			FocusLossVoids: true,
		}},
	}
}

func TestWorstCaseQuestionOpenFitsBudget(t *testing.T) {
	s, _ := newKeys(t)
	text, err := s.Sign(worstCaseQuestionOpen())
	if err != nil {
		t.Fatalf("worst-case question_open must fit the %d B budget: %v", MaxTextBytes, err)
	}
	t.Logf("worst-case question_open cue: %d bytes (budget %d, IVS hard limit 1 KB)", len(text), MaxTextBytes)
}

func TestWorstCasePulseFitsBudget(t *testing.T) {
	s, _ := newKeys(t)
	totals := map[int32]uint64{}
	for k := int32(1); k <= 7; k++ {
		totals[k] = math.MaxUint64
	}
	text, err := s.Sign(&cuev1.CueBody{
		ShowId: uuidLike, Seq: math.MaxUint32, EmittedAtUnixMs: math.MaxInt64,
		Kind: &cuev1.CueBody_Pulse{Pulse: &cuev1.Pulse{
			Viewers: math.MaxUint64, PlayersRemaining: math.MaxUint64, ReactionTotals: totals,
			Phase: cuev1.Phase_PHASE_QUESTION_OPEN, QuestionIndex: 12, RemainingMs: 10_000,
			ReactionsEnabled: true, ChatEnabled: true,
		}},
	})
	if err != nil {
		t.Fatalf("worst-case pulse must fit: %v", err)
	}
	t.Logf("worst-case pulse cue: %d bytes", len(text))
}

func TestRoundTrip(t *testing.T) {
	s, v := newKeys(t)
	in := worstCaseQuestionOpen()
	text, err := s.Sign(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := v.Verify(text)
	if err != nil {
		t.Fatal(err)
	}
	if out.GetQuestionOpen().GetText() != in.GetQuestionOpen().GetText() || out.GetSeq() != in.GetSeq() {
		t.Fatal("round trip changed the body")
	}
}

func TestRejects(t *testing.T) {
	s, v := newKeys(t)
	good, err := s.Sign(&cuev1.CueBody{ShowId: "s", Seq: 1, Kind: &cuev1.CueBody_QuestionClosed{QuestionClosed: &cuev1.QuestionClosed{QuestionId: "q"}}})
	if err != nil {
		t.Fatal(err)
	}
	_, otherPriv, _ := ed25519.GenerateKey(rand.Reader)
	other, _ := NewSigner(7, otherPriv) // same key id, different key: forged cue
	forged, _ := other.Sign(&cuev1.CueBody{ShowId: "s", Seq: 1})
	unknownKey, _ := NewSigner(99, otherPriv)
	unknown, _ := unknownKey.Sign(&cuev1.CueBody{ShowId: "s", Seq: 1})

	// Flip one character in the middle of the payload.
	b := []byte(good)
	mid := len(Prefix) + (len(b)-len(Prefix))/2
	if b[mid] == 'A' {
		b[mid] = 'B'
	} else {
		b[mid] = 'A'
	}
	tampered := string(b)

	cases := map[string]struct {
		text string
		want error
	}{
		"unsigned/plain JSON": {`{"type":"question_open"}`, ErrMalformed},
		"bad base64":          {Prefix + "***", ErrMalformed},
		"forged signature":    {forged, ErrBadSignature},
		"unknown key id":      {unknown, ErrUnknownKey},
		"oversize":            {Prefix + strings.Repeat("A", MaxTextBytes), ErrTooLarge},
	}
	for name, tc := range cases {
		if _, err := v.Verify(tc.text); !errors.Is(err, tc.want) {
			t.Errorf("%s: got %v, want %v", name, err, tc.want)
		}
	}
	if _, err := v.Verify(tampered); err == nil {
		t.Error("tampered cue verified")
	}
}

func TestValidateQuestionContent(t *testing.T) {
	ok := []string{"Tiranë", "Prishtinë", "Shkup"}
	if err := ValidateQuestionContent("Cili është kryeqyteti i Shqipërisë?", ok); err != nil {
		t.Fatalf("valid question rejected: %v", err)
	}
	if err := ValidateQuestionContent(strings.Repeat("ë", MaxQuestionTextBytes/2+1), ok); !errors.Is(err, ErrContent) {
		t.Error("over-long question accepted")
	}
	if err := ValidateQuestionContent("Pyetje?", ok[:2]); !errors.Is(err, ErrContent) {
		t.Error("two options accepted")
	}
}
