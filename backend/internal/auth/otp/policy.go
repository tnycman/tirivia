package otp

import "time"

// Code policy (NIST SP 800-63B-4 §3.1.3.2, §3.2.2; decision ID-2).
const (
	CodeDigits             = 6
	CodeTTL                = 5 * time.Minute
	MaxAttempts            = 5
	MaxConsecutiveFailures = 100
)

// Channel is a code delivery channel.
type Channel string

// Channels (decision ID-1).
const (
	ChannelSMS      Channel = "sms"
	ChannelWhatsApp Channel = "whatsapp"
	ChannelViber    Channel = "viber"
)

var blockedLineTypes = map[string]bool{
	"fixedVoip": true, "nonFixedVoip": true, "tollFree": true, "premium": true,
	"sharedCost": true, "uan": true, "pager": true, "voicemail": true,
}
