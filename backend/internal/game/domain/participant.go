package domain

// Participant is one player's standing in a show.
type Participant struct {
	JoinedAtQ int
	LastOK    int
	LifeUsed  bool
}

// PlaysAlong reports a late joiner: may answer later questions, never wins (AC-8).
func (p Participant) PlaysAlong() bool { return p.JoinedAtQ > 0 }

// IsAlive reports whether the player is in contention for question q.
func (p Participant) IsAlive(q int) bool { return !p.PlaysAlong() && p.LastOK == q-1 }

// CanAnswer reports whether an answer to q is accepted.
func (p Participant) CanAnswer(q int) bool {
	if p.PlaysAlong() {
		return q > p.JoinedAtQ
	}
	return p.IsAlive(q)
}

// CanUseExtraLife applies UF-4: once, never on Q12, only for a question the player was present for.
func (p Participant) CanUseExtraLife(q int) bool {
	return !p.LifeUsed && !p.PlaysAlong() && q >= 1 && q < QuestionsPerShow && p.LastOK == q-1
}

// IsWinner reports an on-time player who passed all questions.
func (p Participant) IsWinner() bool { return !p.PlaysAlong() && p.LastOK == QuestionsPerShow }
