-- Claims, KYC, risk (SPEC §6.6).

-- +goose Up
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
    s3_key_enc      bytea,
    retention_until timestamptz NOT NULL,
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

-- +goose Down
DROP TABLE risk_flags, kyc_records, claims;
