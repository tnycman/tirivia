package domain

import "time"

// Window holds the latency allowances; tuned from rehearsal measurements.
type Window struct {
	LatencyMax  time.Duration
	UploadGrace time.Duration
}

// DefaultWindow is the starting point before rehearsal data (SPEC §6.3).
var DefaultWindow = Window{LatencyMax: 6 * time.Second, UploadGrace: 1500 * time.Millisecond}

// AnswerDeadline is the last accepted receive time; tCue is the PutMetadata success time.
func (w Window) AnswerDeadline(tCue time.Time, timer time.Duration) time.Time {
	return tCue.Add(w.LatencyMax + timer + w.UploadGrace)
}

// CloseAt is when question_closed is emitted.
func (w Window) CloseAt(tCue time.Time, timer time.Duration) time.Time {
	return tCue.Add(w.LatencyMax + timer)
}
