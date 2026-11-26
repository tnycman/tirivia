package cuesign

import (
	"fmt"
	"unicode/utf8"
)

// ValidateQuestionContent enforces the cue budget when a question is saved.
func ValidateQuestionContent(text string, options []string) error {
	if !utf8.ValidString(text) || len(text) == 0 || len(text) > MaxQuestionTextBytes {
		return fmt.Errorf("%w: question text must be 1-%d UTF-8 bytes", ErrContent, MaxQuestionTextBytes)
	}
	if len(options) != OptionsPerQuestion {
		return fmt.Errorf("%w: exactly %d options required", ErrContent, OptionsPerQuestion)
	}
	for i, o := range options {
		if !utf8.ValidString(o) || len(o) == 0 || len(o) > MaxOptionTextBytes {
			return fmt.Errorf("%w: option %d must be 1-%d UTF-8 bytes", ErrContent, i, MaxOptionTextBytes)
		}
	}
	return nil
}
