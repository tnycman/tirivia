# Verification (step 5)

The owner runs every command below. Hand them over as one block, or as a scratchpad script with a single `pwsh -File` line, ideally printing PASS or FAIL per check. Base claims only on the output they paste back.

A claim needs the command, its output, and the number of skipped tests. In this repo the store and live-state tests skip silently when `TEST_DATABASE_URL` and `TEST_VALKEY_URL` are unset, and that is how untested Lua once looked green.

## Real stores

- Start Docker Desktop if it is stopped (`Start-Process "C:\Program Files\Docker\Docker\Docker Desktop.exe"`), then poll `docker info` until it answers.
- Use the pinned images and ports from `backend/Makefile` (`make store-test` does this on Linux). On Windows, run them directly:
  - Valkey: `docker run -d --rm --name tirivia-kv-store -p 56379:6379 <VALKEY_IMAGE from Makefile>`
  - Postgres: `docker run -d --rm --name tirivia-pg-store -e POSTGRES_PASSWORD=test -e POSTGRES_DB=tirivia -p 55433:5432 <PG_IMAGE>`
  - Migrate: `go run github.com/pressly/goose/v3/cmd/goose@v3.30.1 -dir migrations/core postgres <url> up`
- Set `TEST_DATABASE_URL=postgres://postgres:test@localhost:55433/tirivia?sslmode=disable` and `TEST_VALKEY_URL=valkey://localhost:56379`, then run `go test -count=1 ./...`.
- Count the skips with `go test -v ... | Select-String '^--- SKIP'`. The answer must be 0 for the packages you touched.
- To check that keys share a slot, a standalone Valkey isn't enough. Run a cluster-enabled container (`valkey-server --cluster-enabled yes`) and compare `CLUSTER KEYSLOT` for each script's keys.
- Tell the owner the containers are running, and stop them when the work is finished.

## Checks for every change

- `go build ./...`, `go vet ./...`, and golangci-lint v2.15.0 (`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...`): 0 issues.
- Schema changes: `sqlc diff` is clean, the new migration only expands, `goose down` then `up` works, and `make db-test` invariants still hold.
- Protos: `buf lint`, `buf format -d`, breaking-change check, regenerated `gen/`, updated test vectors.

## Mutation check

For each guard or fix, remove it briefly, run its test, and confirm that the test fails with the expected message. Then restore the original and rerun. A test that never failed proves nothing. This follows Google's diff-based mutation testing (Petrović & Ivanković, ICSE-SEIP 2018).

## Measured claims

When a scale or speed question matters (recount at 600k, script throughput), write a benchmark gated on the test env vars. Run it, then record the result and the hardware in the research file. A local single node is not ElastiCache, so label the number that way and leave the final value to the load test.
