-- Initial core schema (SPEC §6.6). Aurora PostgreSQL 17.
-- Conventions:
--   * Primary keys are app-generated UUIDv7 (no serial/nextval: avoids RDS Proxy session pinning).
--   * Money is bigint minor units + ISO 4217 currency. Never float, never the money type.
--   * timestamptz everywhere.
--   * ledger_entries and audit_log are append-only (enforced by trigger, plus REVOKE in bootstrap).

-- +goose Up

CREATE TABLE users (
    id                  uuid PRIMARY KEY,
    phone_hash          bytea       NOT NULL UNIQUE,      -- HMAC-SHA256(phone_e164, pepper)
    phone_enc           bytea       NOT NULL,             -- KMS envelope-encrypted E.164
    display_name        text        NOT NULL DEFAULT '',
    city                text        NOT NULL DEFAULT '',
    country_code        char(2)     NOT NULL,             -- ISO 3166-1 alpha-2; XK = Kosovo
    is_diaspora         boolean     NOT NULL DEFAULT false,
    locale              text        NOT NULL DEFAULT 'sq',
    show_on_winners_wall boolean    NOT NULL DEFAULT false,      -- opt-in, asked at claim (UF-10)
    status              text        NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended','deleted')),
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE devices (
    id              uuid PRIMARY KEY,
    user_id         uuid        NOT NULL REFERENCES users(id),
    platform        text        NOT NULL CHECK (platform IN ('ios','android')),
    attest_key_id   text,                                  -- App Attest key id
    dpop_jkt        text        NOT NULL,                  -- RFC 7638 thumbprint of the DPoP key
    risk_metric     integer,                               -- App Attest risk metric
    first_seen_at   timestamptz NOT NULL DEFAULT now(),
    last_seen_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (dpop_jkt)
);
CREATE INDEX devices_user_idx ON devices (user_id);

-- One prize-eligible account per physical device (AC-6). Other accounts on that device play
-- for points. device_ref: iOS = App Attest key id of the bound account, with the per-device
-- DeviceCheck bit set so other accounts on the device are detected; Android = HMAC of the
-- app-scoped ANDROID_ID (Play Integrity device recall to be evaluated in Phase 1).
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
    refresh_hash        bytea       NOT NULL UNIQUE,      -- hash of the current opaque refresh token
    last_used_at        timestamptz NOT NULL DEFAULT now(),
    rotated_at          timestamptz,
    revoked_at          timestamptz,
    revoke_reason       text
);
CREATE INDEX sessions_family_idx ON sessions (refresh_family_id);

CREATE TABLE sponsors (
    id          uuid PRIMARY KEY,
    name        text NOT NULL,
    branding    jsonb NOT NULL DEFAULT '{}'
);

CREATE TABLE prize_plans (
    id              uuid PRIMARY KEY,
    currency        char(3) NOT NULL,
    pot_minor       bigint  NOT NULL CHECK (pot_minor >= 0),
    sponsor_prizes  jsonb   NOT NULL DEFAULT '[]'
);

CREATE TABLE shows (
    id              uuid PRIMARY KEY,
    starts_at       timestamptz NOT NULL,
    state           text        NOT NULL DEFAULT 'SCHEDULED'
                    CHECK (state IN ('SCHEDULED','LOBBY','QUESTION_OPEN','QUESTION_CLOSED','REVEAL','FINAL','SETTLED')),
    rehearsal       boolean     NOT NULL DEFAULT false,
    sponsor_id      uuid        REFERENCES sponsors(id),
    prize_plan_id   uuid        REFERENCES prize_plans(id),
    ivs_channel_arn text,
    region          text        NOT NULL DEFAULT 'eu-central-1'
);
CREATE INDEX shows_starts_idx ON shows (starts_at);

