package otp

import "github.com/nyaruka/phonenumbers"

var acceptedTypes = map[phonenumbers.PhoneNumberType]bool{
	phonenumbers.MOBILE:               true,
	phonenumbers.FIXED_LINE_OR_MOBILE: true,
}

// parsePhone validates an E.164 number with libphonenumber and returns its region.
func parsePhone(e164 string) (region string, ok bool) {
	p, err := phonenumbers.Parse(e164, "")
	if err != nil || !phonenumbers.IsValidNumber(p) || !acceptedTypes[phonenumbers.GetNumberType(p)] {
		return "", false
	}
	if phonenumbers.Format(p, phonenumbers.E164) != e164 {
		return "", false
	}
	return phonenumbers.GetRegionCodeForNumber(p), true
}
