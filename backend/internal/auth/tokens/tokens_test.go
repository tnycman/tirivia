package tokens

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

const (
	iss = "https://api.tirivia.example"
	aud = "tirivia-api"
	jkt = "0ZcOCORZNYy-DWpqq30jZyJGHTN0d2HglBV3uiguA4I"
)

func setup(t *testing.T) (*Issuer, *Validator, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	is, err := NewIssuer("k1", priv, iss, aud)
	if err != nil {
		t.Fatal(err)
	}
	return is, NewValidator(map[string]ed25519.PublicKey{"k1": pub}, iss, aud), priv
}

func TestIssueAndValidate(t *testing.T) {
	is, v, _ := setup(t)
	tok, ttl, err := is.Issue("user-1", "dev-1", "tirivia-ios", jkt)
	if err != nil || ttl != 10*time.Minute {
		t.Fatalf("issue: %v %v", err, ttl)
	}
	c, err := v.Validate(tok)
	if err != nil {
		t.Fatal(err)
	}
	if c.Subject != "user-1" || c.Cnf.JKT != jkt || c.DeviceID != "dev-1" || c.ClientID != "tirivia-ios" || c.ID == "" {
		t.Fatalf("claims: %+v", c)
	}
}

func TestExpiryWithLeeway(t *testing.T) {
	is, v, _ := setup(t)
	start := time.Now()
	is.now = func() time.Time { return start }
	tok, _, _ := is.Issue("u", "d", "c", jkt)

	v.now = func() time.Time { return start.Add(10*time.Minute + 20*time.Second) } // inside 30 s leeway
	if _, err := v.Validate(tok); err != nil {
		t.Fatalf("within leeway rejected: %v", err)
	}
	v.now = func() time.Time { return start.Add(10*time.Minute + 31*time.Second) }
	if _, err := v.Validate(tok); !errors.Is(err, ErrClaims) {
		t.Fatalf("expired token accepted: %v", err)
	}
}

// RFC 8725 §2.1 / §3.1: "alg":"none" and algorithm substitution must fail.
func TestRejectsAlgNoneAndOtherAlgorithms(t *testing.T) {
	_, v, _ := setup(t)
	b64 := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	none := b64(`{"alg":"none","typ":"at+jwt","kid":"k1"}`) + "." + b64(`{"sub":"attacker","iss":"`+iss+`","aud":"`+aud+`","cnf":{"jkt":"x"}}`) + "."
	if _, err := v.Validate(none); !errors.Is(err, ErrMalformed) {
		t.Fatalf("alg none accepted: %v", err)
	}

	ec, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s, _ := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: ec},
		(&jose.SignerOptions{}).WithType(TypeAccessToken).WithHeader("kid", "k1"))
	es256, _ := jwt.Signed(s).Claims(Claims{Claims: jwt.Claims{Subject: "x", Issuer: iss, Audience: jwt.Audience{aud},
		Expiry: jwt.NewNumericDate(time.Now().Add(time.Minute)), ID: "j"}, Cnf: Confirmation{JKT: "x"}}).Serialize()
	if _, err := v.Validate(es256); !errors.Is(err, ErrMalformed) {
		t.Fatalf("ES256 accepted where only EdDSA is allowed: %v", err)
	}
}

func TestRejectsWrongTypeKeyIssuerAudienceAndUnbound(t *testing.T) {
	_, v, priv := setup(t)
	sign := func(typ, kid string, c Claims) string {
		s, _ := jose.NewSigner(jose.SigningKey{Algorithm: jose.EdDSA, Key: priv},
			(&jose.SignerOptions{}).WithType(jose.ContentType(typ)).WithHeader("kid", kid))
		out, _ := jwt.Signed(s).Claims(c).Serialize()
		return out
	}
	good := Claims{Claims: jwt.Claims{Subject: "u", Issuer: iss, Audience: jwt.Audience{aud}, ID: "j",
		Expiry: jwt.NewNumericDate(time.Now().Add(time.Minute))}, Cnf: Confirmation{JKT: jkt}}
	wrongIss, wrongAud, unbound := good, good, good
	wrongIss.Issuer = "https://evil.example"
	wrongAud.Audience = jwt.Audience{"other-api"}
	unbound.Cnf = Confirmation{}

	cases := map[string]struct {
		tok  string
		want error
	}{
		"id token typ (JWT)":  {sign("JWT", "k1", good), ErrType},
		"unknown kid":         {sign(TypeAccessToken, "k9", good), ErrUnknownKey},
		"wrong issuer":        {sign(TypeAccessToken, "k1", wrongIss), ErrClaims},
		"wrong audience":      {sign(TypeAccessToken, "k1", wrongAud), ErrClaims},
		"not DPoP-bound":      {sign(TypeAccessToken, "k1", unbound), ErrNotBound},
		"garbage":             {"not.a.jwt", ErrMalformed},
		"truncated signature": {strings.TrimRight(sign(TypeAccessToken, "k1", good), "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_")[:20], ErrMalformed},
	}
	for name, tc := range cases {
		if _, err := v.Validate(tc.tok); !errors.Is(err, tc.want) {
			t.Errorf("%s: got %v want %v", name, err, tc.want)
		}
	}
}

func TestSignatureFromOtherKeyRejected(t *testing.T) {
	is, _, _ := setup(t)
	_, v2, _ := setup(t) // validator trusts a different key under the same kid
	tok, _, _ := is.Issue("u", "d", "c", jkt)
	if _, err := v2.Validate(tok); !errors.Is(err, ErrMalformed) {
		t.Fatalf("token signed by another key accepted: %v", err)
	}
}
