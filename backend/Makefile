# Pinned tool versions (docs/research/02, verified 2026-11-25)
GOLANGCI_LINT_VERSION := v2.15.0
GOOSE_VERSION         := v3.30.1
SQLC_VERSION          := v1.31.1
DB_URL                ?= postgres://postgres:test@localhost:55432/trivia?sslmode=disable
PG_IMAGE              := postgres:17-alpine@sha256:f9cf3e1f0066f0eb0c48707d864f0ac5941e9ff454a072077d97a4d278d07d73
VALKEY_IMAGE          := valkey/valkey:9.1-alpine@sha256:56729e68bdda139cbd3f6a2538683dc3523325fc5711ea1a379938f1b37893f0
STORE_DB_URL          := postgres://postgres:test@localhost:55433/trivia?sslmode=disable
STORE_KV_URL          := valkey://localhost:56379

.PHONY: build test lint arch vuln migrate db-test sqlc sqlc-check store-test ci

build:
	go build ./...

test:
	go test -race -count=1 ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

arch:
	go run github.com/fe3dback/go-arch-lint@v1.19.2 check

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

migrate:
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir migrations/core postgres "$(DB_URL)" up

# Runs migrations/tests/invariants.sql against a throwaway Postgres 17 container.
db-test:
	docker run -d --rm --name trivia-pg-test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=trivia -p 55432:5432 $(PG_IMAGE)
	until docker exec trivia-pg-test pg_isready -U postgres -d trivia; do sleep 1; done
	$(MAKE) migrate
	docker cp migrations/tests/invariants.sql trivia-pg-test:/tmp/invariants.sql
	docker exec trivia-pg-test psql -U postgres -d trivia -v ON_ERROR_STOP=0 -f /tmp/invariants.sql > invariants.log 2>&1; \
	  docker stop trivia-pg-test; cat invariants.log; \
	  errors=$$(grep -c 'ERROR:' invariants.log); rm -f invariants.log; \
	  if [ "$$errors" -ne 10 ]; then echo "expected exactly 10 rejected statements (MUST FAIL steps), got $$errors"; exit 1; fi; \
	  echo "schema invariants OK (10/10 rejections, all MUST PASS steps accepted)"

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

sqlc-check:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) diff

# Runs the store and live-state tests against throwaway Postgres 17 and Valkey 9.1 containers.
store-test:
	docker run -d --rm --name trivia-pg-store -e POSTGRES_PASSWORD=test -e POSTGRES_DB=trivia -p 55433:5432 $(PG_IMAGE)
	docker run -d --rm --name trivia-kv-store -p 56379:6379 $(VALKEY_IMAGE)
	until docker exec trivia-pg-store pg_isready -U postgres -d trivia; do sleep 1; done
	$(MAKE) migrate DB_URL="$(STORE_DB_URL)"
	TEST_DATABASE_URL="$(STORE_DB_URL)" TEST_VALKEY_URL="$(STORE_KV_URL)" go test -race -count=1 ./internal/auth/store/... ./internal/game/... ./cmd/game/...; \
	  status=$?; docker stop trivia-pg-store trivia-kv-store; exit $status

ci: build test lint arch vuln vectors sqlc-check store-test

vectors:
	go run ./cmd/testvectors -check
	node ../protocol/testvectors/runners/node/verify-vectors.mjs
