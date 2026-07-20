-- +goose Up
-- Add new columns for split-token authentication
ALTER TABLE api_keys
ADD COLUMN token_id TEXT NOT NULL DEFAULT '';

ALTER TABLE api_keys
ADD COLUMN hashed_secret TEXT NOT NULL DEFAULT '';

-- Backfill existing rows with unique token_ids before adding constraint
UPDATE api_keys SET token_id = 'gdk_' || replace(uuidv7()::text, '-', '') WHERE token_id = '';

-- Add unique constraint and index on token_id
ALTER TABLE api_keys ADD CONSTRAINT api_keys_token_id_key UNIQUE (token_id);

CREATE INDEX idx_api_keys_token_id ON api_keys (token_id);

-- Drop old hashed_key constraint and index
ALTER TABLE api_keys
DROP CONSTRAINT api_keys_hashed_key_key;

DROP INDEX IF EXISTS idx_api_keys_hashed;

-- Drop the old hashed_key column
ALTER TABLE api_keys
DROP COLUMN hashed_key;

