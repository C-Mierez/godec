-- +goose Up
-- +goose StatementBegin
CREATE TABLE tenants (
	id UUID PRIMARY KEY DEFAULT uuidv7 (),
	name TEXT NOT NULL,
	email TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW ()
);

CREATE UNIQUE INDEX idx_tenants_unique_email ON tenants (email)
WHERE
	status = 'active';

CREATE INDEX idx_tenants_active ON tenants (id)
WHERE
	status = 'active';

-- +goose StatementEnd