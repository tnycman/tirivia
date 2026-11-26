package otp

import (
	"errors"
	"time"
)

// Config wires the service.
type Config struct {
	Pepper      []byte
	CoreRegions []string
	CoreLimits  Limits
	OtherLimits Limits
	Channels    map[Channel][]Sender
	Tickets     Tickets
	Limiter     Limiter
	Failures    Failures
	Intel       NumberIntel
	Signups     Signups
	Directory   Directory
	Store       Store
}

// Service runs the code pipeline.
type Service struct {
	cfg  Config
	core map[string]bool
	now  func() time.Time
}

// New validates cfg.
func New(cfg Config) (*Service, error) {
	switch {
	case len(cfg.Pepper) < 32:
		return nil, errors.New("otp: pepper must be at least 32 bytes")
	case len(cfg.Channels[ChannelSMS]) == 0:
		return nil, errors.New("otp: at least one SMS provider is required")
	case !cfg.CoreLimits.valid() || !cfg.OtherLimits.valid():
		return nil, errors.New("otp: limits are not configured")
	case cfg.Failures == nil:
		return nil, errors.New("otp: failure counter is required")
	}
	core := map[string]bool{}
	for _, r := range cfg.CoreRegions {
		core[r] = true
	}
	return &Service{cfg: cfg, core: core, now: time.Now}, nil
}
