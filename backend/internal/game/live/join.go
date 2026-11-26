// Package live runs the hot path in Valkey: joins, answer intake and tallies (SPEC §6.3, §6.7).
package live

import (
	"context"
	"crypto/rand"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/backend/internal/game/domain"
	"github.com/tirivia/backend/internal/platform/livestate"
)

// SeedBytes is the shuffle-seed length: RFC 2104 §2 recommends an HMAC key of at least L bytes, 32 for SHA-256 (SPEC §6.3).
const SeedBytes = 32

// joinScript keeps the first join: rejoining can never turn a late joiner into an on-time player or change the seed.
var joinScript = valkey.NewLuaScript(`if redis.call('HSETNX', KEYS[1], 'joined_at_q', ARGV[1]) == 1 then
  redis.call('HSET', KEYS[1], 'last_ok', 0, 'life_used', 0)
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
redis.call('HSETNX', KEYS[1], 'seed', ARGV[3])
return redis.call('HMGET', KEYS[1], 'joined_at_q', 'last_ok', 'life_used', 'seed')`)

// Players reads and joins players.
type Players struct {
	c   valkey.Client
	ttl time.Duration
}

// NewPlayers returns Players; ttl bounds how long live player state outlives the show.
func NewPlayers(c valkey.Client, ttl time.Duration) *Players { return &Players{c: c, ttl: ttl} }

// Join records the player at questionOpened (0 before Q1) and returns their standing.
func (p *Players) Join(ctx context.Context, show, user string, questionOpened int) (domain.Participant, error) {
	pt, _, err := p.JoinSeeded(ctx, show, user, questionOpened)
	return pt, err
}

// JoinSeeded is Join plus the player's shuffle seed, created at the first join.
func (p *Players) JoinSeeded(ctx context.Context, show, user string, questionOpened int) (domain.Participant, []byte, error) {
	seed := make([]byte, SeedBytes)
	if _, err := rand.Read(seed); err != nil {
		return domain.Participant{}, nil, err
	}
	vals, err := joinScript.Exec(ctx, p.c, []string{livestate.PlayerKey(show, user)},
		[]string{strconv.Itoa(questionOpened), strconv.FormatInt(p.ttl.Milliseconds(), 10), string(seed)}).AsStrSlice()
	if err != nil {
		return domain.Participant{}, nil, err
	}
	if len(vals) != 4 {
		return domain.Participant{}, nil, ErrNotJoined
	}
	pt, err := participant(vals[:3])
	return pt, []byte(vals[3]), err
}

// Get returns the standing without joining; ErrNotJoined if absent.
func (p *Players) Get(ctx context.Context, show, user string) (domain.Participant, error) {
	cmd := p.c.B().Hmget().Key(livestate.PlayerKey(show, user)).Field("joined_at_q", "last_ok", "life_used").Build()
	vals, err := p.c.Do(ctx, cmd).ToArray()
	if err != nil {
		return domain.Participant{}, err
	}
	strs := make([]string, 0, len(vals))
	for _, v := range vals {
		s, err := v.ToString()
		if err != nil {
			return domain.Participant{}, ErrNotJoined
		}
		strs = append(strs, s)
	}
	return participant(strs)
}

func participant(vals []string) (domain.Participant, error) {
	if len(vals) != 3 {
		return domain.Participant{}, ErrNotJoined
	}
	joined, err1 := strconv.Atoi(vals[0])
	lastOK, err2 := strconv.Atoi(vals[1])
	if err1 != nil || err2 != nil {
		return domain.Participant{}, ErrNotJoined
	}
	return domain.Participant{JoinedAtQ: joined, LastOK: lastOK, LifeUsed: vals[2] == "1"}, nil
}
