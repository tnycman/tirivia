# syntax=docker/dockerfile:1
# Multi-stage, ARM64-first (Fargate Graviton is ~20% cheaper; stack/09-infrastructure-aws).

ARG GO_IMAGE=golang:1.27.4-alpine@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280
ARG RUNTIME_IMAGE=gcr.io/distroless/static-debian12:nonroot@sha256:afa5c872c891853ca7fcf1f12c3edb23f7eeef36189728842dd51042ff57f7ab

FROM --platform=$BUILDPLATFORM ${GO_IMAGE} AS build
ARG TARGETOS TARGETARCH CMD
WORKDIR /src
# The protocol SDK is a sibling module (go.mod replace); rewrite the path for the container.
COPY protocol/gen /protocol/gen
COPY backend/go.mod backend/go.sum ./
RUN sed -i 's#=> ../protocol/gen#=> /protocol/gen#' go.mod && go mod download
COPY backend/ .
RUN sed -i 's#=> ../protocol/gen#=> /protocol/gen#' go.mod && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${CMD}

FROM ${RUNTIME_IMAGE}
COPY --from=build /out/app /app
USER nonroot:nonroot
# 8080 public API (game), 8081 core, 9090 internal admin (game; never on the public ALB)
EXPOSE 8080 8081 9090
ENTRYPOINT ["/app"]
