# Example Makefile showing how to use yeet run

# Traditional approach - requires .env file
dev-old:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Run 'yeet fetch' first."; \
		exit 1; \
	fi
	@source .env && npm start

# New approach - no .env file needed!
dev:
	@echo "Starting development server with secrets from Key Vault..."
	@yeet run -- go run ./test-env.go

# Run with a specific vault
dev-staging:
	@yeet run --vault staging-vault npm start

# Run with local overrides from .env
dev-local:
	@yeet run --load-env npm start

# Run tests with secrets
test:
	@yeet run -- go test ./...

# Run docker-compose with secrets
up:
	@yeet run -- docker-compose up -d

down:
	@yeet run -- docker-compose down

# Deploy using secrets
deploy:
	@yeet run --vault production-vault make _deploy

_deploy:
	@echo "Deploying with DATABASE_URL=$$DATABASE_URL"
	@echo "API_KEY is set: $${API_KEY:+yes}"
	# Your actual deployment commands here

.PHONY: dev dev-old dev-staging dev-local test up down deploy _deploy
