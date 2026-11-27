package kvstore

import (
	"context"
	"time"

	"github.com/valkey-io/valkey-go"
)

// Replay is a dpop.ReplayCache: SET NX returns nil when the key already exists.
type Replay struct{ c valkey.Client }

// NewReplay returns a Replay cache.
func NewReplay(c valkey.Client) *Replay { return &Replay{c: c} }

// Seen implements dpop.ReplayCache.
func (r *Replay) Seen(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	err := r.c.Do(ctx, r.c.B().Set().Key("dpop:"+key).Value("1").Nx().Px(ttl).Build()).Error()
	if valkey.IsValkeyNil(err) {
		return true, nil
	}
	return false, err
}
