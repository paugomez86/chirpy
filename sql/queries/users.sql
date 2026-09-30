-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT *
FROM users;

-- name: GetUserFromEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserFromRefreshToken :one
SELECT users.*
FROM users
INNER JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1;

-- name: UpdateUser :one
UPDATE users
SET
    email = $2,
    hashed_password = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpgradeUserRed :one
UPDATE users
SET 
    is_chirpy_red = TRUE,
    updated_at = NOW()
WHERE id = $1
RETURNING *;