# syntax=docker/dockerfile:1
# Multi-stage, ARM64-first (Fargate Graviton is ~20% cheaper; stack/09-infrastructure-aws).
# Build: docker buildx build --platform linux/arm64 --build-arg CMD=game -t game .
# Base images must be pinned by digest in CI (Renovate keeps digests current).

ARG GO_VERSION=1.27.4

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build
ARG TARGETOS TARGETARCH CMD
WORKDIR /src
# The protocol SDK is a sibling module (go.mod replace); build from the repo root context.
COPY protocol/gen /protocol/gen
COPY backend/go.mod backend/go.sum ./
RUN sed -i 's#=> ../protocol/gen#=> /protocol/gen#' go.mod && go mod download
COPY backend/ .
RUN sed -i 's#=> ../protocol/gen#=> /protocol/gen#' go.mod && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${CMD}

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app"]
