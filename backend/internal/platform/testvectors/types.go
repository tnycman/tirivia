// Package testvectors defines the cross-platform corpus format (protocol/testvectors).
package testvectors

// FormatVersion changes on any incompatible JSON layout change.
const FormatVersion = 1

// Expectations and file names.
const (
	ExpectValid          = "valid"
	RejectMalformed      = "reject:malformed"
	RejectTooLarge       = "reject:too_large"
	RejectUnknownKey     = "reject:unknown_key"
	RejectBadSignature   = "reject:bad_signature"
	RejectWrongShow      = "reject:wrong_show"
	RejectReplay         = "reject:replay"
	ValidGapResyncNeeded = "valid:gap_resync"
	TestKeyWarning       = "TEST KEYS ONLY. Derived from a public seed. Never use in any environment."
	CueFile              = "cue_vectors.json"
	ShuffleFile          = "shuffle_vectors.json"
)

// CueSuite is cue_vectors.json.
type CueSuite struct {
	FormatVersion int               `json:"format_version"`
	Warning       string            `json:"warning"`
	PublicKeys    map[string]string `json:"public_keys"`
	MaxTextBytes  int               `json:"max_text_bytes"`
	Vectors       []CueVector       `json:"vectors"`
}

// CueVector is one cue and its expected outcome.
type CueVector struct {
	Name    string            `json:"name"`
	Text    string            `json:"text"`
	Context CueContext        `json:"context"`
	Expect  string            `json:"expect"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// CueContext is the client state a cue is applied to.
type CueContext struct {
	ExpectedShowID string `json:"expected_show_id"`
	LastSeq        uint64 `json:"last_seq,string"`
}

// ShuffleSuite is shuffle_vectors.json.
type ShuffleSuite struct {
	FormatVersion int             `json:"format_version"`
	Algorithm     string          `json:"algorithm"`
	Vectors       []ShuffleVector `json:"vectors"`
}

// ShuffleVector pins one option order.
type ShuffleVector struct {
	Name       string `json:"name"`
	SeedHex    string `json:"seed_hex"`
	QuestionID string `json:"question_id"`
	N          int    `json:"n"`
	Order      []int  `json:"order"`
}
