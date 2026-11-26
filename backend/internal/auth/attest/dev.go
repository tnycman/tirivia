package attest

import (
	"context"

	"github.com/google/uuid"
)

// DevVerifier accepts any non-empty evidence; local development and tests only.
type DevVerifier struct{}

// Verify implements Verifier.
func (DevVerifier) Verify(_ context.Context, _ []byte, ev Evidence) (Result, error) {
	switch {
	case ev.Platform == "ios" && ev.AppAttestKeyID != "":
		return Result{DeviceRef: "ios:" + ev.AppAttestKeyID, PrizeEligible: true}, nil
	case ev.Platform == "android" && ev.PlayIntegrityToken != "":
		id, err := uuid.NewV7()
		if err != nil {
			return Result{}, err
		}
		return Result{DeviceRef: "android:" + id.String(), PrizeEligible: true}, nil
	}
	return Result{}, ErrFailed
}
