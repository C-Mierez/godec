CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE OR REPLACE FUNCTION public.update_updated_at_column()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
	NEW.updated_at = NOW();
	RETURN NEW;
END;
$$;

CREATE TABLE public.tenants (
	id uuid DEFAULT uuidv7() NOT NULL,
	name text NOT NULL,
	email text NOT NULL,
	status text DEFAULT 'active'::text NOT NULL,
	created_at timestamp with time zone DEFAULT now() NOT NULL,
	updated_at timestamp with time zone DEFAULT now() NOT NULL,
	CONSTRAINT tenants_status_check CHECK ((status = ANY (ARRAY['active'::text, 'inactive'::text])))
);

CREATE TABLE public.api_keys (
	id uuid DEFAULT uuidv7() NOT NULL,
	tenant_id uuid NOT NULL,
	name text NOT NULL,
	hashed_key text NOT NULL,
	scopes text[] DEFAULT '{}'::text[] NOT NULL,
	created_at timestamp with time zone DEFAULT now() NOT NULL,
	updated_at timestamp with time zone DEFAULT now() NOT NULL,
	last_used_at timestamp with time zone,
	expires_at timestamp with time zone
);

ALTER TABLE ONLY public.api_keys
	ADD CONSTRAINT api_keys_hashed_key_key UNIQUE (hashed_key);

ALTER TABLE ONLY public.api_keys
	ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.tenants
	ADD CONSTRAINT tenants_pkey PRIMARY KEY (id);

CREATE INDEX idx_api_keys_hashed ON public.api_keys USING btree (hashed_key);

CREATE INDEX idx_api_keys_tenant_id ON public.api_keys USING btree (tenant_id);

CREATE INDEX idx_tenants_active ON public.tenants USING btree (id) WHERE (status = 'active'::text);

CREATE UNIQUE INDEX idx_tenants_unique_email ON public.tenants USING btree (email) WHERE (status = 'active'::text);

CREATE TRIGGER set_api_keys_updated_at BEFORE UPDATE ON public.api_keys FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

CREATE TRIGGER set_tenants_updated_at BEFORE UPDATE ON public.tenants FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();

ALTER TABLE ONLY public.api_keys
	ADD CONSTRAINT api_keys_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;
