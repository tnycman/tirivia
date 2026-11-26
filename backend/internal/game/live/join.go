// Package live runs the hot path in Valkey: joins, answer intake and tallies (SPEC §6.3, §6.7).
package live

import (
	"context"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/backend/internal/game/domain"
	"github.com/tirivia/backend/internal/platform/livestate"
)

// joinScript keeps the first join: rejoining can never turn a late joiner into an on-time player.
var joinScript = valkey.NewLuaScript(`if redis.call('HSETNX', KEYS[1], 'joined_at_q', ARGV[1]) == 1 then
  redis.call('HSET', KEYS[1], 'last_ok', 0, 'life_used', 0)
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return redis.call('HMGET', KEYS[1], 'joined_at_q', 'last_ok', 'life_used')`)

// Players reads and joins players.
type Players struct {
	c   valkey.Client
	ttl time.Duration
}

// NewPlayers returns Players; ttl bounds how long live player state outlives the show.
func NewPlayers(c valkey.Client, ttl time.Duration) *Players { return &Players{c: c, ttl: ttl} }

// Join records the player at questionOpened (0 before Q1) and returns their standing.
func (p *Players) Join(ctx context.Context, show, user string, questionOpened int) (domain.Participant, error) {
	vals, err := joinScript.Exec(ctx, p.c, []string{livestate.PlayerKey(show, user)},
		[]string{strconv.Itoa(questionOpened), strconv.FormatInt(p.ttl.Milliseconds(), 10)}).AsStrSlice()
	if err != nil {
		return domain.Participant{}, err
	}
	return participant(vals)
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
