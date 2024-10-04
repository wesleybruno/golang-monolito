include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations
DB_ADDR = "postgres://$(DB_USER):$(DB_PASSWORD)@localhost/$(DB_NAME)?sslmode=disable"

.PHONY: go
run: 
	@go run cmd/api/*.go

.PHONY: air
air: 
	@air -c .air.toml

.PHONY: git
push: 
	@git push

.PHONY: migrate-create
migration:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down


