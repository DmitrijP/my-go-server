-- name: CreateToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, expires_at, user_id)
VALUES (
    $1, NOW(), NOW(), $2, $3
)
RETURNING *;

-- name: GetOneToken :one
SELECT * FROM refresh_tokens WHERE token = $1 ORDER BY created_at DESC LIMIT 1;

-- name: RevokeToken :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW() WHERE token = $1;