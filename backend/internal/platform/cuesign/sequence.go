package cuesign

import cuev1 "trivia.al/protocol/gen/go/trivia/cue/v1"

// CheckSequence drops cues for other shows and non-increasing seq; gap reports a needed resync.
func CheckSequence(body *cuev1.CueBody, expectedShowID string, lastSeq uint64) (gap bool, err error) {
	if body.GetShowId() != expectedShowID {
		return false, ErrWrongShow
	}
	if body.GetSeq() <= lastSeq {
		return false, ErrReplay
	}
	return body.GetSeq() > lastSeq+1, nil
}
