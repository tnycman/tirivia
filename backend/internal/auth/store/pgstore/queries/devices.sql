-- name: RegisterDevice :one
INSERT INTO devices (id, user_id, platform, attest_key_id, dpop_jkt)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (dpop_jkt) DO UPDATE SET last_seen_at = now() WHERE devices.user_id = EXCLUDED.user_id
RETURNING id;
