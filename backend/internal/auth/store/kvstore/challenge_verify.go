package kvstore

import (
	"context"
	"crypto/hmac"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"

	"trivia.al/backend/internal/auth/otp"
)

var verifyChallenge = valkey.NewLuaScript(`if redis.call('HGET', KEYS[1], 'data') ~= ARGV[1] then return -1 end
local n = redis.call('HINCRBY', KEYS[1], 'attempts', 1)
if n > tonumber(ARGV[2]) then redis.call('DEL', KEYS[1]); return -2 end
if ARGV[3] == '1' or n == tonumber(ARGV[2]) then redis.call('DEL', KEYS[1]) end
return n`)

// Verify implements otp.Store atomically.
func (s *Challenges) Verify(ctx context.Context, id, expectedJKT string, codeMAC []byte, now time.Time, maxAttempts int) (otp.Challenge, error) {
	ch, sealed, err := s.get(ctx, id)
	if err != nil {
		return otp.Challenge{}, err
	}
	if ch.Device.JKT != expectedJKT {
		return otp.Challenge{}, otp.ErrBinding
	}
	if now.After(ch.ExpiresAt) {
		return otp.Challenge{}, otp.ErrExpired
	}
	ok := hmac.Equal(codeMAC, ch.CodeMAC)
	success := "0"
	if ok {
		success = "1"
	}
	n, err := verifyChallenge.Exec(ctx, s.c, []string{challengeKey(id)}, []string{sealed, strconv.Itoa(maxAttempts), success}).AsInt64()
	if err != nil {
		return otp.Challenge{}, err
	}
	if n == -1 {
		return otp.Challenge{}, otp.ErrExpired
	}
	if n == -2 {
		return otp.Challenge{}, otp.ErrTooManyAttempts
	}
	ch.Attempts = int(n)
	if !ok {
		return ch, otp.ErrWrongCode
	}
	return ch, nil
}
