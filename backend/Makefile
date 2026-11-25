# Pinned tool versions (docs/research/02, verified 2026-11-25)
GOLANGCI_LINT_VERSION := v2.15.0
GOOSE_VERSION         := v3.30.1
SQLC_VERSION          := v1.31.1
DB_URL                ?= postgres://postgres:test@localhost:55432/tirivia?sslmode=disable

.PHONY: build test lint arch vuln migrate db-test ci

build:
	go build ./...

test:
	go test -race -count=1 ./...

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

arch:
	go run github.com/fe3dback/go-arch-lint@latest check

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

migrate:
	go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir migrations/core postgres "$(DB_URL)" up

# Runs migrations/tests/invariants.sql against a throwaway Postgres 17 container.
db-test:
	docker run -d --rm --name tirivia-pg-test -e POSTGRES_PASSWORD=test -e POSTGRES_DB=tirivia -p 55432:5432 postgres:17-alpine
	until docker exec tirivia-pg-test pg_isready -U postgres -d tirivia; do sleep 1; done
	$(MAKE) migrate
	docker cp migrations/tests/invariants.sql tirivia-pg-test:/tmp/invariants.sql
	docker exec tirivia-pg-test psql -U postgres -d tirivia -v ON_ERROR_STOP=0 -f /tmp/invariants.sql > invariants.log 2>&1; \
	  docker stop tirivia-pg-test; cat invariants.log; \
	  errors=$$(grep -c 'ERROR:' invariants.log); rm -f invariants.log; \
	  if [ "$$errors" -ne 10 ]; then echo "expected exactly 10 rejected statements (MUST FAIL steps), got $$errors"; exit 1; fi; \
	  echo "schema invariants OK (10/10 rejections, all MUST PASS steps accepted)"

ci: build test lint arch vuln
