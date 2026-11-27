// Package livestate names and reads the Valkey live-state keys shared by studio and game (SPEC §6.7).
package livestate

import "strconv"

// TallyShards spreads tally writes over this many hash slots (SPEC §6.7).
const TallyShards = 64

// ShowKey is the show-state hash; one hash tag per show.
func ShowKey(show string) string { return "show:{" + show + "}:state" }

// PlayerKey holds a player's standing and answers; tagged per player so scripts stay in one slot.
func PlayerKey(show, user string) string { return "player:{" + show + ":" + user + "}" }

// TallyKey is one tally shard; each shard has its own hash tag.
func TallyKey(show string, q, shard int) string {
	return "tally:{" + show + ":" + strconv.Itoa(q) + ":" + strconv.Itoa(shard) + "}"
}

// AnswerField is the player-hash field holding the answer to question q.
func AnswerField(q int) string { return "a" + strconv.Itoa(q) }
