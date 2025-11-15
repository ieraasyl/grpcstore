package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	pb "github.com/ieraasyl/grpcstore/storeproto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// Read the server address from an environment variable
	serverAddr := os.Getenv("SERVER_ADDR")

	// Load the server's public certificate to trust it
	certFile := filepath.Join("tls", "server.crt")
	caCert, err := os.ReadFile(certFile)
	if err != nil {
		log.Fatalf("Failed to read CA certificate: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Failed to add server CA's certificate to pool")
	}

	// Use the hostname (e.g., "localhost" or "server") to verify the cert
	hostname, _, err := net.SplitHostPort(serverAddr)
	if err != nil {
		log.Fatalf("Failed to parse server address: %v", err)
	}

	// Create TLS credentials
	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		ServerName: hostname,
	}

	creds := credentials.NewTLS(tlsConfig)

	// Connect to the server on port 6767
	log.Printf("Connecting to gRPC server at %s...", serverAddr)
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	log.Printf("[OK] Successfully connected to gRPC server at %s", serverAddr)

	// Create a new client stub
	c := pb.NewECommerceStoreClient(conn)

	// Set up a context with a 15-second timeout for the whole run
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	log.Println("=== Starting E2E Test Suite ===")

	log.Println("--- Test 1: Calling SearchProducts (streaming RPC) ---")
	callSearchProducts(ctx, c)

	log.Println("--- Test 2: Calling AddToCart (unary RPC) ---")
	_ = callAddToCart(ctx, c, "user-123", "1", 2) // Add 2 Laptops for user-123
	_ = callAddToCart(ctx, c, "user-123", "2", 1) // Add 1 Mouse for user-123

	log.Println("--- Test 3: Calling GetOrderStatus (unary RPC) ---")
	callGetOrderStatus(ctx, c, "order-xyz-789")

	log.Println("=== All Tests Completed ===")
}

// callSearchProducts handles the streaming response
func callSearchProducts(ctx context.Context, c pb.ECommerceStoreClient) {
	stream, err := c.SearchProducts(ctx, &pb.SearchRequest{
		Query:     "laptop",
		PageSize:  10,
		PageToken: "",
	})
	if err != nil {
		log.Fatalf("could not search products: %v", err)
	}

	productCount := 0
	for {
		product, err := stream.Recv()
		if err == io.EOF {
			log.Printf("[OK] Finished receiving all products. Total: %d", productCount)
			break
		}
		if err != nil {
			log.Fatalf("[ERR] Error while receiving stream: %v", err)
		}
		productCount++
		log.Printf("  Received Product #%d: [ID: %s, Name: %s, Description: %s, Price: $%.2f]", productCount, product.GetId(), product.GetName(), product.GetDescription(), product.GetPrice())
	}
}

// callAddToCart calls the unary RPC and prints the response
func callAddToCart(ctx context.Context, c pb.ECommerceStoreClient, userID, productID string, quantity int32) *pb.Cart {
	log.Printf("  Adding Product %s (Quantity: %d) for User %s", productID, quantity, userID)
	cart, err := c.AddToCart(ctx, &pb.AddItemRequest{
		UserId:    userID,
		ProductId: productID,
		Quantity:  quantity,
	})
	if err != nil {
		log.Fatalf("[ERR] Could not add to cart: %v", err)
	}
	log.Printf("  [OK] Cart Updated: User: %s, Total Items: %d, Total Price: $%.2f", cart.GetUserId(), len(cart.GetItems()), cart.GetTotalPrice())
	for i, item := range cart.GetItems() {
		log.Printf("    Item %d: %s (x%d) - $%.2f each", i+1, item.GetProduct().GetName(), item.GetQuantity(), item.GetProduct().GetPrice())
	}
	return cart
}

// callGetOrderStatus calls the unary RPC and prints the response
func callGetOrderStatus(ctx context.Context, c pb.ECommerceStoreClient, orderID string) {
	log.Printf("  Checking status for Order: %s", orderID)
	status, err := c.GetOrderStatus(ctx, &pb.OrderRequest{OrderId: orderID})
	if err != nil {
		log.Fatalf("[ERR] Could not get order status: %v", err)
	}
	timestampStr := ""
	if status.GetUpdatedAt() > 0 {
		t := time.Unix(status.GetUpdatedAt(), 0)
		timestampStr = t.Format("2006-01-02 15:04:05")
	}
	log.Printf("  [OK] Order Status: [ID: %s, Status: %s, Updated: %s]", status.GetOrderId(), status.GetStatus(), timestampStr)
}
