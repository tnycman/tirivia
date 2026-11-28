package main

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"

	"google.golang.org/protobuf/proto"

	cuev1 "trivia.al/protocol/gen/go/trivia/cue/v1"

	"trivia.al/backend/internal/platform/cuesign"
)

const (
	showID      = "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b"
	otherShowID = "0193a1b2-0000-7000-8000-000000000000"
	questionID  = "0193a1b3-1111-7222-8333-444455556666"
	emittedAt   = int64(1_795_000_000_000)
	keyID       = uint32(7)
)

func testKey(label string) ed25519.PrivateKey {
	seed := sha256.Sum256([]byte("trivia-testvectors-NOT-FOR-PRODUCTION/" + label))
	return ed25519.NewKeyFromSeed(seed[:])
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func rawCue(priv ed25519.PrivateKey, id uint32, body *cuev1.CueBody) string {
	b := must(proto.MarshalOptions{Deterministic: true}.Marshal(body))
	env := must(proto.MarshalOptions{Deterministic: true}.Marshal(&cuev1.SignedCue{Body: b, Signature: ed25519.Sign(priv, b), KeyId: id}))
	return cuesign.Prefix + base64.RawURLEncoding.EncodeToString(env)
}
