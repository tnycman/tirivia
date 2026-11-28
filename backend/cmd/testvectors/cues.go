package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"strconv"
	"strings"

	commonv1 "trivia.al/protocol/gen/go/trivia/common/v1"
	cuev1 "trivia.al/protocol/gen/go/trivia/cue/v1"

	"trivia.al/backend/internal/platform/cuesign"
	tv "trivia.al/backend/internal/platform/testvectors"
)

func questionOpen(seq uint64, text string, opts [3]string) *cuev1.CueBody {
	o := make([]*cuev1.Option, 3)
	for i, t := range opts {
		o[i] = &cuev1.Option{OptionId: string(rune('a' + i)), Text: t}
	}
	return &cuev1.CueBody{ShowId: showID, Seq: seq, EmittedAtUnixMs: emittedAt,
		Kind: &cuev1.CueBody_QuestionOpen{QuestionOpen: &cuev1.QuestionOpen{
			QuestionId: questionID, Index: 5, DurationMs: 8000, Text: text, Options: o, FocusLossVoids: false,
		}}}
}

func cueSuite() tv.CueSuite {
	key := testKey("current")
	signer := must(cuesign.NewSigner(keyID, key))
	sign := func(b *cuev1.CueBody) string { return must(signer.Sign(b)) }
	ctx := tv.CueContext{ExpectedShowID: showID, LastSeq: 41}

	maxQ := strings.Repeat("ë", cuesign.MaxQuestionTextBytes/2)
	maxO := strings.Repeat("ç", cuesign.MaxOptionTextBytes/2)
	worst := questionOpen(42, maxQ, [3]string{maxO, maxO, maxO})
	normal := questionOpen(42, "Cili është kryeqyteti i Shqipërisë?", [3]string{"Tiranë", "Durrës", "Shkodër"})

	pulse := &cuev1.CueBody{ShowId: showID, Seq: 42, EmittedAtUnixMs: emittedAt, Kind: &cuev1.CueBody_Pulse{Pulse: &cuev1.Pulse{
		Viewers: 128_400, PlayersRemaining: 51_230, Phase: cuev1.Phase_PHASE_REVEAL, QuestionIndex: 5,
		ReactionTotals:   map[int32]uint64{int32(commonv1.ReactionKind_REACTION_KIND_HEART): 90_211, int32(commonv1.ReactionKind_REACTION_KIND_FLAG_AL): 12_007},
		ReactionsEnabled: true, ChatEnabled: false,
	}}}
	reveal := &cuev1.CueBody{ShowId: showID, Seq: 42, EmittedAtUnixMs: emittedAt, Kind: &cuev1.CueBody_Reveal{Reveal: &cuev1.Reveal{
		QuestionId: questionID, CorrectOptionId: "a", OptionCounts: map[string]uint64{"a": 40_100, "b": 9_000, "c": 2_130}, PlayersRemaining: 40_100,
	}}}

	good := sign(normal)
	// Tamper: flip one base64 character in the middle of the payload.
	tb := []byte(good)
	mid := len(cuesign.Prefix) + (len(tb)-len(cuesign.Prefix))/2
	if tb[mid] == 'A' {
		tb[mid] = 'B'
	} else {
		tb[mid] = 'A'
	}

	big := questionOpen(42, strings.Repeat("ë", 400), [3]string{"a", "b", "c"})

	return tv.CueSuite{
		FormatVersion: tv.FormatVersion,
		Warning:       tv.TestKeyWarning,
		PublicKeys:    map[string]string{strconv.Itoa(int(keyID)): base64.StdEncoding.EncodeToString(key.Public().(ed25519.PublicKey))},
		MaxTextBytes:  cuesign.MaxTextBytes,
		Vectors: []tv.CueVector{
			{Name: "question_open_albanian", Text: good, Context: ctx, Expect: tv.ExpectValid,
				Fields: map[string]string{"kind": "question_open", "show_id": showID, "seq": "42", "question_id": questionID,
					"index": "5", "duration_ms": "8000", "text": "Cili është kryeqyteti i Shqipërisë?", "option_b_text": "Durrës"}},
			{Name: "question_open_worst_case_size", Text: sign(worst), Context: ctx, Expect: tv.ExpectValid,
				Fields: map[string]string{"kind": "question_open", "seq": "42", "text": maxQ}},
			{Name: "pulse", Text: sign(pulse), Context: ctx, Expect: tv.ExpectValid,
				Fields: map[string]string{"kind": "pulse", "viewers": "128400", "players_remaining": "51230", "chat_enabled": "false", "reactions_enabled": "true", "reaction_total_1": "90211"}},
			{Name: "reveal", Text: sign(reveal), Context: ctx, Expect: tv.ExpectValid,
				Fields: map[string]string{"kind": "reveal", "correct_option_id": "a", "option_count_b": "9000", "players_remaining": "40100"}},
			{Name: "gap_requires_resync", Text: sign(questionOpen(45, "Pyetje?", [3]string{"x", "y", "z"})), Context: ctx, Expect: tv.ValidGapResyncNeeded,
				Fields: map[string]string{"kind": "question_open", "seq": "45"}},
			{Name: "unsigned_plain_json", Text: `{"type":"question_open","text":"Pyetje?"}`, Context: ctx, Expect: tv.RejectMalformed},
			{Name: "bad_base64", Text: cuesign.Prefix + "***not-base64***", Context: ctx, Expect: tv.RejectMalformed},
			{Name: "wrong_prefix_version", Text: "t9." + strings.TrimPrefix(good, cuesign.Prefix), Context: ctx, Expect: tv.RejectMalformed},
			{Name: "tampered_payload", Text: string(tb), Context: ctx, Expect: tv.RejectBadSignature + "|" + tv.RejectMalformed},
			{Name: "forged_same_key_id", Text: rawCue(testKey("attacker"), keyID, normal), Context: ctx, Expect: tv.RejectBadSignature},
			{Name: "unknown_key_id", Text: rawCue(key, 99, normal), Context: ctx, Expect: tv.RejectUnknownKey},
			{Name: "oversize", Text: rawCue(key, keyID, big), Context: ctx, Expect: tv.RejectTooLarge},
			{Name: "other_show", Text: sign(&cuev1.CueBody{ShowId: otherShowID, Seq: 42, Kind: normal.Kind}), Context: ctx, Expect: tv.RejectWrongShow},
			{Name: "replayed_seq", Text: sign(questionOpen(41, "Pyetje?", [3]string{"x", "y", "z"})), Context: ctx, Expect: tv.RejectReplay},
		},
	}
}
