-- Double-entry ledger (append-only, zero-sum) (SPEC §6.6).

-- +goose Up
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
    kind            text        NOT NULL,
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

-- +goose Down
DROP TABLE ledger_entries, ledger_transactions, ledger_accounts;
DROP FUNCTION check_transaction_balanced();
DROP FUNCTION forbid_mutation();
