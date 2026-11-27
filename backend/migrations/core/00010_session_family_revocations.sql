-- Durable refresh-token family revocations.

-- +goose Up
CREATE TABLE session_family_revocations (
    refresh_family_id uuid PRIMARY KEY,
    revoked_at        timestamptz NOT NULL,
    revoke_reason     text        NOT NULL
);

-- +goose Down
DROP TABLE session_family_revocations;
