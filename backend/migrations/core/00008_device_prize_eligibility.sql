-- Latest attestation outcome per device: false means points only (AC-5).

-- +goose Up
ALTER TABLE devices ADD COLUMN prize_eligible boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE devices DROP COLUMN prize_eligible;
