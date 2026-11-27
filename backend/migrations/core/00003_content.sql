-- Clips, likes, winners wall (SPEC §6.6).

-- +goose Up
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
    PRIMARY KEY (item_id, user_id)
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

-- +goose Down
DROP TABLE winners_wall, item_views, item_likes, clips;
