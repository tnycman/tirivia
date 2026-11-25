package cuesign

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"testing"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"

	tv "github.com/tirivia/backend/internal/platform/testvectors"
)

// minCueVectors guards against a runner silently skipping vectors.
const minCueVectors = 14

// outcome runs the shared runner algorithm and returns the expectation string it produced.
func outcome(v *Verifier, vec tv.CueVector) (string, *cuev1.CueBody) {
	body, err := v.Verify(vec.Text)
	switch {
	case errors.Is(err, ErrTooLarge):
		return tv.RejectTooLarge, nil
	case errors.Is(err, ErrMalformed):
		return tv.RejectMalformed, nil
	case errors.Is(err, ErrUnknownKey):
		return tv.RejectUnknownKey, nil
	case errors.Is(err, ErrBadSignature):
		return tv.RejectBadSignature, nil
	case err != nil:
		return "error:" + err.Error(), nil
	}
	gap, err := CheckSequence(body, vec.Context.ExpectedShowID, vec.Context.LastSeq)
	switch {
	case errors.Is(err, ErrWrongShow):
		return tv.RejectWrongShow, body
	case errors.Is(err, ErrReplay):
		return tv.RejectReplay, body
	case gap:
		return tv.ValidGapResyncNeeded, body
	}
	return tv.ExpectValid, body
}

// fields exposes the canonical field names used in cue_vectors.json (same names on every platform).
func fields(b *cuev1.CueBody) map[string]string {
	u := func(n uint64) string { return strconv.FormatUint(n, 10) }
	m := map[string]string{"show_id": b.GetShowId(), "seq": u(b.GetSeq())}
	switch k := b.GetKind().(type) {
	case *cuev1.CueBody_QuestionOpen:
		q := k.QuestionOpen
		m["kind"], m["question_id"], m["text"] = "question_open", q.GetQuestionId(), q.GetText()
		m["index"], m["duration_ms"] = u(uint64(q.GetIndex())), u(uint64(q.GetDurationMs()))
		for _, o := range q.GetOptions() {
			m["option_"+o.GetOptionId()+"_text"] = o.GetText()
		}
	case *cuev1.CueBody_Pulse:
		p := k.Pulse
		m["kind"], m["viewers"], m["players_remaining"] = "pulse", u(p.GetViewers()), u(p.GetPlayersRemaining())
		m["chat_enabled"], m["reactions_enabled"] = strconv.FormatBool(p.GetChatEnabled()), strconv.FormatBool(p.GetReactionsEnabled())
		for kind, n := range p.GetReactionTotals() {
			m["reaction_total_"+strconv.Itoa(int(kind))] = u(n)
		}
	case *cuev1.CueBody_Reveal:
		r := k.Reveal
		m["kind"], m["correct_option_id"], m["players_remaining"] = "reveal", r.GetCorrectOptionId(), u(r.GetPlayersRemaining())
		for opt, n := range r.GetOptionCounts() {
			m["option_count_"+opt] = u(n)
		}
	}
	return m
}

func TestSharedCueVectors(t *testing.T) {
	var suite tv.CueSuite
	if err := tv.Load(tv.CueFile, &suite); err != nil {
		t.Fatal(err)
	}
	if suite.FormatVersion != tv.FormatVersion || suite.MaxTextBytes != MaxTextBytes {
		t.Fatalf("suite format %d / budget %d does not match code", suite.FormatVersion, suite.MaxTextBytes)
	}
	keys := map[uint32]ed25519.PublicKey{}
	for id, b64 := range suite.PublicKeys {
		n, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			t.Fatal(err)
		}
		k, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			t.Fatal(err)
		}
		keys[uint32(n)] = k
	}
	v := NewVerifier(keys)

	if len(suite.Vectors) < minCueVectors {
		t.Fatalf("only %d vectors; expected at least %d", len(suite.Vectors), minCueVectors)
	}
	for _, vec := range suite.Vectors {
		t.Run(vec.Name, func(t *testing.T) {
			got, body := outcome(v, vec)
			// "a|b" means either outcome is acceptable (e.g. tampering may break framing or signature).
			if !strings.Contains("|"+vec.Expect+"|", "|"+got+"|") {
				t.Fatalf("got %s, want %s", got, vec.Expect)
			}
			if body == nil || len(vec.Fields) == 0 {
				return
			}
			have := fields(body)
			for k, want := range vec.Fields {
				if have[k] != want {
					t.Errorf("field %s = %q, want %q", k, have[k], want)
				}
			}
		})
	}
}
