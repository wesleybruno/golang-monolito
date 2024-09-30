.PHONY: go, git

APP_NAME=golang-monolito

run: 
	@go run cmd/api/*.go

br: 
	@air -c .air.toml

push: 
	@git push