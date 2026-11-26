package appattest

import (
	"context"
	"errors"
)

// ErrKeyExists means the key id was already attested; Apple keys are attested once per install.
var ErrKeyExists = errors.New("appattest: key already attested")

// Key is a stored attested key: the uncompressed P-256 point, the last assertion counter and the receipt.
type Key struct {
	ID            []byte
	PublicKey     []byte
	Counter       uint32
	PrizeEligible bool
	Receipt       []byte
}

// Keys persists attested keys.
type Keys interface {
	Insert(ctx context.Context, k Key) error
	Get(ctx context.Context, id []byte) (Key, bool, error)
	AdvanceCounter(ctx context.Context, id []byte, counter uint32) (bool, error)
}
