-- Shows, questions, play (SPEC §6.6).

-- +goose Up
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
    ciphertext      bytea       NOT NULL,
    data_key_enc    bytea       NOT NULL,
    author_id       uuid        NOT NULL,
    approver_1      uuid,
    approver_2      uuid,
    status          text        NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','in_review','approved','used','retired')),
    created_at      timestamptz NOT NULL DEFAULT now(),

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
    t_cue           timestamptz,
    cue_seq         bigint,
    PRIMARY KEY (show_id, idx),
    UNIQUE (show_id, question_id)
);

CREATE TABLE participations (
    show_id             uuid        NOT NULL REFERENCES shows(id),
    user_id             uuid        NOT NULL REFERENCES users(id),
    joined_at           timestamptz NOT NULL DEFAULT now(),
    joined_at_q         smallint    NOT NULL DEFAULT 0,
    last_ok             smallint    NOT NULL DEFAULT 0,
    extra_life_used_q   smallint,
    prize_eligible      boolean     NOT NULL,
    outcome             text        CHECK (outcome IN ('winner','eliminated','play_along','spectator')),

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
    PRIMARY KEY (show_id, question_id, user_id)
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

-- +goose Down
DROP TABLE show_stats, reaction_sets, friendships, referrals, extra_lives, answers, participations, show_questions, questions, shows, prize_plans, sponsors;
