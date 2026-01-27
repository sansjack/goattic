.PHONY: help install build deploy local local-bootstrap local-deploy local-down local-logs dev clean

help:
	@echo "Available commands:"
	@echo "  make install         - Install all dependencies (Go, Bun)"
	@echo "  make build           - Build Lambda binary for deployment"
	@echo "  make deploy          - Deploy to AWS"
	@echo "  make local           - Start LocalStack (run once)"
	@echo "  make local-bootstrap - Bootstrap CDK in LocalStack (run once)"
	@echo "  make local-deploy    - Deploy CDK stack to LocalStack"
	@echo "  make local-down      - Stop LocalStack and clean up"
	@echo "  make local-logs      - View LocalStack logs"
	@echo "  make dev             - Run API locally against LocalStack"
	@echo "  make clean           - Clean build artifacts"

install:
	@echo "Installing dependencies..."
	cd server/lambda && go mod download
	cd server/infra && bun install
	@command -v cdklocal >/dev/null 2>&1 || { echo "Installing cdklocal..."; bun install -g aws-cdk-local aws-cdk; }
	@echo "✓ Dependencies installed"

build:
	@echo "Building Lambda binary..."
	cd server/lambda && $(MAKE) build
	@echo "✓ Lambda binary built"

deploy:
	@echo "Deploying to AWS..."
	cd server/infra && bun deploy
	@echo "✓ Deployed to AWS"

local:
	@echo "Starting LocalStack..."
	cd server/lambda && docker-compose up -d localstack
	@echo "Waiting for LocalStack to be ready (this may take 30-60 seconds)..."
	@counter=0; \
	until curl -s http://localhost:4566/_localstack/health | grep -q '"cloudformation": "available"' && \
	      curl -s http://localhost:4566/_localstack/health | grep -q '"s3": "available"' && \
	      curl -s http://localhost:4566/_localstack/health | grep -q '"dynamodb": "available"'; do \
		counter=$$((counter+1)); \
		if [ $$counter -gt 40 ]; then \
			echo "❌ LocalStack failed to start. Check logs with: make local-logs"; \
			exit 1; \
		fi; \
		printf "."; \
		sleep 3; \
	done
	@echo ""
	@echo "✓ LocalStack is ready"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make local-bootstrap' (first time only)"
	@echo "  2. Run 'make local-deploy' to deploy your stack"
	@echo "  3. Run 'make dev' to start the local API server"

local-bootstrap:
	@echo "Bootstrapping CDK in LocalStack..."
	cd server/infra && bun local:bootstrap
	@echo "✓ Bootstrap complete"

local-deploy:
	@echo "Deploying CDK stack to LocalStack..."
	cd server/infra && bun local:deploy
	@echo ""
	@echo "✓ LocalStack deployed!"
	@echo "  - API Gateway: http://localhost:4566/restapis/"
	@echo "  - DynamoDB: http://localhost:4566"
	@echo "  - S3: http://localhost:4566"
	@echo ""
	@echo "Run 'make dev' to start the local API server"

local-down:
	@echo "Stopping LocalStack..."
	cd server/lambda && docker-compose down -v
	@echo "✓ LocalStack stopped"

local-logs:
	cd server/lambda && docker-compose logs -f localstack

dev:
	@echo "Starting local API server..."
	cd server/lambda && $(MAKE) dev

clean:
	@echo "Cleaning build artifacts..."
	cd server/lambda && $(MAKE) clean
	cd server/infra && rm -rf cdk.out
	@echo "✓ Cleaned"
