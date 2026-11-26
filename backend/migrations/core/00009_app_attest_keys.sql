-- Attested App Attest keys: public key, last assertion counter (Apple assertion step 5) and receipt.

-- +goose Up
CREATE TABLE app_attest_keys (
    key_id          bytea       PRIMARY KEY CHECK (octet_length(key_id) = 32),
    public_key      bytea       NOT NULL CHECK (octet_length(public_key) = 65),
    counter         bigint      NOT NULL DEFAULT 0 CHECK (counter BETWEEN 0 AND 4294967295),
    prize_eligible  boolean     NOT NULL,
    receipt         bytea,
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE app_attest_keys;
