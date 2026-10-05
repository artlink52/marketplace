-- name: CreateUser :one
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetUserByEmail :one
SELECT id, password_hash, role, status
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT email, role, created_at, updated_at
FROM users
WHERE id = $1;
