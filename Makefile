CDK            := bun --env-file=.env cdk
CDK_LOCAL      := bun --env-file=.env.local cdklocal

LAMBDAS := server/lambdas/api server/lambdas/file-validator

.PHONY: help install build deploy diff synth destroy \
        local local-bootstrap local-deploy local-down local-logs \
        dev seed create-key clean

help:
	@echo "Available commands:"
	@echo "  make install          Install all dependencies"
	@echo "  make build            Build all Lambda binaries"
	@echo "  make deploy           Build + deploy to AWS"
	@echo "  make diff             Build + show CDK diff"
	@echo "  make synth            Build + synthesize CloudFormation"
	@echo "  make destroy          Destroy AWS stacks"
	@echo ""
	@echo "  make local            Start LocalStack"
	@echo "  make local-bootstrap  Bootstrap CDK in LocalStack (once)"
	@echo "  make local-deploy     Build + deploy to LocalStack"
	@echo "  make local-down       Stop LocalStack"
	@echo "  make local-logs       Tail LocalStack logs"
	@echo ""
	@echo "  make dev              Run API locally against LocalStack"
	@echo "  make seed             Seed database"
	@echo "  make create-key       Create API key (OWNER=name)"
	@echo "  make clean            Clean all build artifacts"

install:
	@echo "Installing dependencies..."
	@for dir in $(LAMBDAS); do cd $$dir && go mod download && cd -; done
	cd server/lambdas && go work sync
	cd server/infra && bun install
	@echo "Dependencies installed"

build:
	@echo "Building Lambda binaries..."
	@for dir in $(LAMBDAS); do $(MAKE) -C $$dir build; done
	@echo "Build complete"

# --- AWS ---

deploy: build
	@echo "Deploying to AWS..."
	cd server/infra && $(CDK) deploy --all --require-approval never
	@echo "Deployed"

diff: build
	cd server/infra && $(CDK) diff

synth: build
	cd server/infra && $(CDK) synth

destroy:
	cd server/infra && $(CDK) destroy

# --- Local ---

local:
	@echo "Starting LocalStack..."
	cd server/lambdas/api && docker-compose up -d localstack
	@echo "Waiting for LocalStack..."
	@counter=0; \
	until curl -s http://localhost:4566/_localstack/health | grep -q '"cloudformation": "available"' && \
	      curl -s http://localhost:4566/_localstack/health | grep -q '"s3": "available"' && \
	      curl -s http://localhost:4566/_localstack/health | grep -q '"dynamodb": "available"'; do \
		counter=$$((counter+1)); \
		if [ $$counter -gt 40 ]; then \
			echo "LocalStack failed to start. Check logs with: make local-logs"; \
			exit 1; \
		fi; \
		printf "."; \
		sleep 3; \
	done
	@echo ""
	@echo "LocalStack is ready"

local-bootstrap:
	cd server/infra && $(CDK_LOCAL) bootstrap

local-deploy: build
	@echo "Deploying to LocalStack..."
	cd server/infra && $(CDK_LOCAL) deploy --require-approval never
	cd server/infra && bun --env-file=.env.local run scripts/seed-db.ts
	@echo "LocalStack deployed"

local-down:
	cd server/lambdas/api && docker-compose down -v

local-logs:
	cd server/lambdas/api && docker-compose logs -f localstack

dev:
	cd server/lambdas/api && $(MAKE) dev

seed:
	cd server/infra && bun --env-file=.env run scripts/seed-db.ts

create-key:
	@if [ -z "$(OWNER)" ]; then echo "Usage: make create-key OWNER=<name>"; exit 1; fi
	cd server/infra && bun --env-file=.env run scripts/create-key.ts $(OWNER)

clean:
	@for dir in $(LAMBDAS); do $(MAKE) -C $$dir clean; done
	cd server/infra && rm -rf cdk.out
	@echo "Cleaned"
