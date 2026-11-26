package otp

import "time"

// Limit is one fixed-window rate limit.
type Limit struct {
	Max    int
	Window time.Duration
}

// Limits are one tier's send limits; values are config (stack/05-auth-identity).
type Limits struct {
	Resend         Limit
	PerPhone       Limit
	PerDevice      Limit
	PerRange       Limit
	RangeDigits    int
	PumpingBlockAt int
}

func (l Limits) valid() bool {
	for _, x := range []Limit{l.Resend, l.PerPhone, l.PerDevice, l.PerRange} {
		if x.Max <= 0 || x.Window <= 0 {
			return false
		}
	}
	return l.RangeDigits > 0 && l.PumpingBlockAt > 0
}

func (s *Service) limitsFor(region string) Limits {
	if s.core[region] {
		return s.cfg.CoreLimits
	}
	return s.cfg.OtherLimits
}
