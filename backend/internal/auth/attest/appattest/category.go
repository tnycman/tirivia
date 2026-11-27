package appattest

import (
	"encoding/binary"
	"slices"
	"strconv"
)

// Validation categories (Apple, "Validating apps that connect to your server", attestation step 10).
const (
	CategoryTestFlight  uint32 = 2
	CategoryDevelopment uint32 = 3
	CategoryAppStore    uint32 = 4
)

// rejected are Invalid (0), OS executable (1) and the restricted system categories (7, 8, 9).
var rejected = map[uint32]bool{0: true, 1: true, 7: true, 8: true, 9: true}

// CheckCategory rejects Apple's invalid and restricted categories and reports prize eligibility. A client cannot
// strip the category (authData is covered by the Apple-signed nonce), so an absent one (older iOS) defers to the
// production AAGUID; a present one must be in prizeCategories (App Store).
func CheckCategory(e Extensions, prizeCategories []uint32) (prize bool, err error) {
	if !e.HasCategory {
		return true, nil
	}
	if rejected[e.Category] {
		return false, ErrCategory
	}
	return slices.Contains(prizeCategories, e.Category), nil
}

// Extensions are the authData extension values (attestation steps 10-11).
type Extensions struct {
	HasCategory   bool
	Category      uint32
	BundleVersion string
}

// readExtensions accepts a CBOR uint or a 4-byte string; byte order is little-endian, the only reading
// of Apple's sample (01 00 00 00) that lands in Apple's table. Confirm on the first real build (docs/research/08).
func readExtensions(m map[string]any, categoryKey, versionKey string) Extensions {
	var e Extensions
	switch v := m[categoryKey].(type) {
	case uint64:
		if v <= 0xFFFFFFFF {
			e.HasCategory, e.Category = true, uint32(v)
		}
	case []byte:
		if len(v) == 4 {
			e.HasCategory, e.Category = true, binary.LittleEndian.Uint32(v)
		}
	}
	e.BundleVersion, _ = m[versionKey].(string)
	return e
}

// Signals reports the extension values for the risk score.
func (e Extensions) Signals() []string {
	s := []string{}
	if e.HasCategory {
		s = append(s, "aa.category="+strconv.FormatUint(uint64(e.Category), 10))
	} else {
		s = append(s, "aa.category=absent")
	}
	if e.BundleVersion != "" {
		s = append(s, "aa.bundle_version="+e.BundleVersion)
	}
	return s
}
