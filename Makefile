# Project Makefile
.PHONY: help build clean gen
.DEFAULT_GOAL := help

# Directories
WEBSOCKET_DIR := ./websocket
ALERT_LAMBDA_DIR := ./alert-lambda
STATS_LAMBDA_DIR := ./stats-lambda
WEBSITE_DIR := ./website
DUMMIES_DIR := ./dummies/sensor

# Define targets for generated artifacts
WEBSOCKET_GEN := $(WEBSOCKET_DIR)/db/sqlc
ALERT_LAMBDA_GEN := $(ALERT_LAMBDA_DIR)/.venv
STATS_LAMBDA_GEN := $(STATS_LAMBDA_DIR)/.venv
WEBSITE_GEN := $(WEBSITE_DIR)/node_modules
WEBSITE_BUILD := $(WEBSITE_DIR)/dist

help:
	@echo "Usage:"
	@echo "  make gen        - Generate all required files"
	@echo "  make build      - Build all services"
	@echo "  make preview    - Run all services in production mode"
	@echo "  make clean      - Clean up generated files"
	@echo ""
	@echo "Individual commands:"
	@echo "  make gen/websocket        - Generate websocket files"
	@echo "  make gen/alert-lambda     - Setup alert lambda environment"
	@echo "  make gen/stats-lambda     - Setup stats lambda environment"
	@echo "  make gen/website          - Setup website dependencies"
	@echo "  make gen/dummies          - Setup dummy data generator"
	@echo ""
	@echo "  make dev/websocket        - Run websocket in dev mode"
	@echo "  make dev/alert-lambda     - Run alert lambda in dev mode"
	@echo "  make dev/stats-lambda     - Run stats lambda in dev mode"
	@echo "  make dev/website          - Run website in dev mode"
	@echo ""
	@echo "  make build/websocket      - Build websocket service"
	@echo "  make build/website        - Build website"
	@echo ""
	@echo "  make preview/websocket    - Run websocket in production mode"
	@echo "  make preview/alert-lambda - Run alert lambda in production mode"
	@echo "  make preview/stats-lambda - Run stats lambda in production mode"
	@echo "  make preview/website      - Run website in production mode"

# Generation targets
gen/websocket:
	@echo "Generating database queries using sqlc..."
	@cd $(WEBSOCKET_DIR) && \
		sqlc generate && \
		go mod download
	@mkdir -p $(WEBSOCKET_GEN)
	@touch $(WEBSOCKET_GEN)

gen/alert-lambda:
	@echo "Setting up alert lambda environment..."
	@cd $(ALERT_LAMBDA_DIR) && \
		uv venv .venv && \
		. .venv/bin/activate && \
		uv sync

gen/stats-lambda:
	@echo "Setting up stats lambda environment..."
	@cd $(STATS_LAMBDA_DIR) && \
		uv venv .venv && \
		. .venv/bin/activate && \
		uv sync

gen/dummies:
	@echo "Setting up dummy data generator..."
	@cd $(DUMMIES_DIR) && \
		go mod download

gen/website:
	@echo "Setting up website dependencies..."
	@cd $(WEBSITE_DIR) && \
		bun install

gen: gen/websocket gen/alert-lambda gen/stats-lambda gen/website gen/dummies
	@echo "All required files generated successfully."

# Development targets
dev/websocket: gen/websocket
	@echo "Starting websocket in development mode..."
	@cd $(WEBSOCKET_DIR) && air --build.bin "go run main.go"

dev/alert-lambda: gen/alert-lambda
	@echo "Starting alert lambda in development mode..."
	@cd $(ALERT_LAMBDA_DIR) && \
		. .venv/bin/activate && \
		fastapi run main.py --reload --port 8001

dev/stats-lambda: gen/stats-lambda
	@echo "Starting stats lambda in development mode..."
	@cd $(STATS_LAMBDA_DIR) && \
		. .venv/bin/activate && \
		fastapi run main.py --reload --port 8000

dev/website: gen/website
	@echo "Starting website in development mode..."
	@cd $(WEBSITE_DIR) && bun dev

# Build targets
build/websocket: gen/websocket
	@echo "Building websocket service..."
	@mkdir -p $(WEBSOCKET_DIR)/build
	@cd $(WEBSOCKET_DIR) && go build -o ./build/server

build/website: gen/website
	@echo "Building website..."
	@cd $(WEBSITE_DIR) && bun run build

build: build/websocket build/website
	@echo "All services built successfully."

# Preview (production) targets
preview/websocket: build/websocket
	@echo "Starting websocket in production mode..."
	@cd $(WEBSOCKET_DIR) && ./build/server

preview/alert-lambda: gen/alert-lambda
	@echo "Starting alert lambda in production mode..."
	@cd $(ALERT_LAMBDA_DIR) && \
		. .venv/bin/activate && \
		fastapi run main.py

preview/stats-lambda: gen/stats-lambda
	@echo "Starting stats lambda in production mode..."
	@cd $(STATS_LAMBDA_DIR) && \
		. .venv/bin/activate && \
		fastapi run main.py

preview/website: build/website
	@echo "Starting website in production mode..."
	@cd $(WEBSITE_DIR) && bun run preview

# Clean up
clean:
	@echo "Cleaning up..."
	@rm -rf $(WEBSOCKET_DIR)/build
	@rm -rf $(WEBSOCKET_DIR)/db/sqlc
	@rm -rf $(ALERT_LAMBDA_DIR)/.venv
	@rm -rf $(ALERT_LAMBDA_DIR)/alerts.db
	@rm -rf $(STATS_LAMBDA_DIR)/.venv
	@rm -rf ./temperature.db
	@rm -rf $(WEBSITE_DIR)/node_modules
	@rm -rf $(WEBSITE_DIR)/dist
	@echo "Clean up complete."
