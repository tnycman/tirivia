-- Feature switches, eligibility, audit log (SPEC §6.6).

-- +goose Up
CREATE TABLE feature_switches (
    key         text        NOT NULL,
    scope       text        NOT NULL CHECK (scope IN ('global','show')),
    scope_id    text        NOT NULL DEFAULT '',
    enabled     boolean     NOT NULL,
    changed_by  uuid        NOT NULL,
    changed_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (key, scope, scope_id),
    CHECK ((scope = 'global') = (scope_id = ''))
);

CREATE TABLE country_eligibility (
    country_code    char(2)     PRIMARY KEY,
    eligibility     text        NOT NULL CHECK (eligibility IN ('eligible','play_for_points','blocked')),


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
    prev_hash   bytea,
    hash        bytea       NOT NULL
);

CREATE TRIGGER audit_log_append_only
    BEFORE UPDATE OR DELETE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION forbid_mutation();

-- +goose Down
DROP TABLE audit_log, country_eligibility, feature_switches;
