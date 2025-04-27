.PHONY: help dev build preview clean

help:
	@echo "Usage:"
	@echo "  make dev"
	@echo "  make build"
	@echo "  make preview"
	@echo "  make clean"

gen/db:
	@echo "Generating database queries using sqlc..."
	@sqlc generate

dev: gen/db
	@echo "Starting server in development mode..."
	@air --build.bin "go run main.go"

build: gen/db
	@echo "Building server..."
	@go build -o ./build/server

preview: build
	@echo "Starting server in production mode..."
	@./build/server

clean:
	@echo "Cleaning up..."
	@rm -rf ./build
	@rm -rf ./db/sqlc
	@rm -rf temperature.db
