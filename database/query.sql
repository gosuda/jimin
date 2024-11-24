-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: MarkEmailVerified :exec
UPDATE users SET email_verified = true WHERE id = $1 AND email = $2;

-- name: CreateUser :one
INSERT INTO users (id, name, email, password, email_verified) VALUES ($1, $2, $3, $4, $5) RETURNING *;