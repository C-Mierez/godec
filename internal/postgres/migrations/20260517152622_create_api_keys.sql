-- +goose Up
CREATE TABLE api_keys (
	id UUID PRIMARY KEY DEFAULT uuidv7(),
	tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
	name TEXT NOT NULL,
	hashed_key TEXT NOT NULL UNIQUE,
	scopes TEXT[] NOT NULL DEFAULT '{}',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	last_used_at TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_hashed ON api_keys (hashed_key);

CREATE INDEX idx_api_keys_tenant_id ON api_keys (tenant_id);