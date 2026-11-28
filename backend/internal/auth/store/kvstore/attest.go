package kvstore

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/valkey-io/valkey-go"

	"trivia.al/backend/internal/auth/otp"
)

// Attest is an attest.Store; GETDEL makes challenge consumption atomic and single use.
type Attest struct{ c valkey.Client }

// NewAttest returns an Attest store.
func NewAttest(c valkey.Client) *Attest { return &Attest{c: c} }

func challengeKeyAttest(c []byte) string { return "att:c:" + hex.EncodeToString(c) }

// PutChallenge implements attest.Store.
func (a *Attest) PutChallenge(ctx context.Context, c []byte, ttl time.Duration) error {
	return a.c.Do(ctx, a.c.B().Set().Key(challengeKeyAttest(c)).Value("1").Nx().Px(ttl).Build()).Error()
}

// ConsumeChallenge implements attest.Store.
func (a *Attest) ConsumeChallenge(ctx context.Context, c []byte) (bool, error) {
	err := a.c.Do(ctx, a.c.B().Getdel().Key(challengeKeyAttest(c)).Build()).Error()
	if valkey.IsValkeyNil(err) {
		return false, nil
	}
	return err == nil, err
}

// PutTicket implements attest.Store.
func (a *Attest) PutTicket(ctx context.Context, tok string, dev otp.Device, ttl time.Duration) error {
	raw, err := json.Marshal(dev)
	if err != nil {
		return err
	}
	return a.c.Do(ctx, a.c.B().Set().Key("att:t:"+tok).Value(string(raw)).Px(ttl).Build()).Error()
}

// Ticket implements attest.Store.
func (a *Attest) Ticket(ctx context.Context, tok string) (otp.Device, bool, error) {
	raw, err := a.c.Do(ctx, a.c.B().Get().Key("att:t:"+tok).Build()).AsBytes()
	if valkey.IsValkeyNil(err) {
		return otp.Device{}, false, nil
	}
	if err != nil {
		return otp.Device{}, false, err
	}
	var dev otp.Device
	return dev, true, json.Unmarshal(raw, &dev)
}
