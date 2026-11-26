// Package kvstore keeps short-lived auth state in Valkey: rate limits, DPoP replay, OTP challenges, attestation.
package kvstore

import "time"

func ms(d time.Duration) string { return itoa(d.Milliseconds()) }
