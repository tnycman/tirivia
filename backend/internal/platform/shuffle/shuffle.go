// Package shuffle implements the per-player option order (SPEC §6.3 step 7).
//
// Every platform (Go, Swift, Kotlin) must produce the same order for the same inputs;
// protocol/testvectors/answer pins this. Clients always submit option_id, never a position,
// so the server never needs to invert the permutation.
package shuffle

import (
	"crypto/hmac"
	"crypto/sha256"
)

// MaxOptions is the largest n supported (one HMAC byte per swap).
const MaxOptions = sha256.Size + 1

// Order returns a permutation of [0, n): position p shows option Order(...)[p].
//
//	k = HMAC-SHA256(key = seed, msg = questionID)
//	for i = n-1 down to 1: j = k[n-1-i] mod (i+1); swap(idx[i], idx[j])
//
// The modulo bias for n = 3 (256 mod 3 = 1) is irrelevant for display order.
func Order(seed []byte, questionID string, n int) []int {
	if n < 1 || n > MaxOptions {
		panic("shuffle: n out of range")
	}
	mac := hmac.New(sha256.New, seed)
	mac.Write([]byte(questionID))
	k := mac.Sum(nil)

	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	for i := n - 1; i >= 1; i-- {
		j := int(k[n-1-i]) % (i + 1)
		idx[i], idx[j] = idx[j], idx[i]
	}
	return idx
}
