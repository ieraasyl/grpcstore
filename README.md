# grpcstore

Minimal Go gRPC a simple e-commerce store service with a server and a client.

**Prerequisites:**
- Go (latest stable recommended)
```bash
go version
```
- `protoc` and the `protoc-gen-go` / `protoc-gen-go-grpc` plugins to generate protobuf code
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

**Quick Start (PowerShell)**

Build binaries:

```bash
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client
```

Run directly with `go run`:

```bash
go run ./cmd/server
go run ./cmd/client
```

Generate protobufs:

```bash
protoc --go_out=./storeproto --go_opt=paths=source_relative \
       --go-grpc_out=./storeproto --go-grpc_opt=paths=source_relative \
       store.proto
```


**Project layout**
- `store.proto` - protobuf service & messages definition
- `storeproto/` - generated Go protobuf code (`store.pb.go`, `store_grpc.pb.go`)
- `cmd/server/` - server entrypoint
- `cmd/client/` - client entrypoint
- `bin/` - recommended location for built binaries