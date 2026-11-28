package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"trivia.al/backend/internal/platform/shuffle"
	tv "trivia.al/backend/internal/platform/testvectors"
)

func shuffleSuite() tv.ShuffleSuite {
	s := tv.ShuffleSuite{
		FormatVersion: tv.FormatVersion,
		Algorithm:     "k = HMAC-SHA256(key=seed, msg=UTF-8(question_id)); for i = n-1 down to 1: j = k[n-1-i] mod (i+1) (k bytes unsigned); swap(idx[i], idx[j])",
	}
	add := func(name, seedHex, qid string, n int) {
		seed := must(hex.DecodeString(seedHex))
		s.Vectors = append(s.Vectors, tv.ShuffleVector{Name: name, SeedHex: seedHex, QuestionID: qid, N: n, Order: shuffle.Order(seed, qid, n)})
	}
	for i := range 12 {
		seed := sha256.Sum256([]byte(fmt.Sprintf("player-%d", i)))
		add(fmt.Sprintf("player_%02d_q%d", i, i%12+1), hex.EncodeToString(seed[:16]), fmt.Sprintf("0193a1b3-0000-7000-8000-%012d", i), 3)
	}
	add("utf8_question_id", "00112233445566778899aabbccddeeff", "pyetja-ë-ç-12", 3)
	add("high_bytes_mask_check", "ffffffffffffffffffffffffffffffff", "q-high-bytes", 3)
	add("single_option", "01", "q", 1)
	add("four_options", "deadbeef", "q-four", 4)
	return s
}
