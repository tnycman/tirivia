package live

import (
	"context"
	"errors"
	"strconv"

	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/backend/internal/game/domain"
	"github.com/tirivia/backend/internal/platform/livestate"
)

// ErrLifeNotAllowed means UF-4 forbids an extra life here.
var ErrLifeNotAllowed = errors.New("live: extra life not allowed")

// lifeScript mirrors domain.Participant.CanUseExtraLife and counts the life in the same step; a retry on the same question is a no-op.
var lifeScript = valkey.NewLuaScript(`local joined = tonumber(redis.call('HGET', KEYS[1], 'joined_at_q'))
if not joined then return 'not_joined' end
if redis.call('HGET', KEYS[1], 'life_used') == '1' then
  return redis.call('HGET', KEYS[1], 'life_q') == ARGV[1] and 'already' or 'not_allowed'
end
local q, final = tonumber(ARGV[1]), tonumber(ARGV[2])
local last_ok = tonumber(redis.call('HGET', KEYS[1], 'last_ok'))
if joined > 0 or q < 1 or q >= final or last_ok ~= q - 1 then return 'not_allowed' end
redis.call('HSET', KEYS[1], 'last_ok', q, 'life_used', 1, 'life_q', q)
redis.call('HINCRBY', KEYS[2], ARGV[3], 1)
return 'ok'`)

// refundScript reverts a life spent on question q exactly once, tally included.
var refundScript = valkey.NewLuaScript(`if redis.call('HGET', KEYS[1], 'life_q') ~= ARGV[1] then return 0 end
redis.call('HSET', KEYS[1], 'last_ok', tonumber(ARGV[1]) - 1, 'life_used', 0)
redis.call('HDEL', KEYS[1], 'life_q')
redis.call('HINCRBY', KEYS[2], ARGV[2], -1)
return 1`)

// Lives applies the extra-life rule.
type Lives struct{ c valkey.Client }

// NewLives returns Lives.
func NewLives(c valkey.Client) *Lives { return &Lives{c: c} }

// Use spends the player's extra life on question q, or confirms it was already spent there; the caller checks the REVEAL phase.
func (l *Lives) Use(ctx context.Context, show, user string, q int) error {
	res, err := lifeScript.Exec(ctx, l.c, lifeKeys(show, user, q),
		[]string{strconv.Itoa(q), strconv.Itoa(domain.QuestionsPerShow), LifeOption}).ToString()
	switch {
	case err != nil:
		return err
	case res == "ok" || res == "already":
		return nil
	case res == "not_joined":
		return ErrNotJoined
	}
	return ErrLifeNotAllowed
}

// Refund reverts Use when the inventory definitively refuses the debit; repeating it changes nothing.
func (l *Lives) Refund(ctx context.Context, show, user string, q int) error {
	return refundScript.Exec(ctx, l.c, lifeKeys(show, user, q), []string{strconv.Itoa(q), LifeOption}).Error()
}

func lifeKeys(show, user string, q int) []string {
	return []string{livestate.PlayerKey(show, user), livestate.TallyKey(show, q, livestate.Bucket(user))}
}