CREATE TABLE questions (
    id              uuid PRIMARY KEY,
    kind            text        NOT NULL CHECK (kind IN ('general','show_dependent','audio','visual','sequence')),
    difficulty      smallint    NOT NULL CHECK (difficulty BETWEEN 1 AND 5),
    ciphertext      bytea       NOT NULL,                  -- vault: per-question data key (KMS)
    data_key_enc    bytea       NOT NULL,
    author_id       uuid        NOT NULL,
    approver_1      uuid,
    approver_2      uuid,
    status          text        NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','in_review','approved','used','retired')),
    created_at      timestamptz NOT NULL DEFAULT now(),
    -- Four-eyes (NIST 800-53 AC-3(2)): two distinct approvers, neither the author.
    CONSTRAINT questions_four_eyes CHECK (
        status NOT IN ('approved','used') OR (
            approver_1 IS NOT NULL AND approver_2 IS NOT NULL
            AND approver_1 <> approver_2
            AND approver_1 <> author_id AND approver_2 <> author_id
        )
    )
);

CREATE TABLE show_questions (
    show_id         uuid        NOT NULL REFERENCES shows(id),
    idx             smallint    NOT NULL CHECK (idx BETWEEN 1 AND 12),
    question_id     uuid        NOT NULL REFERENCES questions(id),
    timer_ms        integer     NOT NULL,
    released_at     timestamptz,
    t_cue           timestamptz,                           -- PutMetadata success time (deadline anchor)
    cue_seq         bigint,
    PRIMARY KEY (show_id, idx),
    UNIQUE (show_id, question_id)
);

CREATE TABLE participations (
    show_id             uuid        NOT NULL REFERENCES shows(id),
    user_id             uuid        NOT NULL REFERENCES users(id),
    joined_at           timestamptz NOT NULL DEFAULT now(),
    joined_at_q         smallint    NOT NULL DEFAULT 0,      -- > 0 = joined late: plays along, cannot win (AC-8)
    last_ok             smallint    NOT NULL DEFAULT 0,
    extra_life_used_q   smallint,
    prize_eligible      boolean     NOT NULL,                -- false: failed attestation (AC-5), 2nd account on device (AC-6), country switch
    outcome             text        CHECK (outcome IN ('winner','eliminated','play_along','spectator')),
    -- A late joiner can never be a winner, and a life can only be used by an on-time player.
    CHECK (NOT (joined_at_q > 0 AND outcome = 'winner')),
    CHECK (NOT (joined_at_q > 0 AND extra_life_used_q IS NOT NULL)),
    PRIMARY KEY (show_id, user_id)
);

CREATE TABLE answers (
    show_id             uuid        NOT NULL,
    question_id         uuid        NOT NULL,
    user_id             uuid        NOT NULL,
    option_id           text        NOT NULL,
    correct             boolean     NOT NULL,
    accepted            boolean     NOT NULL,
    server_elapsed_ms   integer     NOT NULL,
    client_elapsed_ms   integer,
    signals             jsonb       NOT NULL DEFAULT '{}',
    received_at         timestamptz NOT NULL,
    PRIMARY KEY (show_id, question_id, user_id)            -- natural idempotency key
);

CREATE TABLE extra_lives (
    id              uuid PRIMARY KEY,
    user_id         uuid        NOT NULL REFERENCES users(id),
    source          text        NOT NULL CHECK (source IN ('referral','promo')),
    granted_at      timestamptz NOT NULL DEFAULT now(),
    used_show_id    uuid        REFERENCES shows(id)
);

CREATE TABLE referrals (
    referrer_id     uuid        NOT NULL REFERENCES users(id),
    referee_id      uuid        NOT NULL REFERENCES users(id) UNIQUE,
    status          text        NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','verified','completed_game')),
    granted_at      timestamptz,
    PRIMARY KEY (referrer_id, referee_id),
    CHECK (referrer_id <> referee_id)
);

CREATE TABLE friendships (
    user_id     uuid NOT NULL REFERENCES users(id),
    friend_id   uuid NOT NULL REFERENCES users(id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, friend_id),
    CHECK (user_id <> friend_id)
);

CREATE TABLE reaction_sets (
    id              uuid PRIMARY KEY,
    show_id         uuid NOT NULL REFERENCES shows(id),
    kinds           smallint[] NOT NULL,
    themed_icon_url text
);

