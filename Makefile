.PHONY: help docker.up docker.down build run run.env dev dev.env db.reset tools.mockery mocks

# Help command
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_\.]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Start Docker containers
docker.up: ## Start Docker containers in detached mode
	docker-compose up -d

# Stop Docker containers
docker.down: ## Stop Docker containers
	docker-compose down

# Reset Docker database (useful for SSL errors)
db.reset: docker.down ## Reset database container (resolves SSL issues)
	docker volume rm api_postgres_data || true
	docker-compose up -d

# Build the application
build: ## Build the application
	go build -o bin/api cmd/api/main.go

# Run the application with .env file
run: ## Run the application with .env file
	set -a && source .env && set +a && go run cmd/main.go

# Development mode: Start docker and run the application
dev: docker.up run ## Start Docker containers and run the application

# Start with database reset (for SSL issues)
dev.reset: db.reset run ## Start with fresh database (resolves SSL issues)

# Start with database reset and .env (for SSL issues)
dev.reset.env: db.reset run.env ## Start with fresh database and .env file 

# Install mockery
tools.mockery: ## Install mockery binary
	./tools/install_mockery.sh

# Generate mocks
mocks: tools.mockery ## Generate mocks for interfaces
	mkdir -p tests/mocks
	./tools/bin/mockery --config=.mockery.yaml --with-expecter 