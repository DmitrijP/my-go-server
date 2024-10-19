-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING *;

-- name: SelectUserByEmail :one
SELECT * FROM users WHERE email like $1 LIMIT 1;

-- name: DeleteAllUsers :exec
DELETE FROM users;