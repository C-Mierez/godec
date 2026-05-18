-- name: CreateTenant :one
INSERT INTO
        tenants (name, email)
VALUES
        ($1, $2)
RETURNING
        *;

-- name: GetTenantByID :one
SELECT
        *
FROM
        tenants
WHERE
        id = $1;

-- name: GetActiveTenantByEmail :one
SELECT
        *
FROM
        tenants
WHERE
        email = $1
        AND status = 'active';

-- name: ListTenants :many
SELECT
        *
FROM
        tenants
ORDER BY
        created_at DESC,
        id DESC
LIMIT
        $1
OFFSET
        $2;

-- name: ListTenantsByStatus :many
SELECT
        *
FROM
        tenants
WHERE
        status = $1
ORDER BY
        created_at DESC,
        id DESC
LIMIT
        $2
OFFSET
        $3;

-- name: ListTenantsByEmail :many
SELECT
        *
FROM
        tenants
WHERE
        email = $1
ORDER BY
        created_at DESC,
        id DESC;

-- name: UpdateTenant :one
UPDATE tenants
SET
        name = $1,
        email = $2
WHERE
        id = $3
RETURNING
        *;

-- name: SetTenantStatus :one
UPDATE tenants
SET
        status = $1
WHERE
        id = $2
RETURNING
        *;

-- name: CountTenants :one
SELECT
        COUNT(*)
FROM
        tenants;

-- name: CountTenantsByStatus :one
SELECT
        COUNT(*)
FROM
        tenants
WHERE
        status = $1;
