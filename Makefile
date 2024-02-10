FROM ?= "mongodb://root:root@localhost:27017/?sslmode=disable&authSource=admin"
TO ?= "mongodb://root:root@localhost:27018/?sslmode=disable&authSource=admin"

up:
	@docker compose up -d

down:
	@docker compose down

run-cli:
	@go run main.go --from $(FROM) --to $(TO)

lint:
	@golangci-lint run --config .golangci.yml ./...

pre-release:
	@goreleaser --snapshot --skip=publish --clean

.PHONY: up down run-cli lint pre-release
