# grpcstore

Minimal Go gRPC e-commerce store service with TLS-secured server and client, backed by PostgreSQL, and containerized with Docker.

## Features

- **Protocol Buffers** API definition
- **TLS-secured** server and client implementation
- **PostgreSQL** persistence with automatic schema migration
- **Standard Go Project Layout** structure
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

This starts the **PostgreSQL** database, **gRPC Server**, and runs the **Client** E2E tests.

## Local Development

### Initial Setup

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate TLS certificates
go run ./cmd/gen-cert.go
```

### Generate Protobuf Code

Run after modifying `api/proto/v1/store.proto`:

```bash
protoc --proto_path=api/proto/v1 --go_out=pkg/api/v1 --go_opt=paths=source_relative \
       --go-grpc_out=pkg/api/v1 --go-grpc_opt=paths=source_relative \
       api/proto/v1/store.proto
```

### Run Server

You need a running PostgreSQL instance. The easiest way is to use Docker for the DB:

```bash
# Start Postgres
docker compose up -d postgres

# Run Server (defaults to localhost DB connection)
go run ./cmd/server/main.go
```

Server listens on port `6767` with TLS.

### Run Client

```bash
# Linux/macOS
SERVER_ADDR=localhost:6767 go run ./cmd/client/main.go

# Windows PowerShell
$env:SERVER_ADDR="localhost:6767"; go run ./cmd/client/main.go
```

## Project Structure

```
.
├── api/
│   └── proto/v1/       # Protocol Buffer definitions
├── cmd/
│   ├── client/         # gRPC client entry point
│   ├── server/         # gRPC server entry point
│   └── gen-cert.go     # TLS certificate generator
├── internal/
│   ├── db/             # Database connection & migration
│   └── service/        # Business logic & gRPC implementation
├── pkg/
│   └── api/v1/         # Generated Go code (importable)
├── tls/                # TLS certificates
├── Dockerfile.client   # Client container image
├── Dockerfile.server   # Server container image
├── docker-compose.yml  # Orchestration configuration
└── go.mod              # Go module dependencies
```

## License

MIT