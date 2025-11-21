package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDB connects to the database and runs migrations
func InitDB(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	// Schema Migration
	queries := []string{
		`CREATE TABLE IF NOT EXISTS products (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			price DOUBLE PRECISION NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS carts (
			user_id TEXT PRIMARY KEY,
			updated_at TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS cart_items (
			user_id TEXT REFERENCES carts(user_id),
			product_id TEXT REFERENCES products(id),
			quantity INTEGER NOT NULL,
			PRIMARY KEY (user_id, product_id)
		);`,
		`CREATE TABLE IF NOT EXISTS orders (
			order_id TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);`,
	}

	for _, query := range queries {
		if _, err := pool.Exec(ctx, query); err != nil {
			return nil, fmt.Errorf("failed to execute migration: %v", err)
		}
	}

	// Seed Data
	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check products count: %v", err)
	}

	if count == 0 {
		log.Println("Seeding database with initial products...")
		products := []struct {
			id, name, desc string
			price          float64
		}{
			{"1", "Laptop", "A powerful laptop", 1200.50},
			{"2", "Mouse", "A wireless mouse", 45.99},
			{"3", "Keyboard", "Mechanical keyboard", 150.00},
			{"4", "Laptop Stand", "Ergonomic stand", 89.99},
		}

		for _, p := range products {
			_, err := pool.Exec(ctx, "INSERT INTO products (id, name, description, price) VALUES ($1, $2, $3, $4)", p.id, p.name, p.desc, p.price)
			if err != nil {
				return nil, fmt.Errorf("failed to seed product %s: %v", p.name, err)
			}
		}
	}

	return pool, nil
}
