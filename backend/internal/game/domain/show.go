// Package domain holds the pure game rules: the show state machine, answer deadlines and
// elimination. No I/O lives here, so every rule is unit-testable (SPEC §6.2, §6.3).
package domain

import (
	"errors"
	"fmt"
	"time"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
)

// QuestionsPerShow is fixed by the show format (Master Plan §1).
const QuestionsPerShow = 12

// ErrIllegalTransition is returned when a studio command does not fit the current phase.
var ErrIllegalTransition = errors.New("domain: illegal show transition")

// Show is the authoritative show state. Only the studio module mutates it.
type Show struct {
	Phase         cuev1.Phase
	QuestionIndex int // 0 before Q1; 1..12 afterwards
}

// NewShow returns a show in SCHEDULED.
func NewShow() Show { return Show{Phase: cuev1.Phase_PHASE_SCHEDULED} }

// Advance moves the show to the next phase. Allowed transitions:
//
//	SCHEDULED → LOBBY → QUESTION_OPEN → QUESTION_CLOSED → REVEAL
//	REVEAL → QUESTION_OPEN (if questions remain) | FINAL (after Q12)
//	FINAL → SETTLED
func (s Show) Advance(to cuev1.Phase) (Show, error) {
	next := s
	ok := false
	switch {
	case s.Phase == cuev1.Phase_PHASE_SCHEDULED && to == cuev1.Phase_PHASE_LOBBY:
		ok = true
	case s.Phase == cuev1.Phase_PHASE_LOBBY && to == cuev1.Phase_PHASE_QUESTION_OPEN,
		s.Phase == cuev1.Phase_PHASE_REVEAL && to == cuev1.Phase_PHASE_QUESTION_OPEN && s.QuestionIndex < QuestionsPerShow:
		ok = true
		next.QuestionIndex++
	case s.Phase == cuev1.Phase_PHASE_QUESTION_OPEN && to == cuev1.Phase_PHASE_QUESTION_CLOSED,
		s.Phase == cuev1.Phase_PHASE_QUESTION_CLOSED && to == cuev1.Phase_PHASE_REVEAL:
		ok = true
	case s.Phase == cuev1.Phase_PHASE_REVEAL && to == cuev1.Phase_PHASE_FINAL && s.QuestionIndex == QuestionsPerShow:
		ok = true
	case s.Phase == cuev1.Phase_PHASE_FINAL && to == cuev1.Phase_PHASE_SETTLED:
		ok = true
	}
	if !ok {
		return s, fmt.Errorf("%w: %s → %s at Q%d", ErrIllegalTransition, s.Phase, to, s.QuestionIndex)
	}
	next.Phase = to
	return next, nil
}

// Timer bounds (decision UF-3): producers set each question's timer per show in the studio
// console; the system only enforces these limits.
const (
	MinTimer = 6 * time.Second
	MaxTimer = 10 * time.Second
)

// ErrTimerOutOfBounds is returned when a producer-set timer is outside [MinTimer, MaxTimer].
var ErrTimerOutOfBounds = errors.New("domain: question timer must be between 6 s and 10 s")

// ValidateTimer checks a producer-set timer.
func ValidateTimer(d time.Duration) error {
	if d < MinTimer || d > MaxTimer {
		return fmt.Errorf("%w (got %v)", ErrTimerOutOfBounds, d)
	}
	return nil
}

// TimerFor returns the DEFAULT timer the studio console pre-fills: 10 s for Q1-4, 8 s for
// Q5-8, 7 s for Q9-12. Producers may override each value within the bounds above.
func TimerFor(index int) time.Duration {
	switch {
	case index <= 4:
		return 10 * time.Second
	case index <= 8:
		return 8 * time.Second
	default:
		return 7 * time.Second
	}
}

// Window holds the latency allowances. Values are DEFAULTs tuned from rehearsal data.
type Window struct {
	LatencyMax  time.Duration // L_max: p99 cue-to-render allowance (DEFAULT 6 s)
	UploadGrace time.Duration // U_grace: upload allowance (DEFAULT 1.5 s)
}

// DefaultWindow is the starting point before rehearsal measurements.
var DefaultWindow = Window{LatencyMax: 6 * time.Second, UploadGrace: 1500 * time.Millisecond}

// AnswerDeadline is the last server receive time at which an answer is accepted.
// tCue is when PutMetadata returned success: IVS does not guarantee insertion time, so the
// server never trusts client clocks (docs/research/01).
func (w Window) AnswerDeadline(tCue time.Time, timer time.Duration) time.Time {
	return tCue.Add(w.LatencyMax + timer + w.UploadGrace)
}

// CloseAt is when studio emits question_closed.
func (w Window) CloseAt(tCue time.Time, timer time.Duration) time.Time {
	return tCue.Add(w.LatencyMax + timer)
}

// Participant is one player's standing in a show.
type Participant struct {
	// JoinedAtQ is the index of the latest question that had already opened when the player
	// joined (0 = joined before Q1 opened). Players with JoinedAtQ > 0 play along (AC-8).
	JoinedAtQ int
	// LastOK is the last question passed (answered correctly or saved by the extra life).
	LastOK int
	// LifeUsed records the one extra life per show.
	LifeUsed bool
}

// PlaysAlong reports whether the player joined after Q1 opened. They may answer from the
// current question on, for fun and stats, but can never win and are excluded from tallies.
func (p Participant) PlaysAlong() bool { return p.JoinedAtQ > 0 }

// IsAlive reports whether the player is still in contention for question q.
// Elimination needs no scan: a correct answer (or a used extra life) sets LastOK = q.
func (p Participant) IsAlive(q int) bool { return !p.PlaysAlong() && p.LastOK == q-1 }

// CanAnswer reports whether an answer to question q is accepted at all. In-contention players
// answer; play-along players answer any question that opened after they joined.
func (p Participant) CanAnswer(q int) bool {
	if p.PlaysAlong() {
		return q > p.JoinedAtQ
	}
	return p.IsAlive(q)
}

// CanUseExtraLife applies Master Plan §1 and decision UF-4: one life per show, never on the
// final question, and only for a question the player was present for when it opened. Late
// joiners can never use it to cover questions they missed.
func (p Participant) CanUseExtraLife(q int) bool {
	return !p.LifeUsed && !p.PlaysAlong() && q >= 1 && q < QuestionsPerShow && p.LastOK == q-1
}

// IsWinner reports whether the player won: present from the start and passed all questions.
func (p Participant) IsWinner() bool { return !p.PlaysAlong() && p.LastOK == QuestionsPerShow }
