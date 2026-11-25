// Command testvectors regenerates protocol/testvectors deterministically.
//
//	go run ./cmd/testvectors           # write files
//	go run ./cmd/testvectors -check    # exit 1 if committed files are stale (CI)
//
// Keys are derived from a public seed and are TEST ONLY. Ed25519 signatures and deterministic
// protobuf marshalling make the output byte-for-byte reproducible.
package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	commonv1 "github.com/tirivia/protocol/gen/go/tirivia/common/v1"
	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"

	"github.com/tirivia/backend/internal/platform/cuesign"
	"github.com/tirivia/backend/internal/platform/shuffle"
	tv "github.com/tirivia/backend/internal/platform/testvectors"
)

const (
	showID      = "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b"
	otherShowID = "0193a1b2-0000-7000-8000-000000000000"
	questionID  = "0193a1b3-1111-7222-8333-444455556666"
	emittedAt   = int64(1_795_000_000_000) // fixed so output is reproducible
	keyID       = uint32(7)
)

func testKey(label string) ed25519.PrivateKey {
	seed := sha256.Sum256([]byte("tirivia-testvectors-NOT-FOR-PRODUCTION/" + label))
	return ed25519.NewKeyFromSeed(seed[:])
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// rawCue signs body with priv under keyID without the size guard (for oversize vectors).
func rawCue(priv ed25519.PrivateKey, id uint32, body *cuev1.CueBody) string {
	b := must(proto.MarshalOptions{Deterministic: true}.Marshal(body))
	env := must(proto.MarshalOptions{Deterministic: true}.Marshal(&cuev1.SignedCue{Body: b, Signature: ed25519.Sign(priv, b), KeyId: id}))
	return cuesign.Prefix + base64.RawURLEncoding.EncodeToString(env)
}

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

func shuffleSuite() tv.ShuffleSuite {
	s := tv.ShuffleSuite{
		FormatVersion: tv.FormatVersion,
		Algorithm:     "k = HMAC-SHA256(key=seed, msg=UTF-8(question_id)); for i = n-1 down to 1: j = k[n-1-i] mod (i+1) (k bytes unsigned); swap(idx[i], idx[j])",
	}
	add := func(name, seedHex, qid string, n int) {
		seed := must(hex.DecodeString(seedHex))
		s.Vectors = append(s.Vectors, tv.ShuffleVector{Name: name, SeedHex: seedHex, QuestionID: qid, N: n, Order: shuffle.Order(seed, qid, n)})
	}
	for i := range 12 {
		seed := sha256.Sum256([]byte(fmt.Sprintf("player-%d", i)))
		add(fmt.Sprintf("player_%02d_q%d", i, i%12+1), hex.EncodeToString(seed[:16]), fmt.Sprintf("0193a1b3-0000-7000-8000-%012d", i), 3)
	}
	add("utf8_question_id", "00112233445566778899aabbccddeeff", "pyetja-ë-ç-12", 3)
	add("high_bytes_mask_check", "ffffffffffffffffffffffffffffffff", "q-high-bytes", 3) // Kotlin signed-byte pitfall
	add("single_option", "01", "q", 1)
	add("four_options", "deadbeef", "q-four", 4)
	return s
}

func encode(v any) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func main() {
	check := flag.Bool("check", false, "fail if committed vectors are stale")
	flag.Parse()

	files := map[string][]byte{tv.CueFile: encode(cueSuite()), tv.ShuffleFile: encode(shuffleSuite())}
	root, err := tv.OpenDir() // confine all file access to protocol/testvectors
	if err != nil {
		panic(err)
	}
	defer func() { _ = root.Close() }()
	stale := false
	for name, want := range files {
		path := filepath.Join(tv.Dir(), name)
		if *check {
			got, err := root.ReadFile(name)
			if err != nil || !bytes.Equal(got, want) {
				fmt.Fprintf(os.Stderr, "stale: %s (run: go run ./cmd/testvectors)\n", path)
				stale = true
			}
			continue
		}
		if err := root.WriteFile(name, want, 0o600); err != nil {
			panic(err)
		}
		fmt.Println("wrote", path)
	}
	if stale {
		os.Exit(1)
	}
}
