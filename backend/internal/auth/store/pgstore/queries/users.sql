-- name: UserExists :one
SELECT EXISTS (SELECT 1 FROM users WHERE phone_hash = $1);

-- name: InsertUser :one
INSERT INTO users (id, phone_hash, phone_enc, country_code)
VALUES ($1, $2, $3, $4)
ON CONFLICT (phone_hash) DO NOTHING
RETURNING id;

-- name: UserByPhone :one
SELECT id FROM users WHERE phone_hash = $1;
