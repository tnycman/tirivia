// Package dpop verifies DPoP proofs (RFC 9449), which bind every token to a key held in the
// phone's Secure Enclave / Android Keystore (ADR-0005). It implements the checks of RFC 9449
// §4.3 plus jti replay protection (§11.1) and access-token binding via "ath" and "cnf.jkt".
package dpop

import (
	"context"
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
)

// TypeDPoP is the required JWS "typ" header (RFC 9449 §4.2).
const TypeDPoP = "dpop+jwt"

// AllowedAlgorithms lists accepted proof algorithms. Device keys are P-256 in the Secure Enclave and Android Keystore, so only
// ES256 is accepted. Symmetric algorithms and "none" are impossible by construction.
var AllowedAlgorithms = []jose.SignatureAlgorithm{jose.ES256}

// DefaultWindow is how far iat may deviate from server time (RFC 9449 §11.1 lets servers choose).
const DefaultWindow = 60 * time.Second

// Errors; callers map all of them to CodeUnauthenticated (and set DPoP-Nonce on ErrNonce).
var (
	ErrMissing    = errors.New("dpop: proof missing")
	ErrMalformed  = errors.New("dpop: malformed proof")
	ErrType       = errors.New("dpop: typ must be dpop+jwt")
	ErrKey        = errors.New("dpop: jwk header missing, invalid or private")
	ErrSignature  = errors.New("dpop: signature invalid")
	ErrMethod     = errors.New("dpop: htm does not match the request method")
	ErrURL        = errors.New("dpop: htu does not match the request URL")
	ErrTime       = errors.New("dpop: iat outside the acceptable window")
	ErrReplay     = errors.New("dpop: jti already used")
	ErrNonce      = errors.New("dpop: nonce missing or wrong")
	ErrAccessHash = errors.New("dpop: ath does not match the access token")
	ErrBinding    = errors.New("dpop: proof key does not match the token binding")
)

