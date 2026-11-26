module github.com/tirivia/backend

go 1.27.0

toolchain go1.27.4

require (
	connectrpc.com/connect/v2 v2.0.2
	github.com/go-jose/go-jose/v4 v4.1.9
	github.com/google/uuid v1.6.0
	github.com/nyaruka/phonenumbers v1.8.1
	github.com/tirivia/protocol/gen v0.0.0
	google.golang.org/protobuf v1.36.12
)

require golang.org/x/text v0.23.0 // indirect

// Local development: the protocol repo is a sibling folder. In CI the generated SDK is
// consumed at a tagged version instead (see protocol/README.md).
replace github.com/tirivia/protocol/gen => ../protocol/gen
