-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
VALUES (sqlc.arg('user_id'), sqlc.arg('token_hash'), sqlc.arg('expires_at'))
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT *
FROM refresh_tokens
WHERE token_hash = sqlc.arg('token_hash');

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE id = sqlc.arg('id');

-- name: RevokeAllUserTokens :exec
UPDATE refresh_tokens
SET revoked_at = now()
WHERE user_id = sqlc.arg('user_id') AND revoked_at IS NULL;