// ReplayCache remembers jti values. Seen must atomically record key and report whether it was
// already present (Valkey: SET NX EX). Entries live at least as long as the iat window.
type ReplayCache interface {
	Seen(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

// Options are the per-request expectations.
type Options struct {
	// AccessToken, when the request carries one: the proof must contain ath = hash(token).
	AccessToken string
	// ExpectedJKT, when the token or session is bound: the proof key must match it.
	ExpectedJKT string
	// Nonce, when the server requires one (claims, payouts; SPEC §7 step-up).
	Nonce string
}

// Proof is a verified DPoP proof.
type Proof struct {
	// JKT is the RFC 7638 SHA-256 thumbprint of the proof key, base64url without padding.
	JKT      string
	JTI      string
	IssuedAt time.Time
}

type claims struct {
	JTI   string `json:"jti"`
	HTM   string `json:"htm"`
	HTU   string `json:"htu"`
	IAT   *int64 `json:"iat"`
	ATH   string `json:"ath,omitempty"`
	Nonce string `json:"nonce,omitempty"`
}

// Verifier checks proofs.
type Verifier struct {
	replay ReplayCache
	window time.Duration
	now    func() time.Time
}

// NewVerifier returns a Verifier using replay for jti tracking.
func NewVerifier(replay ReplayCache) *Verifier {
	return &Verifier{replay: replay, window: DefaultWindow, now: time.Now}
}

// Verify runs RFC 9449 §4.3 for a proof sent with an HTTP request to method + requestURL.
func (v *Verifier) Verify(ctx context.Context, proof, method, requestURL string, o Options) (*Proof, error) {
	if proof == "" {
		return nil, ErrMissing
	}
	// §4.3 (2,5): a single well-formed JWS; only asymmetric, supported algorithms.
	jws, err := jose.ParseSignedCompact(proof, AllowedAlgorithms)
	if err != nil || len(jws.Signatures) != 1 {
		return nil, ErrMalformed
	}
	h := jws.Signatures[0].Protected
	// §4.3 (4): typ.
	if typ, _ := h.ExtraHeaders[jose.HeaderType].(string); typ != TypeDPoP {
		return nil, ErrType
	}
	// §4.3 (6,7): jwk present, valid, public only; the signature verifies with it.
	jwk := h.JSONWebKey
	if jwk == nil || !jwk.Valid() || !jwk.IsPublic() {
		return nil, ErrKey
	}
	payload, err := jws.Verify(jwk.Key)
	if err != nil {
		return nil, ErrSignature
	}
	var c claims
	if err := json.Unmarshal(payload, &c); err != nil || c.JTI == "" || c.HTM == "" || c.HTU == "" || c.IAT == nil {
		return nil, ErrMalformed // §4.3 (3): all required claims present
	}
	// §4.3 (8,9): method and URL (htu compared without query and fragment).
	if c.HTM != method {
		return nil, ErrMethod
	}
	if !sameURL(c.HTU, requestURL) {
		return nil, ErrURL
	}
	// §4.3 (10): server-provided nonce.
	if o.Nonce != "" && c.Nonce != o.Nonce {
		return nil, ErrNonce
	}
	// §4.3 (11): freshness.
	iat := time.Unix(*c.IAT, 0)
	if d := v.now().Sub(iat); d > v.window || d < -v.window {
		return nil, ErrTime
	}
	// §4.3 (12): access-token binding.
	if o.AccessToken != "" {
		sum := sha256.Sum256([]byte(o.AccessToken))
		if c.ATH != base64.RawURLEncoding.EncodeToString(sum[:]) {
			return nil, ErrAccessHash
		}
	}
	tp, err := jwk.Thumbprint(crypto.SHA256)
	if err != nil {
		return nil, ErrKey
	}
	jkt := base64.RawURLEncoding.EncodeToString(tp)
	if o.ExpectedJKT != "" && jkt != o.ExpectedJKT {
		return nil, ErrBinding
	}
	// §11.1: single use. Keyed by key thumbprint + jti so clients cannot collide with each other.
	seen, err := v.replay.Seen(ctx, "dpop:"+jkt+":"+c.JTI, 2*v.window)
	if err != nil {
		return nil, fmt.Errorf("dpop: replay cache: %w", err)
	}
	if seen {
		return nil, ErrReplay
	}
	return &Proof{JKT: jkt, JTI: c.JTI, IssuedAt: iat}, nil
}

// sameURL compares scheme, host (case-insensitive, default ports normalised) and path, ignoring
// query and fragment (RFC 9449 §4.3 step 9, RFC 3986 §6.2.2–6.2.3 normalisation).
func sameURL(a, b string) bool {
	ua, err1 := url.Parse(a)
	ub, err2 := url.Parse(b)
	if err1 != nil || err2 != nil || ua.Scheme == "" || ua.Host == "" {
		return false
	}
	norm := func(u *url.URL) string {
		host := strings.ToLower(u.Hostname())
		port := u.Port()
		if (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") {
			port = ""
		}
		if port != "" {
			host += ":" + port
		}
		path := u.EscapedPath()
		if path == "" {
			path = "/"
		}
		return strings.ToLower(u.Scheme) + "://" + host + path
	}
	return norm(ua) == norm(ub)
}

// MemoryReplayCache is an in-process ReplayCache for tests and single-task development.
// Production uses Valkey (SET key 1 NX EX ttl) so all tasks share one view.
type MemoryReplayCache struct {
	mu   sync.Mutex
	seen map[string]time.Time
	now  func() time.Time
}

// NewMemoryReplayCache returns an empty cache.
func NewMemoryReplayCache() *MemoryReplayCache {
	return &MemoryReplayCache{seen: map[string]time.Time{}, now: time.Now}
}

// Seen implements ReplayCache.
func (m *MemoryReplayCache) Seen(_ context.Context, key string, ttl time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	if exp, ok := m.seen[key]; ok && now.Before(exp) {
		return true, nil
	}
	m.seen[key] = now.Add(ttl)
	return false, nil
}