CREATE TABLE show_stats (
    show_id             uuid PRIMARY KEY REFERENCES shows(id),
    peak_viewers        bigint,
    players_per_q       bigint[],
    reactions_total     bigint
);

CREATE TABLE clips (
    id          uuid PRIMARY KEY,
    show_id     uuid        NOT NULL REFERENCES shows(id),
    s3_key      text        NOT NULL,
    duration_ms integer     NOT NULL,
    views       bigint      NOT NULL DEFAULT 0,
    likes       bigint      NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE item_likes (
    item_id     uuid NOT NULL,
    user_id     uuid NOT NULL REFERENCES users(id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (item_id, user_id)                          -- one like per account per item
);

CREATE TABLE item_views (
    item_id     uuid        NOT NULL,
    user_id     uuid        NOT NULL,
    viewed_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE winners_wall (
    show_id         uuid        NOT NULL REFERENCES shows(id),
    user_id         uuid        NOT NULL REFERENCES users(id),
    first_name      text        NOT NULL,
    city            text        NOT NULL,
    currency        char(3)     NOT NULL,
    amount_minor    bigint      NOT NULL,
    opted_out       boolean     NOT NULL DEFAULT false,
    PRIMARY KEY (show_id, user_id)
);

-- ---------------------------------------------------------------- ledger (double entry)
CREATE TABLE ledger_accounts (
    id              uuid PRIMARY KEY,
    owner_kind      text    NOT NULL CHECK (owner_kind IN ('user','prize_pool','payout_clearing','charity','sponsor')),
    owner_id        uuid,
    currency        char(3) NOT NULL,
    balance_minor   bigint  NOT NULL DEFAULT 0,
    version         bigint  NOT NULL DEFAULT 0,
    allow_negative  boolean NOT NULL DEFAULT false,
    CHECK (allow_negative OR balance_minor >= 0)
);

CREATE TABLE ledger_transactions (
    id              uuid PRIMARY KEY,
    idempotency_key text        NOT NULL UNIQUE,
    kind            text        NOT NULL,                  -- prize_credit, payout, expiry, reversal, ...
    reverses_id     uuid        REFERENCES ledger_transactions(id),
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE ledger_entries (
    id                  uuid PRIMARY KEY,
    transaction_id      uuid    NOT NULL REFERENCES ledger_transactions(id),
    account_id          uuid    NOT NULL REFERENCES ledger_accounts(id),
    amount_minor        bigint  NOT NULL CHECK (amount_minor <> 0),
    prev_balance_minor  bigint  NOT NULL,
    new_balance_minor   bigint  NOT NULL,
    account_version     bigint  NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now(),
    CHECK (new_balance_minor = prev_balance_minor + amount_minor),
    UNIQUE (account_id, account_version)
);
CREATE INDEX ledger_entries_txn_idx ON ledger_entries (transaction_id);

-- +goose StatementBegin
CREATE FUNCTION forbid_mutation() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION '% is append-only', TG_TABLE_NAME;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER ledger_entries_append_only
    BEFORE UPDATE OR DELETE ON ledger_entries
    FOR EACH ROW EXECUTE FUNCTION forbid_mutation();

-- Every transaction's entries must sum to zero; checked at COMMIT (deferred).
-- +goose StatementBegin
CREATE FUNCTION check_transaction_balanced() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE
    total bigint;
BEGIN
    SELECT COALESCE(SUM(amount_minor), 0) INTO total
    FROM ledger_entries WHERE transaction_id = NEW.transaction_id;
    IF total <> 0 THEN
        RAISE EXCEPTION 'ledger transaction % is unbalanced (sum %)', NEW.transaction_id, total;
    END IF;
    RETURN NULL;
END;
$$;
-- +goose StatementEnd

CREATE CONSTRAINT TRIGGER ledger_entries_balanced
    AFTER INSERT ON ledger_entries
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW EXECUTE FUNCTION check_transaction_balanced();

-- ---------------------------------------------------------------- claims, KYC, risk
CREATE TABLE claims (
    id              uuid PRIMARY KEY,
    user_id         uuid        NOT NULL REFERENCES users(id),
    currency        char(3)     NOT NULL,
    amount_minor    bigint      NOT NULL CHECK (amount_minor > 0),
    tier            char(1)     NOT NULL CHECK (tier IN ('A','B','C')),
    deadline        timestamptz NOT NULL,
    status          text        NOT NULL DEFAULT 'open'
                    CHECK (status IN ('open','verification_required','under_review','approved','paid','expired','rejected')),
    collection      text        CHECK (collection IN ('bank_transfer_al','in_person_office','representative_poa','charity_donation','sponsor_pickup_or_delivery')),
    iban_enc        bytea,
    approver_1      uuid,
    approver_2      uuid,
    created_at      timestamptz NOT NULL DEFAULT now(),
    -- Payout release is four-eyes.
    CONSTRAINT claims_four_eyes CHECK (
        status <> 'paid' OR (approver_1 IS NOT NULL AND approver_2 IS NOT NULL AND approver_1 <> approver_2)
    )
);

CREATE TABLE kyc_records (
    id              uuid PRIMARY KEY,
    user_id         uuid        NOT NULL REFERENCES users(id),
    vendor          text        NOT NULL,
    vendor_ref      text        NOT NULL,
    result          text        NOT NULL CHECK (result IN ('approved','declined','resubmit','review')),
    liveness_score  real,
    s3_key_enc      bytea,                                  -- pointer to SSE-KMS object (KYC CMK)
    retention_until timestamptz NOT NULL,                   -- set by the DPIA retention schedule
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE risk_flags (
    id          uuid PRIMARY KEY,
    user_id     uuid        NOT NULL REFERENCES users(id),
    show_id     uuid        REFERENCES shows(id),
    score       real        NOT NULL,
    verdict     text        NOT NULL CHECK (verdict IN ('clear','review','block')),
    reasons     text[]      NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------- switches, eligibility, audit
CREATE TABLE feature_switches (
    key         text        NOT NULL,                       -- e.g. chat.enabled
    scope       text        NOT NULL CHECK (scope IN ('global','show')),
    scope_id    text        NOT NULL DEFAULT '',            -- show id when scope = 'show'
    enabled     boolean     NOT NULL,
    changed_by  uuid        NOT NULL,
    changed_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (key, scope, scope_id),
    CHECK ((scope = 'global') = (scope_id = ''))
);

CREATE TABLE country_eligibility (
    country_code    char(2)     PRIMARY KEY,
    eligibility     text        NOT NULL CHECK (eligibility IN ('eligible','play_for_points','blocked')),
    -- Reference for the change: an owner decision (e.g. 'LG-1 owner 2026-11-25') or a counsel
    -- opinion id. Required for every row so each country's status is traceable (LG-1).
    decision_ref    text        NOT NULL CHECK (length(trim(decision_ref)) > 0),
    changed_by      uuid        NOT NULL,
    changed_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE audit_log (
    id          uuid PRIMARY KEY,
    at          timestamptz NOT NULL DEFAULT now(),
    actor_id    uuid        NOT NULL,
    action      text        NOT NULL,
    entity      text        NOT NULL,
    entity_id   text        NOT NULL,
    before      jsonb,
    after       jsonb,
    approver_id uuid,
    ip          inet,
    device      text,
    request_id  text,
    prev_hash   bytea,                                      -- hash chain: tamper evidence
    hash        bytea       NOT NULL
);

CREATE TRIGGER audit_log_append_only
    BEFORE UPDATE OR DELETE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION forbid_mutation();

-- +goose Down
DROP TABLE audit_log, country_eligibility, feature_switches, risk_flags, kyc_records, claims,
    ledger_entries, ledger_transactions, ledger_accounts, winners_wall, item_views, item_likes,
    clips, show_stats, reaction_sets, friendships, referrals, extra_lives, answers,
    participations, show_questions, questions, shows, prize_plans, sponsors, sessions,
    push_tokens, device_prize_bindings, devices, users;
DROP FUNCTION check_transaction_balanced();
DROP FUNCTION forbid_mutation();
