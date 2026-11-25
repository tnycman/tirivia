# ADR-0007: Connect RPC as the protobuf-over-HTTPS transport

- **Status:** Accepted (owner, 2026-11-25)
- **Date:** 2026-11-25

## Context ([research 02](../research/02-backend-protocol-data.md))
The Master Plan calls for "Protobuf over HTTPS", with one definition shared by Go, Swift, Kotlin and the web consoles.

Connect RPC fits that requirement:
- Unary calls are plain HTTP `POST` with `application/proto`.
- There are no trailers, so calls pass cleanly through an ALB with HTTP/1.1 targets.
- Read-only calls can use `GET`.
- The Go server library (connect-go) also speaks gRPC and gRPC-Web.
- Client libraries exist for Swift (URLSession), Kotlin (OkHttp) and TypeScript.

## Decision
- Server: connect-go v2.0.2. If v2 causes problems, fall back to v1.21.1, which is "supported indefinitely".
- Mobile clients: connect-swift 1.2.3 and connect-kotlin 0.10.0, both pinned.
- Web consoles: protobuf-es 2.x.
- Schemas: a Buf v2 workspace with lint rules `STANDARD` and breaking-change rules `FILE`.

## Consequences
- connect-kotlin is still pre-1.0, so the protocol test vectors serve as the guard against regressions.
- `connectrpc/conformance` checks protocol behaviour only, not our schemas, so we keep our own test vectors as well.
