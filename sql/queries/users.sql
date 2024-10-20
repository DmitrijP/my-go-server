-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING *;

-- name: SelectUserByEmail :one
SELECT * FROM users WHERE email like $1 LIMIT 1;

-- name: SelectUserById :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: UpdateEmailAndPassword :exec
UPDATE users SET hashed_password = $1, email =$2, updated_at = NOW() WHERE id = $3;

-- name: UpdateChirpyRed :exec
UPDATE users SET is_chirpy_red = $1, updated_at = NOW() WHERE id = $2;