// Package livestate names and reads the Valkey live-state keys shared by studio and game (SPEC §6.7).
package livestate

import (
	"hash/fnv"
	"strconv"
)

// TallyShards is the number of buckets per show; policy value proven in the 600k load test (TS-16).
const TallyShards = 64

// ShowKey is the show-state hash; one hash tag per show.
func ShowKey(show string) string { return "show:{" + show + "}:state" }

// Bucket places a player; a player's keys and its bucket's tallies share one hash tag, so one script stays in one slot.
func Bucket(user string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(user))
	return int(h.Sum32() % TallyShards)
}

func tag(show string, bucket int) string { return "{" + show + ":" + strconv.Itoa(bucket) + "}" }

// PlayerKey holds a player's standing and answers.
func PlayerKey(show, user string) string { return "player:" + tag(show, Bucket(user)) + ":" + user }

// TallyKey is one bucket's option → count hash for question q.
func TallyKey(show string, q, bucket int) string {
	return "tally:" + tag(show, bucket) + ":" + strconv.Itoa(q)
}

// CountedKey is one bucket's user → option hash of the answers the tally counted, used by the recount.
func CountedKey(show string, q, bucket int) string {
	return "counted:" + tag(show, bucket) + ":" + strconv.Itoa(q)
}

// AnswerField is the player-hash field holding the answer to question q.
func AnswerField(q int) string { return "a" + strconv.Itoa(q) }
