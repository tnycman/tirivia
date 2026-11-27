package kvstore

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/backend/internal/auth/otp"
	"github.com/tirivia/backend/internal/platform/seal"
)

var putHash = valkey.NewLuaScript(`redis.call('HSET', KEYS[1], 'data', ARGV[1], 'attempts', 0)
return redis.call('PEXPIRE', KEYS[1], ARGV[2])`)

// Challenges is an otp.Store; each challenge is sealed because it holds the phone number.
type Challenges struct {
	c    valkey.Client
	seal *seal.Sealer
}

// NewChallenges returns a Challenges store.
func NewChallenges(c valkey.Client, s *seal.Sealer) *Challenges { return &Challenges{c: c, seal: s} }

func challengeKey(id string) string { return "otp:" + id }

// Put implements otp.Store.
func (s *Challenges) Put(ctx context.Context, ch otp.Challenge, ttl time.Duration) error {
	ch.Attempts = 0
	raw, err := json.Marshal(ch)
	if err != nil {
		return err
	}
	k := challengeKey(ch.ID)
	sealed := s.seal.Seal(raw, []byte(k))
	return putHash.Exec(ctx, s.c, []string{k}, []string{string(sealed), ms(ttl)}).Error()
}

// Get implements otp.Store.
func (s *Challenges) Get(ctx context.Context, id string) (otp.Challenge, error) {
	ch, _, err := s.get(ctx, id)
	return ch, err
}

func (s *Challenges) get(ctx context.Context, id string) (otp.Challenge, string, error) {
	k := challengeKey(id)
	m, err := s.c.Do(ctx, s.c.B().Hgetall().Key(k).Build()).AsStrMap()
	if err != nil {
		return otp.Challenge{}, "", err
	}
	if len(m) == 0 {
		return otp.Challenge{}, "", otp.ErrExpired
	}
	raw, err := s.seal.Open([]byte(m["data"]), []byte(k))
	if err != nil {
		return otp.Challenge{}, "", err
	}
	var ch otp.Challenge
	if err := json.Unmarshal(raw, &ch); err != nil {
		return otp.Challenge{}, "", err
	}
	ch.Attempts, err = strconv.Atoi(m["attempts"])
	return ch, m["data"], err
}

// Delete implements otp.Store.
func (s *Challenges) Delete(ctx context.Context, id string) error {
	return s.c.Do(ctx, s.c.B().Del().Key(challengeKey(id)).Build()).Error()
}
