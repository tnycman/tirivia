// Package accounts finds or creates users and registers their attested devices.
package accounts

import "errors"

// ErrKeyBound means the DPoP key already belongs to another account.
var ErrKeyBound = errors.New("accounts: device key bound to another user")

// NewUser is what a verified code proves about a phone.
type NewUser struct {
	PhoneHash, PhoneE164, Country string
}
