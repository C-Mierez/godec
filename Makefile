# Makefile

-include .env

TOOLS_DIR := tools
GO_TOOL := go -C $(TOOLS_DIR) tool

SQLC := $(GO_TOOL) sqlc
GOOSE := $(GO_TOOL) goose
LEFTHOOK := $(GO_TOOL) lefthook

SQLC_CONFIG := $(abspath sqlc.yaml)
LEFTHOOK_CONFIG := $(abspath lefthook.yml)
MIGRATIONS_DIR_ABS := $(abspath internal/postgres/migrations)
PG_DUMP := pg_dump

MIGRATIONS_DIR := internal/postgres/migrations
SCHEMA_FILE := internal/postgres/db/schema.sql


.PHONY: help api-codegen codegen-check sqlc sqlc-check dump goose-new goose-up goose-down goose-status migrate hooks-install

help:
	@echo "Usage: make <target> [VAR=value]"
	@echo
	@echo "Common targets:"
	@echo "  hooks-install   - install git hooks with lefthook"
	@echo "  api-codegen     - run API code generation"
	@echo "  codegen-check   - run API code generation and verify it is up to date"
	@echo "  sqlc            - run sqlc generate"
	@echo "  sqlc-check      - run sqlc generate and verify it is up to date"
	@echo "  goose-new       - create new migration: NAME=desc"
	@echo "  goose-up        - run goose up"
	@echo "  goose-down      - run goose down (one migration)"
	@echo "  goose-status    - show migration status"
	@echo "  migrate         - run migrations then dump schema to $(SCHEMA_FILE)"
	@echo "  dump     - pg_dump --schema-only to $(SCHEMA_FILE)"



dump:
	@echo "Dumping schema to $(SCHEMA_FILE)"
	@mkdir -p $(dir $(SCHEMA_FILE))
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	$(PG_DUMP) --schema-only --schema=public --exclude-table=goose_db --no-owner --no-privileges --no-comments --dbname="$(DATABASE_URL)" \
		| grep -v '^\\\\' \
		> "$(SCHEMA_FILE)"

api-codegen:
	@echo "Running API code generation..."
	go generate ./internal/api

codegen-check: api-codegen
	@git diff --exit-code -- internal/api/gen.go


hooks-install:
	@echo "Installing git hooks with lefthook..."
	LEFTHOOK_CONFIG=$(LEFTHOOK_CONFIG) $(LEFTHOOK) install

sqlc: dump
	@echo "Running sqlc generate..."
	$(SQLC) generate -f $(SQLC_CONFIG)

sqlc-check:
	@echo "Running sqlc drift check..."
	$(GO_TOOL) sqlc generate -f $(SQLC_CONFIG)
	@git diff --exit-code -- internal/postgres/db/api_keys.sql.go internal/postgres/db/db.go internal/postgres/db/models.go internal/postgres/db/tenants.sql.go

goose-new:
	@if [ -z "$(NAME)" ]; then echo "Usage: make goose-new NAME=descriptive_name"; exit 1; fi
	GOOSE_MIGRATION_DIR=$(MIGRATIONS_DIR_ABS) $(GOOSE) create $(NAME) sql

goose-up:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DATABASE_URL)" GOOSE_MIGRATION_DIR=$(MIGRATIONS_DIR_ABS) $(GOOSE) up

goose-down:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DATABASE_URL)" GOOSE_MIGRATION_DIR=$(MIGRATIONS_DIR_ABS) $(GOOSE) down

goose-status:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DATABASE_URL)" GOOSE_MIGRATION_DIR=$(MIGRATIONS_DIR_ABS) $(GOOSE) status

migrate: goose-up sqlc
	@echo "Migrations applied and schema dumped to $(SCHEMA_FILE)"
