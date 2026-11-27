-- Users, devices, sessions (SPEC §6.6).

-- +goose Up
CREATE TABLE users (
    id                  uuid PRIMARY KEY,
    phone_hash          bytea       NOT NULL UNIQUE,
    phone_enc           bytea       NOT NULL,
    display_name        text        NOT NULL DEFAULT '',
    city                text        NOT NULL DEFAULT '',
    country_code        char(2)     NOT NULL,
    is_diaspora         boolean     NOT NULL DEFAULT false,
    locale              text        NOT NULL DEFAULT 'sq',
    show_on_winners_wall boolean    NOT NULL DEFAULT false,
    status              text        NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended','deleted')),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE devices (
    id              uuid PRIMARY KEY,
    user_id         uuid        NOT NULL REFERENCES users(id),
    platform        text        NOT NULL CHECK (platform IN ('ios','android')),
    attest_key_id   text,
    dpop_jkt        text        NOT NULL,
    risk_metric     integer,
    first_seen_at   timestamptz NOT NULL DEFAULT now(),
    last_seen_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (dpop_jkt)
);

CREATE INDEX devices_user_idx ON devices (user_id);

CREATE TABLE device_prize_bindings (
    user_id     uuid        PRIMARY KEY REFERENCES users(id),
    platform    text        NOT NULL CHECK (platform IN ('ios','android')),
    device_ref  text        NOT NULL UNIQUE,
    bound_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE push_tokens (
    device_id   uuid        NOT NULL REFERENCES devices(id),
    provider    text        NOT NULL CHECK (provider IN ('apns','fcm','apns_live_activity_start')),
    token       text        NOT NULL,
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (device_id, provider)
);

CREATE TABLE sessions (
    id                  uuid PRIMARY KEY,
    user_id             uuid        NOT NULL REFERENCES users(id),
    device_id           uuid        NOT NULL REFERENCES devices(id),
    refresh_family_id   uuid        NOT NULL,
    refresh_hash        bytea       NOT NULL UNIQUE,
    last_used_at        timestamptz NOT NULL DEFAULT now(),
    rotated_at          timestamptz,
    revoked_at          timestamptz,
    revoke_reason       text
);

CREATE INDEX sessions_family_idx ON sessions (refresh_family_id);

-- +goose Down
DROP TABLE sessions, push_tokens, device_prize_bindings, devices, users;
