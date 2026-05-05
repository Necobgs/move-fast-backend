
-- name: FindUserByEmail :one
SELECT 
    users.id, 
    name, 
    email, 
    password,
    photo_url,
    drivers.id as "driver_id"
FROM 
    users
left join drivers on drivers.user_id = users.id
WHERE email ilike $1;

-- name: FindSafeUserByEmail :one
SELECT id,name,email,created_at,deleted_at, photo_url FROM users WHERE email = $1;

-- name: ExistsUserByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: CreateUser :one
INSERT INTO users (id, name, email, password,photo_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, email,photo_url;

-- name: UpdateUser :exec
UPDATE users
SET name = $2, email = $3, password = $4, deleted_at = $5
WHERE id = $1;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = $1;
