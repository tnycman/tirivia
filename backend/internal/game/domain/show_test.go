package domain

import (
	"errors"
	"testing"
	"time"

	cuev1 "github.com/tirivia/protocol/gen/go/tirivia/cue/v1"
)

func TestFullShowRunsTwelveQuestions(t *testing.T) {
	s := NewShow()
	var err error
	step := func(to cuev1.Phase) {
		t.Helper()
		if s, err = s.Advance(to); err != nil {
			t.Fatal(err)
		}
	}
	step(cuev1.Phase_PHASE_LOBBY)
	for q := 1; q <= QuestionsPerShow; q++ {
		step(cuev1.Phase_PHASE_QUESTION_OPEN)
		if s.QuestionIndex != q {
			t.Fatalf("question index %d, want %d", s.QuestionIndex, q)
		}
		step(cuev1.Phase_PHASE_QUESTION_CLOSED)
		step(cuev1.Phase_PHASE_REVEAL)
	}
	step(cuev1.Phase_PHASE_FINAL)
	step(cuev1.Phase_PHASE_SETTLED)
}

func TestIllegalTransitions(t *testing.T) {
	cases := []struct {
		name string
		from Show
		to   cuev1.Phase
	}{
		{"skip lobby", Show{Phase: cuev1.Phase_PHASE_SCHEDULED}, cuev1.Phase_PHASE_QUESTION_OPEN},
		{"reveal before close", Show{Phase: cuev1.Phase_PHASE_QUESTION_OPEN, QuestionIndex: 1}, cuev1.Phase_PHASE_REVEAL},
		{"final too early", Show{Phase: cuev1.Phase_PHASE_REVEAL, QuestionIndex: 11}, cuev1.Phase_PHASE_FINAL},
		{"thirteenth question", Show{Phase: cuev1.Phase_PHASE_REVEAL, QuestionIndex: 12}, cuev1.Phase_PHASE_QUESTION_OPEN},
		{"reopen settled", Show{Phase: cuev1.Phase_PHASE_SETTLED, QuestionIndex: 12}, cuev1.Phase_PHASE_LOBBY},
	}
	for _, tc := range cases {
		got, err := tc.from.Advance(tc.to)
		if !errors.Is(err, ErrIllegalTransition) {
			t.Errorf("%s: expected ErrIllegalTransition, got %v", tc.name, err)
		}
		if got != tc.from {
			t.Errorf("%s: state changed on rejected transition", tc.name)
		}
	}
}

func TestTimers(t *testing.T) {
	want := map[int]time.Duration{1: 10 * time.Second, 4: 10 * time.Second, 5: 8 * time.Second, 8: 8 * time.Second, 9: 7 * time.Second, 12: 7 * time.Second}
	for q, d := range want {
		if got := TimerFor(q); got != d {
			t.Errorf("Q%d: %v want %v", q, got, d)
		}
	}
}

func TestAnswerDeadlineAnchoredOnServerCue(t *testing.T) {
	tCue := time.Date(2026, 11, 25, 21, 3, 0, 0, time.UTC)
	got := DefaultWindow.AnswerDeadline(tCue, 10*time.Second)
	if want := tCue.Add(17500 * time.Millisecond); !got.Equal(want) {
		t.Fatalf("deadline %v want %v", got, want)
	}
	if !DefaultWindow.CloseAt(tCue, 10*time.Second).Before(got) {
		t.Fatal("question_closed must precede the acceptance deadline (upload grace)")
	}
}

func TestEliminationAndLives(t *testing.T) {
	onTime := func(lastOK int, used bool) Participant { return Participant{LastOK: lastOK, LifeUsed: used} }
	if !onTime(0, false).IsAlive(1) || !onTime(6, false).IsAlive(7) || onTime(5, false).IsAlive(7) {
		t.Fatal("IsAlive wrong")
	}
	if onTime(11, false).CanUseExtraLife(12) {
		t.Fatal("extra life must not be usable on the final question")
	}
	if onTime(4, true).CanUseExtraLife(5) || !onTime(4, false).CanUseExtraLife(5) {
		t.Fatal("one life per show")
	}
	if !onTime(12, false).IsWinner() || onTime(11, false).IsWinner() {
		t.Fatal("winner = present from the start and passed all 12")
	}
}

// Decision AC-8 + UF-4: late joiners play along from the current question but can never win,
// and an extra life cannot cover questions that opened before they joined.
func TestLateJoinerPlaysAlongButCannotWin(t *testing.T) {
	late := Participant{JoinedAtQ: 2} // joined while Q2 was open
	if !late.PlaysAlong() {
		t.Fatal("joined after Q1 opened must play along")
	}
	if late.CanAnswer(2) {
		t.Fatal("the question already open at join time is missed")
	}
	if !late.CanAnswer(3) || !late.CanAnswer(12) {
		t.Fatal("play-along players answer later questions")
	}
	if late.IsAlive(3) {
		t.Fatal("play-along players are never in contention (excluded from tallies)")
	}
	if late.CanUseExtraLife(3) {
		t.Fatal("extra life cannot rescue a late joiner")
	}
	late.LastOK = QuestionsPerShow // even answering everything afterwards
	if late.IsWinner() {
		t.Fatal("late joiners can never win")
	}
}

func TestProducerTimerBounds(t *testing.T) {
	for _, d := range []time.Duration{6 * time.Second, 7 * time.Second, 10 * time.Second} {
		if err := ValidateTimer(d); err != nil {
			t.Errorf("%v rejected: %v", d, err)
		}
	}
	for _, d := range []time.Duration{5 * time.Second, 11 * time.Second, 0} {
		if err := ValidateTimer(d); !errors.Is(err, ErrTimerOutOfBounds) {
			t.Errorf("%v accepted", d)
		}
	}
	for q := 1; q <= QuestionsPerShow; q++ {
		if err := ValidateTimer(TimerFor(q)); err != nil {
			t.Errorf("default for Q%d out of bounds", q)
		}
	}
}
