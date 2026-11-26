-- name: FailureCount :one
SELECT consecutive FROM otp_failures WHERE phone_hash = $1;

-- name: IncrementFailures :one
INSERT INTO otp_failures (phone_hash, consecutive) VALUES ($1, 1)
ON CONFLICT (phone_hash) DO UPDATE SET consecutive = otp_failures.consecutive + 1, updated_at = now()
RETURNING consecutive;

-- name: ResetFailures :exec
DELETE FROM otp_failures WHERE phone_hash = $1;
