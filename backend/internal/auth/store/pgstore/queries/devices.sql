-- name: RegisterDevice :one
INSERT INTO devices (id, user_id, platform, attest_key_id, dpop_jkt, prize_eligible)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (dpop_jkt) DO UPDATE SET last_seen_at = now(), prize_eligible = EXCLUDED.prize_eligible
WHERE devices.user_id = EXCLUDED.user_id
RETURNING id;
