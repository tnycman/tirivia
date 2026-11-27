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

// lifeScript mirrors domain.Participant.CanUseExtraLife; it covers question q by setting last_ok = q.
var lifeScript = valkey.NewLuaScript(`local joined = tonumber(redis.call('HGET', KEYS[1], 'joined_at_q'))
if not joined then return 'not_joined' end
local q, final = tonumber(ARGV[1]), tonumber(ARGV[2])
local last_ok = tonumber(redis.call('HGET', KEYS[1], 'last_ok'))
if redis.call('HGET', KEYS[1], 'life_used') == '1' or joined > 0 or q < 1 or q >= final or last_ok ~= q - 1 then
  return 'not_allowed'
end
redis.call('HSET', KEYS[1], 'last_ok', q, 'life_used', 1)
return 'ok'`)

// Lives applies the extra-life rule.
type Lives struct{ c valkey.Client }

// NewLives returns Lives.
func NewLives(c valkey.Client) *Lives { return &Lives{c: c} }

// Use spends the player's extra life on question q; the caller checks the REVEAL phase and the inventory.
func (l *Lives) Use(ctx context.Context, show, user string, q int) error {
	res, err := lifeScript.Exec(ctx, l.c, []string{livestate.PlayerKey(show, user)},
		[]string{strconv.Itoa(q), strconv.Itoa(domain.QuestionsPerShow)}).ToString()
	switch {
	case err != nil:
		return err
	case res == "ok":
		return nil
	case res == "not_joined":
		return ErrNotJoined
	}
	return ErrLifeNotAllowed
}

// Refund reverts Use when the inventory debit fails afterwards.
func (l *Lives) Refund(ctx context.Context, show, user string, q int) error {
	return l.c.Do(ctx, l.c.B().Hset().Key(livestate.PlayerKey(show, user)).FieldValue().
		FieldValue("last_ok", strconv.Itoa(q-1)).FieldValue("life_used", "0").Build()).Error()
}
