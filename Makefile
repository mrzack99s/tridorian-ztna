.PHONY: help build-all clean test proto setup generate-keys
.PHONY: build-management build-controlplane build-auth build-gateway-agent build-legacy
.PHONY: run-management run-controlplane run-auth run-gateway-agent
.PHONY: build-frontend build-tenant-admin build-backoffice
.PHONY: docker-build docker-push docker-build-push

# Configuration
PROJECT_ID ?= trivpn-demo-prj
REGION ?= asia-southeast1
REGISTRY ?= $(REGION)-docker.pkg.dev
REPO_NAME ?= triztna
IMAGE_TAG ?= latest

# Image names
MGMT_IMAGE = $(REGISTRY)/$(PROJECT_ID)/$(REPO_NAME)/management-api:$(IMAGE_TAG)
CP_IMAGE = $(REGISTRY)/$(PROJECT_ID)/$(REPO_NAME)/gateway-controlplane:$(IMAGE_TAG)
AUTH_IMAGE = $(REGISTRY)/$(PROJECT_ID)/$(REPO_NAME)/auth-api:$(IMAGE_TAG)
TENANT_IMAGE = $(REGISTRY)/$(PROJECT_ID)/$(REPO_NAME)/tenant-admin:$(IMAGE_TAG)
BACKOFFICE_IMAGE = $(REGISTRY)/$(PROJECT_ID)/$(REPO_NAME)/backoffice:$(IMAGE_TAG)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-25s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Setup targets
setup: ## Initial setup (generate keys + copy .env.example)
	@echo "🚀 Setting up Tridorian ZTNA..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ Created .env from .env.example"; \
	else \
		echo "⚠️  .env already exists, skipping..."; \
	fi
	@./scripts/generate-keys.sh
	@echo ""
	@echo "✅ Setup complete! Next steps:"
	@echo "   1. Review and update .env file if needed"
	@echo "   2. Run: make docker-build-push"

generate-keys: ## Generate new EdDSA key pair
	@./scripts/generate-keys.sh

# Build targets (local Go builds)
build-management: ## Build Management API (HTTP)
	@echo "🔨 Building Management API..."
	@go build -o bin/management-api cmd/management-api/main.go

build-controlplane: ## Build Gateway Control Plane (gRPC)
	@echo "🔨 Building Gateway Control Plane..."
	@go build -o bin/gateway-controlplane cmd/gateway-controlpane/main.go

build-auth: ## Build Authentication API
	@echo "🔨 Building Authentication API..."
	@go build -o bin/auth-api cmd/auth-api/main.go

build-gateway-agent: ## Build Gateway Agent
	@echo "🔨 Building Gateway Agent..."
	@go build -o bin/gateway cmd/gateway/main.go

build-all: build-management build-controlplane build-auth build-gateway-agent ## Build all backend services

# Frontend build targets
build-tenant-admin: ## Build Tenant Admin frontend
	@echo "🎨 Building Tenant Admin..."
	@cd apps/tenant-admin && npm run build

build-backoffice: ## Build Backoffice frontend
	@echo "🎨 Building Backoffice..."
	@cd apps/backoffice && npm run build

build-frontend: build-tenant-admin build-backoffice ## Build all frontend apps

# Docker build targets
docker-build-management: ## Build Management API Docker image
	@echo "� Building Management API image..."
	@docker build -t $(MGMT_IMAGE) -f Dockerfile.management-api .

docker-build-controlplane: ## Build Gateway Control Plane Docker image
	@echo "� Building Gateway Control Plane image..."
	@docker build -t $(CP_IMAGE) -f Dockerfile.gateway-controlplane .

docker-build-auth: ## Build Auth API Docker image
	@echo "� Building Auth API image..."
	@docker build -t $(AUTH_IMAGE) -f Dockerfile.auth-api .

docker-build-tenant-admin: ## Build Tenant Admin Docker image
	@echo "🐳 Building Tenant Admin image..."
	@docker build -t $(TENANT_IMAGE) -f Dockerfile.tenant-admin .

docker-build-backoffice: ## Build Backoffice Docker image
	@echo "🐳 Building Backoffice image..."
	@docker build -t $(BACKOFFICE_IMAGE) -f Dockerfile.backoffice .

docker-build: ## Build all Docker images
	@echo "🐳 Building all Docker images..."
	@$(MAKE) docker-build-management
	@$(MAKE) docker-build-controlplane
	@$(MAKE) docker-build-auth
	@$(MAKE) docker-build-tenant-admin
	@$(MAKE) docker-build-backoffice
	@echo "✅ All images built successfully!"

# Docker push targets
docker-push-management: ## Push Management API image
	@echo "📤 Pushing Management API image..."
	@docker push $(MGMT_IMAGE)

docker-push-controlplane: ## Push Gateway Control Plane image
	@echo "📤 Pushing Gateway Control Plane image..."
	@docker push $(CP_IMAGE)

docker-push-auth: ## Push Auth API image
	@echo "📤 Pushing Auth API image..."
	@docker push $(AUTH_IMAGE)

docker-push-tenant-admin: ## Push Tenant Admin image
	@echo "� Pushing Tenant Admin image..."
	@docker push $(TENANT_IMAGE)

docker-push-backoffice: ## Push Backoffice image
	@echo "📤 Pushing Backoffice image..."
	@docker push $(BACKOFFICE_IMAGE)

docker-push: ## Push all Docker images
	@echo "📤 Pushing all Docker images..."
	@$(MAKE) docker-push-management
	@$(MAKE) docker-push-controlplane
	@$(MAKE) docker-push-auth
	@$(MAKE) docker-push-tenant-admin
	@$(MAKE) docker-push-backoffice
	@echo "✅ All images pushed successfully!"

# Combined build and push
docker-build-push: docker-build docker-push ## Build and push all Docker images

# Artifact Registry setup
setup-artifact-registry: ## Create Artifact Registry repository
	@echo "🏗️  Setting up Artifact Registry..."
	@gcloud artifacts repositories create $(REPO_NAME) \
		--repository-format=docker \
		--location=$(REGION) \
		--description="Tridorian ZTNA Docker images" \
		--project=$(PROJECT_ID) || echo "Repository may already exist"
	@echo "✅ Artifact Registry ready!"
	@echo "📝 Configure Docker authentication:"
	@echo "   gcloud auth configure-docker $(REGION)-docker.pkg.dev"

# Docker authentication
docker-auth: ## Configure Docker authentication for Artifact Registry
	@echo "� Configuring Docker authentication..."
	@gcloud auth configure-docker $(REGION)-docker.pkg.dev

# Utility targets
clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	@rm -rf bin/
	@rm -rf apps/tenant-admin/dist
	@rm -rf apps/backoffice/dist

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test ./...

proto: ## Generate protobuf code
	@echo "🔨 Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto
