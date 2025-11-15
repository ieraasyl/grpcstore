# grpcstore

Minimal Go gRPC e-commerce store service with TLS-secured server and client, containerized with Docker.

## Features

- **Protocol Buffers** API definition
- **TLS-secured** server and client implementation
- **Self-contained** certificate generator
- **Multi-stage** Docker builds
- **Docker Compose** orchestration for E2E testing

## Prerequisites

- Go 1.25+
- Protocol Buffer Compiler (`protoc`)
- Docker & Docker Compose

## Quick Start (Recommended)

### 1. Generate TLS Certificates

Run once to create self-signed certificates:

```bash
go run ./cmd/gen-cert.go
```

This creates `tls/server.crt` and `tls/server.key`.

### 2. Run with Docker Compose

```bash
# Build and run secure E2E test
docker compose up --build --exit-code-from client

# Clean up
docker compose down
```

## Local Development

### Initial Setup

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate TLS certificates
go run ./cmd/gen-cert
```

### Generate Protobuf Code

Run after modifying `store.proto`:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       storeproto/store.proto
```

### Run Server

```bash
go run ./cmd/server
```

Server listens on port `6767` with TLS.

### Run Client

```bash
# Linux/macOS
SERVER_ADDR=localhost:6767 go run ./cmd/client

# Windows PowerShell
$env:SERVER_ADDR="localhost:6767"; go run ./cmd/client
```

## Project Structure

```
.
├── cmd/
│   ├── client/         # gRPC client implementation
│   ├── gen-cert/       # TLS certificate generator
│   └── server/         # gRPC server implementation
├── storeproto/         # Generated protobuf code
├── tls/                # TLS certificates
├── Dockerfile.client   # Client container image
├── Dockerfile.server   # Server container image
├── docker-compose.yml  # Orchestration configuration
├── store.proto         # API definition
└── go.mod              # Go module dependencies
```

## License

MIT