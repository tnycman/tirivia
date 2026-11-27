package appattest

import (
	"encoding/binary"
)

type authData struct {
	rpIDHash   []byte
	flags      byte
	counter    uint32
	aaguid     []byte
	credID     []byte
	extensions map[string]any
}

// parseAuthData reads rpIdHash(32) flags(1) counter(4), credential data (attestations only), then extensions.
// Apple sets flags inconsistently (assertions carry 0x40 without credential data; extensions come without 0x80),
// so the object type, not the flags, decides the layout; trailing bytes are the extensions map.
func parseAuthData(b []byte, withCredential bool) (authData, error) {
	if len(b) < 37 {
		return authData{}, ErrFormat
	}
	a := authData{rpIDHash: b[:32], flags: b[32], counter: binary.BigEndian.Uint32(b[33:37])}
	rest := b[37:]
	if withCredential {
		if len(rest) < 18 {
			return authData{}, ErrFormat
		}
		a.aaguid = rest[:16]
		n := int(binary.BigEndian.Uint16(rest[16:18]))
		if len(rest) < 18+n {
			return authData{}, ErrFormat
		}
		a.credID, rest = rest[18:18+n], rest[18+n:]
		var cose any
		var err error
		if rest, err = decoder.UnmarshalFirst(rest, &cose); err != nil {
			return authData{}, ErrFormat
		}
	}
	if len(rest) > 0 {
		var err error
		if rest, err = decoder.UnmarshalFirst(rest, &a.extensions); err != nil {
			return authData{}, ErrFormat
		}
	}
	if len(rest) != 0 {
		return authData{}, ErrFormat
	}
	return a, nil
}
