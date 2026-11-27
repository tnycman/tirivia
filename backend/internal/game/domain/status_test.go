package domain

import "testing"

func TestStatusAt(t *testing.T) {
	for _, tc := range []struct {
		p        Participant
		q        int
		revealed bool
		want     Status
	}{
		{Participant{}, 0, false, StatusAlive},
		{Participant{LastOK: 2}, 3, false, StatusAlive},
		{Participant{LastOK: 2}, 3, true, StatusEliminated},
		{Participant{LastOK: 3}, 3, true, StatusAlive},
		{Participant{LastOK: 1}, 3, false, StatusEliminated},
		{Participant{JoinedAtQ: 2}, 5, true, StatusPlayAlong},
		{Participant{LastOK: 12}, 12, true, StatusAlive},
		{Participant{LastOK: 3}, 3, false, StatusAlive},
	} {
		if got := tc.p.StatusAt(tc.q, tc.revealed); got != tc.want {
			t.Errorf("%+v Q%d revealed=%v: got %v want %v", tc.p, tc.q, tc.revealed, got, tc.want)
		}
	}
}

func TestStatusNeverLeaksCorrectnessBeforeReveal(t *testing.T) {
	for q := 1; q <= QuestionsPerShow; q++ {
		right, wrong := Participant{LastOK: q}, Participant{LastOK: q - 1}
		if right.StatusAt(q, false) != wrong.StatusAt(q, false) {
			t.Fatalf("Q%d: status differs before the reveal", q)
		}
		if right.StatusAt(q, true) == wrong.StatusAt(q, true) {
			t.Fatalf("Q%d: status must differ after the reveal", q)
		}
	}
}
