-- Auth persistence: session creation time and the NIST 800-63B-4 §3.2.2 consecutive-failure counter.

-- +goose Up
ALTER TABLE sessions ADD COLUMN created_at timestamptz NOT NULL DEFAULT now();

CREATE TABLE otp_failures (
    phone_hash   bytea       PRIMARY KEY,
    consecutive  integer     NOT NULL CHECK (consecutive > 0),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE otp_failures;
ALTER TABLE sessions DROP COLUMN created_at;
