-- name: InsertAppAttestKey :execrows
INSERT INTO app_attest_keys (key_id, public_key, prize_eligible, receipt)
VALUES ($1, $2, $3, $4)
ON CONFLICT (key_id) DO NOTHING;

-- name: AppAttestKey :one
SELECT public_key, counter, prize_eligible, receipt FROM app_attest_keys WHERE key_id = $1;

-- name: AdvanceAppAttestCounter :execrows
UPDATE app_attest_keys SET counter = $2 WHERE key_id = $1 AND counter < $2;
