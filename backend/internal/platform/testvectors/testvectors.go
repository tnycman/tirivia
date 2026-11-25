// Package testvectors defines the JSON format of the cross-platform conformance corpus in
// protocol/testvectors (Master Plan §5 "one shared definition and conformance test suite").
// Go, Swift and Kotlin runners all read these files; the Go generator in cmd/testvectors
// writes them deterministically, and CI fails if the committed files differ from a fresh run.
package testvectors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// FormatVersion is bumped on any incompatible change to the JSON layout.
const FormatVersion = 1

// Expectation values used in vectors.
const (
	ExpectValid           = "valid"
	RejectMalformed       = "reject:malformed"
	RejectTooLarge        = "reject:too_large"
	RejectUnknownKey      = "reject:unknown_key"
	RejectBadSignature    = "reject:bad_signature"
	RejectWrongShow       = "reject:wrong_show"
	RejectReplay          = "reject:replay"
	ValidGapResyncNeeded  = "valid:gap_resync"
	TestKeyWarning        = "TEST KEYS ONLY. Derived from a public seed. Never use in any environment."
	CueFile               = "cue_vectors.json"
	ShuffleFile           = "shuffle_vectors.json"
	MaxCueTextBytesInFile = 900
)

// CueSuite is protocol/testvectors/cue_vectors.json.
type CueSuite struct {
	FormatVersion int    `json:"format_version"`
	Warning       string `json:"warning"`
	// PublicKeys maps key_id (decimal string) to the base64 (std, padded) Ed25519 public key.
	PublicKeys   map[string]string `json:"public_keys"`
	MaxTextBytes int               `json:"max_text_bytes"`
	Vectors      []CueVector       `json:"vectors"`
}

// CueVector is one text cue and the outcome every platform must produce.
//
// Runner algorithm (all platforms):
//  1. Verify(text) with PublicKeys: size, prefix, base64url, SignedCue decode, key lookup,
//     Ed25519 over body bytes, CueBody decode. Failures map to the reject:* values.
//  2. CheckSequence(body, Context.ExpectedShowID, Context.LastSeq) -> wrong_show / replay,
//     or valid:gap_resync when seq > last_seq + 1.
//  3. For valid vectors, compare Fields against the decoded body.
type CueVector struct {
	Name    string            `json:"name"`
	Text    string            `json:"text"`
	Context CueContext        `json:"context"`
	Expect  string            `json:"expect"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// CueContext is the client state the cue is applied to.
type CueContext struct {
	ExpectedShowID string `json:"expected_show_id"`
	LastSeq        uint64 `json:"last_seq,string"`
}

// ShuffleSuite is protocol/testvectors/shuffle_vectors.json.
type ShuffleSuite struct {
	FormatVersion int             `json:"format_version"`
	Algorithm     string          `json:"algorithm"`
	Vectors       []ShuffleVector `json:"vectors"`
}

// ShuffleVector pins OptionShuffle.order(seed, questionID, n).
type ShuffleVector struct {
	Name       string `json:"name"`
	SeedHex    string `json:"seed_hex"`
	QuestionID string `json:"question_id"`
	N          int    `json:"n"`
	Order      []int  `json:"order"`
}

// Dir returns the absolute path of protocol/testvectors, located relative to this source file
// so tests and the generator work from any working directory.
func Dir() string {
	_, file, _, _ := runtime.Caller(0)
	// file = <repo>/backend/internal/platform/testvectors/testvectors.go
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "protocol", "testvectors")
}

// OpenDir opens Dir() as an os.Root so file access cannot escape the vectors directory.
func OpenDir() (*os.Root, error) { return os.OpenRoot(Dir()) }

// Load decodes a suite file from Dir().
func Load(name string, v any) error {
	root, err := OpenDir()
	if err != nil {
		return fmt.Errorf("testvectors: open dir: %w", err)
	}
	defer func() { _ = root.Close() }()
	b, err := root.ReadFile(name)
	if err != nil {
		return fmt.Errorf("testvectors: read %s: %w", name, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("testvectors: decode %s: %w", name, err)
	}
	return nil
}
