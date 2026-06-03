package database

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    string    `json:"image_url"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var (
	ErrNameRequired = errors.New("name is required")
	ErrInvalidPrice = errors.New("price must be non-negative")
	ErrInvalidStock = errors.New("stock must be non-negative")
)

const (
	maxListProducts = 100

	addProductSQL = `
INSERT INTO products (name, description, price, stock, image_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, description, price, stock, image_url, is_active, created_at, updated_at
`

	listProductsSQL = `
SELECT id, name, description, price, stock, image_url, is_active, created_at, updated_at
FROM products
ORDER BY created_at DESC
LIMIT $1
`
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProduct(s rowScanner) (Product, error) {
	var p Product
	err := s.Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock,
		&p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

func AddProduct(ctx context.Context, db *DB, p Product) (Product, error) {
	if db == nil || db.Write == nil {
		return Product{}, errors.New("database write pool is not initialized")
	}

	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return Product{}, ErrNameRequired
	}
	if p.Price < 0 {
		return Product{}, ErrInvalidPrice
	}
	if p.Stock < 0 {
		return Product{}, ErrInvalidStock
	}

	created, err := scanProduct(db.Write.QueryRow(ctx, addProductSQL,
		p.Name, p.Description, p.Price, p.Stock, p.ImageURL,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, errors.New("failed to insert product")
		}
		return Product{}, err
	}
	return created, nil
}

func ListProducts(ctx context.Context, db *DB) ([]Product, error) {
	if db == nil {
		return nil, errors.New("database pool not initialized")
	}

	rows, err := db.Reader().Query(ctx, listProductsSQL, maxListProducts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]Product, 0, maxListProducts)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}
