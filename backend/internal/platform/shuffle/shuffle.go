// Package shuffle implements the per-player option order (SPEC §6.3), pinned by protocol/testvectors.
package shuffle

import (
	"crypto/hmac"
	"crypto/sha256"
)

// MaxOptions is the largest n supported (one HMAC byte per swap).
const MaxOptions = sha256.Size + 1

// Order returns a permutation of [0, n) via HMAC-SHA256(seed, questionID)-driven Fisher-Yates.
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
