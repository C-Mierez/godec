-- name: CreateApiKey :one
INSERT INTO
        api_keys (tenant_id, name, hashed_key, scopes)
VALUES
        ($1, $2, $3, $4)
RETURNING
        *;

-- name: GetApiKeyByID :one
SELECT
        *
FROM
        api_keys
WHERE
        id = $1;

-- name: GetApiKeyByHashedKey :one
SELECT
        *
FROM
        api_keys
WHERE
        hashed_key = $1;

-- name: ListApiKeysByTenantID :many
SELECT
        *
FROM
        api_keys
WHERE
        tenant_id = $1
ORDER BY
        created_at DESC,
        id DESC;

-- name: ListActiveApiKeysByTenantID :many
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

-- name: ListExpiredApiKeys :many
SELECT
        *
FROM
        api_keys
WHERE
        expires_at IS NOT NULL
        AND expires_at <= NOW()
ORDER BY
        expires_at DESC;

-- name: UpdateApiKeyName :one
UPDATE
        api_keys
SET
        name = $1
WHERE
        id = $2
RETURNING
        *;

-- name: UpdateApiKeyScopes :one
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

-- name: SetApiKeyExpiration :one
UPDATE
        api_keys
SET
        expires_at = $1
WHERE
        id = $2
RETURNING
        *;

-- name: DeleteApiKey :exec
DELETE FROM
        api_keys
WHERE
        id = $1;

-- name: DeleteApiKeysByTenantID :exec
DELETE FROM
        api_keys
WHERE
        tenant_id = $1;

-- name: CountApiKeysByTenantID :one
SELECT
        COUNT(*)
FROM
        api_keys
WHERE
        tenant_id = $1;

-- name: CountActiveApiKeysByTenantID :one
SELECT
        COUNT(*)
FROM
        api_keys
WHERE
        tenant_id = $1
        AND (expires_at IS NULL OR expires_at > NOW());

-- name: ListStaleApiKeys :many
SELECT
        *
FROM
        api_keys
WHERE
        (last_used_at IS NULL OR last_used_at < NOW() - INTERVAL '30 days')
        AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY
        last_used_at ASC NULLS FIRST;
