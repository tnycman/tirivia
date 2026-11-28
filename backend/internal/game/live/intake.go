package live

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"

	"trivia.al/backend/internal/platform/livestate"
)

// Errors for answers that are not recorded.
var (
	ErrNotJoined   = errors.New("live: player has not joined")
	ErrNotAlive    = errors.New("live: player cannot answer this question")
	ErrDuplicate   = errors.New("live: question already answered")
	ErrClosed      = errors.New("live: answer window closed")
	ErrNotQuestion = errors.New("live: not the open question")
)

// Outcome of a recorded answer; Counted is false for play-along players (excluded from tallies, AC-8).
type Outcome struct {
	Correct bool
	Counted bool
}

// intakeScript mirrors domain.Participant.CanAnswer and counts a new on-time answer in the same atomic step (TS-16).
// Every check runs before the first write, because a script is not rolled back on error.
var intakeScript = valkey.NewLuaScript(guards + `local joined = tonumber(redis.call('HGET', KEYS[1], 'joined_at_q'))
if not joined then return 'not_joined' end
local q = tonumber(ARGV[1])
if redis.call('HEXISTS', KEYS[1], ARGV[4]) == 1 then return 'duplicate' end
if not counter(KEYS[2], ARGV[2], 1) or not hashOrNone(KEYS[3]) then return redis.error_reply(malformed) end
local last_ok = tonumber(redis.call('HGET', KEYS[1], 'last_ok'))
local late = joined > 0
if late then
  if q <= joined then return 'not_alive' end
elseif last_ok ~= q - 1 then return 'not_alive' end
redis.call('HSET', KEYS[1], ARGV[4], ARGV[2])
local correct = ARGV[2] == ARGV[3]
if correct and not late then redis.call('HSET', KEYS[1], 'last_ok', q) end
if late then return correct and 'late_correct' or 'late_wrong' end
if ARGV[5] == '1' then
  redis.call('HINCRBY', KEYS[2], ARGV[2], 1)
  redis.call('HSET', KEYS[3], ARGV[6], ARGV[2])
end
return correct and 'correct' or 'wrong'`)

var outcomes = map[string]Outcome{
	"correct": {Correct: true, Counted: true}, "wrong": {Counted: true},
	"late_correct": {Correct: true}, "late_wrong": {},
}

var rejections = map[string]error{"not_joined": ErrNotJoined, "duplicate": ErrDuplicate, "not_alive": ErrNotAlive}

// Answers records answers against the open question.
type Answers struct{ c valkey.Client }

// NewAnswers returns Answers.
func NewAnswers(c valkey.Client) *Answers { return &Answers{c: c} }

// Submit records option for question q if received by the server deadline; correctness is computed on intake.
// A voided answer is recorded but never tallied.
func (a *Answers) Submit(ctx context.Context, show, user string, st livestate.Show, questionID, option string, voided bool, received time.Time) (Outcome, error) {
	if questionID != st.QuestionID {
		return Outcome{}, ErrNotQuestion
	}
	if received.After(st.Deadline) {
		return Outcome{}, ErrClosed
	}
	q, b, count := st.QuestionIndex, livestate.Bucket(user), "1"
	if voided {
		count = "0"
	}
	res, err := intakeScript.Exec(ctx, a.c,
		[]string{livestate.PlayerKey(show, user), livestate.TallyKey(show, q, b), livestate.CountedKey(show, q, b)},
		[]string{strconv.Itoa(q), option, st.CorrectOption, livestate.AnswerField(q), count, user}).ToString()
	if err != nil {
		return Outcome{}, err
	}
	if o, ok := outcomes[res]; ok {
		o.Counted = o.Counted && !voided
		return o, nil
	}
	if e, ok := rejections[res]; ok {
		return Outcome{}, e
	}
	return Outcome{}, errors.New("live: unexpected intake result " + res)
}
