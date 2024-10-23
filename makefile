include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations
DB_ADDR = "postgres://$(DB_USER):$(DB_PASSWORD)@localhost/$(DB_NAME)?sslmode=disable"

.PHONY: run
run: 
	@go run cmd/api/*.go

.PHONY: air
air: 
	@air -c .air.toml

.PHONY: push
push: 
	@git push

.PHONY: migrate
migrate:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDR) down

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: docs
docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt 

.PHONY: build
build:
	@go build -o /cmd/api/*.go ./cmd/web