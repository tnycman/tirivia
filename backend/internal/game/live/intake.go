package live

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/tirivia/backend/internal/platform/livestate"
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

// intakeScript mirrors domain.Participant.CanAnswer; a test checks the two agree on every state.
var intakeScript = valkey.NewLuaScript(`local joined = tonumber(redis.call('HGET', KEYS[1], 'joined_at_q'))
if not joined then return 'not_joined' end
local q = tonumber(ARGV[1])
if redis.call('HEXISTS', KEYS[1], ARGV[4]) == 1 then return 'duplicate' end
local last_ok = tonumber(redis.call('HGET', KEYS[1], 'last_ok'))
local late = joined > 0
if late then
  if q <= joined then return 'not_alive' end
elseif last_ok ~= q - 1 then return 'not_alive' end
redis.call('HSET', KEYS[1], ARGV[4], ARGV[2])
local correct = ARGV[2] == ARGV[3]
if correct and not late then redis.call('HSET', KEYS[1], 'last_ok', q) end
if late then return correct and 'late_correct' or 'late_wrong' end
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
func (a *Answers) Submit(ctx context.Context, show, user string, st livestate.Show, questionID, option string, received time.Time) (Outcome, error) {
	if questionID != st.QuestionID {
		return Outcome{}, ErrNotQuestion
	}
	if received.After(st.Deadline) {
		return Outcome{}, ErrClosed
	}
	res, err := intakeScript.Exec(ctx, a.c, []string{livestate.PlayerKey(show, user)},
		[]string{strconv.Itoa(st.QuestionIndex), option, st.CorrectOption, livestate.AnswerField(st.QuestionIndex)}).ToString()
	if err != nil {
		return Outcome{}, err
	}
	if o, ok := outcomes[res]; ok {
		return o, nil
	}
	if e, ok := rejections[res]; ok {
		return Outcome{}, e
	}
	return Outcome{}, errors.New("live: unexpected intake result " + res)
}
