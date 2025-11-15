package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/ieraasyl/grpcstore/storeproto"

	"google.golang.org/grpc"
)

// server struct embeds the generated unimplemented server
type server struct {
	pb.UnimplementedECommerceStoreServer

	// In a real app, you'd have a DB connection.
	// For now, we'll use in-memory mock data.
	mockProducts map[string]*pb.Product
	mockCarts    map[string]*pb.Cart
}

// newServer creates a server instance with mock data
func newServer() *server {
	return &server{
		mockProducts: map[string]*pb.Product{
			"1": {Id: "1", Name: "Laptop", Description: "A powerful laptop", Price: 1200.50},
			"2": {Id: "2", Name: "Mouse", Description: "A wireless mouse", Price: 45.99},
			"3": {Id: "3", Name: "Keyboard", Description: "Mechanical keyboard", Price: 150.00},
			"4": {Id: "4", Name: "Laptop Stand", Description: "Ergonomic stand", Price: 89.99},
		},
		mockCarts: make(map[string]*pb.Cart),
	}
}

// Implement the 'SearchProducts' streaming RPC
func (s *server) SearchProducts(req *pb.SearchRequest, stream pb.ECommerceStore_SearchProductsServer) error {
	log.Printf("Received search request with query: %s", req.GetQuery())

	// Simulate streaming results
	for _, product := range s.mockProducts {
		// In a real search, you'd filter by req.GetQuery()
		// Here we just stream all products for the demo.
		log.Printf("Streaming product: %s", product.GetName())
		if err := stream.Send(product); err != nil {
			return err
		}
		// Simulate network delay for a clear streaming effect
		time.Sleep(500 * time.Millisecond)
	}

	log.Println("Finished streaming products.")
	return nil
}

// Implement the 'AddToCart' unary RPC
func (s *server) AddToCart(ctx context.Context, req *pb.AddItemRequest) (*pb.Cart, error) {
	log.Printf("Received add to cart request for user %s, product %s", req.GetUserId(), req.GetProductId())

	// Find the product
	product, exists := s.mockProducts[req.GetProductId()]
	if !exists {
		return nil, fmt.Errorf("product with id %s not found", req.GetProductId())
	}

	// Get or create cart for the user
	cart, ok := s.mockCarts[req.GetUserId()]
	if !ok {
		cart = &pb.Cart{UserId: req.GetUserId(), Items: []*pb.CartItem{}}
	}

	// Add item to cart (simple logic, doesn't check for existing item)
	cart.Items = append(cart.Items, &pb.CartItem{
		Product:  product,
		Quantity: req.GetQuantity(),
	})

	// Recalculate total price
	cart.TotalPrice = 0
	for _, item := range cart.Items {
		cart.TotalPrice += item.GetProduct().GetPrice() * float64(item.GetQuantity())
	}

	// Save and return the cart
	s.mockCarts[req.GetUserId()] = cart
	log.Printf("Updated cart for user %s: %d items, Total: %.2f", cart.GetUserId(), len(cart.GetItems()), cart.GetTotalPrice())
	return cart, nil
}

// Implement the 'GetOrderStatus' unary RPC
func (s *server) GetOrderStatus(ctx context.Context, req *pb.OrderRequest) (*pb.OrderStatus, error) {
	log.Printf("Received get order status request for order: %s", req.GetOrderId())

	// Return mock data with current timestamp
	return &pb.OrderStatus{
		OrderId:   req.GetOrderId(),
		Status:    "IN_PROGRESS",
		UpdatedAt: time.Now().Unix(),
	}, nil
}

func main() {
	// Set up the listener on port 6767
	port := ":6767"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create a new gRPC server instance
	grpcServer := grpc.NewServer()

	// Create our server implementation
	s := newServer()

	// Register our service implementation with the gRPC server
	pb.RegisterECommerceStoreServer(grpcServer, s)

	log.Printf("Server listening at %v", lis.Addr())

	// Start serving requests. This is a blocking call.
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
