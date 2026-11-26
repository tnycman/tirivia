// Package domain holds the pure game rules (SPEC §6.2, §6.3).
package domain

import (
	"errors"
	"fmt"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
)

// QuestionsPerShow is fixed by the show format.
const QuestionsPerShow = 12

// ErrIllegalTransition rejects a studio command that does not fit the phase.
var ErrIllegalTransition = errors.New("domain: illegal show transition")

// Show is the authoritative show state.
type Show struct {
	Phase         cuev1.Phase
	QuestionIndex int
}

// NewShow returns a scheduled show.
func NewShow() Show { return Show{Phase: cuev1.Phase_PHASE_SCHEDULED} }

// Advance moves to the next phase if the transition is allowed.
func (s Show) Advance(to cuev1.Phase) (Show, error) {
	next, ok := s, false
	switch {
	case s.Phase == cuev1.Phase_PHASE_SCHEDULED && to == cuev1.Phase_PHASE_LOBBY,
		s.Phase == cuev1.Phase_PHASE_QUESTION_OPEN && to == cuev1.Phase_PHASE_QUESTION_CLOSED,
		s.Phase == cuev1.Phase_PHASE_QUESTION_CLOSED && to == cuev1.Phase_PHASE_REVEAL,
		s.Phase == cuev1.Phase_PHASE_REVEAL && to == cuev1.Phase_PHASE_FINAL && s.QuestionIndex == QuestionsPerShow,
		s.Phase == cuev1.Phase_PHASE_FINAL && to == cuev1.Phase_PHASE_SETTLED:
		ok = true
	case s.Phase == cuev1.Phase_PHASE_LOBBY && to == cuev1.Phase_PHASE_QUESTION_OPEN,
		s.Phase == cuev1.Phase_PHASE_REVEAL && to == cuev1.Phase_PHASE_QUESTION_OPEN && s.QuestionIndex < QuestionsPerShow:
		ok = true
		next.QuestionIndex++
	}
	if !ok {
		return s, fmt.Errorf("%w: %s → %s at Q%d", ErrIllegalTransition, s.Phase, to, s.QuestionIndex)
	}
	next.Phase = to
	return next, nil
}
