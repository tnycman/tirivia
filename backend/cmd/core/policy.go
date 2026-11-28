package main

import (
	"time"

	"trivia.al/backend/internal/auth/attest"
	"trivia.al/backend/internal/auth/otp"
)

// Sources for every value: stack/05-auth-identity "Sourced values" table.
var (
	coreRegions = []string{"AL", "XK", "MK", "IT", "GR", "DE", "GB", "US", "FR", "CH", "SI", "HR", "FI"}

	coreLimits = otp.Limits{
		Resend:         otp.Limit{Max: 1, Window: 30 * time.Second},
		PerPhone:       otp.Limit{Max: 5, Window: 10 * time.Minute},
		PerDevice:      otp.Limit{Max: 5, Window: 10 * time.Minute},
		PerRange:       otp.Limit{Max: 100, Window: 10 * time.Minute},
		RangeDigits:    7,
		PumpingBlockAt: 90,
	}

	otherLimits = otp.Limits{
		Resend:         otp.Limit{Max: 1, Window: time.Minute},
		PerPhone:       otp.Limit{Max: 3, Window: 10 * time.Minute},
		PerDevice:      otp.Limit{Max: 3, Window: 10 * time.Minute},
		PerRange:       otp.Limit{Max: 20, Window: 10 * time.Minute},
		RangeDigits:    7,
		PumpingBlockAt: 75,
	}

	attestTTLs = attest.TTLs{Challenge: 5 * time.Minute, Ticket: 10 * time.Minute}
	dpopWindow = 60 * time.Second
)
