package kvstore

import (
	"context"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"
)

// incrWindow is the Redis INCR "rate limiter 2" pattern: the expiry is set atomically with the first count.
var incrWindow = valkey.NewLuaScript(`local c = redis.call('INCR', KEYS[1])
if c == 1 then redis.call('PEXPIRE', KEYS[1], ARGV[1]) end
return c`)

// Limiter is a fixed-window otp.Limiter.
type Limiter struct{ c valkey.Client }

// NewLimiter returns a Limiter.
func NewLimiter(c valkey.Client) *Limiter { return &Limiter{c: c} }

// Allow implements otp.Limiter.
func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	n, err := incrWindow.Exec(ctx, l.c, []string{"rl:" + key}, []string{ms(window)}).AsInt64()
	if err != nil {
		return false, err
	}
	return n <= int64(limit), nil
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
