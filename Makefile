# Makefile

-include .env

SQLC := sqlc
GOOSE := goose
PG_DUMP := pg_dump

MIGRATIONS_DIR := internal/postgres/migrations
SCHEMA_FILE := internal/postgres/db/schema.sql


.PHONY: help sqlc dump goose-new goose-up goose-down goose-status migrate

help:
	@echo "Usage: make <target> [VAR=value]"
	@echo
	@echo "Common targets:"
	@echo "  sqlc            - run sqlc generate"
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
	$(PG_DUMP) --schema-only --schema=public --exclude-table=goose_db --no-owner --no-privileges --no-comments --dbname="$(DATABASE_URL)" -f "$(SCHEMA_FILE)"

sqlc: dump
	@echo "Running sqlc generate..."
	$(SQLC) generate

goose-new:
	@if [ -z "$(NAME)" ]; then echo "Usage: make goose-new NAME=descriptive_name"; exit 1; fi
	GOOSE_MIGRATION_DIR=$(MIGRATIONS_DIR) $(GOOSE) create $(NAME) sql

goose-up:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DATABASE_URL)" GOOSE_MIGRATION_DIR=$(MIGRATIONS_DIR) $(GOOSE) up

goose-down:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DATABASE_URL)" GOOSE_MIGRATION_DIR=$(MIGRATIONS_DIR) $(GOOSE) down

goose-status:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DATABASE_URL)" GOOSE_MIGRATION_DIR=$(MIGRATIONS_DIR) $(GOOSE) status

migrate: goose-up sqlc
	@echo "Migrations applied and schema dumped to $(SCHEMA_FILE)"
