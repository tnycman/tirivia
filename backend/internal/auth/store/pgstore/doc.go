// Package pgstore persists users, devices, sessions and OTP failure counts in Postgres via sqlc (pgq).
package pgstore

import (
	"encoding/hex"

	"github.com/google/uuid"
)

func phoneKey(phoneHash string) ([]byte, error) { return hex.DecodeString(phoneHash) }

func parseIDs(ids ...string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, len(ids))
	for i, s := range ids {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		out[i] = id
	}
	return out, nil
}
