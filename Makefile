# Variables
BINARY_NAME=porticos
MAIN_PATH=./cmd/api

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

.PHONY: help start start-dev install-air

help: ## Show available commands
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}'

install-air: ## Install Air for hot reload
	@echo "$(YELLOW)Installing Air...$(NC)"
	@go install github.com/air-verse/air@latest
	@echo "$(GREEN)Air installed successfully!$(NC)"

start: ## Start the application (normal mode)
	@echo "$(YELLOW)Starting application...$(NC)"
	@go run $(MAIN_PATH)

start-dev: ## Start the application with hot reload
	@echo "$(YELLOW)Starting application with hot reload...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		air -c .air.toml; \
	else \
		echo "$(YELLOW)Air not found. Installing...$(NC)"; \
		make install-air; \
		$(shell go env GOPATH)/bin/air -c .air.toml; \
	fi

docs-generate: ## Generate OpenAPI documentation from code annotations
	@echo "$(YELLOW)Generating API documentation from code...$(NC)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o docs/; \
		echo "$(GREEN)✅ Documentation generated at docs/$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest$(NC)"; \
		exit 1; \
	fi

docs-serve: ## Serve API documentation with ReDoc  
	@echo "$(YELLOW)Starting application with ReDoc documentation...$(NC)"
	@echo "$(GREEN)Documentation available at: http://localhost:3200/docs$(NC)"
	@make start-dev

docs-validate: ## Validate OpenAPI specification
	@echo "$(YELLOW)Validating OpenAPI specification...$(NC)"
	@if [ -f docs/openapi.yaml ]; then \
		echo "$(GREEN)OpenAPI spec found at docs/openapi.yaml$(NC)"; \
	else \
		echo "$(RED)OpenAPI spec not found. Create docs/openapi.yaml$(NC)"; \
	fi


# Testing commands (solo modules)
test: ## Run all module tests
	@echo "$(YELLOW)Running module tests...$(NC)"
	@go test ./internal/modules/... -v

test-coverage: ## Run module tests with coverage
	@echo "$(YELLOW)Running module tests with coverage...$(NC)"
	@mkdir -p coverage
	@go test -coverprofile=coverage/coverage.out ./internal/modules/...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@go tool cover -func=coverage/coverage.out
	@echo "$(GREEN)Coverage report: coverage/coverage.html$(NC)"

test-watch: ## Run module tests in watch mode
	@echo "$(YELLOW)Running module tests in watch mode...$(NC)"
	@find ./internal/modules -name "*.go" | entr -c go test ./internal/modules/... -v

test-unit: ## Run only unit tests (domain layer)
	@echo "$(YELLOW)Running unit tests...$(NC)"
	@go test ./internal/modules/*/domain/... -v

test-integration: ## Run integration tests (infrastructure layer)
	@echo "$(YELLOW)Running integration tests...$(NC)"
	@go test ./internal/modules/*/infrastructure/... -v

test-business: ## Run business logic tests (application layer)
	@echo "$(YELLOW)Running business logic tests...$(NC)"
	@go test ./internal/modules/*/application/... -v


# Security commands
secrets: ## Detect secrets in code using gitleaks
	@echo "$(YELLOW)🔍 Scanning for secrets in code...$(NC)"
	@if command -v gitleaks >/dev/null 2>&1; then \
		gitleaks detect --source . --verbose --no-git; \
		echo "$(GREEN) Secret scan completed$(NC)"; \
	else \
		echo "$(YELLOW)  gitleaks not found. Installing...$(NC)"; \
		echo "$(YELLOW)Run: brew install gitleaks$(NC)"; \
		exit 1; \
	fi

secrets-baseline: ## Create gitleaks baseline (ignore existing secrets)
	@echo "$(YELLOW)Creating gitleaks baseline...$(NC)"
	@if command -v gitleaks >/dev/null 2>&1; then \
		gitleaks detect --source . --baseline-path .gitleaksignore; \
		echo "$(GREEN) Baseline created: .gitleaksignore$(NC)"; \
	else \
		echo "$(YELLOW)  gitleaks not found. Run: brew install gitleaks$(NC)"; \
		exit 1; \
	fi

security-scan: secrets ## Run complete security scan
	@echo "$(GREEN) Security scan completed$(NC)"

deps-check: ## Check and clean dependencies
	@echo "$(YELLOW)Checking Go dependencies...$(NC)"
	@go mod verify
	@go mod tidy
	@echo "$(GREEN) Dependencies verified and cleaned$(NC)"

audit: ## Audit Go dependencies for vulnerabilities
	@echo "$(YELLOW)🔍 Auditing Go dependencies...$(NC)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
		echo "$(GREEN)✅ Dependency audit completed$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  govulncheck not found. Installing...$(NC)"; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck ./...; \
	fi

audit-ci: ## Audit for CI/CD (strict mode)
	@echo "$(YELLOW)🔍 Auditing dependencies (CI mode)...$(NC)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck -json ./... > audit-report.json; \
		govulncheck ./...; \
	else \
		echo "$(YELLOW)Installing govulncheck...$(NC)"; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		govulncheck -json ./... > audit-report.json; \
		govulncheck ./...; \
	fi