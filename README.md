# GoAttic

Go-based serverless API with AWS CDK infrastructure.

## Prerequisites

- [Go](https://golang.org/) 1.23+
- [Bun](https://bun.sh/)
- [Docker](https://www.docker.com/)
- AWS CLI configured with credentials

## Setup

1. Copy the example environment file:

```bash
cp server/infra/.env.example server/infra/.env
```

2. Edit `server/infra/.env` with your AWS profile and region

3. Install dependencies:

```bash
make install
```

## Local Development

```bash
make local              # Start LocalStack
make local-bootstrap    # Bootstrap CDK (first time only)
make local-deploy       # Deploy stack to LocalStack
make dev                # Run API server on http://localhost:8080
```

## AWS Deployment

```bash
make build              # Build Lambda binary
make deploy             # Deploy to AWS
```

## Cleanup

```bash
make local-down         # Stop LocalStack
make clean              # Clean build artifacts
```
