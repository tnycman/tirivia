package domain

import "time"

// AC-3: on the last three questions an answer is voided by going to background or an inactive gap above 1.5 s.
const (
	FocusRuleFromQuestion = QuestionsPerShow - 2
	InactiveVoidAfter     = 1500 * time.Millisecond
)

// FocusLossVoids applies AC-3; shorter gaps are risk signals only.
func FocusLossVoids(q int, wentBackground bool, longestInactive time.Duration) bool {
	return q >= FocusRuleFromQuestion && (wentBackground || longestInactive > InactiveVoidAfter)
}
