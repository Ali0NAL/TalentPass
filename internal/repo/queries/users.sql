-- name: CreateUser :one
INSERT INTO users (email, password_hash)
VALUES (sqlc.arg('email'), sqlc.arg('password_hash'))
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = sqlc.arg('email');
