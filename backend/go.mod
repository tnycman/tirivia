module github.com/tirivia/backend

go 1.27.0

toolchain go1.27.4

require (
	connectrpc.com/connect/v2 v2.0.2
	github.com/tirivia/protocol/gen v0.0.0
	google.golang.org/protobuf v1.36.12
)

// Local development: the protocol repo is a sibling folder. In CI the generated SDK is
// consumed at a tagged version instead (see protocol/README.md).
replace github.com/tirivia/protocol/gen => ../protocol/gen
