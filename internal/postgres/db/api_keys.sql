-- name: CreateAPIKey :one
INSERT INTO
        api_keys (tenant_id, name, token_id, hashed_secret, scopes)
VALUES
        ($1, $2, $3, $4, $5)
RETURNING
        *;

-- name: GetAPIKeyByID :one
SELECT
        *
FROM
        api_keys
WHERE
        id = $1;

-- name: GetAPIKeyByTokenID :one
SELECT
        *
FROM
        api_keys
WHERE
        token_id = $1;

-- name: ListAPIKeysByTenantID :many
SELECT
        *
FROM
        api_keys
WHERE
        tenant_id = $1
ORDER BY
        created_at DESC,
        id DESC;

-- name: ListActiveAPIKeysByTenantID :many
SELECT
        *
FROM
        api_keys
WHERE
        tenant_id = $1
        AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY
        created_at DESC,
        id DESC;

-- name: ListExpiredAPIKeys :many
SELECT
        *
FROM
        api_keys
WHERE
        expires_at IS NOT NULL
        AND expires_at <= NOW()
ORDER BY
        expires_at DESC;

-- name: UpdateAPIKeyName :one
UPDATE
        api_keys
SET
        name = $1
WHERE
        id = $2
RETURNING
        *;

-- name: UpdateAPIKeyScopes :one
UPDATE
        api_keys
SET
        scopes = $1
WHERE
        id = $2
RETURNING
        *;

-- name: UpdateLastUsedAtIfStale :exec
UPDATE
        api_keys
SET
        last_used_at = NOW()
WHERE
        id = $1
        AND (last_used_at IS NULL OR last_used_at < NOW() - INTERVAL '5 minutes');

-- name: SetAPIKeyExpiration :one
UPDATE
        api_keys
SET
        expires_at = $1
WHERE
        id = $2
RETURNING
        *;

-- name: DeleteAPIKey :exec
DELETE FROM
        api_keys
WHERE
        id = $1;

-- name: DeleteAPIKeysByTenantID :exec
DELETE FROM
        api_keys
WHERE
        tenant_id = $1;

-- name: CountAPIKeysByTenantID :one
SELECT
        COUNT(*)
FROM
        api_keys
WHERE
        tenant_id = $1;

-- name: CountActiveAPIKeysByTenantID :one
SELECT
        COUNT(*)
FROM
        api_keys
WHERE
        tenant_id = $1
        AND (expires_at IS NULL OR expires_at > NOW());

-- name: ListStaleAPIKeys :many
SELECT
        *
FROM
        api_keys
WHERE
        (last_used_at IS NULL OR last_used_at < NOW() - INTERVAL '30 days')
        AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY
        last_used_at ASC NULLS FIRST;
