package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	pb "github.com/ieraasyl/grpcstore/pkg/api/v1"
)

// StoreService implements the ECommerceStoreServer interface
type StoreService struct {
	pb.UnimplementedECommerceStoreServer
	db *pgxpool.Pool
}

// NewStoreService creates a new service instance
func NewStoreService(db *pgxpool.Pool) *StoreService {
	return &StoreService{
		db: db,
	}
}

// SearchProducts streams products matching the query
func (s *StoreService) SearchProducts(req *pb.SearchRequest, stream pb.ECommerceStore_SearchProductsServer) error {
	log.Printf("Received search request with query: %s", req.GetQuery())

	// Simple ILIKE search
	rows, err := s.db.Query(context.Background(), "SELECT id, name, description, price FROM products WHERE name ILIKE $1", "%"+req.GetQuery()+"%")
	if err != nil {
		return fmt.Errorf("failed to query products: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p pb.Product
		if err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.Price); err != nil {
			return fmt.Errorf("failed to scan product: %v", err)
		}
		log.Printf("Streaming product: %s", p.GetName())
		if err := stream.Send(&p); err != nil {
			return err
		}
		// Simulate network delay
		time.Sleep(100 * time.Millisecond)
	}

	return rows.Err()
}

// AddToCart adds an item to the user's cart
func (s *StoreService) AddToCart(ctx context.Context, req *pb.AddItemRequest) (*pb.Cart, error) {
	log.Printf("Received add to cart request for user %s, product %s", req.GetUserId(), req.GetProductId())

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// Ensure cart exists
	_, err = tx.Exec(ctx, "INSERT INTO carts (user_id, updated_at) VALUES ($1, NOW()) ON CONFLICT (user_id) DO UPDATE SET updated_at = NOW()", req.GetUserId())
	if err != nil {
		return nil, fmt.Errorf("failed to upsert cart: %v", err)
	}

	// Upsert cart item
	_, err = tx.Exec(ctx, `
		INSERT INTO cart_items (user_id, product_id, quantity) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, product_id) 
		DO UPDATE SET quantity = cart_items.quantity + $3
	`, req.GetUserId(), req.GetProductId(), req.GetQuantity())
	if err != nil {
		return nil, fmt.Errorf("failed to upsert cart item: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Fetch updated cart to return
	return s.getCart(ctx, req.GetUserId())
}

func (s *StoreService) getCart(ctx context.Context, userID string) (*pb.Cart, error) {
	cart := &pb.Cart{UserId: userID}

	rows, err := s.db.Query(ctx, `
		SELECT p.id, p.name, p.description, p.price, ci.quantity
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart items: %v", err)
	}
	defer rows.Close()

	var totalPrice float64
	for rows.Next() {
		var p pb.Product
		var qty int32
		if err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.Price, &qty); err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %v", err)
		}
		cart.Items = append(cart.Items, &pb.CartItem{
			Product:  &p,
			Quantity: qty,
		})
		totalPrice += p.Price * float64(qty)
	}
	cart.TotalPrice = totalPrice

	return cart, nil
}

// GetOrderStatus retrieves the status of an order
func (s *StoreService) GetOrderStatus(ctx context.Context, req *pb.OrderRequest) (*pb.OrderStatus, error) {
	log.Printf("Received get order status request for order: %s", req.GetOrderId())

	// For demo purposes, if order doesn't exist, create a fake one
	var status string
	var updatedAt time.Time

	err := s.db.QueryRow(ctx, "SELECT status, updated_at FROM orders WHERE order_id = $1", req.GetOrderId()).Scan(&status, &updatedAt)
	if err == pgx.ErrNoRows {
		// Create fake order
		status = "IN_PROGRESS"
		updatedAt = time.Now()
		_, err = s.db.Exec(ctx, "INSERT INTO orders (order_id, status, updated_at) VALUES ($1, $2, $3)", req.GetOrderId(), status, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to create fake order: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to query order: %v", err)
	}

	return &pb.OrderStatus{
		OrderId:   req.GetOrderId(),
		Status:    status,
		UpdatedAt: updatedAt.Unix(),
	}, nil
}
