-- name: InsertSession :exec
INSERT INTO sessions (id, user_id, device_id, refresh_family_id, refresh_hash, created_at, last_used_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: SessionByRefreshHash :one
SELECT s.id, s.user_id, s.device_id, s.refresh_family_id, d.dpop_jkt, s.created_at, s.last_used_at,
       s.rotated_at, s.revoked_at, s.revoke_reason
FROM sessions s JOIN devices d ON d.id = s.device_id
WHERE s.refresh_hash = $1;

-- name: MarkSessionRotated :execrows
UPDATE sessions SET rotated_at = $2 WHERE id = $1 AND rotated_at IS NULL;

-- name: RevokeSessionFamily :exec
UPDATE sessions SET revoked_at = $3, revoke_reason = $2
WHERE refresh_family_id = $1 AND revoked_at IS NULL;
