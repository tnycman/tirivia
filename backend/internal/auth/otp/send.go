package otp

import "context"

func (s *Service) send(ctx context.Context, channel Channel, phone, code, locale string) (Channel, string, error) {
	order := []Channel{channel}
	if channel != ChannelSMS {
		order = append(order, ChannelSMS)
	}
	for _, c := range order {
		for _, p := range s.cfg.Channels[c] {
			if err := p.Send(ctx, phone, code, locale); err == nil {
				return c, p.Name(), nil
			}
		}
	}
	return "", "", ErrSendFailed
}
