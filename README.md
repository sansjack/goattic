# GoAttic

Personal media storage — Go serverless API on AWS + CLI client with automatic compression.

## Prerequisites

- [Go](https://golang.org/) 1.23+
- [Bun](https://bun.sh/)
- [Docker](https://www.docker.com/)
- AWS CLI configured with credentials

## Server

### Setup

```bash
cp server/infra/.env.example server/infra/.env
# Edit .env with your AWS profile and region

make install       # Install dependencies
```

### Local Development

```bash
make local              # Start LocalStack
make local-bootstrap    # Bootstrap CDK (first time only)
make local-deploy       # Deploy stack to LocalStack
make dev                # Run API server on http://localhost:8080
```

### Deploy

```bash
make build         # Build Lambda binaries
make deploy        # Deploy to AWS
```

### Cleanup

```bash
make local-down    # Stop LocalStack
make clean         # Clean build artifacts
```

---

## Client

CLI tool to upload files to GoAttic with automatic compression.

### Build

```bash
make fetch-ffmpeg         # Download ffmpeg binary for your platform
make -C client build      # Build the goattic binary → client/build/goattic
```

Add `client/build/` to your `PATH` or copy the binary somewhere convenient.

### Configure

```bash
goattic configure
```

Prompts for your API URL and API key, saved to your platform config directory.

```bash
goattic config    # Show current config
```

### Upload

```bash
goattic <file>
```

Images (`.jpg`, `.jpeg`, `.png`) and videos (`.mov`, `.mp4`) are automatically compressed before upload. The public URL is printed on success and a desktop notification is shown.

### Supported formats

| Format | Compression |
|--------|-------------|
| `.jpg` / `.jpeg` | Re-encoded at quality 80 |
| `.png` | Best compression |
| `.mp4` / `.mov` | H.264 CRF 28, AAC 128k (requires ffmpeg) |
