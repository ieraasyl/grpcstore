package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/ieraasyl/grpcstore/internal/db"
	"github.com/ieraasyl/grpcstore/internal/service"
	pb "github.com/ieraasyl/grpcstore/pkg/api/v1"
)

func main() {
	// Set up the listener on port 6767
	port := ":6767"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Load TLS credentials
	certFile := filepath.Join("tls", "server.crt")
	keyFile := filepath.Join("tls", "server.key")

	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load TLS certificates: %v", err)
	}

	// Connect to DB
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Println("DB_URL not set, using default for local development")
		dbURL = "postgres://postgres:postgres@localhost:5432/grpcstore"
	}

	// Wait for DB to be ready (simple retry loop)
	var pool *pgxpool.Pool
	for i := 0; i < 10; i++ {
		pool, err = db.InitDB(context.Background(), dbURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to DB (attempt %d/10): %v. Retrying in 2s...", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Could not connect to database after retries: %v", err)
	}
	defer pool.Close()

	// Create a new gRPC server instance
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	// Create our server implementation
	storeService := service.NewStoreService(pool)

	// Register our service implementation with the gRPC server
	pb.RegisterECommerceStoreServer(grpcServer, storeService)

	// Channel to listen for interrupt signals and shut down gracefully
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server gracefully...")
		grpcServer.GracefulStop()
	}()

	log.Printf("Server listening at %v", lis.Addr())

	// Start serving requests. This is a blocking call.
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
