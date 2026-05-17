-- +goose Up
-- +goose StatementBegin
ALTER TABLE api_keys
ADD COLUMN expires_at TIMESTAMPTZ DEFAULT NULL;

-- Trigger
CREATE OR REPLACE FUNCTION update_updated_at_column () RETURNS TRIGGER AS $$
BEGIN
	NEW.updated_at = NOW();
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_tenants_updated_at BEFORE
UPDATE ON TENANTS FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column ();

CREATE TRIGGER set_api_keys_updated_at BEFORE
UPDATE ON API_KEYS FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column ();

-- +goose StatementEnd