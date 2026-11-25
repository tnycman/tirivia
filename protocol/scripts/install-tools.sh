#!/usr/bin/env sh
# Installs the pinned protocol toolchain. Versions verified 2026-11-25 (docs/research/02).
set -eu
export GOTOOLCHAIN=auto
go install github.com/bufbuild/buf/cmd/buf@v1.75.0
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
go install connectrpc.com/connect/v2/cmd/protoc-gen-connect-go@v2.0.2
