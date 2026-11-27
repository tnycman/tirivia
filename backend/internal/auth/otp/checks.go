package otp

import "context"

func rangeKey(phone string, digits int) string {
	if n := digits + 1; len(phone) > n {
		return phone[:n]
	}
	return phone
}

func (s *Service) rateLimited(ctx context.Context, l Limits, phoneHash, deviceID, phone string) (bool, error) {
	for _, c := range []struct {
		key string
		lim Limit
	}{
		{"otp:resend:" + phoneHash, l.Resend},
		{"otp:phone:" + phoneHash, l.PerPhone},
		{"otp:device:" + deviceID, l.PerDevice},
		{"otp:range:" + rangeKey(phone, l.RangeDigits), l.PerRange},
	} {
		ok, err := s.cfg.Limiter.Allow(ctx, c.key, c.lim.Max, c.lim.Window)
		if err != nil || !ok {
			return true, err
		}
	}
	n, err := s.cfg.Failures.Count(ctx, phoneHash)
	return n >= MaxConsecutiveFailures, err
}

func (s *Service) numberBlocked(ctx context.Context, l Limits, phone string) (bool, error) {
	lineType, risk, err := s.cfg.Intel.Lookup(ctx, phone)
	if err != nil {
		return false, err
	}
	return blockedLineTypes[lineType] || risk >= l.PumpingBlockAt, nil
}
