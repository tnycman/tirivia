package attest

import "context"

// ByPlatform routes evidence to the verifier for its platform ("ios", "android").
type ByPlatform map[string]Verifier

// Verify implements Verifier.
func (b ByPlatform) Verify(ctx context.Context, challenge []byte, ev Evidence) (Result, error) {
	v, ok := b[ev.Platform]
	if !ok {
		return Result{}, ErrFailed
	}
	return v.Verify(ctx, challenge, ev)
}
