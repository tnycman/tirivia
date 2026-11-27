-- name: InsertSession :exec
INSERT INTO sessions (id, user_id, device_id, refresh_family_id, refresh_hash, created_at, last_used_at,
                      revoked_at, revoke_reason)
SELECT $1, $2, $3, $4, $5, $6, $7, r.revoked_at, r.revoke_reason
FROM (SELECT 1) AS seed
LEFT JOIN session_family_revocations r ON r.refresh_family_id = $4;

-- name: SessionByRefreshHash :one
SELECT s.id, s.user_id, s.device_id, s.refresh_family_id, d.dpop_jkt, s.created_at, s.last_used_at,
       s.rotated_at, s.revoked_at, s.revoke_reason
FROM sessions s JOIN devices d ON d.id = s.device_id
WHERE s.refresh_hash = $1;

-- name: MarkSessionRotated :execrows
UPDATE sessions SET rotated_at = $2
WHERE id = $1 AND rotated_at IS NULL AND revoked_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM session_family_revocations r WHERE r.refresh_family_id = sessions.refresh_family_id
  );

-- name: RecordSessionFamilyRevocation :exec
INSERT INTO session_family_revocations (refresh_family_id, revoke_reason, revoked_at)
VALUES ($1, $2, $3)
ON CONFLICT (refresh_family_id) DO NOTHING;

-- name: ApplySessionFamilyRevocation :exec
UPDATE sessions s SET revoked_at = r.revoked_at, revoke_reason = r.revoke_reason
FROM session_family_revocations r
WHERE s.refresh_family_id = $1 AND r.refresh_family_id = s.refresh_family_id AND s.revoked_at IS NULL;
